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

package cluster

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	grpc "k8s.io/kubernetes/globalscheduler/grpc/cluster"
	clienteset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	clusterscheme "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned/scheme"
	informers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1"
	listers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/listers/cluster/v1"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	"k8s.io/kubernetes/pkg/controller"
)

const ControllerAgentName = "globalscheduler-cluster-controller"
const (
	SuccessSynched         = "Synched"
	MessageResourceSynched = "Cluster synced successfully"
	ClusterKind            = "Cluster"
	ClusterStatusCreated   = "Created"
	ClusterStatusUpdated   = "Updated"
	ClusterStatusDeleted   = "Deleted"
)

type EventType int

const (
	EventType_Create EventType = 0
	EventType_Update EventType = 1
	EventType_Delete EventType = 2
)

type KeyWithEventType struct {
	EventType EventType
	Key       string
}

const (
	ClusterUpdateNo  int = 1
	ClusterUpdateYes int = 2
)

// Cluster Controller Struct
type ClusterController struct {
	kubeclientset          kubernetes.Interface
	apiextensionsclientset apiextensionsclientset.Interface
	clusterclientset       clienteset.Interface
	clusterlister          listers.ClusterLister
	clusterSynced          cache.InformerSynced
	workqueue              workqueue.RateLimitingInterface
	recorder               record.EventRecorder
	grpcHost               string
	deletedClusters        map[string]*clusterv1.Cluster
}

func NewClusterController(
	kubeclientset kubernetes.Interface,
	apiextensionsclientset apiextensionsclientset.Interface,
	clusterclientset clienteset.Interface,
	clusterInformer informers.ClusterInformer,
	grpcHost string) *ClusterController {
	utilruntime.Must(clusterscheme.AddToScheme(clusterscheme.Scheme))
	klog.Info("Creating cluster event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(clusterscheme.Scheme, corev1.EventSource{Component: ControllerAgentName})
	workqueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Cluster")
	klog.Info("Creating a new cluster controller")
	c := &ClusterController{
		kubeclientset:          kubeclientset,
		apiextensionsclientset: apiextensionsclientset,
		clusterclientset:       clusterclientset,
		clusterlister:          clusterInformer.Lister(),
		clusterSynced:          clusterInformer.Informer().HasSynced,
		workqueue:              workqueue,
		recorder:               recorder,
		grpcHost:               grpcHost,
		deletedClusters:        make(map[string]*clusterv1.Cluster),
	}

	//KeyFunc is defined at controller.lookup_cache.go
	klog.Infof("Setting up event handlers")
	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addCluster,
		UpdateFunc: c.updateCluster,
		DeleteFunc: c.deleteCluster,
	})
	return c
}

func (c *ClusterController) addCluster(object interface{}) {
	resource := object.(*clusterv1.Cluster)
	clusterCopy := resource.DeepCopy()
	if verifyClusterInfo(clusterCopy) == false {
		klog.Infof(" Cluster data is not correct: %v", cluster)
	}
	key, err := controller.KeyFunc(object)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object: %v, error: %v", object, err))
		return
	}
	c.Enqueue(key, EventType_Create)
	klog.Infof("Enqueue Create cluster: %v", key)
}

func (c *ClusterController) updateCluster(oldObject, newObject interface{}) {
	oldResource := oldObject.(*clusterv1.Cluster)
	newResource := newObject.(*clusterv1.Cluster)
	oldClusterCopy := oldResource.DeepCopy()
	newClusterCopy := newResource.DeepCopy()
	if verifyClusterInfo(newClusterCopy) {
		klog.Infof(" Cluster data is not correct: %v", newResource)
	}
	key1, err1 := controller.KeyFunc(oldObject)
	key2, err2 := controller.KeyFunc(newObject)
	if key1 == "" || key2 == "" || err1 != nil || err2 != nil {
		klog.Errorf("Unexpected string in queue; discarding: %v", key2)
		return
	}

	eventType, err := c.determineEventType(oldClusterCopy, newClusterCopy)
	if err != nil {
		klog.Errorf("Unexpected string in queue; discarding: %v", key2)
		return
	}
	switch eventType {
	case ClusterUpdateNo:
		{
			klog.Infof("No actual change in clusters, discarding: %v", newClusterCopy.Name)
			break
		}
	case ClusterUpdateYes:
		{
			c.Enqueue(key2, EventType_Update)
			klog.Infof("Enqueue Update Cluster: %v", key2)
			break
		}
	default:
		{
			klog.Errorf("Unexpected cluster update event; discarding: %v", key2)
			return
		}
	}
}

func (c *ClusterController) deleteCluster(object interface{}) {
	resource := object.(*clusterv1.Cluster)
	clusterCopy := resource.DeepCopy()
	if verifyClusterInfo(clusterCopy) == false {
		klog.Infof(" Cluster data is not correct: %v", cluster)
		return
	}	
	key, err := controller.KeyFunc(object)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object: %v, error: %v", object, err))
		return
	}
	c.Enqueue(key, EventType_Delete)
	c.deletedClusters[key] = clusterCopy
	klog.Infof("Enqueue Delete Cluster: %v", key)
}

// Enqueue puts key of the cluster object in the work queue
// EventType: Create=0, Update=1, Delete=2
func (c *ClusterController) Enqueue(key string, eventType EventType) {
	c.workqueue.Add(KeyWithEventType{Key: key, EventType: eventType})
}

func (c *ClusterController) RunController(workers int, stopCh <-chan struct{}, wg *sync.WaitGroup) {
	klog.Info("gs-controllers-manager starts cluster controller")
	defer wg.Done()
	c.Run(workers, stopCh)
}

// Run starts an asynchronous loop that detects events of cluster clusters.
func (c *ClusterController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	klog.Infof("Starting cluster controller")
	klog.Infof("Waiting informer caches to synce")
	if ok := cache.WaitForCacheSync(stopCh, c.clusterSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers...")
	//perform runworker function until stopCh is closed
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	klog.Info("Started workers")
	<-stopCh
	klog.Infof("Shutting down cluster controller")
	return nil
}

func (c *ClusterController) runWorker() {
	klog.Info("Starting a worker")
	for c.processNextWorkItem() {
	}
}

func (c *ClusterController) processNextWorkItem() bool {
	workItem, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	klog.Infof("Process an item in work queue %v ", workItem)
	eventKey := workItem.(KeyWithEventType)
	key := eventKey.Key
	defer c.workqueue.Done(key)
	if err := c.syncHandler(eventKey); err != nil {
		c.workqueue.AddRateLimited(eventKey)
		utilruntime.HandleError(fmt.Errorf("Handle %v of key %v failed with %v", "serivce", key, err))
	}
	c.workqueue.Forget(key)
	klog.Infof("Successfully synced %s", key)
	return true
}

func (c *ClusterController) syncHandler(keyWithEventType KeyWithEventType) error {
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

	//This performs controller logic such as gRPC handling
	klog.Infof("gRPC processing - event: %v, cluster name: %v", keyWithEventType.EventType, clusterName)
	result, err := c.gRPCRequest(key, keyWithEventType.EventType, nameSpace, clusterName)
	if !result {
		klog.Errorf("Failed a cluster processing - event: %v, key: %v, error: %v", keyWithEventType, key, err)
		c.workqueue.AddRateLimited(keyWithEventType)
	} else {
		klog.Infof(" Processed a cluster: %v", key)
		c.workqueue.Forget(key)
	}
	klog.Infof("Cluster was handled by ClusterController - event: %v, cluster name: %v", keyWithEventType.EventType, clusterName)
	if keyWithEventType.EventType != EventType_Delete {
		cluster, err := c.clusterlister.Clusters(nameSpace).Get(clusterName)
		clusterCopy := cluster.DeepCopy()
		clusterCopy.Status = "HandledByClusterController"
		if err != nil || cluster == nil {
			klog.Errorf("Failed to retrieve cluster in local cache by cluster name: %s", clusterName)
			return err
		}
		//c.recorder.Event(clusterCopy, corev1.EventTypeNormal, SuccessSynched, MessageResourceSynched)
	}
	return nil
}

//This function determines if there is any actual change in cluster
//to improve performance by avoiding unnecessary update
func (c *ClusterController) determineEventType(cluster1, cluster2 *clusterv1.Cluster) (event int, err error) {
	clusterName1, clusterSpec1, clusterStatus1, err1 := c.getclusterInfo(cluster1)
	clusterName2, clusterSpec2, clusterStatus2, err2 := c.getclusterInfo(cluster2)
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
func (c *ClusterController) getclusterInfo(cluster *clusterv1.Cluster) (clusterName string, clusterSpec clusterv1.ClusterSpec, clusterStatus string, err error) {
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

//This is gRPC client, and performs controller logic including gRPC handling
func (c *ClusterController) gRPCRequest(key string, event EventType, clusterNameSpace string, clusterName string) (response bool, err error) {
	switch event {
	case EventType_Create:
		cluster, err := c.clusterlister.Clusters(clusterNameSpace).Get(clusterName)
		clusterCopy := cluster.DeepCopy()
		if err != nil || cluster == nil {
			klog.Errorf("Failed to retrieve cluster in local cache by cluster name: %s", clusterName)
			return false, err
		}
		if c.grpcHost != "" {
			klog.Infof("gRPC request - create a cluster, host: %v, cluster profile: %v", c.grpcHost, clusterCopy)
			clusterCopy.Status = ClusterStatusCreated
			response := grpc.GrpcSendClusterProfile(c.grpcHost, clusterCopy)
			klog.Infof("gRPC response - create a cluster, response: %v", response)
		}
		break
	case EventType_Update:
		cluster, err := c.clusterlister.Clusters(clusterNameSpace).Get(clusterName)
		clusterCopy := cluster.DeepCopy()
		if err != nil || cluster == nil {
			klog.Errorf("Failed to retrieve cluster in local cache by cluster name - %s", clusterName)
			return false, err
		}
		if c.grpcHost != "" {
			clusterCopy.Status = ClusterStatusUpdated
		}
	case EventType_Delete:
		//When deleting a cluster, API Server deletes the cluster before cluster controller watches the event.
		//so, ClusterController cannot get the deleted cluster's info from etcd.
		//To solve this issue, ClusterController stores/retrieve deleted cluster objects in a map named deletedClusters
		cluster := c.deletedClusters[key]
		clusterCopy := cluster.DeepCopy()
		if cluster == nil {
			klog.Errorf("Failed to retrieve cluster in map by cluster name - %s", clusterName)
			return false, err
		}
		if c.grpcHost != "" {
			klog.Infof("gRPC request - delete a cluster, host: %v, cluster profile: %v ", c.grpcHost, clusterCopy)
			clusterCopy.Status = ClusterStatusDeleted
			response := grpc.GrpcSendClusterProfile(c.grpcHost, clusterCopy)
			delete(c.deletedClusters, key)
			klog.Infof("gRPC response - delete a cluster, response: %v", response)
		}
		break
	default:
		klog.Infof("cluster event %v is not correct", event)
		err = fmt.Errorf("cluster event %v is not correct", event)
		return false, err
	}
	return true, nil
}

func (c *ClusterController) verifyClusterInfo(cluster *clusterv1.Cluster) (verified bool) {
	verified = false
	ipAddress := cluster.Spec.IpAddress
	region := cluster.Spec.Region.Region
	az := cluster.Spec.Region.AvailabilityZone
	siteID := fmt.Sprintf("%s|%s", region, az)
	clusterName := cluster.Name
	if ipAddress == nil || region == nil || az == nil || clusterName == nil  {
		err = fmt.Errorf("required cluster spec data is null")
		return verified
	}
	verified = (clusterName == siteID)
	return verified
}

//create dummy cluster for grpc delete request
/*func (c *ClusterController) newCluster(namespace string, name string) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{Kind: clusterv1.Kind, APIVersion: clusterv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: clusterv1.ClusterSpec{
			IpAddress: "0.0.0.0",
		},
	}
}*/
