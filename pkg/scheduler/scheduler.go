package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/kubernetes/pkg/scheduler/algorithmprovider"
	"k8s.io/kubernetes/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/pkg/scheduler/common/config"
	"k8s.io/kubernetes/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/pkg/scheduler/framework/interfaces"
	frameworkplugins "k8s.io/kubernetes/pkg/scheduler/framework/plugins"
	internalcache "k8s.io/kubernetes/pkg/scheduler/internal/cache"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/pkg/scheduler/types"
	"k8s.io/kubernetes/pkg/scheduler/utils"
	"k8s.io/kubernetes/pkg/scheduler/utils/wait"
	"k8s.io/kubernetes/pkg/scheduler/utils/workqueue"
)

var schedulerCache internalcache.Cache

func GetScheduler() *Scheduler {
	sched, err := NewScheduler(schedulerCache)
	if err != nil {
		fmt.Printf("NewScheduler failed!, err: %s", err)
		return nil
	}

	return sched
}

// ScheduleResult represents the result of one pod scheduled. It will contain
// the final selected Node, along with the selected intermediate information.
type ScheduleResult struct {
	// Name of the scheduler suggest host
	SuggestedHost string
	Stacks        []types.Stack
	// Number of nodes scheduler evaluated on one stack scheduled
	EvaluatedNodes int
	// Number of feasible nodes on one stack scheduled
	FeasibleNodes int
}

// Scheduler watches for new unscheduled pods. It attempts to find
// nodes that they fit on and writes bindings back to the api server.
type Scheduler struct {
	// It is expected that changes made via SchedulerCache will be observed
	// by NodeLister and Algorithm.
	SchedulerCache internalcache.Cache

	nodeInfoSnapshot *internalcache.Snapshot

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	Plugins *types.Plugins
	// policy are the scheduling policy.
	SchedFrame interfaces.Framework
}

// Cache returns the cache in scheduler for test to check the data in scheduler.
func (sched *Scheduler) Cache() internalcache.Cache {
	return sched.SchedulerCache
}

// Run begins watching and scheduling. It waits for cache to be synced, then starts scheduling
// and blocked until the context is done.
func (sched *Scheduler) Run(ctx context.Context) {

}

// snapshot snapshots scheduler cache and node infos for all fit and priority
// functions.
func (sched *Scheduler) snapshot() error {
	// Used for all fit and priority funcs.
	return sched.SchedulerCache.UpdateSnapshot(sched.nodeInfoSnapshot)
}

// stackPassesFiltersOnNode checks whether a node given by NodeInfo satisfies the
// filter plugins.
// This function is called from two different places: Schedule and Preempt.
// When it is called from Schedule, we want to test whether the pod is
// schedulable on the node with all the existing pods on the node plus higher
// and equal priority pods nominated to run on the node.
// When it is called from Preempt, we should remove the victims of preemption
// and add the nominated pods. Removal of the victims is done by
// SelectVictimsOnNode(). Preempt removes victims from PreFilter state and
// NodeInfo before calling this function.
func (sched *Scheduler) stackPassesFiltersOnNode(
	ctx context.Context,
	state *interfaces.CycleState,
	stack *types.Stack,
	info *schedulernodeinfo.NodeInfo,
) (bool, *interfaces.Status, error) {
	var status *interfaces.Status

	statusMap := sched.SchedFrame.RunFilterPlugins(ctx, state, stack, info)
	status = statusMap.Merge()
	if !status.IsSuccess() && !status.IsUnschedulable() {
		return false, status, status.AsError()
	}

	return status.IsSuccess(), status, nil
}

// findNodesThatPassFilters finds the nodes that fit the filter plugins.
func (sched *Scheduler) findNodesThatPassFilters(ctx context.Context, state *interfaces.CycleState,
	stack *types.Stack, statuses interfaces.NodeToStatusMap) ([]*types.SiteNode, error) {
	allNodes, err := sched.nodeInfoSnapshot.NodeInfos().List()
	if err != nil {
		return nil, err
	}

	// Create filtered list with enough space to avoid growing it
	// and allow assigning.
	filtered := make([]*types.SiteNode, len(allNodes))

	if !sched.SchedFrame.HasFilterPlugins() {
		for i := range filtered {
			filtered[i] = allNodes[i].Node()
		}
		return filtered, nil
	}

	errCh := utils.NewErrorChannel()
	var statusesLock sync.Mutex
	var filteredLen int32
	ctx, cancel := context.WithCancel(ctx)
	checkNode := func(i int) {
		nodeInfo := allNodes[i]
		fits, status, err := sched.stackPassesFiltersOnNode(ctx, state, stack, nodeInfo)
		if err != nil {
			errCh.SendErrorWithCancel(err, cancel)
			return
		}
		if fits {
			length := atomic.AddInt32(&filteredLen, 1)
			filtered[length-1] = nodeInfo.Node()
		} else {
			statusesLock.Lock()
			if !status.IsSuccess() {
				statuses[nodeInfo.Node().SiteID] = status
			}
			statusesLock.Unlock()
		}
	}

	// Stops searching for more nodes once the configured number of feasible nodes
	// are found.
	workqueue.ParallelizeUntil(ctx, 16, len(allNodes), checkNode)

	filtered = filtered[:filteredLen]
	if err := errCh.ReceiveError(); err != nil {
		return nil, err
	}
	return filtered, nil
}

// prioritizeNodes prioritizes the nodes by running the score plugins,
// which return a score for each node from the call to RunScorePlugins().
// The scores from each plugin are added together to make the score for that node, then
// any extenders are run as well.
// All scores are finally combined (added) to get the total weighted scores of all nodes
func (sched *Scheduler) prioritizeNodes(
	ctx context.Context,
	state *interfaces.CycleState,
	pod *types.Stack,
	nodes []*types.SiteNode,
) (interfaces.NodeScoreList, error) {
	// If no priority configs are provided, then all nodes will have a score of one.
	// This is required to generate the priority list in the required format
	if !sched.SchedFrame.HasScorePlugins() {
		result := make(interfaces.NodeScoreList, 0, len(nodes))
		for i := range nodes {
			result = append(result, interfaces.NodeScore{
				Name:  nodes[i].SiteID,
				Score: 1,
			})
		}
		return result, nil
	}

	// Run the Score plugins.
	scoresMap, scoreStatus := sched.SchedFrame.RunScorePlugins(ctx, state, pod, nodes)
	if !scoreStatus.IsSuccess() {
		return interfaces.NodeScoreList{}, scoreStatus.AsError()
	}

	// Summarize all scores.
	result := make(interfaces.NodeScoreList, 0, len(nodes))

	for i := range nodes {
		result = append(result, interfaces.NodeScore{Name: nodes[i].SiteID, AZ: nodes[i].AvailabilityZone, Score: 0})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}

	// sort by score.
	sort.Sort(sort.Reverse(result))

	logger.Debug(ctx, "score nodes: %v", result)

	return result, nil
}

// selectHost takes a prioritized list of nodes and then picks one
// in a reservoir sampling manner from the nodes that had the highest score.
func (sched *Scheduler) selectHost(nodeScoreList interfaces.NodeScoreList) (string, error) {
	if len(nodeScoreList) == 0 {
		return "", fmt.Errorf("empty priorityList")
	}
	maxScore := nodeScoreList[0].Score
	selected := nodeScoreList[0].Name
	cntOfMaxScore := 1
	for _, ns := range nodeScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selected = ns.Name
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selected = ns.Name
			}
		}
	}
	return selected, nil
}

// bind binds a pod to a given node defined in a binding object.
// The precedence for binding is: (1) extenders and (2) framework plugins.
// We expect this to run asynchronously, so we handle binding metrics internally.
func (sched *Scheduler) bind(ctx context.Context, stack *types.Stack, targetNode string,
	state *interfaces.CycleState) (err error) {
	bindStatus := sched.SchedFrame.RunBindPlugins(ctx, state, stack, targetNode)
	if bindStatus.IsSuccess() {
		return nil
	}
	if bindStatus.Code() == interfaces.Error {
		return bindStatus.AsError()
	}
	return fmt.Errorf("bind status: %s, %v", bindStatus.Code().String(), bindStatus.Message())
}

// Schedule Run begins watching and scheduling. It waits for cache to be synced ,
// then starts scheduling and blocked until the context is done.
func (sched *Scheduler) Schedule2(ctx context.Context, allocation *types.Allocation) (result ScheduleResult, err error) {
	logger.Debug(ctx, "Attempting to schedule allocation: %v", allocation.ID)

	state := interfaces.NewCycleState()
	schedulingCycleCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	start := time.Now()
	logger.Debug(ctx, "[START] snapshot nodes...")
	err = sched.snapshot()
	if err != nil {
		logger.Error(ctx, "sched snapshot failed! err : %s", err)
		return result, err
	}
	end := time.Now()
	logger.Debug(ctx, "[DONE] snapshot nodes, use_time: %d us", end.Sub(start))

	start = end
	logger.Debug(ctx, "[START] Running prefilter plugins...")
	// Run "prefilter" plugins.
	preFilterStatus := sched.SchedFrame.RunPreFilterPlugins(schedulingCycleCtx, state, &allocation.Stack)
	if !preFilterStatus.IsSuccess() {
		return result, preFilterStatus.AsError()
	}

	end = time.Now()
	logger.Debug(ctx, "[DONE] Running prefilter plugins, use_time: %d us", end.Sub(start))
	start = end

	logger.Debug(ctx, "[START] Running filter plugins...")
	filteredNodesStatuses := make(interfaces.NodeToStatusMap)
	allocation.Stack.Selector = allocation.Selector
	filteredNodes, err := sched.findNodesThatPassFilters(ctx, state, &allocation.Stack, filteredNodesStatuses)
	if err != nil {
		logger.Error(ctx, "findNodesThatPassFilters failed! err: %s", err)
		return result, err
	}
	end = time.Now()
	logger.Debug(ctx, "[DONE] Running filter plugins, use_time: %d us", end.Sub(start))
	start = end

	logger.Debug(ctx, "filteredNodesStatuses = %v", filteredNodesStatuses.ToString())
	if len(filteredNodes) <= 0 {
		logger.Error(ctx, "filter none nodes. resultStatus: %s", filteredNodesStatuses.ToString())
		return result, nil
	}

	logger.Debug(ctx, "[START] Running preScore plugins...")
	// Run "prescore" plugins.
	prescoreStatus := sched.SchedFrame.RunPreScorePlugins(ctx, state, &allocation.Stack, filteredNodes)
	if !prescoreStatus.IsSuccess() {
		return result, prescoreStatus.AsError()
	}

	end = time.Now()
	logger.Debug(ctx, "[DONE] Running preScore plugins, use_time: %d us", end.Sub(start))
	start = end

	logger.Debug(ctx, "[START] Running prioritizeNodes plugins...")
	priorityList, err := sched.prioritizeNodes(ctx, state, &allocation.Stack, filteredNodes)
	if err != nil {
		logger.Error(ctx, "prioritizeNodes failed! err: %s", err)
		return result, err
	}
	end = time.Now()
	logger.Debug(ctx, "[DONE] Running prioritizeNodes plugins, use_time: %d us", end.Sub(start))
	start = end

	logger.Debug(ctx, "[START] Running StrategyPlugins plugins...")
	nodeCount, strategyStatus := sched.SchedFrame.RunStrategyPlugins(ctx, state, allocation, priorityList)
	if !strategyStatus.IsSuccess() {
		logger.Error(ctx, "RunStrategyPlugins failed! err: %s", err)
		return result, err
	}
	end = time.Now()
	logger.Debug(ctx, "[DONE] Running StrategyPlugins plugins, use_time: %d us", end.Sub(start))
	logger.Debug(ctx, "selected Hosts : %#v", nodeCount)
	start = end

	var count = 0

	for _, value := range nodeCount {
		for i := 0; i < value.StackCount; i++ {
			newStack := allocation.Stack
			//bind
			err = sched.bind(ctx, &newStack, value.Name, state)
			if err != nil {
				logger.Error(ctx, "bind host(%s) failed! err: %s", value.Name, err)
				return result, err
			}
			result.Stacks = append(result.Stacks, newStack)
			count++
			if count >= allocation.Replicas {
				break
			}
		}

		if count >= allocation.Replicas {
			break
		}
	}

	if count < allocation.Replicas {
		logger.Error(ctx, "not find suit host")
		return result, fmt.Errorf("not find suit host")
	}

	end = time.Now()
	logger.Debug(ctx, "allocation(%s) success, use_time: %d us", allocation.ID, end.Sub(start))

	return
}

// Schedule Run begins watching and scheduling. It waits for cache to be synced,
// then starts scheduling and blocked until the context is done.
func (sched *Scheduler) Schedule(ctx context.Context, stack *types.Stack) (result ScheduleResult, err error) {
	logger.Info(ctx, "Attempting to schedule stack: %v/%v", stack.Name, stack.UID)

	state := interfaces.NewCycleState()
	schedulingCycleCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = sched.snapshot()
	if err != nil {
		logger.Error(ctx, "sched snapshot failed! err : %s", err)
		return result, err
	}

	logger.Info(ctx, "Snapshotting scheduler cache and node infos done")
	logger.Info(ctx, "Running prefilter plugins...")
	// Run "prefilter" plugins.
	preFilterStatus := sched.SchedFrame.RunPreFilterPlugins(schedulingCycleCtx, state, stack)
	if !preFilterStatus.IsSuccess() {
		return result, preFilterStatus.AsError()
	}
	logger.Info(ctx, "Running prefilter plugins done")

	filteredNodesStatuses := make(interfaces.NodeToStatusMap)
	filteredNodes, err := sched.findNodesThatPassFilters(ctx, state, stack, filteredNodesStatuses)
	if err != nil {
		logger.Error(ctx, "findNodesThatPassFilters failed! err: %s", err)
		return result, err
	}
	logger.Debug(ctx, "Computing predicates done, filteredNodesStatuses = %v", filteredNodesStatuses.ToString())
	if len(filteredNodes) <= 0 {
		logger.Debug(ctx, "filter none nodes. resultStatus: %s", filteredNodesStatuses.ToString())
		return result, nil
	}

	// Run "prescore" plugins.
	prescoreStatus := sched.SchedFrame.RunPreScorePlugins(ctx, state, stack, filteredNodes)
	if !prescoreStatus.IsSuccess() {
		return result, prescoreStatus.AsError()
	}
	logger.Info(ctx, "Running prescore plugins done")

	priorityList, err := sched.prioritizeNodes(ctx, state, stack, filteredNodes)
	if err != nil {
		logger.Error(ctx, "prioritizeNodes failed! err: %s", err)
		return result, err
	}
	host, err := sched.selectHost(priorityList)
	logger.Info(ctx, "selectHost is %s", host)

	// bind
	err = sched.bind(ctx, stack, host, state)
	if err != nil {
		logger.Error(ctx, "bind host(%s) failed! err: %s", host, err)
		return result, err
	}

	return ScheduleResult{
		SuggestedHost:  host,
		EvaluatedNodes: len(filteredNodes) + len(filteredNodesStatuses),
		FeasibleNodes:  len(filteredNodes),
	}, err
}

func (sched *Scheduler) buildFramework() error {
	registry := frameworkplugins.NewRegistry()

	policyFile := config.String(constants.ConfPolicyFile)
	if policyFile == "" {
		logger.Errorf("policyFile(%s) not set!", constants.ConfPolicyFile)
		return fmt.Errorf("policyFile(%s) not set", constants.ConfPolicyFile)
	}

	policy := &types.Policy{}
	err := config.InitPolicyFromFile(policyFile, policy)
	if err != nil {
		logger.Errorf("InitPolicyFromFile %s failed! err: %s", constants.ConfPolicyFile, err)
		return err
	}

	defaultPlugins := algorithmprovider.GetPlugins(*policy)
	sched.SchedFrame, err = interfaces.NewFramework(registry, defaultPlugins,
		interfaces.WithSnapshotSharedLister(sched.nodeInfoSnapshot),
		interfaces.WithCache(sched.SchedulerCache))
	if err != nil {
		logger.Errorf("NewFramework failed! err : %s", err)
		return err
	}

	return nil
}

func InitSchedulerCache(stopCh <-chan struct{}) internalcache.Cache {
	stopEverything := stopCh
	if stopEverything == nil {
		stopEverything = wait.NeverStop
	}

	schedulerCache = internalcache.New(30*time.Second, stopEverything)
	return schedulerCache
}

func NewScheduler(cache internalcache.Cache) (*Scheduler, error) {

	if cache == nil {
		logger.Errorf("cache not init, please wait...")
		return nil, fmt.Errorf("cache not init, please wait")
	}
	sched := &Scheduler{
		SchedulerCache:   cache,
		nodeInfoSnapshot: internalcache.NewEmptySnapshot(),
	}

	err := sched.buildFramework()
	if err != nil {
		logger.Errorf("buildFramework by %s failed! err: %s", types.SchedulerDefaultProviderName, err)
		return nil, err
	}

	return sched, nil
}

// Init resource cache informer
func InitSchedulerInformer(stopCh <-chan struct{}) {
	go func(stopCh2 <-chan struct{}) {
		// init informer
		informers.InformerFac = informers.NewSharedInformerFactory(nil, 60*time.Second)

		// init volume type informer
		volumetypeInterval := config.DefaultInt(constants.ConfVolumeTypeInterval, 600)
		informers.InformerFac.VolumeType(informers.VOLUMETYPE, "ID",
			time.Duration(volumetypeInterval)*time.Second).Informer()

		// init site informer
		siteInterval := config.DefaultInt(constants.ConfSiteInterval, 600)
		informers.InformerFac.Sites(informers.SITES, "ID",
			time.Duration(siteInterval)*time.Second).Informer()

		// init flavor informer
		flavorInterval := config.DefaultInt(constants.ConfFlavorInterval, 600)
		informers.InformerFac.Flavor(informers.FLAVOR, "RegionFlavorID",
			time.Duration(flavorInterval)*time.Second).Informer()

		// init eip pool informer
		eipPoolInterval := config.DefaultInt(constants.ConfEipPoolInterval, 60)
		eipPoolInformer := informers.InformerFac.EipPools(informers.EIPPOOLS, "Region",
			time.Duration(eipPoolInterval)*time.Second).Informer()
		eipPoolInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: updateEipPools,
			})

		// init volume pool informer
		volumePoolInterval := config.DefaultInt(constants.ConfVolumePoolInterval, 60)
		volumePoolInformer := informers.InformerFac.VolumePools(informers.VOLUMEPOOLS, "Region",
			time.Duration(volumePoolInterval)*time.Second).Informer()
		volumePoolInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: updateVolumePools,
			})

		// init node informer
		nodeInterval := config.DefaultInt(constants.ConfCommonHypervisorInterval, 86400)
		nodeInformer := informers.InformerFac.Nodes(informers.NODES, "EdgeSiteID",
			time.Duration(nodeInterval)*time.Second).Informer()
		nodeInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: addSiteNodesToCache,
			})

		informers.InformerFac.Start(stopCh2)

		// wait until node informer synced
		for {
			if nodeInformer.HasSynced() {
				break
			}

			time.Sleep(2 * time.Second)
		}

		// need sync once before start
		volumePoolInformer.SyncOnce()
		eipPoolInformer.SyncOnce()
	}(stopCh)
}

// update EipPools with sched cache
func updateEipPools(obj []interface{}) {
	if obj == nil {
		return
	}

	for _, eipPoolObj := range obj {
		eipPool, ok := eipPoolObj.(typed.EipPool)
		if !ok {
			logger.Warnf("convert interface to (typed.EipPool) failed.")
			continue
		}

		err := schedulerCache.UpdateEipPool(&eipPool)
		if err != nil {
			logger.Infof("UpdateEipPool failed! err: %s", err)
		}
	}
}

// update VolumePools with sched cache
func updateVolumePools(obj []interface{}) {
	if obj == nil {
		return
	}

	for _, volumePoolObj := range obj {
		volumePool, ok := volumePoolObj.(typed.RegionVolumePool)
		if !ok {
			logger.Warnf("convert interface to (typed.VolumePools) failed.")
			continue
		}

		err := schedulerCache.UpdateVolumePool(&volumePool)
		if err != nil {
			logger.Infof("updateVolumePools failed! err: %s", err)
		}
	}
}

// add site node to cache
func addSiteNodesToCache(obj []interface{}) {
	if obj == nil {
		return
	}

	sites := informers.InformerFac.GetInformer(informers.SITES).GetStore().List()

	for _, sn := range obj {
		siteNode, ok := sn.(typed.SiteNode)
		if !ok {
			logger.Warnf("convert interface to (typed.SiteNode) failed.")
			continue
		}

		var isFind = false
		for _, site := range sites {
			siteInfo, ok := site.(typed.Site)
			if !ok {
				continue
			}

			if siteInfo.Region == siteNode.Region && siteInfo.Az == siteNode.AvailabilityZone {
				info := convertToSiteNode(siteInfo, siteNode)
				err := schedulerCache.AddNode(info)
				if err != nil {
					logger.Infof("add node to cache failed! err: %s", err)
				}

				isFind = true
				break
			}
		}

		if !isFind {
			site := &types.SiteNode{
				SiteID: siteNode.Region + "--" + siteNode.AvailabilityZone,
				RegionAzMap: types.RegionAzMap{
					Region:           siteNode.Region,
					AvailabilityZone: siteNode.AvailabilityZone,
				},
				Status: constants.SiteStatusNormal,
			}

			site.Nodes = append(site.Nodes, siteNode.Nodes...)
			err := schedulerCache.AddNode(site)
			if err != nil {
				logger.Infof("add node to cache failed! err: %s", err)
			}
		}
	}

	schedulerCache.PrintString()
}

func convertToSiteNode(site typed.Site, node typed.SiteNode) *types.SiteNode {
	siteNode := &types.SiteNode{
		SiteID: site.ID,
		GeoLocation: types.GeoLocation{
			Country:  site.Country,
			Area:     site.Area,
			Province: site.Province,
			City:     site.City,
		},
		RegionAzMap: types.RegionAzMap{
			Region:           site.Region,
			AvailabilityZone: site.Az,
		},
		Operator:      site.Operator.Name,
		EipTypeName:   site.EipTypeName,
		Status:        site.Status,
		SiteAttribute: site.SiteAttributes,
	}

	siteNode.Nodes = append(siteNode.Nodes, node.Nodes...)
	return siteNode
}
