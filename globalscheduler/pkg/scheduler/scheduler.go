/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	internalinformers "k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/algorithmprovider"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/config"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/factory"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	frameworkplugins "k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins"
	internalcache "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache"
	internalqueue "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/queue"
	schedulersitecacheinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/wait"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/workqueue"
)

// single scheduler instance
var scheduler *Scheduler
var once sync.Once

// InitScheduler
func InitScheduler(config *types.GSSchedulerConfiguration, stopCh <-chan struct{}) error {
	var err error
	scheduler, err = NewScheduler(config, stopCh)
	return err
}

// GetScheduler gets single scheduler instance. New Scheduler will only run once,
// if it runs failed, nil will be return.
func GetScheduler() *Scheduler {
	if scheduler == nil {
		logger.Errorf("Scheduler need to be init correctly")
		return scheduler
	}

	return scheduler
}

// ScheduleResult represents the result of one pod scheduled. It will contain
// the final selected Site, along with the selected intermediate information.
type ScheduleResult struct {
	// Name of the scheduler suggest host
	SuggestedHost string
	Stacks        []types.Stack
	// Number of site scheduler evaluated on one stack scheduled
	EvaluatedSites int
	// Number of feasible site on one stack scheduled
	FeasibleSites int
}

// Scheduler watches for new unscheduled pods. It attempts to find
// site that they fit on and writes bindings back to the api server.
type Scheduler struct {
	// Name of the current scheduler
	SchedulerName string

	// Resource Collector API URL
	ResourceCollectorApiUrl string

	// It is expected that changes made via SchedulerCache will be observed
	// by NodeLister and Algorithm.
	SchedulerCache internalcache.Cache

	siteCacheInfoSnapshot *internalcache.Snapshot

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	Plugins *types.Plugins
	// policy are the scheduling policy.
	SchedFrame interfaces.Framework

	// queue for stacks that need scheduling
	StackQueue      internalqueue.SchedulingQueue
	PodInformer     coreinformers.PodInformer
	Client          clientset.Interface
	InformerFactory internalinformers.SharedInformerFactory

	// NextStack should be a function that blocks until the next stack
	// is available. We don't use a channel for this, because scheduling
	// a stack may take some amount of time and we don't want pods to get
	// stale while they sit in a channel.
	NextStack func() *types.Stack

	mu sync.RWMutex
}

// Cache returns the cache in scheduler for test to check the data in scheduler.
func (sched *Scheduler) Cache() internalcache.Cache {
	return sched.SchedulerCache
}

// scheduleOne does the entire scheduling workflow for a single pod.
func (sched *Scheduler) scheduleOne() {
	stack := sched.NextStack()

	// generate allocation from stack
	allocation, err := sched.generateAllocationFromStack(stack)

	tmpContext := context.Background()

	// do scheduling process
	result, err := sched.Schedule2(tmpContext, allocation)
	if err != nil {
		logger.Errorf("Schedule failed, err: %s", err)
		return
	}

	logger.Infof("Scheduler result: %v", result)

	// bind scheduler result to pod
	logger.Infof("Try to bind to site, stacks:%v", result.Stacks)
	sched.bindStacks(result.Stacks)
}

// generateAllocationFromStack generate a new allocation obj from one single stack
func (sched *Scheduler) generateAllocationFromStack(stack *types.Stack) (*types.Allocation, error) {
	allocation := &types.Allocation{
		ID:       uuid.NewV4().String(),
		Stack:    *stack,
		Replicas: 1,
		Selector: stack.Selector,
	}

	return allocation, nil
}

// Run begins watching and scheduling. It waits for cache to be synced, then starts scheduling
// and blocked until the context is done.
func (sched *Scheduler) Run() {
	go wait.Until(sched.scheduleOne, 0, sched.StopEverything)
}

func (sched *Scheduler) GetResourceSnapshot(resourceCollectorApiUrl string) (internalcache.Snapshot, error) {
	snapshotEndpoint := "http://" + resourceCollectorApiUrl + constants.ResourceCollecotrSnapshotURL
	resp, err := utils.SendHTTPRequest("GET", snapshotEndpoint, nil, nil, false)
	if err != nil {
		// snapshot api error
		return internalcache.Snapshot{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// read snapshot resp error
		return internalcache.Snapshot{}, err
	}

	bodyStr := string(body)
	var snapshot internalcache.Snapshot
	if err = json.Unmarshal([]byte(bodyStr), &snapshot); err != nil {
		return internalcache.Snapshot{}, err
	}

	// update flavor map
	internalcache.FlavorCache.UpdateFlavorMap(snapshot.RegionFlavorMap, snapshot.FlavorMap)

	logger.Infof("snapshot : %v", snapshot)
	return snapshot, nil
}

func (sched *Scheduler) updateSnapshot() error {
	snapshot, err := sched.GetResourceSnapshot(sched.ResourceCollectorApiUrl)
	if err != nil {
		return err
	}

	// set snapshot
	sched.mu.Lock()
	sched.mu.Unlock()
	sched.siteCacheInfoSnapshot = &snapshot
	return nil
}

// snapshot snapshots scheduler cache and site cache infos for all fit and priority
// functions.
func (sched *Scheduler) snapshot() error {
	// Used for all fit and priority funcs.
	return sched.Cache().UpdateSnapshot(sched.siteCacheInfoSnapshot)
}

// stackPassesFiltersOnSite checks whether a site given by Host satisfies the
// filter plugins.
// This function is called from two different places: Schedule and Preempt.
// When it is called from Schedule, we want to test whether the pod is
// schedulable on the site with all the existing pods on the site plus higher
// and equal priority pods nominated to run on the site.
// When it is called from Preempt, we should remove the victims of preemption
// and add the nominated pods. Removal of the victims is done by
// SelectVictimsOnNode(). Preempt removes victims from PreFilter state and
// Host before calling this function.
func (sched *Scheduler) stackPassesFiltersOnSite(
	ctx context.Context,
	state *interfaces.CycleState,
	stack *types.Stack,
	info *schedulersitecacheinfo.SiteCacheInfo,
) (bool, *interfaces.Status, error) {
	var status *interfaces.Status

	statusMap := sched.SchedFrame.RunFilterPlugins(ctx, state, stack, info)
	status = statusMap.Merge()
	if !status.IsSuccess() && !status.IsUnschedulable() {
		return false, status, status.AsError()
	}

	return status.IsSuccess(), status, nil
}

// findSitesThatPassFilters finds the site that fit the filter plugins.
func (sched *Scheduler) findSitesThatPassFilters(ctx context.Context, state *interfaces.CycleState,
	stack *types.Stack, statuses interfaces.SiteToStatusMap) ([]*types.Site, error) {
	allSiteCacheInfos, err := sched.siteCacheInfoSnapshot.SiteCacheInfos().List()
	if err != nil {
		return nil, err
	}

	// Create filtered list with enough space to avoid growing it
	// and allow assigning.
	filtered := make([]*types.Site, len(allSiteCacheInfos))

	if !sched.SchedFrame.HasFilterPlugins() {
		for i := range filtered {
			filtered[i] = allSiteCacheInfos[i].GetSite()
		}
		return filtered, nil
	}

	errCh := utils.NewErrorChannel()
	var statusesLock sync.Mutex
	var filteredLen int32
	ctx, cancel := context.WithCancel(ctx)
	checkSite := func(i int) {
		siteCacheInfo := allSiteCacheInfos[i]
		fits, status, err := sched.stackPassesFiltersOnSite(ctx, state, stack, siteCacheInfo)
		if err != nil {
			errCh.SendErrorWithCancel(err, cancel)
			return
		}
		if fits {
			length := atomic.AddInt32(&filteredLen, 1)
			filtered[length-1] = siteCacheInfo.GetSite()
		} else {
			statusesLock.Lock()
			if !status.IsSuccess() {
				statuses[siteCacheInfo.GetSite().SiteID] = status
			}
			statusesLock.Unlock()
		}
	}

	// Stops searching for more site once the configured number of feasible site
	// are found.
	workqueue.ParallelizeUntil(ctx, 16, len(allSiteCacheInfos), checkSite)

	filtered = filtered[:filteredLen]
	if err := errCh.ReceiveError(); err != nil {
		return nil, err
	}
	return filtered, nil
}

// prioritizeSites prioritizes the site by running the score plugins,
// which return a score for each site from the call to RunScorePlugins().
// The scores from each plugin are added together to make the score for that site, then
// any extenders are run as well.
// All scores are finally combined (added) to get the total weighted scores of all site
func (sched *Scheduler) prioritizeSites(
	ctx context.Context,
	state *interfaces.CycleState,
	pod *types.Stack,
	sites []*types.Site,
) (interfaces.SiteScoreList, error) {
	// If no priority configs are provided, then all sites will have a score of one.
	// This is required to generate the priority list in the required format
	if !sched.SchedFrame.HasScorePlugins() {
		result := make(interfaces.SiteScoreList, 0, len(sites))
		for i := range sites {
			result = append(result, interfaces.SiteScore{
				SiteID: sites[i].SiteID,
				Score:  1,
			})
		}
		return result, nil
	}

	// Run the Score plugins.
	scoresMap, scoreStatus := sched.SchedFrame.RunScorePlugins(ctx, state, pod, sites,
		sched.siteCacheInfoSnapshot.SiteCacheInfoMap)
	if !scoreStatus.IsSuccess() {
		return interfaces.SiteScoreList{}, scoreStatus.AsError()
	}

	// Summarize all scores.
	result := make(interfaces.SiteScoreList, 0, len(sites))

	for i := range sites {
		result = append(result, interfaces.SiteScore{SiteID: sites[i].SiteID, AZ: sites[i].AvailabilityZone, Score: 0, Region: sites[i].Region})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}

	// sort by score.
	sort.Sort(sort.Reverse(result))

	logger.Debug(ctx, "score sites: %v", result)

	return result, nil
}

// selectHost takes a prioritized list of site and then picks one
// in a reservoir sampling manner from the site that had the highest score.
func (sched *Scheduler) selectHost(siteScoreList interfaces.SiteScoreList) (string, error) {
	if len(siteScoreList) == 0 {
		return "", fmt.Errorf("empty priorityList")
	}
	maxScore := siteScoreList[0].Score
	selected := siteScoreList[0].SiteID
	cntOfMaxScore := 1
	for _, ns := range siteScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selected = ns.SiteID
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selected = ns.SiteID
			}
		}
	}
	return selected, nil
}

// bind binds a pod to a given site defined in a binding object.
// The precedence for binding is: (1) extenders and (2) framework plugins.
// We expect this to run asynchronously, so we handle binding metrics internally.
func (sched *Scheduler) bind(ctx context.Context, stack *types.Stack, targetSiteID string,
	state *interfaces.CycleState) (err error) {
	bindStatus := sched.SchedFrame.RunBindPlugins(ctx, state, stack,
		sched.siteCacheInfoSnapshot.SiteCacheInfoMap[targetSiteID])
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
	logger.Debug(ctx, "[START] snapshot site...")
	err = sched.updateSnapshot()
	if err != nil {
		logger.Error(ctx, "sched snapshot failed! err : %s", err)
		return result, err
	}
	end := time.Now()
	logger.Debug(ctx, "[DONE] snapshot site, use_time: %d us", end.Sub(start))

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
	filteredSitesStatuses := make(interfaces.SiteToStatusMap)
	allocation.Stack.Selector = allocation.Selector
	filteredSites, err := sched.findSitesThatPassFilters(ctx, state, &allocation.Stack, filteredSitesStatuses)
	if err != nil {
		logger.Error(ctx, "findSitesThatPassFilters failed! err: %s", err)
		return result, err
	}
	end = time.Now()
	logger.Debug(ctx, "[DONE] Running filter plugins, use_time: %d us", end.Sub(start))
	start = end

	logger.Debug(ctx, "filteredSitesStatuses = %v", filteredSitesStatuses.ToString())
	if len(filteredSites) <= 0 {
		logger.Error(ctx, "filter none site. resultStatus: %s", filteredSitesStatuses.ToString())
		return result, nil
	}

	logger.Debug(ctx, "[START] Running preScore plugins...")
	// Run "prescore" plugins.
	prescoreStatus := sched.SchedFrame.RunPreScorePlugins(ctx, state, &allocation.Stack, filteredSites)
	if !prescoreStatus.IsSuccess() {
		return result, prescoreStatus.AsError()
	}

	end = time.Now()
	logger.Debug(ctx, "[DONE] Running preScore plugins, use_time: %d us", end.Sub(start))
	start = end

	logger.Debug(ctx, "[START] Running prioritizeSites plugins...")
	priorityList, err := sched.prioritizeSites(ctx, state, &allocation.Stack, filteredSites)
	if err != nil {
		logger.Error(ctx, "prioritizeSites failed! err: %s", err)
		return result, err
	}
	end = time.Now()
	logger.Debug(ctx, "[DONE] Running prioritizeSites plugins, use_time: %d us", end.Sub(start))
	start = end

	logger.Debug(ctx, "[START] Running StrategyPlugins plugins...")
	siteCount, strategyStatus := sched.SchedFrame.RunStrategyPlugins(ctx, state, allocation, priorityList)
	if !strategyStatus.IsSuccess() {
		logger.Error(ctx, "RunStrategyPlugins failed! err: %s", err)
		return result, err
	}
	end = time.Now()
	logger.Debug(ctx, "[DONE] Running StrategyPlugins plugins, use_time: %d us", end.Sub(start))
	logger.Debug(ctx, "selected Hosts : %#v", siteCount)
	start = end

	var count = 0

	for _, value := range siteCount {
		for i := 0; i < value.StackCount; i++ {
			newStack := allocation.Stack
			//bind
			err = sched.bind(ctx, &newStack, value.SiteID, state)
			if err != nil {
				logger.Error(ctx, "bind host(%s) failed! err: %s", value.SiteID, err)
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

	logger.Info(ctx, "Snapshotting scheduler cache and site infos done")
	logger.Info(ctx, "Running prefilter plugins...")
	// Run "prefilter" plugins.
	preFilterStatus := sched.SchedFrame.RunPreFilterPlugins(schedulingCycleCtx, state, stack)
	if !preFilterStatus.IsSuccess() {
		return result, preFilterStatus.AsError()
	}
	logger.Info(ctx, "Running prefilter plugins done")

	filteredSitesStatuses := make(interfaces.SiteToStatusMap)
	filteredSites, err := sched.findSitesThatPassFilters(ctx, state, stack, filteredSitesStatuses)
	if err != nil {
		logger.Error(ctx, "findSitesThatPassFilters failed! err: %s", err)
		return result, err
	}
	logger.Debug(ctx, "Computing predicates done, filteredSitesStatuses = %v", filteredSitesStatuses.ToString())
	if len(filteredSites) <= 0 {
		logger.Debug(ctx, "filter none site. resultStatus: %s", filteredSitesStatuses.ToString())
		return result, nil
	}

	// Run "prescore" plugins.
	prescoreStatus := sched.SchedFrame.RunPreScorePlugins(ctx, state, stack, filteredSites)
	if !prescoreStatus.IsSuccess() {
		return result, prescoreStatus.AsError()
	}
	logger.Info(ctx, "Running prescore plugins done")

	priorityList, err := sched.prioritizeSites(ctx, state, stack, filteredSites)
	if err != nil {
		logger.Error(ctx, "prioritizeSites failed! err: %s", err)
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
		EvaluatedSites: len(filteredSites) + len(filteredSitesStatuses),
		FeasibleSites:  len(filteredSites),
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
		interfaces.WithSnapshotSharedLister(sched.siteCacheInfoSnapshot),
		interfaces.WithCache(sched.SchedulerCache))
	if err != nil {
		logger.Errorf("NewFramework failed! err : %s", err)
		return err
	}

	return nil
}

func NewScheduler(config *types.GSSchedulerConfiguration, stopCh <-chan struct{}) (*Scheduler, error) {
	stopEverything := stopCh
	if stopEverything == nil {
		stopEverything = wait.NeverStop
	}

	sched := &Scheduler{
		SchedulerCache:          internalcache.New(30*time.Second, stopEverything),
		siteCacheInfoSnapshot:   internalcache.NewEmptySnapshot(),
		SchedulerName:           config.SchedulerName,
		ResourceCollectorApiUrl: config.ResourceCollectorApiUrl,
	}

	err := sched.buildFramework()
	if err != nil {
		return nil, fmt.Errorf("buildFramework by %s failed! err: %v", types.SchedulerDefaultProviderName, err)
	}

	// init pod informers for scheduler
	err = sched.initPodInformers(stopEverything)
	if err != nil {
		return nil, err
	}

	// add event handler
	AddAllEventHandlers(sched)

	return sched, nil
}

// initPodInformers init scheduler with podInformer
func (sched *Scheduler) initPodInformers(stopCh <-chan struct{}) error {
	masterURL := config.DefaultString("master", "127.0.0.1:8080")
	kubeconfig := config.DefaultString("kubeconfig", "/var/run/kubernetes/admin.kubeconfig")

	// init client
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		return err
	}

	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	sched.StackQueue = internalqueue.NewSchedulingQueue(stopCh, sched.SchedFrame)
	sched.InformerFactory = internalinformers.NewSharedInformerFactory(client, 0)
	sched.PodInformer = factory.NewPodInformer(sched.SchedulerName, client, 0)
	sched.NextStack = internalqueue.MakeNextStackFunc(sched.StackQueue)
	sched.Client = client
	return nil
}

// start Scheduler
func (sched *Scheduler) StartPodInformerAndRun(stopCh <-chan struct{}) {
	go func(stopCh2 <-chan struct{}) {
		// start pod informers
		if sched.PodInformer != nil && sched.InformerFactory != nil {
			go sched.PodInformer.Informer().Run(stopCh2)
			sched.InformerFactory.Start(stopCh2)

			// Wait for all caches to sync before scheduling.
			sched.InformerFactory.WaitForCacheSync(stopCh2)

			// Do scheduling
			sched.Run()
		}

	}(stopCh)
}

// start resource cache informer and run
func (sched *Scheduler) StartInformersAndRun(stopCh <-chan struct{}) {
	go func(stopCh2 <-chan struct{}) {
		// init informer
		informers.InformerFac = informers.NewSharedInformerFactory(nil, 60*time.Second)

		// init volume type informer
		volumetypeInterval := config.DefaultInt(constants.ConfVolumeTypeInterval, 600)
		informers.InformerFac.VolumeType(informers.VOLUMETYPE, "ID",
			time.Duration(volumetypeInterval)*time.Second).Informer()

		// init site informer
		siteInfoInterval := config.DefaultInt(constants.ConfSiteInfoInterval, 600)
		informers.InformerFac.SiteInfo(informers.SITEINFOS, "SITEID",
			time.Duration(siteInfoInterval)*time.Second).Informer()

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

		// init site resource informer
		siteResourceInterval := config.DefaultInt(constants.ConfCommonHypervisorInterval, 86400)
		siteResourceInformer := informers.InformerFac.SiteResource(informers.SITERESOURCES, "SiteID",
			time.Duration(siteResourceInterval)*time.Second).Informer()
		siteResourceInformer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				ListFunc: addSitesToCache,
			})

		informers.InformerFac.Start(stopCh2)

		// wait until site resource informer synced
		for {
			if siteResourceInformer.HasSynced() {
				break
			}

			time.Sleep(2 * time.Second)
		}

		// need sync once before start
		volumePoolInformer.SyncOnce()
		eipPoolInformer.SyncOnce()

		// start pod informers
		if sched.PodInformer != nil && sched.InformerFactory != nil {
			go sched.PodInformer.Informer().Run(stopCh2)
			sched.InformerFactory.Start(stopCh2)

			// Wait for all caches to sync before scheduling.
			sched.InformerFactory.WaitForCacheSync(stopCh2)

			// Do scheduling
			sched.Run()
		}

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

		err := scheduler.Cache().UpdateEipPool(&eipPool)
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

		err := scheduler.Cache().UpdateVolumePool(&volumePool)
		if err != nil {
			logger.Infof("updateVolumePools failed! err: %s", err)
		}
	}
}

// add site to cache
func addSitesToCache(obj []interface{}) {
	if obj == nil {
		return
	}

	siteInfos := informers.InformerFac.GetInformer(informers.SITEINFOS).GetStore().List()

	for _, sn := range obj {
		siteResource, ok := sn.(typed.SiteResource)
		if !ok {
			logger.Warnf("convert interface to (typed.SiteResource) failed.")
			continue
		}

		var isFind = false
		for _, site := range siteInfos {
			siteInfo, ok := site.(typed.SiteInfo)
			if !ok {
				continue
			}

			if siteInfo.Region == siteResource.Region && siteInfo.AvailabilityZone == siteResource.AvailabilityZone {
				info := convertToSite(siteInfo, siteResource)
				err := scheduler.Cache().AddSite(info)
				if err != nil {
					logger.Infof("add site to cache failed! err: %s", err)
				}

				isFind = true
				break
			}
		}

		if !isFind {
			site := &types.Site{
				SiteID: siteResource.Region + "--" + siteResource.AvailabilityZone,
				RegionAzMap: types.RegionAzMap{
					Region:           siteResource.Region,
					AvailabilityZone: siteResource.AvailabilityZone,
				},
				Status: constants.SiteStatusNormal,
			}

			site.Hosts = append(site.Hosts, siteResource.Hosts...)
			err := scheduler.Cache().AddSite(site)
			if err != nil {
				logger.Infof("add site to cache failed! err: %s", err)
			}
		}
	}

	scheduler.Cache().PrintString()
}

func convertToSite(site typed.SiteInfo, siteResource typed.SiteResource) *types.Site {
	result := &types.Site{
		SiteID: site.SiteID,
		GeoLocation: types.GeoLocation{
			Country:  site.Country,
			Area:     site.Area,
			Province: site.Province,
			City:     site.City,
		},
		RegionAzMap: types.RegionAzMap{
			Region:           site.Region,
			AvailabilityZone: site.AvailabilityZone,
		},
		Operator:      site.Operator.Name,
		EipTypeName:   site.EipTypeName,
		Status:        site.Status,
		SiteAttribute: site.SiteAttributes,
	}

	result.Hosts = append(result.Hosts, siteResource.Hosts...)
	return result
}
