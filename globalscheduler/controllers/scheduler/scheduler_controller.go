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
	"fmt"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/util/union"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1"
	clusterlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/listers/cluster/v1"
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	schedulerscheme "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned/scheme"
	schedulerinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions/scheduler/v1"
	schedulerlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/listers/scheduler/v1"
	schedulercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"

	"k8s.io/kubernetes/globalscheduler/controllers/util/consistenthashing"
	gld "k8s.io/kubernetes/globalscheduler/controllers/util/geoLocationDistribute"
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

	consistentHash        *consistenthashing.ConsistentHash
	geoLocationDistribute *gld.GeoLocationDistribute

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
	// mutex for updating schedulers
	mu sync.Mutex
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
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &SchedulerController{
		kubeclientset:          kubeclientset,
		apiextensionsclientset: apiextensionsclientset,
		schedulerclient:        schedulerclient,
		clusterclient:          clusterclient,
		schedulerInformer:      schedulerInformer.Lister(),
		clusterInformer:        clusterInformer.Lister(),
		schedulerSynced:        schedulerInformer.Informer().HasSynced,
		consistentHash:         consistenthashing.New(),
		geoLocationDistribute:  gld.New(),
		workqueue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Scheduler"),
		recorder:               recorder,
	}

	klog.Info("Setting up scheduler event handlers")
	schedulerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addScheduler,
		UpdateFunc: controller.updateScheduler,
		DeleteFunc: controller.deleteScheduler,
	})

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addCluster,
		UpdateFunc: controller.updateCluster,
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

		sc.geoLocationDistribute.SchedulersName = append(sc.geoLocationDistribute.SchedulersName, schedulerCopy.Name)

		err = sc.balanceScheduler(namespace)
		if err != nil {
			return fmt.Errorf("balanceScheduler function failed")
		}

		// We don't need homescheduler anymore. Will delete it in the future.
		// Update cluster HomeScheduler
		//err = sc.updateClusterBinding(schedulerCopy, namespace)
		//if err != nil {
		//	return fmt.Errorf("cluster HomeScheduler update failed")
		//}

		// Start Scheduler Process
		// command := "./hack/globalscheduler/start_scheduler.sh " + schedulerCopy.Spec.Tag + " " + schedulerCopy.Name
		// err = runCommand(command)
		// if err != nil {
		// 	return fmt.Errorf("start scheduler process failed")
		// }
		args := strings.Split(fmt.Sprintf("-n %s", name), " ")

		//	Format the command
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			klog.Fatalf("Failed to get the path to the process with the err %v", err)
		}

		cmd := exec.Command(path.Join(dir, "mock_scheduler"), args...)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		//	Run the command
		go cmd.Run()

		klog.V(2).Infof("Running process with the result: %v / %v\n", out.String(), stderr.String())

		sc.recorder.Event(schedulerCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeUpdateScheduler:
		klog.Infof("Event Type '%s'", EventTypeUpdateScheduler)
		schedulerCopy, err := sc.getScheduler(namespace, name)
		if err != nil {
			return err
		}

		err = sc.balanceScheduler(namespace)
		if err != nil {
			return fmt.Errorf("balanceScheduler function failed")
		}

		sc.recorder.Event(schedulerCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeDeleteScheduler:
		klog.Infof("Event Type '%s'", EventTypeDeleteScheduler)
		err = sc.balanceScheduler(namespace)
		if err != nil {
			return fmt.Errorf("balanceScheduler function failed")
		}
	case EventTypeAddCluster:
		klog.Infof("Event Type '%s'", EventTypeAddCluster)
		clusterCopy, err := sc.getCluster(namespace, name)
		if err != nil {
			return err
		}

		sc.geoLocationDistribute.ClustersName = append(sc.geoLocationDistribute.ClustersName, clusterCopy.Name)

		// Store the data according to Cluster geoLocation different fields
		sc.dataStoreClassified(clusterCopy)

		err = sc.balanceScheduler(namespace)
		if err != nil {
			return fmt.Errorf("balanceScheduler function failed")
		}

		sc.recorder.Event(clusterCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeUpdateCluster:
		klog.Infof("Event Type '%s'", EventTypeUpdateCluster)
		clusterCopy, err := sc.getCluster(namespace, name)
		if err != nil {
			return err
		}
		err = sc.balanceScheduler(namespace)
		if err != nil {
			return fmt.Errorf("balanceScheduler function failed")
		}

		sc.recorder.Event(clusterCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeDeleteCluster:
		klog.Infof("Event Type '%s'", EventTypeDeleteCluster)
		err = sc.balanceScheduler(namespace)
		if err != nil {
			return fmt.Errorf("balanceScheduler function failed")
		}
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

func (sc *SchedulerController) getCluster(namespace string, name string) (*clustercrdv1.Cluster, error) {
	// Get the Cluster resource with this namespace/name
	cluster, err := sc.clusterInformer.Clusters(namespace).Get(name)
	if err != nil {
		// The Cluster resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("failed to list cluster by: %s/%s", namespace, name))
			return cluster, nil
		}
		return nil, err
	}
	clusterCopy := cluster.DeepCopy()
	return clusterCopy, nil
}

func (sc *SchedulerController) addCluster(clusterObj interface{}) {
	sc.enqueue(clusterObj, EventTypeAddCluster)
}

func (sc *SchedulerController) addScheduler(schedulerObj interface{}) {
	sc.enqueue(schedulerObj, EventTypeCreateScheduler)
}

func (sc *SchedulerController) updateScheduler(old, new interface{}) {
	oldScheduler := old.(*schedulercrdv1.Scheduler)
	newScheduler := new.(*schedulercrdv1.Scheduler)
	if oldScheduler.ResourceVersion == newScheduler.ResourceVersion {
		return
	}
	sc.enqueue(new, EventTypeUpdateScheduler)
}

//deleteScheduler takes a deleted Scheduler resource and converts it into a namespace/name
//string which is then put into the work queue. This method should *not* be
//passed resources of any type other than Scheduler.
func (sc *SchedulerController) deleteScheduler(schedulerObj interface{}) {
	scheduler, ok := schedulerObj.(*schedulercrdv1.Scheduler)
	if !ok {
		klog.Warningf("Failed to convert an deleted object  %+v to a scheduler", schedulerObj)
		return
	}

	sc.geoLocationDistribute = gld.RemoveSchedulerName(scheduler, sc.geoLocationDistribute)
	sc.recorder.Event(scheduler, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(schedulerObj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	keyWithEventType := NewKeyWithEventType(EventTypeDeleteScheduler, key)
	sc.workqueue.AddRateLimited(keyWithEventType)
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

func (sc *SchedulerController) updateCluster(old, new interface{}) {
	oldCluster := old.(*clustercrdv1.Cluster)
	newCluster := new.(*clustercrdv1.Cluster)
	if oldCluster.ResourceVersion == newCluster.ResourceVersion {
		return
	}
	sc.enqueue(new, EventTypeUpdateCluster)
}

func (sc *SchedulerController) deleteCluster(obj interface{}) {
	cluster, ok := obj.(*clustercrdv1.Cluster)
	if !ok {
		klog.Warningf("Failed to convert an deleted object  %+v to a cluster", obj)
		return
	}

	sc.geoLocationDistribute = gld.RemoveClusterName(cluster, sc.geoLocationDistribute)
	sc.recorder.Event(cluster, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	keyWithEventType := NewKeyWithEventType(EventTypeDeleteCluster, key)
	sc.workqueue.AddRateLimited(keyWithEventType)
}

func (sc *SchedulerController) updateUnion(obj *schedulercrdv1.Scheduler, namespace string) error {
	newUnion := schedulercrdv1.ClusterUnion{}
	for _, v := range obj.Spec.Cluster {
		clusterObj, err := sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(v, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newUnion = union.UpdateUnion(newUnion, clusterObj)
	}
	obj.Spec.Union = newUnion
	_, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(obj)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SchedulerController) updateClusterBinding(schedulerCopy *schedulercrdv1.Scheduler, namespace string) error {
	clusterIdList := sc.consistentHash.GetIdList(schedulerCopy.Name)
	for _, v := range clusterIdList {
		clusterObj, err := sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(v, metav1.GetOptions{})
		if err != nil {
			return err
		}
		clusterObj.Spec.HomeScheduler = schedulerCopy.Name
		_, err = sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Update(clusterObj)
		if err != nil {
			return err
		}
	}

	return nil
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

func (sc *SchedulerController) dataStoreClassified(clusterCopy *clustercrdv1.Cluster) {
	clusterLocation := clusterCopy.Spec.GeoLocation
	clusterCountry := clusterLocation.Country
	clusterArea := clusterLocation.Area
	clusterProvince := clusterLocation.Province
	clusterCity := clusterLocation.City

	sc.geoLocationDistribute.CountryMap[clusterCountry] = append(sc.geoLocationDistribute.CountryMap[clusterCountry], clusterCopy.Name)
	sc.geoLocationDistribute.AreaMap[clusterArea] = append(sc.geoLocationDistribute.AreaMap[clusterArea], clusterCopy.Name)
	sc.geoLocationDistribute.ProvinceMap[clusterProvince] = append(sc.geoLocationDistribute.ProvinceMap[clusterProvince], clusterCopy.Name)
	sc.geoLocationDistribute.CityMap[clusterCity] = append(sc.geoLocationDistribute.CityMap[clusterCity], clusterCopy.Name)
}

func (sc *SchedulerController) balanceScheduler(namespace string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	schedulersLength := len(sc.geoLocationDistribute.SchedulersName)
	if len(sc.geoLocationDistribute.CountryMap) == 0 || schedulersLength == 0 {
		return nil
	}

	if len(sc.geoLocationDistribute.CountryMap) >= schedulersLength {
		if schedulersLength == 1 {
			schedulerName := sc.geoLocationDistribute.SchedulersName[0]
			schedulerObj, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Get(schedulerName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("scheduler object Get Failed")
			}

			schedulerObj.Spec.Cluster = sc.geoLocationDistribute.ClustersName

			unions := schedulercrdv1.ClusterUnion{}
			for _, cluster := range schedulerObj.Spec.Cluster {
				clusterObj, err := sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(cluster, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("cluster object get failed")
				}
				unions = union.UpdateUnion(unions, clusterObj)
			}
			schedulerObj.Spec.Union = unions

			_, err = sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(schedulerObj)
			if err != nil {
				return fmt.Errorf("scheduler object Update Failed")
			}
		} else {
			err := sc.updateSchedulersWithClusters(sc.geoLocationDistribute.CountryMap, namespace, schedulersLength)
			if err != nil {
				return err
			}
		}
	} else if len(sc.geoLocationDistribute.AreaMap) >= schedulersLength {
		err := sc.updateSchedulersWithClusters(sc.geoLocationDistribute.AreaMap, namespace, schedulersLength)
		if err != nil {
			return err
		}
	} else if len(sc.geoLocationDistribute.ProvinceMap) >= schedulersLength {
		err := sc.updateSchedulersWithClusters(sc.geoLocationDistribute.ProvinceMap, namespace, schedulersLength)
		if err != nil {
			return err
		}
	} else if len(sc.geoLocationDistribute.CityMap) >= schedulersLength {
		err := sc.updateSchedulersWithClusters(sc.geoLocationDistribute.CityMap, namespace, schedulersLength)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sc *SchedulerController) updateSchedulersWithClusters(dataMap map[string][]string, namespace string, schedulersLength int) error {
	schedulerIdx := 0
	mapClustersToScheduler := make(map[string][]string)
	for _, clusters := range dataMap {
		schedulerName := sc.geoLocationDistribute.SchedulersName[schedulerIdx]

		_, exist := mapClustersToScheduler[schedulerName]
		if exist {
			mapClustersToScheduler[schedulerName] = append(mapClustersToScheduler[schedulerName], clusters...)
		} else {
			mapClustersToScheduler[schedulerName] = clusters
		}

		schedulerIdx++
		if schedulerIdx == schedulersLength {
			schedulerIdx = 0
		}
	}

	for schedulerName, clusters := range mapClustersToScheduler {
		schedulerObj, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Get(schedulerName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("scheduler object Get Failed")
		}

		schedulerObj.Spec.Cluster = clusters
		unions := schedulercrdv1.ClusterUnion{}
		for _, clusterName := range clusters {
			clusterObj, err := sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(clusterName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("cluster object get failed")
			}
			unions = union.UpdateUnion(unions, clusterObj)
		}
		schedulerObj.Spec.Union = unions

		_, err = sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(schedulerObj)
		if err != nil {
			return fmt.Errorf("scheduler object Update Failed")
		}
	}

	return nil
}
