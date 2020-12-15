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
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	clusterclient "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client"
	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1"
	clusterlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/listers/cluster/v1"
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	schedulerclient "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client"
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
	kubeclientset   *kubernetes.Clientset
	schedulerclient *schedulerclient.SchedulerClient
	clusterclient   *clusterclient.ClusterClient

	schedulerInformer schedulerlisters.SchedulerLister
	clusterInformer   clusterlisters.ClusterLister
	schedulerSynced   cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewSchedulerController returns a new scheduler controller
func NewSchedulerController(
	kubeclientset *kubernetes.Clientset,
	schedulerclient *schedulerclient.SchedulerClient,
	clusterclient *clusterclient.ClusterClient,
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
		kubeclientset:     kubeclientset,
		schedulerclient:   schedulerclient,
		clusterclient:     clusterclient,
		schedulerInformer: schedulerInformer.Lister(),
		clusterInformer:   clusterInformer.Lister(),
		schedulerSynced:   schedulerInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Scheduler"),
		recorder:          recorder,
	}

	klog.Info("Setting up scheduler event handlers")
	schedulerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addScheduler,
		UpdateFunc: controller.updateScheduler,
		DeleteFunc: controller.deleteScheduler,
	})

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addClusterToScheduler,
		UpdateFunc: controller.updateClusterFromScheduler,
		DeleteFunc: controller.deleteClusterFromScheduler,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (sc *SchedulerController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer sc.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Scheduler control loop")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, sc.schedulerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
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
		klog.Infof("Successfully synced '%s'", key.Value)
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

	klog.Infof("Event Type '%s'", key.EventType)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key.Value)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key.Value))
		return nil
	}

	// Get the Scheduler resource with this namespace/name
	scheduler, err := sc.schedulerInformer.Schedulers(namespace).Get(name)
	if err != nil {
		// The Scheduler resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			klog.Warningf("SchedulerCRD: %s/%s does not exist in local cache, will delete it from Scheduler ...",
				namespace, name)

			klog.Infof("[SchedulerCRD] Deleting scheduler: %s/%s ...", namespace, name)

			return nil
		}

		runtime.HandleError(fmt.Errorf("failed to list scheduler by: %s/%s", namespace, name))

		return err
	}

	klog.Infof("[SchedulerCRD] Try to process scheduler: %#v ...", scheduler)

	sc.recorder.Event(scheduler, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (sc *SchedulerController) addClusterToScheduler(clusterObj interface{}) {
	fmt.Println("Watching Cluster Create")
	cluster := clusterObj.(*clustercrdv1.Cluster)

	schedulerList, err := sc.schedulerclient.List(metav1.ListOptions{})
	if err != nil {
		klog.Infof("Error listing all schedulers")
		return
	}

	for i := 0; i < len(schedulerList.Items); i++ {
		scheduler := schedulerList.Items[i]
		clusterArray := scheduler.Spec.Cluster
		if cluster.Spec.GeoLocation.Area == scheduler.Spec.Location.Area {
			scheduler.Spec.Cluster = append(clusterArray, cluster)
			_, err = sc.schedulerclient.Update(&scheduler)
			if err != nil {
				klog.Infof("Fail to update scheduler")
			}
			break
		}
	}
}

func (sc *SchedulerController) updateClusterFromScheduler(old, new interface{}) {
	oldCluster := old.(*clustercrdv1.Cluster)
	newCluster := new.(*clustercrdv1.Cluster)
	if oldCluster.ResourceVersion == newCluster.ResourceVersion {
		return
	}
	fmt.Println("Watching Cluster Update")
}

func (sc *SchedulerController) deleteClusterFromScheduler(clusterObj interface{}) {
	fmt.Println("Watching Cluster Delete")
	cluster := clusterObj.(*clustercrdv1.Cluster)

	schedulerList, err := sc.schedulerclient.List(metav1.ListOptions{})
	if err != nil {
		klog.Infof("Error listing all schedulers")
		return
	}

	for i := 0; i < len(schedulerList.Items); i++ {
		scheduler := schedulerList.Items[i]
		if cluster.Spec.GeoLocation.Area == scheduler.Spec.Location.Area {
			scheduler.Spec.Cluster = removeCluster(scheduler.Spec.Cluster, cluster)
			_, err = sc.schedulerclient.Update(&scheduler)
			if err != nil {
				klog.Infof("Fail to update scheduler")
			}
			break
		}
	}
}

func (sc *SchedulerController) addScheduler(schedulerObj interface{}) {
	sc.enqueueScheduler(schedulerObj, EventTypeCreate)
}

func (sc *SchedulerController) updateScheduler(old, new interface{}) {
	oldScheduler := old.(*schedulercrdv1.Scheduler)
	newScheduler := new.(*schedulercrdv1.Scheduler)
	if oldScheduler.ResourceVersion == newScheduler.ResourceVersion {
		return
	}
	sc.enqueueScheduler(new, EventTypeUpdate)
}

// deleteScheduler takes a deleted Scheduler resource and converts it into a namespace/name
// string which is then put into the work queue. This method should *not* be
// passed resources of any type other than Scheduler.
func (sc *SchedulerController) deleteScheduler(schedulerObj interface{}) {
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(schedulerObj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	keyWithEventType := NewKeyWithEventType(EventTypeDelete, key)
	sc.workqueue.AddRateLimited(keyWithEventType)
}

// enqueueScheduler takes a Scheduler resource and converts it into a namespace/name
// string which is then put into the work queue. This method should *not* be
// passed resources of any type other than Scheduler.
func (sc *SchedulerController) enqueueScheduler(obj interface{}, eventType EventType) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	keyWithEventType := NewKeyWithEventType(eventType, key)
	sc.workqueue.AddRateLimited(keyWithEventType)
}

func removeCluster(clusterArray []*clustercrdv1.Cluster, cluster *clustercrdv1.Cluster) []*clustercrdv1.Cluster {
	for idx, val := range clusterArray {
		if reflect.DeepEqual(val, cluster) {
			return append(clusterArray[:idx], clusterArray[idx+1:]...)
		}
	}

	return clusterArray
}
