/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

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
	"reflect"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	//"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
	"k8s.io/kubernetes/pkg/controller"
	statusutil "k8s.io/kubernetes/pkg/util/pod"
)

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

// AddAllEventHandlers is a helper function used in tests and in Scheduler
// to add event handlers for various informers.
func AddAllEventHandlers(sched *Scheduler) {
	sched.PodInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return assignedPod(t) && responsibleForPod(t, sched.SchedulerName)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return assignedPod(pod) && responsibleForPod(pod, sched.SchedulerName)
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, sched))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    sched.addPodToCache,
				UpdateFunc: sched.updatePodInCache,
				DeleteFunc: sched.deletePodFromCache,
			},
		})
	// unscheduled pod queue
	sched.PodInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return needToSchedule(t) && responsibleForPod(t, sched.SchedulerName)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return !assignedPod(pod) && responsibleForPod(pod, sched.SchedulerName)
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, sched))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    sched.addPodToSchedulingQueue,
				UpdateFunc: sched.updatePodInSchedulingQueue,
				DeleteFunc: sched.deletePodFromSchedulingQueue,
			},
		},
	)
	// failed pod queue
	sched.PodInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return failedToSchedule(t) && responsibleForPod(t, sched.SchedulerName)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return failedToSchedule(pod) && responsibleForPod(pod, sched.SchedulerName)
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, sched))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    sched.addPodWithdrawResource,
				UpdateFunc: sched.updatePodWithdrawResource,
				DeleteFunc: sched.deletePodWithdrawResource,
			},
		},
	)
	sched.ClusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sched.addCluster,
		UpdateFunc: sched.updateCluster,
		DeleteFunc: sched.deleteCluster,
	})
}

// needToSchedule selects pods that need to be scheduled
func needToSchedule(pod *v1.Pod) bool {
	return pod.Spec.VirtualMachine != nil && pod.Status.Phase == v1.PodAssigned
}

// assignedPod selects pods that are assigned (scheduled and running).
func assignedPod(pod *v1.Pod) bool {
	return pod.Spec.VirtualMachine != nil && pod.Status.Phase == v1.PodBound
}

// responsibleForPod returns true if the pod has asked to be scheduled by the given scheduler.
func responsibleForPod(pod *v1.Pod, schedulerName string) bool {
	return schedulerName == pod.Status.AssignedScheduler.Name
}

// failedToSchedule selects pods that scheduled but failed to create vm
func failedToSchedule(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodFailed
}

// addPodToCache add pod to the stack cache of the scheduler
func (sched *Scheduler) addPodToCache(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	klog.V(4).Infof("Add a pod: %v", pod.Name)
	if !ok {
		klog.Errorf("cannot convert to *v1.Pod: %v", obj)
		return
	}

	// add pod resource to a stack
	stack := getStackFromPod(pod)

	// add stack to cache
	if err := sched.SchedulerCache.AddStack(stack); err != nil {
		klog.Errorf("scheduler cache AddStack failed: %v", err)
	}
}

func (sched *Scheduler) updatePodInCache(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert oldObj to *v1.Pod: %v", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	klog.V(4).Infof("Update a pod: %v", newPod)
	if !ok {
		klog.Errorf("cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}

	oldStack := getStackFromPod(oldPod)
	newStack := getStackFromPod(newPod)
	if err := sched.SchedulerCache.UpdateStack(oldStack, newStack); err != nil {
		klog.Errorf("scheduler cache UpdatePod failed: %v", err)
	}
}

func (sched *Scheduler) deletePodFromCache(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
		klog.V(4).Infof("Delete a pod: %v", pod.Name)
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			klog.Errorf("cannot convert to *v1.Pod: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("cannot convert to *v1.Pod: %v", t)
		return
	}

	// get stack from pod
	stack := getStackFromPod(pod)

	// NOTE: Updates must be written to scheduler cache before invalidating
	// equivalence cache, because we could snapshot equivalence cache after the
	// invalidation and then snapshot the cache itself. If the cache is
	// snapshotted before updates are written, we would update equivalence
	// cache with stale information which is based on snapshot of old cache.
	if err := sched.SchedulerCache.RemoveStack(stack); err != nil {
		klog.Errorf("scheduler cache RemoveStack failed: %v", err)
	}
}

func getStackFromPod(pod *v1.Pod) *types.Stack {
	stack := &types.Stack{
		PodName:      pod.Name,
		Tenant:       pod.Tenant,
		PodNamespace: pod.Namespace,
		UID:          string(pod.UID),
		Selector:     getStackSelector(&pod.Spec.VirtualMachine.ResourceCommonInfo.Selector),
		Resources:    getStackResources(pod),
	}

	return stack
}

// getStackResources change pod resources to stack Resource
func getStackResources(pod *v1.Pod) []*types.Resource {
	vmSpec := pod.Spec.VirtualMachine
	flavors := make([]types.Flavor, 0)
	for _, value := range vmSpec.Flavors {
		flavors = append(flavors, types.Flavor{
			FlavorID: value.FlavorID,
			// TODO(nkwangjun): nee to add spot value
			Spot: nil,
		})
	}

	resource := &types.Resource{
		Name: pod.Name,
		// currently we just support resource type is vm
		ResourceType: "vm",
		// TODO(nkwangjun): need to add storage in vmSpec
		Storage: nil,
		Flavors: flavors,
		NeedEip: vmSpec.NeedEIP,
		Count:   1,
	}

	return []*types.Resource{resource}
}

// getStackSelector change vm selector to stack selector
func getStackSelector(selector *v1.ResourceSelector) types.Selector {
	// depress empty slice warning
	var siteID string
	newRegions := make([]types.CloudRegion, 0)
	for _, value := range selector.Regions {
		newRegions = append(newRegions, types.CloudRegion{
			Region:           value.Region,
			AvailabilityZone: value.AvailablityZone,
		})
		/// the following check is to avoid an out of index error when pod doesn't have az
		siteID = value.Region + constants.SiteDelimiter
		if len(value.AvailablityZone) > 0 {
			siteID = siteID + value.AvailablityZone[0]
		}
	}

	newSelector := types.Selector{
		GeoLocation: types.GeoLocation{
			Country:  selector.GeoLocation.Country,
			Area:     selector.GeoLocation.Area,
			Province: selector.GeoLocation.Province,
			City:     selector.GeoLocation.City,
		},
		Regions:  newRegions,
		Operator: selector.Operator,
		SiteID:   siteID,
		Strategy: types.Strategy{
			LocationStrategy: selector.Strategy.LocalStrategy,
		},
	}

	return newSelector
}

func (sched *Scheduler) addPodToSchedulingQueue(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert to *v1.Pod: %v", obj)
		return
	}

	// add pod resource to a stack
	stack := getStackFromPod(pod)
	stack.CreateTime = time.Now().UnixNano()

	if err := sched.StackQueue.Add(stack); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to queue %T: %v", obj, err))
	}
}

func (sched *Scheduler) updatePodInSchedulingQueue(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert oldObj to *v1.Pod: %v", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	klog.V(4).Infof("updatePodToSchedulingQueue : %v", newPod)
	if !ok {
		klog.Errorf("cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}
	oldStack := getStackFromPod(oldPod)
	newStack := getStackFromPod(newPod)
	if sched.skipStackUpdate(newStack) {
		return
	}
	if err := sched.StackQueue.Update(oldStack, newStack); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to update %T: %v", newObj, err))
	}
}

func (sched *Scheduler) deletePodFromSchedulingQueue(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = obj.(*v1.Pod)
		klog.V(4).Infof("deletePodToSchedulingQueue : %v", pod.Name)
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, sched))
			return
		}
	default:
		utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
		return
	}

	stack := getStackFromPod(pod)
	if err := sched.StackQueue.Delete(stack); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to dequeue %T: %v", obj, err))
	}
}

// skipStackUpdate checks whether the specified pod update should be ignored.
// This function will return true if
//   - The pod has already been assumed, AND
//   - The pod has only its ResourceVersion, Spec.NodeName and/or Annotations
//     updated.
func (sched *Scheduler) skipStackUpdate(stack *types.Stack) bool {
	// Non-assumed stacks should never be skipped.
	isAssumed, err := sched.SchedulerCache.IsAssumedStack(stack)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to check whether stack %s/%s/%s is assumed: %v", stack.Tenant, stack.PodNamespace, stack.PodName, err))
		return false
	}
	if !isAssumed {
		return false
	}

	// Gets the assumed stack from the cache.
	assumedStack, err := sched.SchedulerCache.GetStack(stack)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to get assumed stack %s/%s/%s from cache: %v", stack.Tenant, stack.PodNamespace, stack.PodName, err))
		return false
	}

	// TODO(wangjun): We should re-define stack as pods, and implement the DeepCopy function.
	//                Here if stack is changed, we just go the stack update process without compare
	//                site name or something else.
	assumedStackCopy, stackCopy := assumedStack.DeepCopy(), stack.DeepCopy()
	if !reflect.DeepEqual(assumedStackCopy, stackCopy) {
		return false
	}
	klog.V(4).Infof("Skipping stack %s/%s/%s update", stack.Tenant, stack.PodNamespace, stack.PodName)
	return true
}

func (sched *Scheduler) bindStacks(assumedStacks []types.Stack) {
	klog.V(4).Infof("assumedStacks: %v", assumedStacks)
	for _, newStack := range assumedStacks {
		klog.V(4).Infof("newStack: %v", newStack)
		clusterName := newStack.Selected.ClusterName
		sched.bindToSite(clusterName, &newStack)
	}
}

func (sched *Scheduler) setPodScheduleErr(reqStack *types.Stack) error {
	// get pod first
	pod, err := sched.Client.CoreV1().PodsWithMultiTenancy(reqStack.PodNamespace, reqStack.Tenant).Get(reqStack.PodName, metav1.GetOptions{})
	if err != nil {
		klog.Warningf("Failed to get status for pod %q: %v", reqStack.PodName+"/"+reqStack.PodNamespace+"/"+
			reqStack.Tenant+"/"+reqStack.UID, err)
		return err
	}

	// update pod status to NoSchedule
	newStatus := v1.PodStatus{
		Phase: v1.PodNoSchedule,
	}
	klog.V(4).Infof("Attempting to update pod status from %v to %v", pod.Status, newStatus)
	_, _, err = statusutil.PatchPodStatus(sched.Client, reqStack.Tenant, reqStack.PodNamespace, reqStack.PodName, pod.Status, newStatus)
	if err != nil {
		klog.Warningf("PatchPodStatus for pod %q: %v", reqStack.PodName+"/"+reqStack.PodNamespace+"/"+
			reqStack.Tenant+"/"+reqStack.UID, err)
		return err
	}

	klog.V(4).Infof("Update pod status from %v to %v success", pod.Status, newStatus)
	return nil
}

func (sched *Scheduler) bindToSite(clusterName string, assumedStack *types.Stack) error {
	binding := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Tenant:    assumedStack.Tenant,
			Namespace: assumedStack.PodNamespace,
			Name:      assumedStack.PodName,
			UID:       apitypes.UID(assumedStack.UID),
		},
		Target: v1.ObjectReference{
			Kind: "Cluster",
			Name: clusterName,
		},
	}
	klog.V(4).Infof("binding: %v", binding)
	// do api server update here
	klog.V(4).Infof("Attempting to bind %v to %v", binding.Name, binding.Target.Name)
	err := sched.Client.CoreV1().PodsWithMultiTenancy(binding.Namespace, binding.Tenant).Bind(binding)
	if err != nil {
		klog.Errorf("Failed to bind stack: %v/%v/%v", assumedStack.Tenant, assumedStack.PodNamespace,
			assumedStack.PodName)
		if err := sched.SchedulerCache.ForgetStack(assumedStack); err != nil {
			klog.Errorf("scheduler cache ForgetStack failed: %v", err)
		}
		return err
	}
	//
	return nil
}

func (sched *Scheduler) addCluster(object interface{}) {
	resource := object.(*clusterv1.Cluster)
	clusterCopy := resource.DeepCopy()
	if sched.verifyClusterInfo(clusterCopy) == false {
		klog.V(4).Infof(" Cluster data is not correct: %v", clusterCopy)
	}
	key, err := controller.KeyFunc(object)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object: %v, error: %v", object, err))
		return
	}
	sched.Enqueue(key, EventType_Create)
	klog.V(4).Infof("Enqueue Create cluster: %v", key)
}

func (sched *Scheduler) updateCluster(oldObject, newObject interface{}) {
	oldResource := oldObject.(*clusterv1.Cluster)
	newResource := newObject.(*clusterv1.Cluster)
	oldClusterCopy := oldResource.DeepCopy()
	newClusterCopy := newResource.DeepCopy()
	if sched.verifyClusterInfo(newClusterCopy) {
		klog.V(4).Infof(" Cluster data is not correct: %v", newResource)
	}
	key1, err1 := controller.KeyFunc(oldObject)
	key2, err2 := controller.KeyFunc(newObject)
	if key1 == "" || key2 == "" || err1 != nil || err2 != nil {
		klog.Errorf("Unexpected string in queue, discarding: %v", key2)
		return
	}

	eventType, err := sched.determineEventType(oldClusterCopy, newClusterCopy)
	if err != nil {
		klog.Errorf("Unexpected string in queue, discarding: %v", key2)
		return
	}
	switch eventType {
	case ClusterUpdateNo:
		{
			klog.V(4).Infof("No actual change in clusters, discarding: %v", newClusterCopy.Name)
			break
		}
	case ClusterUpdateYes:
		{
			sched.Enqueue(key2, EventType_Update)
			klog.V(4).Infof("Enqueue Update Cluster: %v", key2)
			break
		}
	default:
		{
			klog.Errorf("Unexpected cluster update event, discarding: %v", key2)
			return
		}
	}
}

func (sched *Scheduler) deleteCluster(object interface{}) {
	resource := object.(*clusterv1.Cluster)
	clusterCopy := resource.DeepCopy()
	if sched.verifyClusterInfo(clusterCopy) == false {
		klog.V(4).Infof(" Cluster data is not correct: %v", clusterCopy)
		return
	}
	key, err := controller.KeyFunc(object)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object: %v, error: %v", object, err))
		return
	}
	sched.Enqueue(key, EventType_Delete)
	siteID := clusterCopy.Spec.Region.Region + constants.SiteDelimiter + clusterCopy.Spec.Region.AvailabilityZone
	sched.deletedClusters[key] = siteID
	klog.V(4).Infof("Enqueue Delete Cluster: %v", key)
}

// Enqueue puts key of the cluster object in the work queue
// EventType: Create=0, Update=1, Delete=2
func (sched *Scheduler) Enqueue(key string, eventType EventType) {
	sched.ClusterQueue.Add(KeyWithEventType{Key: key, EventType: eventType})
}

func (sched *Scheduler) verifyClusterInfo(cluster *clusterv1.Cluster) (verified bool) {
	verified = false
	ipAddress := cluster.Spec.IpAddress
	region := cluster.Spec.Region.Region
	az := cluster.Spec.Region.AvailabilityZone
	clusterName := cluster.Name
	if ipAddress == "" || region == "" || az == "" || clusterName == "" {
		klog.Errorf("cluster ipAddress:%s, region:%s, az:%s, or custer name:%s is null", ipAddress, region, az, clusterName)
		return verified
	}
	verified = true
	return verified
}

func (sched *Scheduler) verifyPodInfo(pod *v1.Pod) (verified bool) {
	verified = false
	name := pod.Name
	if pod.Name == "" {
		klog.Errorf("pod name:%s is null", name)
		return verified
	}
	verified = true
	return verified
}

func (sched *Scheduler) addPodWithdrawResource(object interface{}) {
	pod, ok := object.(*v1.Pod)
	klog.V(4).Infof("Add a pod to withdraw resource: %v", pod.Name)
	if !ok {
		klog.Errorf("cannot convert to *v1.Pod: %v", object)
		return
	}
	podCopy := pod.DeepCopy()
	if sched.verifyPodInfo(podCopy) == false {
		klog.V(4).Infof(" Pod data is not correct: %v", podCopy)
	}
	err := sched.withdrawResource(pod.Name)
	if err != nil {
		klog.Errorf("withdraw resource of pod %s failed", pod.Name)
	}
}

func (sched *Scheduler) updatePodWithdrawResource(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert oldObj to *v1.Pod: %v", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	klog.V(4).Infof("Update a pod: %v", newPod)
	if !ok {
		klog.Errorf("cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}
	if oldPod.Name != newPod.Name {
		klog.Errorf("old pod name and new pod name should be equal: %s, %s", oldPod.Name, newPod.Name)
		return
	}
	err := sched.withdrawResource(newPod.Name)
	if err != nil {
		klog.Errorf("withdraw resource of pod %s failed", oldPod.Name)
	}
}

func (sched *Scheduler) deletePodWithdrawResource(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
		klog.V(4).Infof("Delete a pod: %v", pod.Name)
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			klog.Errorf("cannot convert to *v1.Pod: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("cannot convert to *v1.Pod: %v", t)
		return
	}
	err := sched.withdrawResource(pod.Name)
	if err != nil {
		klog.Errorf("withdraw resource of pod %s failed", pod.Name)
	}
}

//withdraw reserved resources to a pod & add it to cash to other pods
func (sched *Scheduler) withdrawResource(podName string) error {
	resource := sched.PodSiteResourceMap[podName]
	if resource == nil {
		klog.V(4).Infof("there is no preserved resource for pod: %s", podName)
		return nil
	}
	allResInfo := resource.Resource
	regionName := utils.GetRegionName(resource.SiteID)
	regionFlavor, err := sched.siteCacheInfoSnapshot.GetRegionFlavors(regionName)
	if err != nil {
		klog.Errorf("there is no valid flavor for region: %s", regionName)
		return err
	}
	siteCacheInfo := sched.siteCacheInfoSnapshot.SiteCacheInfoMap[resource.SiteID]
	siteCacheInfo.UpdateSiteResInfo(allResInfo, regionFlavor, false)
	delete(sched.PodSiteResourceMap, podName)
	return nil
}
