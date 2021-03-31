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
	"fmt"
	uuid "github.com/satori/go.uuid"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/cmd/conf"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	internalinformers "k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/algorithmprovider"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/config"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	frameworkplugins "k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins"
	internalcache "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache"
	internalqueue "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/queue"
	schedulersitecacheinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/wait"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/workqueue"
	//cluster
	clusterworkqueue "k8s.io/client-go/util/workqueue"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	clusterscheme "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned/scheme"
	externalinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions"
	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1"
	clusterlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/listers/cluster/v1"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

// ScheduleResult represents the result of one pod scheduled. It will contain
// the final selected Site, along with the selected intermediate information.
type ScheduleResult struct {
	SuggestedHost  string // Name of the scheduler suggest host
	Stacks         []types.Stack
	EvaluatedSites int // Number of site scheduler evaluated on one stack scheduled
	FeasibleSites  int // Number of feasible site on one stack scheduled
}

// Scheduler watches for new unscheduled pods. It attempts to find
// site that they fit on and writes bindings back to the api server.
type Scheduler struct {
	SchedulerName           string                  // Name of the current scheduler
	ResourceCollectorApiUrl string                  // Resource Collector API URL
	SchedulerCache          internalcache.Cache     // Scheduler's internal cache such as SiteTree or SiteList
	siteCacheInfoSnapshot   *internalcache.Snapshot // Sites' updated resource info cache
	ConfigFilePath          string                  // scheduling plugins list config

	StopEverything <-chan struct{} // Close this to shut down the scheduler.

	Plugins    *types.Plugins
	SchedFrame interfaces.Framework // policy are the scheduling policy.

	StackQueue      internalqueue.SchedulingQueue // queue for stacks that need scheduling
	PodInformer     coreinformers.PodInformer
	PodLister       corelisters.PodLister
	PodSynced       cache.InformerSynced
	Client          clientset.Interface
	InformerFactory internalinformers.SharedInformerFactory

	// NextStack should be a function that blocks until the next stack
	// is available. We don't use a channel for this, because scheduling
	// a stack may take some amount of time and we don't want pods to get
	// stale while they sit in a channel.
	NextStack func() *types.Stack

	mu sync.RWMutex

	//Cluster
	KubeClientset          clientset.Interface //kubernetes.Interface
	ApiextensionsClientset apiextensionsclientset.Interface
	ClusterClientset       clusterclientset.Interface
	ClusterInformerFactory externalinformers.SharedInformerFactory
	ClusterInformer        clusterinformers.ClusterInformer
	ClusterLister          clusterlisters.ClusterLister
	ClusterSynced          cache.InformerSynced
	ClusterQueue           clusterworkqueue.RateLimitingInterface
	deletedClusters        map[string]string //<key:namespace/name, value:region--az>
}

// single scheduler instance
var scheduler *Scheduler
var once sync.Once

func NewScheduler(gsconfig *types.GSSchedulerConfiguration, stopCh <-chan struct{}) (*Scheduler, error) {
	stopEverything := stopCh
	if stopEverything == nil {
		stopEverything = wait.NeverStop
	}

	sched := &Scheduler{
		SchedulerName:           gsconfig.SchedulerName,
		ResourceCollectorApiUrl: gsconfig.ResourceCollectorApiUrl,
		SchedulerCache:          internalcache.New(30*time.Second, stopEverything),
		siteCacheInfoSnapshot:   internalcache.NewEmptySnapshot(),
		ConfigFilePath:          gsconfig.ConfigFilePath,
		deletedClusters:         make(map[string]string),
	}

	err := sched.buildFramework()
	if err != nil {
		return nil, fmt.Errorf("buildFramework by %s failed! err: %v", types.SchedulerDefaultProviderName, err)
	}

	//build entire FlavorMap map<flovorid, flavorinfo>
	sched.UpdateFlavor()
	//sched.siteCacheInfoSnapshot.FlavorMap = config.ReadFlavorConf()
	klog.Infof("FlavorMap: %v", sched.siteCacheInfoSnapshot.FlavorMap)
	// init pod informers & cluster informers for scheduler
	err = sched.initPodClusterInformers(stopEverything)
	if err != nil {
		return nil, err
	}

	// add event handler
	AddAllEventHandlers(sched)
	return sched, nil
}

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
		klog.Errorf("Scheduler need to be init correctly")
		return scheduler
	}
	return scheduler
}

// start Scheduler - server.go calls this function to start Scheduler
func (sched *Scheduler) StartInformersAndRun(stopCh <-chan struct{}) {
	// start cluster informers
	if sched.ClusterInformer != nil && sched.ClusterInformerFactory != nil {
		//perform go informer.Run(stopCh) internally
		sched.ClusterInformerFactory.Start(stopCh) //perform go informer.Run(stopCh) internally
		// Wait for all caches to sync before scheduling.
		sched.ClusterInformerFactory.WaitForCacheSync(stopCh)
	}
	// start pod informers
	if sched.PodInformer != nil && sched.InformerFactory != nil {
		klog.Infof("Starting scheduler %s informer", sched.SchedulerName)
		//go sched.PodInformer.Informer().Run(stopCh2)
		sched.InformerFactory.Start(stopCh) //perform go informer.Run(stopCh) internally
		// Wait for all caches to sync before scheduling.
		sched.InformerFactory.WaitForCacheSync(stopCh)
	}
	// Do scheduling
	sched.Run(1, 1)
}

// Run begins watching and scheduling. It waits for cache to be synced, then starts scheduling
// and blocked until the context is done.
func (sched *Scheduler) Run(clusterWorkers int, podWorkers int) {
	klog.Infof("Starting scheduler %s", sched.SchedulerName)
	defer utilruntime.HandleCrash()

	//cluster
	if clusterWorkers > 0 {
		defer sched.ClusterQueue.ShutDown()
		klog.Infof("Waiting informer caches to sync")
		if ok := cache.WaitForCacheSync(sched.StopEverything, sched.ClusterSynced); !ok {
			klog.Errorf("failed to wait for caches to sync")
		}
		klog.Info("Starting cluster workers...")
		//perform runworker function until stopCh is closed
		for i := 0; i < clusterWorkers; i++ {
			go wait.Until(sched.runClusterWorker, time.Second, sched.StopEverything)
		}
	}

	//pod
	//defer sched.StackQueue.ShutDown()
	klog.Infof("Waiting informer caches to sync")
	if ok := cache.WaitForCacheSync(sched.StopEverything, sched.PodSynced); !ok {
		klog.Errorf("failed to wait for caches to sync")
	}
	klog.Info("Starting pod workers...")
	//perform runworker function until stopCh is closed
	for i := 0; i < podWorkers; i++ {
		go wait.Until(sched.runPodWorker, time.Second, sched.StopEverything)
	}
	//go wait.Until(sched.scheduleOne, 0, sched.StopEverything)
	klog.Info("Started cluster & pod workers")
	<-sched.StopEverything
	klog.Infof("Shutting down scheduler %s", sched.SchedulerName)
	//return nil
}

// Cache returns the cache in scheduler for test to check the data in scheduler.
func (sched *Scheduler) Cache() internalcache.Cache {
	return sched.SchedulerCache
}

func (sched *Scheduler) runPodWorker() {
	klog.Info("Starting a worker")
	for sched.scheduleOne() {
	}
}

// scheduleOne does the entire scheduling workflow for a single pod.
func (sched *Scheduler) scheduleOne() bool {
	// 1.pop queue and generate allocation from scheduler.StackQueue
	stack := sched.NextStack()
	klog.Infof("*** 1. Stack: %v stack selector: %v***", stack.Selector)

	//stack := sched.NextStack()
	allocation, err := sched.generateAllocationFromStack(stack)
	klog.Infof("*** 2. Allocation: %v , allocation selector: %v***", allocation.Selector)
	if err != nil {
		return false
	}
	start := stack.CreateTime
	end := time.Now().UnixNano()
	klog.Infof("=== done pop queue, time consumption: %vms ===", (end-start)/int64(time.Millisecond))

	// 2.do scheduling process
	start = end
	tmpContext := context.Background()
	result, err := sched.Schedule(tmpContext, allocation)
	if err != nil {
		klog.Errorf("Schedule failed, err: %s", err)
		sched.setPodScheduleErr(stack)
		return true
	}
	end = time.Now().UnixNano()
	klog.Infof("=== done Scheduling pipline, time consumption: %vms ===", (end-start)/int64(time.Millisecond))
	klog.Infof("Scheduler result: %v", result) //result is assumed stacks
	klog.Infof("*** 3. Assumed Stacks: %v", result)

	// 3.bind scheduler result to pod
	start = end
	klog.Infof("Try to bind to site, stacks:%v", result.Stacks)
	sched.bindStacks(result.Stacks)
	end = time.Now().UnixNano()
	klog.Infof("=== done bind pod to cluster, time consumption: %vms ===", (end-start)/int64(time.Millisecond))

	// log the elapsed time for the entire schedule
	if stack.CreateTime != 0 {
		spendTime := time.Now().UnixNano() - stack.CreateTime
		klog.Infof("@@@ Finished Schedule, time consumption: %vms @@@", spendTime/int64(time.Millisecond))
	}
	return true
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
	///allSiteCacheInfos, err := sched.siteCacheInfoSnapshot.SiteCacheInfos().List()
	siteID := stack.Selector.SiteID
	var allSiteCacheInfos [1]*schedulersitecacheinfo.SiteCacheInfo
	klog.Infof("sched.siteCacheInfoSnapshot.SiteCacheInfoMap ==> %v", sched.siteCacheInfoSnapshot.SiteCacheInfoMap)
	if sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteID] == nil {
		return nil, nil
	}
	klog.Infof("siteID ==> %v", siteID)
	allSiteCacheInfos[0] = sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteID]
	/*(if allSiteCacheInfos == nil {
		return nil, err
	}*/

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
		result = append(result, interfaces.SiteScore{SiteID: sites[i].SiteID, AZ: sites[i].RegionAzMap.AvailabilityZone, Score: 0, Region: sites[i].RegionAzMap.Region})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}

	// sort by score.
	sort.Sort(sort.Reverse(result))

	klog.Infof("score sites: %v", result)

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
func (sched *Scheduler) Schedule(ctx context.Context, allocation *types.Allocation) (result ScheduleResult, err error) {
	klog.Infof("Attempting to schedule allocation: %v", allocation.ID)

	state := interfaces.NewCycleState()
	schedulingCycleCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 1. Snapshot site resource cache
	start := time.Now()
	klog.Infof("[START] snapshot site...")

	///UpdateFlavorMap updates FlavorCache.RegionFlavorMap, FlavorCache.FlavorMap)
	///FlavorMap is updated when scheduler starts, RegionFlavorMap is updated
	///when cluster is added/updated. AllocatableFlavor is computed after binding
	sched.UpdateFlavor() //update sched.siteCacheInfoSnapshot.FlavorMap
	internalcache.FlavorCache.UpdateFlavorMap(sched.siteCacheInfoSnapshot.RegionFlavorMap, sched.siteCacheInfoSnapshot.FlavorMap)

	// 2. Run "prefilter" plugins.
	start = time.Now()
	klog.Infof("[START] Running prefilter plugins...")
	preFilterStatus := sched.SchedFrame.RunPreFilterPlugins(schedulingCycleCtx, state, &allocation.Stack)
	if !preFilterStatus.IsSuccess() {
		return result, preFilterStatus.AsError()
	}
	klog.Infof("[DONE] Running prefilter plugins, use_time: %s", time.Since(start).String())

	// 3. Run "filter" plugins.
	start = time.Now()
	klog.Infof("[START] Running filter plugins...")
	filteredSitesStatuses := make(interfaces.SiteToStatusMap)
	allocation.Stack.Selector = allocation.Selector
	filteredSites, err := sched.findSitesThatPassFilters(ctx, state, &allocation.Stack, filteredSitesStatuses)
	if err != nil {
		klog.Errorf("findSitesThatPassFilters failed! err: %s", err)
		return result, err
	}
	klog.Infof("[DONE] Running filter plugins, use_time: %s", time.Since(start).String())

	klog.Infof("filteredSitesStatuses = %v", filteredSitesStatuses.ToString())
	if len(filteredSites) <= 0 {
		err := fmt.Errorf("filter none site. resultStatus: %s", filteredSitesStatuses.ToString())
		klog.Error(err)
		return result, err
	}

	// 4. Run "prescore" plugins.
	start = time.Now()
	klog.Infof("[START] Running preScore plugins...")
	prescoreStatus := sched.SchedFrame.RunPreScorePlugins(ctx, state, &allocation.Stack, filteredSites)
	if !prescoreStatus.IsSuccess() {
		return result, prescoreStatus.AsError()
	}
	klog.Infof("[DONE] Running preScore plugins, use_time: %s", time.Since(start).String())

	// 5. Run "prioritizeSites" plugins.
	start = time.Now()
	klog.Infof("[START] Running prioritizeSites plugins...")
	priorityList, err := sched.prioritizeSites(ctx, state, &allocation.Stack, filteredSites)
	if err != nil {
		klog.Errorf("prioritizeSites failed! err: %s", err)
		return result, err
	}
	klog.Infof("[DONE] Running prioritizeSites plugins, use_time: %s", time.Since(start).String())

	// 6. Run "strategy" plugins.
	start = time.Now()
	klog.Infof("[START] Running strategy plugins...")
	siteCount, strategyStatus := sched.SchedFrame.RunStrategyPlugins(ctx, state, allocation, priorityList)
	if !strategyStatus.IsSuccess() {
		klog.Errorf("RunStrategyPlugins failed! err: %s", err)
		return result, err
	}
	klog.Infof("[DONE] Running StrategyPlugins plugins, use_time: %s", time.Since(start).String())

	klog.Infof("selected Hosts : %#v", siteCount)

	// 7. reserve resource
	start = time.Now()
	var count = 0
	for _, value := range siteCount {
		for i := 0; i < value.StackCount; i++ {
			newStack := allocation.Stack
			//bind
			err = sched.bind(ctx, &newStack, value.SiteID, state)
			if err != nil {
				klog.Errorf("bind host(%s) failed! err: %s", value.SiteID, err)
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
		klog.Errorf("not find suit host")
		return result, fmt.Errorf("not find suit host")
	}

	klog.Infof("reserve resource(%s) success, use_time: %s", allocation.ID, time.Since(start).String())
	return
}

func (sched *Scheduler) buildFramework() error {
	registry := frameworkplugins.NewRegistry()
	policyFile := config.String(constants.ConfPolicyFile)
	if policyFile == "" {
		klog.Errorf("policyFile(%s) not set!", constants.ConfPolicyFile)
		return fmt.Errorf("policyFile(%s) not set", constants.ConfPolicyFile)
	}

	policy := &types.Policy{}
	err := config.InitPolicyFromFile(policyFile, policy)
	if err != nil {
		klog.Errorf("InitPolicyFromFile %s failed! err: %s", constants.ConfPolicyFile, err)
		return err
	}

	defaultPlugins := algorithmprovider.GetPlugins(*policy)
	sched.SchedFrame, err = interfaces.NewFramework(registry, defaultPlugins,
		interfaces.WithSnapshotSharedLister(sched.siteCacheInfoSnapshot),
		interfaces.WithCache(sched.SchedulerCache))
	if err != nil {
		klog.Errorf("NewFramework failed! err : %s", err)
		return err
	}

	return nil
}

// initPodInformers init scheduler with podInformer
func (sched *Scheduler) initPodClusterInformers(stopCh <-chan struct{}) error {
	masterURL := config.DefaultString("master", "127.0.0.1:8080")
	kubeconfig := config.DefaultString("kubeconfig", "/var/run/kubernetes/admin.kubeconfig")

	// init kubeclient
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		return err
	}
	conf.AddQPSFlags(cfg, conf.GetInstance().Scheduler)
	client, err := clientset.NewForConfig(cfg) //kubeclientset
	if err != nil {
		return err
	}

	//pod
	sched.StackQueue = internalqueue.NewSchedulingQueue(stopCh, sched.SchedFrame)
	sched.InformerFactory = internalinformers.NewSharedInformerFactory(client, 0)
	sched.PodInformer = sched.InformerFactory.Core().V1().Pods()
	sched.PodLister = sched.PodInformer.Lister()
	sched.PodSynced = sched.PodInformer.Informer().HasSynced
	sched.NextStack = internalqueue.MakeNextStackFunc(sched.StackQueue)
	sched.Client = client

	///cluster, apiextensions clientset to create crd programmatically
	apiextensionsClientset, err := apiextensionsclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("error - building global scheduler apiextensions client: %s", err.Error())
	}
	clusterClientset, err := clusterclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building clusterclientset: %s", err.Error())
	}
	utilruntime.Must(clusterscheme.AddToScheme(clusterscheme.Scheme))
	sched.ClusterInformerFactory = externalinformers.NewSharedInformerFactory(clusterClientset, 0)
	sched.ClusterInformer = sched.ClusterInformerFactory.Globalscheduler().V1().Clusters()
	sched.ApiextensionsClientset = apiextensionsClientset
	sched.ClusterClientset = clusterClientset
	sched.ClusterLister = sched.ClusterInformer.Lister()
	sched.ClusterSynced = sched.ClusterInformer.Informer().HasSynced
	sched.ClusterQueue = clusterworkqueue.NewNamedRateLimitingQueue(clusterworkqueue.DefaultControllerRateLimiter(), "Cluster")

	return nil
}

func convertClusterToSite(cluster *clusterv1.Cluster) *types.Site {
	result := &types.Site{
		SiteID:           cluster.Spec.Region.Region + "--" + cluster.Spec.Region.AvailabilityZone,
		ClusterName:      cluster.ObjectMeta.Name,
		ClusterNamespace: cluster.ObjectMeta.Namespace,
		GeoLocation: types.GeoLocation{
			City:     cluster.Spec.GeoLocation.City,
			Province: cluster.Spec.GeoLocation.Province,
			Area:     cluster.Spec.GeoLocation.Area,
			Country:  cluster.Spec.GeoLocation.Country,
		},
		RegionAzMap: types.RegionAzMap{
			Region:           cluster.Spec.Region.Region,
			AvailabilityZone: cluster.Spec.Region.AvailabilityZone,
		},
		Operator:      cluster.Spec.Operator.Operator,
		Status:        cluster.Status,
		SiteAttribute: nil,
	}

	return result
}

func (sched *Scheduler) runClusterWorker() {
	klog.Info("Starting a worker")
	for sched.processNextClusterItem() {
	}
}

func (sched *Scheduler) processNextClusterItem() bool {
	workItem, shutdown := sched.ClusterQueue.Get()
	if shutdown {
		return false
	}
	klog.Infof("Process an item in work queue %v ", workItem)
	eventKey := workItem.(KeyWithEventType)
	key := eventKey.Key
	defer sched.ClusterQueue.Done(key)
	if err := sched.clusterSyncHandler(eventKey); err != nil {
		sched.ClusterQueue.AddRateLimited(eventKey)
		utilruntime.HandleError(fmt.Errorf("Handle %v of key %v failed with %v", "serivce", key, err))
	}
	sched.ClusterQueue.Forget(key)
	klog.Infof("Successfully processed & synced %s", key)
	return true
}

func (sched *Scheduler) clusterSyncHandler(keyWithEventType KeyWithEventType) error {
	if keyWithEventType.EventType < 0 {
		err := fmt.Errorf("cluster event is not create, update, or delete")
		return err
	}
	key := keyWithEventType.Key
	klog.Infof("sync cache for key %v", key)
	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing  %q (%v)", key, time.Since(startTime))
	}()
	nameSpace, clusterName, err := cache.SplitMetaNamespaceKey(key)

	//This performs controller logic - create site's static info
	klog.Infof("cluster processing - event: %v, cluster name: %v", keyWithEventType.EventType, clusterName)
	result, err := sched.updateStaticSiteResourceInfo(key, keyWithEventType.EventType, nameSpace, clusterName)
	if !result {
		klog.Errorf("Failed a cluster processing - event: %v, key: %v, error: %v", keyWithEventType, key, err)
		sched.ClusterQueue.AddRateLimited(keyWithEventType)
	} else {
		klog.Infof(" Processed a cluster: %v", key)
		sched.ClusterQueue.Forget(key)
	}
	klog.Infof("Cluster was handled by ClusterController - event: %v, cluster name: %v", keyWithEventType.EventType, clusterName)
	if keyWithEventType.EventType != EventType_Delete {
		cluster, err := sched.ClusterLister.Clusters(nameSpace).Get(clusterName)
		clusterCopy := cluster.DeepCopy()
		clusterCopy.Status = "HandledByClusterController"
		if err != nil || cluster == nil {
			klog.Errorf("Failed to retrieve cluster in local cache by cluster name: %s", clusterName)
			return err
		}
	}
	return nil
}

//This function determines if there is any actual change in cluster
//to improve performance by avoiding unnecessary update
func (sched *Scheduler) determineEventType(cluster1, cluster2 *clusterv1.Cluster) (event int, err error) {
	clusterName1, clusterSpec1, clusterStatus1, err1 := sched.getclusterInfo(cluster1)
	clusterName2, clusterSpec2, clusterStatus2, err2 := sched.getclusterInfo(cluster2)
	if cluster1 == nil || cluster2 == nil || err1 != nil || err2 != nil {
		err = fmt.Errorf("It cannot determine null clusters event type - cluster1: %v, cluster2:%v", cluster1, cluster2)
		return
	}
	event = ClusterUpdateYes
	if clusterName1 == clusterName2 && clusterStatus1 == clusterStatus2 && reflect.DeepEqual(clusterSpec1, clusterSpec2) == true {
		event = ClusterUpdateNo
	}
	return
}

// Retrieve cluster info
func (sched *Scheduler) getclusterInfo(cluster *clusterv1.Cluster) (clusterName string, clusterSpec clusterv1.ClusterSpec, clusterStatus string, err error) {
	if cluster == nil {
		err = fmt.Errorf("cluster is null")
		return
	}
	clusterName = cluster.ObjectMeta.Name
	if clusterName == "" {
		err = fmt.Errorf("cluster name is not valid - %s", clusterName)
		return
	}
	clusterSpec = cluster.Spec
	clusterStatus = cluster.Status
	return
}

//This function updates sites' static resource informaton
func (sched *Scheduler) updateStaticSiteResourceInfo(key string, event EventType, clusterNameSpace string, clusterName string) (response bool, err error) {
	switch event {
	case EventType_Create:
		cluster, err := sched.ClusterLister.Clusters(clusterNameSpace).Get(clusterName)
		clusterCopy := cluster.DeepCopy()
		if err != nil || clusterCopy == nil {
			klog.Errorf("Failed to retrieve cluster in local cache by cluster name: %s", clusterName)
			return false, err
		}
		klog.Infof("create a site static info, cluster profile: %v", clusterCopy)
		clusterCopy.Status = ClusterStatusCreated
		site := convertClusterToSite(clusterCopy)
		siteCacheInfo := schedulersitecacheinfo.NewSiteCacheInfo()
		//siteCacheInfo.SetSite(site)
		siteCacheInfo.Site = site
		sched.siteCacheInfoSnapshot.SiteCacheInfoMap[site.SiteID] = siteCacheInfo
		for _, flavor := range clusterCopy.Spec.Flavors {
			sched.siteCacheInfoSnapshot.SiteCacheInfoMap[site.SiteID].AllocatableFlavor[flavor.FlavorID] = flavor.TotalCapacity
			sched.UpdateRegionFlavor(clusterCopy.Spec.Region.Region, flavor.FlavorID)
		}
		//sched.UpdateSiteDynamicResource_Temp(clusterCopy.Spec.Region.Region, clusterCopy.Spec.Region.AvailabilityZone)
		//klog.Infof("created a site, site , id - site: %v", site.SiteID)
		klog.Infof("created a site, site id - site: %v", *(sched.siteCacheInfoSnapshot.SiteCacheInfoMap[site.SiteID].Site))
		klog.Infof("created a site, site id - map: %v", sched.siteCacheInfoSnapshot.SiteCacheInfoMap[site.SiteID])
		break
	case EventType_Update:
		cluster, err := sched.ClusterLister.Clusters(clusterNameSpace).Get(clusterName)
		clusterCopy := cluster.DeepCopy()
		if err != nil || clusterCopy == nil {
			klog.Errorf("Failed to retrieve cluster in local cache by cluster name - %s", clusterName)
			return false, err
		}
		klog.Infof("update a site static info, cluster profile: %v", clusterCopy)
		clusterCopy.Status = ClusterStatusUpdated
		site := convertClusterToSite(clusterCopy)
		siteCacheInfo := schedulersitecacheinfo.NewSiteCacheInfo()
		//siteCacheInfo.SetSite(site)
		siteCacheInfo.Site = site
		sched.siteCacheInfoSnapshot.SiteCacheInfoMap[site.SiteID] = siteCacheInfo
		for _, flavor := range clusterCopy.Spec.Flavors {
			sched.siteCacheInfoSnapshot.SiteCacheInfoMap[site.SiteID].AllocatableFlavor[flavor.FlavorID] = flavor.TotalCapacity
			sched.UpdateRegionFlavor(clusterCopy.Spec.Region.Region, flavor.FlavorID)
		}
		//sched.UpdateSiteDynamicResource_Temp(clusterCopy.Spec.Region.Region, clusterCopy.Spec.Region.AvailabilityZone)
		klog.Infof("created a site, site id: %v", site.SiteID)
		break
	case EventType_Delete:
		cluster, err := sched.ClusterLister.Clusters(clusterNameSpace).Get(clusterName)
		clusterCopy := cluster.DeepCopy()
		if clusterCopy == nil {
			klog.Errorf("Failed to retrieve cluster in map by cluster name - %s", clusterName)
			return false, err
		}
		siteID := sched.deletedClusters[key]
		delete(sched.siteCacheInfoSnapshot.SiteCacheInfoMap, siteID)
		delete(sched.deletedClusters, key)
		klog.Infof("created a site, site id: %v", siteID)
		break
	default:
		klog.Infof("cluster event %v is not correct", event)
		err = fmt.Errorf("cluster event %v is not correct", event)
		return false, err
	}
	return true, nil
}

//This function updates sites' dynamic resource informaton
func (sched *Scheduler) UpdateSiteDynamicResource(region string, resource *types.SiteResource) (err error) {
	klog.Infof("UpdateSiteDynamicResource1 region: %s, resource:%v", region, resource)
	var siteID string
	for _, siteresource := range resource.CPUMemResources {
		siteID = region + "--" + siteresource.AvailabilityZone
		klog.Infof("UpdateSiteDynamicResource2 siteID: %s", siteID)
		klog.Infof("UpdateSiteDynamicResource3 site: %v", siteresource)
		siteCacheInfo := sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteID]
		if siteCacheInfo == nil {
			siteCacheInfo = schedulersitecacheinfo.NewSiteCacheInfo()
			sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteID] = siteCacheInfo
		}
		sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteID].TotalResources[siteID] = &types.CPUAndMemory{VCPU: siteresource.CpuCapacity, Memory: siteresource.MemCapacity}
		for _, storage := range resource.VolumeResources {
			klog.Infof("UpdateSiteDynamicResource4 storage: %v", storage)
			sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteID].TotalStorage[storage.TypeId] = storage.StorageCapacity
			klog.Infof("UpdateSiteDynamicResource5 storage: %v", sched.siteCacheInfoSnapshot.SiteCacheInfoMap[siteID].TotalStorage[storage.TypeId])
			klog.Infof("UpdateSiteDynamicResource6 site: %v", sched.siteCacheInfoSnapshot.SiteCacheInfoMap)
		}
	}
	return nil
}

//This function updates sites' flavor
func (sched *Scheduler) UpdateFlavor() error {
	//sched.siteCacheInfoSnapshot.FlavorMap = config.ReadFlavorConf()
	sched.siteCacheInfoSnapshot.FlavorMap = config.FlavorMap
	return nil
}

//This function updates sites' flavor
func (sched *Scheduler) UpdateRegionFlavor(region string, flavorId string) (err error) {
	regionFlavorId := region + "--" + flavorId
	flavor := sched.siteCacheInfoSnapshot.FlavorMap[flavorId]
	if sched.siteCacheInfoSnapshot.RegionFlavorMap == nil {
		sched.siteCacheInfoSnapshot.RegionFlavorMap = make(map[string]*typed.RegionFlavor)
	}
	sched.siteCacheInfoSnapshot.RegionFlavorMap[regionFlavorId] = flavor
	err = nil
	return
}
