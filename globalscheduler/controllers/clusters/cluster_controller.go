/*
Copyright 2020 Authors of Global Scheduler.

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

package clusters

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/controller"

	clusterclienteset "globalscheduler/pkg/apis/clientset/versioned"
	clusterscheme "globalscheduler/pkg/apis/clientset/versioned/scheme"
	clusterlisters "globalscheduler/pkg/apis/listers/clusters/v1"
)

const (
	ClusterKind          string = "Cluster"
	ClusterStatusMessage string = "HANDLED"
)

// Cluster Controller Struct
type ClusterController struct {
	kubeclientset          *kubernetes.Clientset
	apiextensionsclientset apiextensionsclientset.Interface
	clusterclientset       clusterclienteset.Interface
	informers              cache.SharedIndexInformer
	informerSynced         cache.InformerSynced
	syncHandler            func(eventKey KeyWithEventType) error
	listers                clusterlisters.ClusterLister
	recorder               record.EventRecorder
	queue                  workqueue.RateLimitingInterface
}

func NewClusterController(kubeclientset *kubernetes.Clientset, clusterInformer informers.ClusterInformer) (*ClusterController, error) {
	informer := clusterInformer
	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(clusterscheme.Scheme, v1.EventSource{Component: "globalscheduler-cluster-controller"})
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&v1core.EventSinkImpl{Interface: kubeclientset.CoreV1().EventsWithMultiTenancy(metav1.NamespaceAll, metav1.TenantAll)})
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	c := &ClusterController{
		kubeclientset:  kubeclientset,
		informer:       informer,
		informerSynced: informer.Informer().HasSynced,
		lister:         informer.Lister(),
		recorder:       recorder,
		queue:          queue,
		grpcHost:       grpcHost,
	}

	klog.Infof("Sending events to api server")
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object interface{}) {
			key, err := controller.KeyFunc(object)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", object, err))
				return
			}
			c.Enqueue(key, EventType_Create)
			klog.Infof("Create cluster -%v ", key)
		},
		UpdateFunc: func(oldObject, newObject interface{}) {
			key1, err1 := controller.KeyFunc(oldObject)
			key2, err2 := controller.KeyFunc(newObject)
			if key1 == "" || key2 == "" || err1 != nil || err2 != nil {
				klog.Errorf("Unexpected string in queue; discarding - %v", key2)
				return
			}
			oldResource := oldObject.(*v1.cluster)
			newResource := newObject.(*v1.cluster)
			eventType, err := c.determineEventType(oldResource, newResource)
			if err != nil {
				klog.Errorf("Unexpected string in queue; discarding - %v ", key2)
				return
			}
			switch eventType {
			case CusterCreate:
				{
					klog.Infof("No actual change in clusters, discarding -%v ", newResource.Name)
					break
				}
			case ClusterDelete:
				{
					c.Enqueue(key2, EventType_Update)
					klog.Infof("Delete Cluster - %v", key2)
					break
				}
			case ClusterUpdate:
				{
					c.Enqueue(key2, EventType_Update)
					klog.Infof("Update Cluster - %v", key2)
					break
				}
			default:
				{
					klog.Errorf("Unexpected cluster event; discarding - %v", key2)
					return
				}
			}
		},
		DeleteFunc: func(object interface{}) {
			key, err := controller.KeyFunc(object)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", object, err))
				return
			}
			c.Enqueue(key, EventType_Delete)
			klog.Infof("Delete Cluster - %v", key)
		},
	})

	c.syncHandler = c.synccluster
	return c, nil
}

// Run starts an asynchronous loop that detects events of cluster clusters.
func (c *ClusterController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()
	klog.Infof("Starting cluster controller")
	klog.Infof("Waiting cache to be synced")

	ok := cache.WaitForCacheSync(stopCh, c.informerSynced)
	klog.Infof("111 - sync done")
	if !ok {
		klog.Infof("Timeout expired during waiting for caches to sync.")
	}
	klog.Infof("Starting workers...")
	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
		//go wait.Until(c.runWorker, time.Second, stopCh)
	}
	<-stopCh
	klog.Infof("Shutting down cluster controller")
}

// Enqueue puts key of the cluster object in the work queue
// EventType: Create=0, Update=1, Delete=2, Resume=3
func (c *ClusterController) Enqueue(key string, eventType EventType) {
	c.queue.Add(KeyWithEventType{Key: key, EventType: eventType})
}

func (c *ClusterController) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *ClusterController) processNextWorkItem() bool {
	workItem, quit := c.queue.Get()
	if quit {
		return false
	}

	eventKey := workItem.(KeyWithEventType)
	key := eventKey.Key
	defer c.queue.Done(key)

	err := c.syncHandler(eventKey)
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("Handle %v of key %v failed with %v", "serivce", key, err))
	c.queue.AddRateLimited(eventKey)

	return true
}

// Dequeue an item and process it
/*func (c *ClussterController) runWorker() {
	for {
		item, queueIsEmpty := c.queue.Get()
		if queueIsEmpty {
			return
		}
		if (item != nil){
			c.process(item)
		}
		//c.process(item)
	}
}
*/

func (c *ClusterController) syncCluster(keyWithEventType KeyWithEventType) error {
	//defer c.queue.Done(item)
	// keyWithEventType, ok := item.(KeyWithEventType)
	// if !ok {
	// 	klog.Errorf("Unexpected item in queue - %v", keyWithEventType)
	// 	c.queue.Forget(item)
	// 	return
	// }
	key := keyWithEventType.Key
	eventType := keyWithEventType.EventType

	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing service %q (%v)", key, time.Since(startTime))
	}()
	_, _, clusterName, err := cache.SplitMetaTenantNamespaceKey(key)
	cluster, err := c.lister.Get(clusterName)
	if err != nil || cluster == nil {
		klog.Errorf("Failed to retrieve cluster in local cache by cluster name - %s", clusterName)
		c.queue.AddRateLimited(keyWithEventType)
		return err
	}
	_, _, _, clusterAddress, err := c.getclusterInfo(cluster)
	if err != nil {
		klog.Errorf("Failed to retrieve cluster address in local cache by cluster name %v", clusterName)
		c.queue.AddRateLimited(keyWithEventType)
		return err
	}
	result, err := c.gRPCRequest(eventType, clusterName, clusterAddress)
	if !result {
		klog.Errorf("Failed a cluster processing - %v", key)
		c.queue.AddRateLimited(keyWithEventType)
	} else {
		klog.Infof(" Processed a cluster - %v", key)
		c.queue.Forget(key)
	}
	return nil
}

// Parsing a item key and call gRPC request
// func (c *GlobalSchedulerclusterController) process(item interface{}) {
// 	defer c.queue.Done(item)
// 	keyWithEventType, ok := item.(KeyWithEventType)
// 	if !ok {
// 		klog.Errorf("Unexpected item in queue - %v", keyWithEventType)
// 		c.queue.Forget(item)
// 		return
// 	}
// 	key := keyWithEventType.Key
// 	eventType := keyWithEventType.EventType
// 	_, _, clusterName, err := cache.SplitMetaTenantNamespaceKey(key)
// 	cluster, err := c.lister.Get(clusterName)
// 	if err != nil || cluster == nil {
// 		klog.Errorf("Failed to retrieve cluster in local cache by cluster name - %s", clusterName)
// 		c.queue.AddRateLimited(item)
// 		return
// 	}
// 	_, _, _, clusterAddress, err := c.getclusterInfo(cluster)
// 	if err != nil {
// 		klog.Errorf("Failed to retrieve cluster address in local cache by cluster name %v", clusterName)
// 		c.queue.AddRateLimited(item)
// 		return
// 	}
// 	result, err := c.gRPCRequest(eventType, clusterName, clusterAddress)
// 	if !result {
// 		klog.Errorf("Failed a cluster processing - %v", key)
// 		c.queue.AddRateLimited(item)
// 	} else {
// 		klog.Infof(" Processed a cluster - %v", key)
// 		c.queue.Forget(item)
// 	}
// }

// Retrieve cluster info
func (c *ClusterController) getclusterInfo(cluster *v1.cluster) (clusterTenant, clusterName, clusterStatus, clusterAddress string, err error) {
	if cluster == nil {
		err = fmt.Errorf("cluster is null")
		return
	}
	clusterTenant = cluster.GetTenant()
	clusterName = cluster.GetName()
	if clusterName == "" {
		err = fmt.Errorf("cluster name is not valid - %s", clusterName)
		return
	}
	conditions := cluster.Status.Conditions
	if conditions == nil {
		err = fmt.Errorf("cluster status information is not available - %s", clusterName)
		return
	}
	var clusterStatusType string
	for i := 0; i < len(conditions); i++ {
		clusterStatusType = fmt.Sprintf("%v", conditions[i].Type)
		clusterStatus = fmt.Sprintf("%v", conditions[i].Status)
		if clusterStatusType == clusterStatusType {
			break
		}
	}
	addresses := cluster.Status.Addresses
	if addresses == nil {
		err = fmt.Errorf("cluster address information is not available - %v", clusterName)
		return
	}
	var clusterAddressType string
	for i := 0; i < len(addresses); i++ {
		clusterAddressType = fmt.Sprintf("%s", addresses[i].Type)
		clusterAddress = fmt.Sprintf("%s", addresses[i].Address)
		if clusterAddressType == clusterInternalIP {
			break
		}
	}
	return
}

func (c *ClusterController) determineEventType(cluster1, cluster2 *v1.cluster) (event int, err error) {
	clusterTenant1, clusterName1, clusterStatus1, clusterAddress1, err1 := c.getclusterInfo(cluster1)
	clusterTenant2, clusterName2, clusterStatus2, clusterAddress2, err2 := c.getclusterInfo(cluster1)
	if cluster1 == nil || cluster2 == nil || err1 != nil || err2 != nil {
		err = fmt.Errorf("It cannot determine null clusters event type - cluster1: %v, cluster2:%v", cluster1, cluster2)
		return
	}
	event = clusterUpdate
	if clusterTenant1 == clusterTenant2 && clusterName1 == clusterName2 && clusterAddress1 == clusterAddress2 && clusterStatus1 == clusterStatus2 {
		event = clusterNoChange
	} else if clusterStatus1 != clusterStatus2 && clusterStatus2 == clusterReadyTrue {
		event = clusterResume
	}
	return
}
