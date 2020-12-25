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

package dispatcher

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/util"
	"k8s.io/kubernetes/globalscheduler/controllers/util/consistenthashing"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1"
	clusterlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/listers/cluster/v1"
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	dispatcherclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned"
	dispatcherscheme "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned/scheme"
	dispatcherinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/informers/externalversions/dispatcher/v1"
	dispatcherlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/listers/dispatcher/v1"
	dispatchercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/v1"
	"strings"
	"time"
)

const (
	controllerAgentName   = "dispatcher-controller"
	SuccessSynced         = "Synced"
	MessageResourceSynced = "dispatcher synced successfully"
)

type DispatcherController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset    kubernetes.Interface
	dispatcherclient dispatcherclientset.Interface
	clusterclient    clusterclientset.Interface

	dispatcherInformer dispatcherlisters.DispatcherLister
	clusterInformer    clusterlisters.ClusterLister
	dispatcherSynced   cache.InformerSynced

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

// NewDispatcherController returns a new dispatcher controller
func NewDispatcherController(
	kubeclientset kubernetes.Interface,
	dispatcherclient dispatcherclientset.Interface,
	clusterclient clusterclientset.Interface,
	dispatcherInformer dispatcherinformers.DispatcherInformer,
	clusterInformer clusterinformers.ClusterInformer) *DispatcherController {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(dispatcherscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &DispatcherController{
		kubeclientset:      kubeclientset,
		dispatcherclient:   dispatcherclient,
		clusterclient:      clusterclient,
		dispatcherInformer: dispatcherInformer.Lister(),
		clusterInformer:    clusterInformer.Lister(),
		dispatcherSynced:   dispatcherInformer.Informer().HasSynced,
		consistentHash:     consistenthashing.New(),
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Dispatcher"),
		recorder:           recorder,
	}

	klog.Info("Setting up dispatcher event handlers")
	dispatcherInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addDispatcher,
		UpdateFunc: controller.updateDispatcher,
		DeleteFunc: controller.deleteDispatcher,
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
func (dc *DispatcherController) Run(workers int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer dc.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Dispatcher control loop")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, dc.dispatcherSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Infof("Starting %d workers", workers)
	for i := 0; i < workers; i++ {
		go wait.Until(dc.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (dc *DispatcherController) runWorker() {
	for dc.processNextWorkItem() {
	}
}

func (dc *DispatcherController) processNextWorkItem() bool {
	obj, shutdown := dc.workqueue.Get()

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
		defer dc.workqueue.Done(obj)
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
			dc.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected *KeyWithEventType in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Dispatcher resource to be synced.
		if err := dc.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key.Value, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		dc.workqueue.Forget(obj)
		klog.Infof("[ Dispatcher Controller ]Successfully synced '%s'", key.Value)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Dispatcher resource
// with the current status of the resource.
func (dc *DispatcherController) syncHandler(key *KeyWithEventType) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key.Value)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key.Value))
		return nil
	}

	switch key.EventType {
	case EventTypeCreateDispatcher:
		klog.Infof("Event Type '%s'", EventTypeCreateDispatcher)
		dispatcherCopy, err := dc.getDispatcher(namespace, name)
		if err != nil {
			return fmt.Errorf("dispatcher object get failed")
		}

		dispatcherInput := []string{dispatcherCopy.Name}
		dc.consistentHash.Add(dispatcherInput)
		for key, val := range dc.consistentHash.Results {
			dispatcherObj, err := dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Get(key, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("dispatcher object get failed")
			}

			dispatcherObj.Spec.Cluster = val
			_, err = dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Update(dispatcherObj)
			if err != nil {
				return fmt.Errorf("dispatcher object update failed")
			}
		}

		// Update cluster HomeDispatcher
		err = dc.updateClusterBinding(dispatcherCopy, namespace)
		if err != nil {
			return fmt.Errorf("cluster HomeDispatcher update failed")
		}

		// Start Dispatcher Process

		dc.recorder.Event(dispatcherCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeUpdateDispatcher:
		klog.Infof("Event Type '%s'", EventTypeUpdateDispatcher)
		dispatcherCopy, err := dc.getDispatcher(namespace, name)
		if err != nil {
			return fmt.Errorf("dispatcher object get failed")
		}

		if dispatcherCopy.State == "Delete" {
			clusterIds := dc.consistentHash.Results[dispatcherCopy.Name]
			dc.consistentHash.Remove(dispatcherCopy.Name)
			for key, val := range dc.consistentHash.Results {
				dispatcherObj, err := dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Get(key, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("dispatcher object get failed")
				}
				dispatcherObj.Spec.Cluster = val
				_, err = dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Update(dispatcherObj)
				if err != nil {
					return fmt.Errorf("dispatcher object update failed")
				}
			}

			for _, v := range clusterIds {
				idx := strings.Index(v, "&")
				clusterName := v[:idx]
				clusterObj, err := dc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(clusterName, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("cluster object get failed")
				}
				newHomeDispatcher := dc.consistentHash.Members[v]
				clusterObj.Spec.HomeDispatcher = newHomeDispatcher
				_, err = dc.clusterclient.GlobalschedulerV1().Clusters(namespace).Update(clusterObj)
				if err != nil {
					return fmt.Errorf("dispatcher object update failed")
				}
			}

			err = dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Delete(dispatcherCopy.Name, &metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("dispatcher object delete failed")
			}
		}
		dc.recorder.Event(dispatcherCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeAddCluster:
		klog.Infof("Event Type '%s'", EventTypeAddCluster)
		clusterCopy, err := dc.getCluster(namespace, name)
		if err != nil {
			return fmt.Errorf("cluster object get failed")
		}

		str := clusterCopy.Name + "&" + clusterCopy.Spec.IpAddress
		clusterInput := []string{str}
		err = dc.consistentHash.Insert(clusterInput)
		if err != nil {
			return fmt.Errorf("cluster name failed to insert into consistent hash circle")
		}

		dispatcherName := dc.consistentHash.Members[str]
		if dispatcherName != "nil" {
			clusterCopy.Spec.HomeDispatcher = dispatcherName
			_, err = dc.clusterclient.GlobalschedulerV1().Clusters(namespace).Update(clusterCopy)
			if err != nil {
				return fmt.Errorf("cluster object update failed")
			}

			dispatcherObj, err := dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Get(dispatcherName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("dispatcher object get failed")
			}

			dispatcherObj.Spec.Cluster = append(dispatcherObj.Spec.Cluster, str)
			_, err = dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Update(dispatcherObj)
			if err != nil {
				return fmt.Errorf("dispatcher object update failed")
			}
		}
		dc.recorder.Event(clusterCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	case EventTypeUpdateCluster:
		klog.Infof("Event Type '%s'", EventTypeUpdateCluster)
		clusterCopy, err := dc.getCluster(namespace, name)
		if err != nil {
			return fmt.Errorf("cluster object get failed")
		}

		if clusterCopy.Status == "Delete" {
			str := clusterCopy.Name + "&" + clusterCopy.Spec.IpAddress
			dc.consistentHash.Delete(str)
			dispatcher, err := dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Get(clusterCopy.Spec.HomeDispatcher, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("dispatcher object get failed")
			}

			dispatcherCopy := dispatcher.DeepCopy()
			dispatcherCopy.Spec.Cluster = util.RemoveCluster(dispatcherCopy.Spec.Cluster, str)

			_, err = dc.dispatcherclient.GlobalschedulerV1().Dispatchers(namespace).Update(dispatcherCopy)
			if err != nil {
				return fmt.Errorf("dispatcher object update failed")
			}

			err = dc.clusterclient.GlobalschedulerV1().Clusters(namespace).Delete(clusterCopy.Name, &metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("cluster object delete failed")
			}
		}
		dc.recorder.Event(clusterCopy, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	}

	return nil
}

func (dc *DispatcherController) getDispatcher(namespace string, name string) (*dispatchercrdv1.Dispatcher, error) {
	// Get the Dispatcher resource with this namespace/name
	dispatcher, err := dc.dispatcherInformer.Dispatchers(namespace).Get(name)
	if err != nil {
		// The Dispatcher resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("failed to list Dispatcher by: %s/%s", namespace, name))
			return dispatcher, nil
		}
		return nil, err
	}
	dispatcherCopy := dispatcher.DeepCopy()
	return dispatcherCopy, nil
}

func (dc *DispatcherController) addDispatcher(obj interface{}) {
	dc.enqueue(obj, EventTypeCreateDispatcher)
}

// enqueue takes a Dispatcher resource and converts it into a namespace/name
// string which is then put into the work queue. This method should *not* be
// passed resources of any type other than Dispatcher.
func (dc *DispatcherController) enqueue(obj interface{}, eventType EventType) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	keyWithEventType := NewKeyWithEventType(eventType, key)
	dc.workqueue.AddRateLimited(keyWithEventType)
}

func (dc *DispatcherController) updateClusterBinding(dispatcherCopy *dispatchercrdv1.Dispatcher, namespace string) error {
	clusterIdList := dc.consistentHash.GetIdList(dispatcherCopy.Name)
	for _, v := range clusterIdList {
		idx := strings.Index(v, "&")
		clusterName := v[:idx]
		clusterObj, err := dc.clusterclient.GlobalschedulerV1().Clusters(namespace).Get(clusterName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		clusterObj.Spec.HomeDispatcher = dispatcherCopy.Name
		_, err = dc.clusterclient.GlobalschedulerV1().Clusters(namespace).Update(clusterObj)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dc *DispatcherController) updateDispatcher(old, new interface{}) {
	oldDispatcher := old.(*dispatchercrdv1.Dispatcher)
	newDispatcher := new.(*dispatchercrdv1.Dispatcher)
	if oldDispatcher.ResourceVersion == newDispatcher.ResourceVersion {
		return
	}
	dc.enqueue(new, EventTypeUpdateDispatcher)
}

func (dc *DispatcherController) deleteDispatcher(obj interface{}) {
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	keyWithEventType := NewKeyWithEventType(EventTypeDeleteDispatcher, key)
	dc.workqueue.AddRateLimited(keyWithEventType)
}

func (dc *DispatcherController) addCluster(obj interface{}) {
	dc.enqueue(obj, EventTypeAddCluster)
}

func (dc *DispatcherController) updateCluster(old, new interface{}) {
	oldCluster := old.(*clustercrdv1.Cluster)
	newCluster := new.(*clustercrdv1.Cluster)
	if oldCluster.ResourceVersion == newCluster.ResourceVersion {
		return
	}
	dc.enqueue(new, EventTypeUpdateCluster)
}

func (dc *DispatcherController) deleteCluster(obj interface{}) {
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	keyWithEventType := NewKeyWithEventType(EventTypeDeleteCluster, key)
	dc.workqueue.AddRateLimited(keyWithEventType)
}

func (dc *DispatcherController) getCluster(namespace string, name string) (*clustercrdv1.Cluster, error) {
	// Get the Cluster resource with this namespace/name
	cluster, err := dc.clusterInformer.Clusters(namespace).Get(name)
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
