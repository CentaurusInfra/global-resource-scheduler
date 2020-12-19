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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/util/union"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	"os/exec"
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
)

const (
	controllerAgentName   = "scheduler-controller"
	SuccessSynced         = "Synced"
	MessageResourceSynced = "scheduler synced successfully"
)

type SchedulerController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset   kubernetes.Interface
	schedulerclient schedulerclientset.Interface
	clusterclient   clusterclientset.Interface

	schedulerInformer schedulerlisters.SchedulerLister
	clusterInformer   clusterlisters.ClusterLister
	schedulerSynced   cache.InformerSynced

	consistentHash *consistenthashing.ConsistentHash

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
	kubeclientset kubernetes.Interface,
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
		kubeclientset:     kubeclientset,
		schedulerclient:   schedulerclient,
		clusterclient:     clusterclient,
		schedulerInformer: schedulerInformer.Lister(),
		clusterInformer:   clusterInformer.Lister(),
		schedulerSynced:   schedulerInformer.Informer().HasSynced,
		consistentHash:    consistenthashing.New(),
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
		AddFunc:    controller.addCluster,
		UpdateFunc: controller.updateCluster,
		DeleteFunc: controller.deleteCluster,
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
		schedulerCopy, err := sc.getSchedulerFromCache(namespace, name)
		if err != nil {
			klog.Infof("Scheduler Object Get From Cache Failed")
			return err
		}

		schedulerInput := []string{schedulerCopy.Name}
		sc.consistentHash.Add(schedulerInput)
		for key, val := range sc.consistentHash.Results {
			schedulerObj, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Get(key, metav1.GetOptions{})
			if err != nil {
				klog.Infof("Scheduler Object Get Failed")
				return err
			}

			schedulerObj.Spec.Cluster = val
			newUnion := schedulercrdv1.ClusterUnion{}
			for _, v := range schedulerObj.Spec.Cluster {
				clusterObj, err := sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(v, metav1.GetOptions{})
				if err != nil {
					klog.Infof("Cluster Object Get Failed")
					return err
				}
				newUnion = union.UpdateUnion(newUnion, clusterObj)
			}
			schedulerObj.Spec.Union = newUnion

			_, err = sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(schedulerObj)
			if err != nil {
				klog.Infof("Scheduler Object Update Failed")
				return err
			}
		}

		// Update cluster HomeScheduler
		err = sc.updateClusterBinding(schedulerCopy, namespace)
		if err != nil {
			klog.Infof("Cluster HomeScheduler Update Failed")
			return err
		}

		// Start Scheduler Process
		command := "./hack/globalscheduler/start_scheduler.sh " + schedulerCopy.Spec.Tag + " " + schedulerCopy.Name
		err = runCommand(command)
		if err != nil {
			klog.Infof("ERROR HERE")
			return err
		}
		sc.recorder.Event(schedulerCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeUpdateScheduler:
		klog.Infof("Event Type '%s'", EventTypeUpdateScheduler)
		schedulerCopy, err := sc.getSchedulerFromCache(namespace, name)
		if err != nil {
			klog.Infof("Scheduler Object Get From Cache Failed")
			return err
		}

		if schedulerCopy.Status == schedulercrdv1.SchedulerDelete {
			sc.consistentHash.Remove(schedulerCopy.Name)
			for key, val := range sc.consistentHash.Results {
				schedulerObj, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Get(key, metav1.GetOptions{})
				if err != nil {
					klog.Infof("Scheduler Object Get Failed")
					return err
				}
				schedulerObj.Spec.Cluster = val
				newUnion := schedulercrdv1.ClusterUnion{}
				for _, v := range schedulerObj.Spec.Cluster {
					clusterObj, err := sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(v, metav1.GetOptions{})
					if err != nil {
						klog.Infof("Cluster Object Get Failed")
						return err
					}
					newUnion = union.UpdateUnion(newUnion, clusterObj)
				}
				schedulerObj.Spec.Union = newUnion

				_, err = sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(schedulerObj)
				if err != nil {
					klog.Infof("Scheduler Object Update Failed")
					return err
				}
			}

			err = sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Delete(schedulerCopy.Name, &metav1.DeleteOptions{})
			if err != nil {
				klog.Infof("Scheduler Object Delete Failed")
				return err
			}

			// Delete scheduler process
			// TBD: Will add logic which will wait for the scheduler process to complete processing all PODs in its internal queue and then close the scheduler process
			command := "./hack/globalscheduler/close_scheduler.sh " + schedulerCopy.Spec.Tag
			err = runCommand(command)
			if err != nil {
				klog.Infof("ERROR HERE")
				return err
			}
		}
		sc.recorder.Event(schedulerCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeAddCluster:
		klog.Infof("Event Type '%s'", EventTypeAddCluster)
		clusterCopy, err := sc.getClusterFromCache(namespace, name)
		if err != nil {
			klog.Infof("Cluster Object Get From Cache Failed")
			return err
		}
		clusterInput := []string{clusterCopy.Name}
		err = sc.consistentHash.Insert(clusterInput)
		if err != nil {
			klog.Infof("Cluster Name Failed to Insert Into Consistent Hash Circle")
			return err
		}

		schedulerName := sc.consistentHash.Members[clusterCopy.Name]
		if schedulerName != "nil" {
			clusterCopy.Spec.HomeScheduler = schedulerName
			_, err = sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Update(clusterCopy)
			if err != nil {
				klog.Infof("Cluster Object Update Failed")
				return err
			}

			schedulerObj, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Get(schedulerName, metav1.GetOptions{})
			if err != nil {
				klog.Infof("Scheduler Object Get Failed")
				return err
			}

			schedulerObj.Spec.Cluster = append(schedulerObj.Spec.Cluster, clusterCopy.Name)

			newUnion := union.UpdateUnion(schedulerObj.Spec.Union, clusterCopy)
			schedulerObj.Spec.Union = newUnion
			_, err = sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(schedulerObj)
			if err != nil {
				klog.Infof("Scheduler Object Update Failed")
				return err
			}
		}
		sc.recorder.Event(clusterCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeUpdateCluster:
		klog.Infof("Event Type '%s'", EventTypeUpdateCluster)
		clusterCopy, err := sc.getClusterFromCache(namespace, name)
		if err != nil {
			klog.Infof("Cluster Object Get From Cache Failed")
			return err
		}

		if clusterCopy.State == "Delete" {
			sc.consistentHash.Delete(clusterCopy.Name)
			scheduler, err := sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Get(clusterCopy.Spec.HomeScheduler, metav1.GetOptions{})
			if err != nil {
				klog.Infof("Scheduler Object Get Failed")
				return err
			}

			schedulerCopy := scheduler.DeepCopy()
			schedulerCopy.Spec.Cluster = removeCluster(schedulerCopy.Spec.Cluster, clusterCopy.Name)
			newUnion := schedulercrdv1.ClusterUnion{}
			for _, v := range schedulerCopy.Spec.Cluster {
				clusterObj, err := sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(v, metav1.GetOptions{})
				if err != nil {
					klog.Infof("Cluster Object Get Failed")
					return err
				}
				newUnion = union.UpdateUnion(newUnion, clusterObj)
			}
			schedulerCopy.Spec.Union = newUnion

			_, err = sc.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(schedulerCopy)
			if err != nil {
				klog.Infof("Scheduler Object Update Failed")
				return err
			}

			err = sc.clusterclient.GlobalschedulerV1().Clusters(namespace).Delete(clusterCopy.Name, &metav1.DeleteOptions{})
			if err != nil {
				klog.Infof("Cluster Object Delete Failed")
				return err
			}
		}
		sc.recorder.Event(clusterCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	}
	return nil
}

func removeCluster(clusters []string, clusterName string) []string {
	var idx int
	for i, v := range clusters {
		if v == clusterName {
			idx = i
			break
		}
	}
	clusters[idx], clusters[len(clusters)-1] = clusters[len(clusters)-1], clusters[idx]
	return clusters[:len(clusters)-1]
}

func (sc *SchedulerController) getSchedulerFromCache(namespace string, name string) (*schedulercrdv1.Scheduler, error) {
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

func (sc *SchedulerController) getClusterFromCache(namespace string, name string) (*clustercrdv1.Cluster, error) {
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
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(schedulerObj)
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
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
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
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
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
