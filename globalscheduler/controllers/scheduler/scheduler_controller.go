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
	"bytes"
	"encoding/json"
	"fmt"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/util"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1"
	clusterlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/listers/cluster/v1"
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	schedulerscheme "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned/scheme"
	schedulerinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions/scheduler/v1"
	schedulerlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/listers/scheduler/v1"
	schedulercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
)

const (
	controllerAgentName   = "scheduler-controller"
	SuccessSynced         = "Synced"
	MessageResourceSynced = "scheduler synced successfully"
)

type SchedulerController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset          kubernetes.Interface
	apiextensionsclientset apiextensionsclientset.Interface
	schedulerclient        schedulerclientset.Interface
	clusterclient          clusterclientset.Interface

	schedulerInformer schedulerlisters.SchedulerLister
	clusterInformer   clusterlisters.ClusterLister
	schedulerSynced   cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// mutex for updating schedulers
	mu sync.Mutex
	// Create a tree of all the geolocations and their union information
	countryNode *ClusterInfoNode
	// sccheduler list
	schedulers []*schedulercrdv1.Scheduler
}

// NewSchedulerController returns a new scheduler controller
func NewSchedulerController(
	kubeclientset kubernetes.Interface,
	apiextensionsclientset apiextensionsclientset.Interface,
	schedulerclient schedulerclientset.Interface,
	clusterclient clusterclientset.Interface,
	schedulerInformer schedulerinformers.SchedulerInformer,
	clusterInformer clusterinformers.ClusterInformer) *SchedulerController {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(schedulerscheme.AddToScheme(scheme.Scheme))

	controller := &SchedulerController{
		kubeclientset:          kubeclientset,
		apiextensionsclientset: apiextensionsclientset,
		schedulerclient:        schedulerclient,
		clusterclient:          clusterclient,
		schedulerInformer:      schedulerInformer.Lister(),
		clusterInformer:        clusterInformer.Lister(),
		schedulerSynced:        schedulerInformer.Informer().HasSynced,
		workqueue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Scheduler"),
		countryNode:            NewClusterInfoNode(),
		schedulers:             make([]*schedulercrdv1.Scheduler, 0),
	}

	klog.Info("Setting up scheduler event handlers")
	schedulerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addScheduler,
		DeleteFunc: controller.deleteScheduler,
	})

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addCluster,
		DeleteFunc: controller.deleteCluster,
	})

	return controller
}

func (sc *SchedulerController) RunController(threadiness int, stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	sc.Run(threadiness, stopCh)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (sc *SchedulerController) Run(workers int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer sc.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Scheduler control loop")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, sc.schedulerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Infof("Starting %d workers", workers)
	for i := 0; i < workers; i++ {
		go wait.Until(sc.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (sc *SchedulerController) runWorker() {
	for sc.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (sc *SchedulerController) processNextWorkItem() bool {
	obj, shutdown := sc.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer sc.workqueue.Done(obj)
		var key *KeyWithEventType
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(*KeyWithEventType); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			sc.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected *KeyWithEventType in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Scheduler resource to be synced.
		if err := sc.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key.Value, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		sc.workqueue.Forget(obj)
		klog.Infof("[ Scheduler Controller ]Successfully synced '%s'", key.Value)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Scheduler resource
// with the current status of the resource.
func (sc *SchedulerController) syncHandler(key *KeyWithEventType) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key.Value)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key.Value))
		return nil
	}

	switch key.EventType {
	case EventTypeCreateScheduler:
		klog.Infof("Event Type '%s'", EventTypeCreateScheduler)
		schedulerCopy, err := sc.getScheduler(namespace, name)
		if err != nil {
			return err
		}

		// We don't need homescheduler anymore. Will delete it in the future.
		// Update cluster HomeScheduler
		//err = sc.updateClusterBinding(schedulerCopy, namespace)
		//if err != nil {
		//	return fmt.Errorf("cluster HomeScheduler update failed")
		//}

		// Start Scheduler Process
		command := "./hack/globalscheduler/start_scheduler.sh " + schedulerCopy.Spec.Tag + " " + schedulerCopy.Name + " " + schedulerCopy.Spec.IpAddress + " " + schedulerCopy.Spec.PortNumber
		err = runCommand(command)
		if err != nil {
			return fmt.Errorf("start scheduler process failed")
		}

		// call resource collector API to send scheduler IP and Port
		finalData, _ := json.Marshal(schedulerCopy)
		schedulerEndpoint := "http://" + constants.DefaultResourceCollectorAPIURL + constants.SchedulerRegisterURL
		req, _ := http.NewRequest("POST", schedulerEndpoint, bytes.NewBuffer(finalData))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			klog.V(3).Infof("HTTP Post Instance Request Failed: %v", err)
			return err
		}
		defer resp.Body.Close()

		//args := strings.Split(fmt.Sprintf("-n %s", name), " ")
		//
		////	Format the command
		//dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		//if err != nil {
		//	klog.Fatalf("Failed to get the path to the process with the err %v", err)
		//}
		//
		//cmd := exec.Command(path.Join(dir, "mock_scheduler"), args...)
		//var out bytes.Buffer
		//var stderr bytes.Buffer
		//cmd.Stdout = &out
		//cmd.Stderr = &stderr
		//
		////	Run the command
		//go cmd.Run()

		//klog.V(2).Infof("Running process with the result: %v / %v\n", out.String(), stderr.String())

		//sc.recorder.Event(schedulerCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeDeleteScheduler:
		klog.Infof("Event Type '%s'", EventTypeDeleteScheduler)
	}
	return nil
}

func (sc *SchedulerController) getScheduler(namespace string, name string) (*schedulercrdv1.Scheduler, error) {
	// Get the Scheduler resource with this namespace/name
	scheduler, err := sc.schedulerInformer.Schedulers(namespace).Get(name)
	if err != nil {
		// The Scheduler resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("failed to list scheduler by: %s/%s", namespace, name))
			return scheduler, nil
		}
		return nil, err
	}
	schedulerCopy := scheduler.DeepCopy()
	return schedulerCopy, nil
}

func (sc *SchedulerController) addCluster(clusterObj interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if cluster, ok := clusterObj.(*clustercrdv1.Cluster); ok {
		sc.countryNode.AddCluster(cluster, 0)
		if err := sc.balance(cluster.ObjectMeta.GetNamespace(), []*ClusterInfoNode{sc.countryNode}); err != nil {
			klog.Warningf("Failed to balance schedulers with error %v", err)
		}
	} else {
		klog.Warningf("Failed to convert the added object %v to a cluster", clusterObj)
	}
}

func (sc *SchedulerController) addScheduler(schedulerObj interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if scheduler, ok := schedulerObj.(*schedulercrdv1.Scheduler); ok {
		sc.schedulers = append(sc.schedulers, scheduler)
		if err := sc.balance(scheduler.ObjectMeta.GetNamespace(), []*ClusterInfoNode{sc.countryNode}); err != nil {
			klog.Warningf("Failed to balance schedulers with error %v", err)
		} else {
			klog.V(2).Infof("The scheduler %s has been balanced\n", scheduler.GetName())
			sc.enqueue(schedulerObj, EventTypeCreateScheduler)
		}

	} else {
		klog.Warningf("Failed to convert the added object %v to a scheduler", schedulerObj)
	}
}

//deleteScheduler takes a deleted Scheduler resource and converts it into a namespace/name
//string which is then put into the work queue. This method should *not* be
//passed resources of any type other than Scheduler.
func (sc *SchedulerController) deleteScheduler(schedulerObj interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if scheduler, ok := schedulerObj.(*schedulercrdv1.Scheduler); ok {
		schedulerLen := len(sc.schedulers)
		for idx, sched := range sc.schedulers {
			if scheduler.Name == sched.Name && scheduler.Namespace == sched.Namespace {
				sc.schedulers[idx] = sc.schedulers[schedulerLen-1]
				sc.schedulers = sc.schedulers[:schedulerLen-1]
				break
			}
		}
		if err := sc.balance(scheduler.ObjectMeta.GetNamespace(), []*ClusterInfoNode{sc.countryNode}); err != nil {
			klog.Warningf("Failed to balance schedulers with error %v", err)
		}
	} else {
		klog.Warningf("Failed to convert the added object %v to a scheduler", schedulerObj)
	}
}

// enqueue takes a Scheduler resource and converts it into a namespace/name
// string which is then put into the work queue. This method should *not* be
// passed resources of any type other than Scheduler.
func (sc *SchedulerController) enqueue(obj interface{}, eventType EventType) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	keyWithEventType := NewKeyWithEventType(eventType, key)
	sc.workqueue.AddRateLimited(keyWithEventType)
}

func (sc *SchedulerController) deleteCluster(clusterObj interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if cluster, ok := clusterObj.(*clustercrdv1.Cluster); ok {
		sc.countryNode.DelCluster(cluster, 0)
		if err := sc.balance(cluster.ObjectMeta.GetNamespace(), []*ClusterInfoNode{sc.countryNode}); err != nil {
			klog.Warningf("Failed to balance schedulers with error %v", err)
		}
	} else {
		klog.Warningf("Failed to convert the deleted object %v to a cluster", clusterObj)
	}
}

func (sc *SchedulerController) balance(namespace string, nodes []*ClusterInfoNode) error {
	schedulersLen := len(sc.schedulers)
	nodesLen := len(nodes)

	if nodesLen == 0 || schedulersLen == 0 {
		return nil
	}

	if schedulersLen > nodesLen {
		nextLevelNodes := make([]*ClusterInfoNode, 0)
		for _, node := range nodes {
			if len(node.infoMap) > 0 {
				for _, nextLevelNode := range node.infoMap {
					nextLevelNodes = append(nextLevelNodes, nextLevelNode)
				}
			}
		}
		return sc.balance(namespace, nextLevelNodes)
	}

	nodeIdxBounds := util.EvenlyDivide(schedulersLen, int64(nodesLen-1))

	for schedulerIdx, nodeIdxBound := range nodeIdxBounds {
		clusters := make(map[string]bool)
		unionGeoLocation := make(map[clustercrdv1.GeolocationInfo]bool)
		unionRegion := make(map[clustercrdv1.RegionInfo]bool)
		unionOperator := make(map[clustercrdv1.OperatorInfo]bool)
		unionFlavor := make(map[clustercrdv1.FlavorInfo]bool)
		unionStorage := make(map[clustercrdv1.StorageSpec]bool)
		unionEipCapacity := make(map[int64]bool)
		unionCPUCapacity := make(map[int64]bool)
		unionMemCapacity := make(map[int64]bool)
		unionServerPrice := make(map[int64]bool)

		for idx := nodeIdxBound[0]; idx <= nodeIdxBound[1]; idx++ {
			for key := range nodes[idx].clusterMap {
				clusters[key.(string)] = true
			}
			for key := range nodes[idx].clusterUnionMap.GeoLocationMap {
				unionGeoLocation[key.(clustercrdv1.GeolocationInfo)] = true
			}
			for key := range nodes[idx].clusterUnionMap.RegionMap {
				unionRegion[key.(clustercrdv1.RegionInfo)] = true
			}
			for key := range nodes[idx].clusterUnionMap.OperatorMap {
				unionOperator[key.(clustercrdv1.OperatorInfo)] = true
			}
			for key := range nodes[idx].clusterUnionMap.FlavorMap {
				unionFlavor[key.(clustercrdv1.FlavorInfo)] = true
			}
			for key := range nodes[idx].clusterUnionMap.StorageMap {
				unionStorage[key.(clustercrdv1.StorageSpec)] = true
			}
			for key := range nodes[idx].clusterUnionMap.EipCapacityMap {
				unionEipCapacity[key.(int64)] = true
			}
			for key := range nodes[idx].clusterUnionMap.CPUCapacityMap {
				unionCPUCapacity[key.(int64)] = true
			}
			for key := range nodes[idx].clusterUnionMap.MemCapacityMap {
				unionMemCapacity[key.(int64)] = true
			}
			for key := range nodes[idx].clusterUnionMap.ServerPriceMap {
				unionServerPrice[key.(int64)] = true
			}
		}
		latestScheduler, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Get(sc.schedulers[schedulerIdx].GetName(), metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Failed to get scheduler %s", sc.schedulers[schedulerIdx].GetName())
		}
		latestScheduler.Spec.Union = schedulercrdv1.ClusterUnion{}

		combinedCluster := make([]string, 0)
		for key := range clusters {
			combinedCluster = append(combinedCluster, key)
		}
		latestScheduler.Spec.Cluster = combinedCluster

		combinedGeoLcation := make([]*clustercrdv1.GeolocationInfo, 0)
		for key := range unionGeoLocation {
			combinedGeoLcation = append(combinedGeoLcation, key.DeepCopy())
		}
		latestScheduler.Spec.Union.GeoLocation = combinedGeoLcation

		combinedRegion := make([]*clustercrdv1.RegionInfo, 0)
		for key := range unionRegion {
			combinedRegion = append(combinedRegion, key.DeepCopy())
		}
		latestScheduler.Spec.Union.Region = combinedRegion

		combinedOperator := make([]*clustercrdv1.OperatorInfo, 0)
		for key := range unionOperator {
			combinedOperator = append(combinedOperator, key.DeepCopy())
		}
		latestScheduler.Spec.Union.Operator = combinedOperator

		combinedFlavor := make([]*clustercrdv1.FlavorInfo, 0)
		for key := range unionFlavor {
			combinedFlavor = append(combinedFlavor, key.DeepCopy())
		}
		latestScheduler.Spec.Union.Flavors = combinedFlavor

		combinedStorage := make([]*clustercrdv1.StorageSpec, 0)
		for key := range unionStorage {
			combinedStorage = append(combinedStorage, key.DeepCopy())
		}
		latestScheduler.Spec.Union.Storage = combinedStorage
		latestScheduler.Spec.Union.EipCapacity = getMapKeys(unionEipCapacity)
		latestScheduler.Spec.Union.CPUCapacity = getMapKeys(unionCPUCapacity)
		latestScheduler.Spec.Union.MemCapacity = getMapKeys(unionMemCapacity)
		latestScheduler.Spec.Union.ServerPrice = getMapKeys(unionServerPrice)

		updatedScheduler, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(latestScheduler)
		if err != nil {
			return err
		}
		sc.schedulers[schedulerIdx] = updatedScheduler
	}
	return nil
}

func getMapKeys(m map[int64]bool) []int64 {
	res := make([]int64, 0)
	for key := range m {
		res = append(res, key)
	}
	return res
}

func runCommand(command string) error {
	cmd := exec.Command("/bin/bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
