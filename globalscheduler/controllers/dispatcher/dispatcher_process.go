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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/cmd/conf"
	"k8s.io/kubernetes/globalscheduler/controllers/util"
	"k8s.io/kubernetes/globalscheduler/controllers/util/openstack"
	allocclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/allocation/client/clientset/versioned"
	allocv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/allocation/v1"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	dispatcherclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned"
	dispatcherv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/v1"
	"os"
	"reflect"
	"sync"
	"syscall"
	"time"
)

const dispatcherName = "dispatcher"

type Process struct {
	namespace                string
	name                     string
	allocClientset           *allocclientset.Clientset
	dispatcherClientset      *dispatcherclientset.Clientset
	clusterClientset         *clusterclientset.Clientset
	clientset                *kubernetes.Clientset
	podQueue                 chan *v1.Pod
	clusterIpMap             map[string]string
	tokenMap                 map[string]string
	clusterRange             dispatcherv1.DispatcherRange
	totalCreateLatency       int64
	totalDeleteLatency       int64
	totalPodCreateNum        int
	totalPodDeleteNum        int
	totalAllocationCreateNum int
	totalAllocationDeleteNum int
}

func NewProcess(config *rest.Config, namespace string, name string, quit chan struct{}) Process {
	podQueue := make(chan *v1.Pod, 300)

	dispatcherClientset, err := dispatcherclientset.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	dispatcher, err := dispatcherClientset.GlobalschedulerV1().Dispatchers(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		klog.Fatal(err)
	}

	clusterClientset, err := clusterclientset.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	conf.AddQPSFlags(config, conf.GetInstance().Dispatcher.Allocation)
	allocationClientset, err := allocclientset.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	conf.AddQPSFlags(config, conf.GetInstance().Dispatcher.Pod)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}
	return Process{
		namespace:                namespace,
		name:                     name,
		clientset:                clientset,
		allocClientset:           allocationClientset,
		clusterClientset:         clusterClientset,
		dispatcherClientset:      dispatcherClientset,
		podQueue:                 podQueue,
		clusterIpMap:             make(map[string]string),
		tokenMap:                 make(map[string]string),
		clusterRange:             dispatcher.Spec.ClusterRange,
		totalCreateLatency:       0,
		totalDeleteLatency:       0,
		totalPodCreateNum:        0,
		totalPodDeleteNum:        0,
		totalAllocationCreateNum: 0,
		totalAllocationDeleteNum: 0,
	}
}

func (p *Process) Run(quit chan struct{}) {
	p.init()

	dispatcherSelector := fields.ParseSelectorOrDie("metadata.name=" + p.name)
	dispatcherLW := cache.NewListWatchFromClient(p.dispatcherClientset.GlobalschedulerV1(), "dispatchers", p.namespace, dispatcherSelector)
	dispatcherInformer := cache.NewSharedIndexInformer(dispatcherLW, &dispatcherv1.Dispatcher{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	dispatcherInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			klog.V(3).Infof("The dispatcher %s process is going to be killed...", p.name)
			os.Exit(0)
		},
		UpdateFunc: func(old, new interface{}) {
			newDispatcher, ok := new.(*dispatcherv1.Dispatcher)
			if !ok {
				klog.Warningf("Failed to convert a new object  %+v to a dispatcher", new)
				return
			}
			if !reflect.DeepEqual(p.clusterRange, newDispatcher.Spec.ClusterRange) {
				p.clusterRange = newDispatcher.Spec.ClusterRange
				if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
					klog.Fatal(err)
				}
			}
		},
	})

	go dispatcherInformer.Run(quit)

	boundAllocationInformer := p.initAllocationInformer(allocv1.AllocationBound, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			alloc, ok := obj.(*allocv1.Allocation)
			if !ok {
				klog.Warningf("Failed to convert  object  %+v to an allocation", obj)
				return
			}
			klog.V(4).Infof("An allocation %s has been added", alloc.Name)

			util.CheckTime(alloc.GetName(), "dispatcher", "CreateAllocation-Start", 1)
			go func() {
				p.SendAllocationToCluster(alloc)
			}()
		},
	})
	go boundAllocationInformer.Run(quit)

	scheduledAllocationInformer := p.initAllocationInformer(allocv1.AllocationScheduled, cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			alloc, ok := obj.(*allocv1.Allocation)
			if !ok {
				klog.Warningf("Failed to convert an deleted object  %+v to an allocation", obj)
				return
			}
			klog.V(4).Infof("An allocation %s has been deleted", alloc.Name)
			if alloc.ObjectMeta.DeletionTimestamp == nil {
				alloc.ObjectMeta.SetDeletionTimestamp(&metav1.Time{})
			}
			util.CheckTime(alloc.Name, "dispatcher", "DeleteAllocation-Start", 1)
			go func() {
				p.SendAllocationToCluster(alloc)
			}()
		},
	})
	go scheduledAllocationInformer.Run(quit)

	boundPodInformer := p.initPodInformer(v1.PodBound, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert an added object  %+v to a pod", obj)
				return
			}
			klog.V(4).Infof("Pod %s with cluster %s has been added", pod.Name, pod.Spec.ClusterName)
			util.CheckTime(pod.Name, "dispatcher", "CreatePod-Start", 1)
			go func() {
				p.SendPodToCluster(pod)
			}()
		},
	})
	go boundPodInformer.Run(quit)
	scheduledPodInformer := p.initPodInformer(v1.ClusterScheduled, cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert an deleted object  %+v to a pod", obj)
				return
			}
			klog.V(4).Infof("Pod %s with cluster %s has been deleted", pod.Name, pod.Spec.ClusterName)
			util.CheckTime(pod.Name, "dispatcher", "DeletePod-Start", 1)
			go func() {
				p.SendPodToCluster(pod)
			}()
		},
	})
	go scheduledPodInformer.Run(quit)

	go p.refreshToken()
	<-quit
}

func (p *Process) initAllocationInformer(phase allocv1.AllocationPhase, funcs cache.ResourceEventHandlerFuncs) cache.SharedIndexInformer {
	allocSelector := fields.ParseSelectorOrDie(fmt.Sprintf("status.phase=%s,status.cluster_name=gte:%s,status.cluster_name=lte:%s", string(phase),
		p.clusterRange.Start, p.clusterRange.End))
	allocLW := cache.NewListWatchFromClient(p.allocClientset.GlobalschedulerV1(), "allocations", p.namespace, allocSelector)
	allocInformer := cache.NewSharedIndexInformer(allocLW, &allocv1.Allocation{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	allocInformer.AddEventHandler(funcs)
	return allocInformer
}

func (p *Process) initPodInformer(phase v1.PodPhase, funcs cache.ResourceEventHandlerFuncs) cache.SharedIndexInformer {
	podSelector := fields.ParseSelectorOrDie(fmt.Sprintf("status.phase=%s,spec.clusterName=gte:%s,spec.clusterName=lte:%s", string(phase),
		p.clusterRange.Start, p.clusterRange.End))
	lw := cache.NewListWatchFromClient(p.clientset.CoreV1(), string(v1.ResourcePods), metav1.NamespaceAll, podSelector)
	podInformer := cache.NewSharedIndexInformer(lw, &v1.Pod{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podInformer.AddEventHandler(funcs)
	return podInformer
}

func (p *Process) SendPodToCluster(pod *v1.Pod) {
	if pod != nil {
		klog.V(3).Infof("Processing the item %v", pod)
		host, err := p.getHostIP(pod.Spec.ClusterName)
		if err != nil {
			klog.Warningf("Failed to get host from the cluster %v", pod.Spec.ClusterName)
			return
		}
		token, err := p.getToken(host)
		if err != nil {
			klog.Warningf("Failed to get token from host %v", host)
			return
		}
		pod.Status.DispatcherName = p.name
		if pod.ObjectMeta.DeletionTimestamp != nil {
			p.totalPodDeleteNum += 1
			// Calculate delete latency
			podDeleteTime := pod.DeletionTimestamp
			currentTime := time.Now().UTC()
			duration := (currentTime.UnixNano() - podDeleteTime.UnixNano()) / 1000000
			p.totalDeleteLatency += duration
			deleteLatency := int(duration)
			klog.V(2).Infof("************************************ Pod Name: %s, Delete Latency: %d Millisecond ************************************", pod.Name, deleteLatency)

			// Calculate average delete latency
			averageDeleteLatency := int(p.totalDeleteLatency) / p.totalPodDeleteNum
			klog.V(2).Infof("%%%%%%%%%%%%%%%%%%%%%%%%%% Total Number of Pods Deleted: %d, Average Delete Latency: %d Millisecond %%%%%%%%%%%%%%%%%%%%%%%%%%", p.totalPodDeleteNum, averageDeleteLatency)
			go func() {
				err = openstack.DeleteInstance(host, token, pod.Status.ClusterInstanceId)
				if err == nil {
					klog.V(3).Infof("The openstack vm for the pod %v has been deleted at the host %v", pod.ObjectMeta.Name, host)
				} else {
					klog.Warningf("The openstack vm for the pod %v failed to delete with the error %v", pod.ObjectMeta.Name, err)
				}
				util.CheckTime(pod.Name, "dispatcher", "DeletePod-End", 2)
			}()
		} else {
			util.CheckTime(pod.Name, "dispatcher", "CreatePod-End", 2)
			p.totalPodCreateNum += 1
			// Calculate create latency
			podCreateTime := pod.CreationTimestamp
			currentTime := time.Now().UTC()
			duration := (currentTime.UnixNano() - podCreateTime.UnixNano()) / 1000000
			p.totalCreateLatency += duration
			createLatency := int(duration)
			klog.V(2).Infof("************************************ Pod Name: %s, Create Latency: %d Millisecond ************************************", pod.Name, createLatency)

			// Calculate average create latency
			averageCreateLatency := int(p.totalCreateLatency) / p.totalPodCreateNum
			klog.V(2).Infof("%%%%%%%%%%%%%%%%%%%%%%%%%% Total Number of Pods Created: %d, Average Create Latency: %d Millisecond %%%%%%%%%%%%%%%%%%%%%%%%%%", p.totalPodCreateNum, averageCreateLatency)
			go func() {
				instanceId, err := openstack.ServerCreate(host, token, &pod.Spec)
				if err == nil {
					klog.V(3).Infof("The openstack vm for the pod %v has been created at the host %v", pod.ObjectMeta.Name, host)
					pod.Status.ClusterInstanceId = instanceId
					pod.Status.Phase = v1.ClusterScheduled
					updatedPod, err := p.clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).UpdateStatus(pod)
					if err == nil {
						klog.V(3).Infof("The pod %v has been updated its apiserver database status to scheduled successfully with the instance id %v", updatedPod, instanceId)

					} else {
						klog.Warningf("The pod %v failed to update its apiserver database status to scheduled with the error %v", pod.ObjectMeta.Name, err)
					}
				} else {
					klog.Warningf("The openstack vm for the  pod %v failed to create with the error %v", pod.ObjectMeta.Name, err)
					pod.Status.Phase = v1.PodFailed
					if _, err := p.clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).UpdateStatus(pod); err != nil {
						klog.Warningf("The pod %v failed to update its apiserver dtatbase status to failed with the error %v", pod.ObjectMeta.Name, err)
					}
				}
			}()
		}
	}
}

func (p *Process) getToken(ip string) (string, error) {
	if token, ok := p.tokenMap[ip]; ok {
		return token, nil
	}
	token, err := openstack.RequestToken(ip)
	if err != nil {
		return "", err
	}
	p.tokenMap[ip] = token
	return token, nil
}

func (p *Process) getHostIP(clusterName string) (string, error) {
	if ipAddress, ok := p.clusterIpMap[clusterName]; ok {
		return ipAddress, nil
	}
	cluster, err := p.clusterClientset.GlobalschedulerV1().Clusters(metav1.NamespaceDefault).Get(clusterName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	p.clusterIpMap[clusterName] = cluster.Spec.IpAddress
	return p.clusterIpMap[clusterName], nil
}

func (p *Process) refreshToken() {
	for range time.Tick(time.Hour * 2) {
		for ip, token := range p.tokenMap {
			if openstack.TokenExpired(ip, token) {
				newToken, err := openstack.RequestToken(ip)
				if err == nil {
					p.tokenMap[ip] = newToken
				}
			}
		}
	}
}

func (p *Process) init() {
	clusters, err := p.clusterClientset.GlobalschedulerV1().Clusters(metav1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {
		klog.Warningf("Failed to get clusters with error %v", err)
	}
	for _, cluster := range clusters.Items {
		p.clusterIpMap[cluster.Name] = cluster.Spec.IpAddress
		if _, err = p.getToken(cluster.Spec.IpAddress); err != nil {
			klog.Warningf("Failed to get token of cluster %s with error %v", cluster.Name, err)
		}
	}
}

func (p *Process) SendAllocationToCluster(alloc *allocv1.Allocation) {
	if alloc != nil {
		klog.V(2).Infof("Processing the item %v", alloc)

		alloc.Status.DispatcherName = p.name
		if alloc.ObjectMeta.DeletionTimestamp != nil {
			util.CheckTime(alloc.Name, "dispatcher", "DeleteAllocation-End", 2)
			p.totalAllocationDeleteNum += 1
			p.deleteAllocationResources(alloc)
			// Calculate delete latency
			allocDeleteTime := alloc.DeletionTimestamp
			currentTime := time.Now().UTC()
			duration := (currentTime.UnixNano() - allocDeleteTime.UnixNano()) / 1000000
			p.totalDeleteLatency += duration
			deleteLatency := int(duration)
			klog.V(2).Infof("************************************ Allocation Name: %s, Delete Latency: %d Millisecond ************************************", alloc.Name, deleteLatency)

			// Calculate average delete latency
			averageDeleteLatency := int(p.totalDeleteLatency) / p.totalPodDeleteNum
			klog.V(2).Infof("%%%%%%%%%%%%%%%%%%%%%%%%%% Total Number of Allocation Deleted: %d, Average Delete Latency: %d Millisecond %%%%%%%%%%%%%%%%%%%%%%%%%%", p.totalAllocationDeleteNum, averageDeleteLatency)
			p.totalAllocationCreateNum += 1
		} else {
			p.totalAllocationCreateNum += 1
			// Calculate create latency
			allocCreateTime := alloc.CreationTimestamp
			currentTime := time.Now().UTC()
			duration := (currentTime.UnixNano() - allocCreateTime.UnixNano()) / 1000000
			p.totalCreateLatency += duration
			createLatency := int(duration)
			klog.V(2).Infof("************************************ Allocation Name: %s, Create Latency: %d Millisecond ************************************", alloc.Name, createLatency)

			// Calculate average create latency
			averageCreateLatency := int(p.totalCreateLatency) / p.totalAllocationCreateNum
			klog.V(2).Infof("%%%%%%%%%%%%%%%%%%%%%%%%%% Total Number of Allocation Created: %d, Average Create Latency: %d Millisecond %%%%%%%%%%%%%%%%%%%%%%%%%%", p.totalAllocationCreateNum, averageCreateLatency)

			if isAllResourcesCreated := p.isAllAllocationResourcesCreated(alloc); isAllResourcesCreated {
				alloc.Status.Phase = allocv1.AllocationScheduled
			} else {
				alloc.Status.Phase = allocv1.AllocationFailed
				p.deleteAllocationResources(alloc)
			}
			if _, err := p.allocClientset.GlobalschedulerV1().AllocationsWithMultiTenancy(alloc.Namespace, alloc.ObjectMeta.Tenant).Update(alloc); err != nil {
				klog.Warningf("The allocation %v failed to update its status phase to %v with the error %v", alloc.Name, alloc.Status.Phase, err)
			} else {
				klog.V(3).Infof("The allocation %s updated its  status phase to %v successfully", alloc.Name, alloc.Status.Phase)
			}
			util.CheckTime(alloc.Name, "dispatcher", "CreateAllocation-End", 2)
		}

	}
}
func (p *Process) isAllAllocationResourcesCreated(alloc *allocv1.Allocation) bool {
	for idx, resource := range alloc.Spec.ResourceGroup.Resources {
		for ciIdx, clusterInstance := range resource.VirtualMachine.ClusterInstances {
			host, err := p.getHostIP(clusterInstance.ClusterName)
			if err != nil {
				klog.Warningf("Failed to get host from the cluster %v", clusterInstance.ClusterName)
				return false
			}
			token, err := p.getToken(host)
			if err != nil {
				klog.Warningf("Failed to get token from host %v", host)
				return false
			}
			instanceId, err := openstack.ServerCreateResources(host, token, clusterInstance.ClusterName, &resource)
			if err == nil {
				alloc.Spec.ResourceGroup.Resources[idx].VirtualMachine.ClusterInstances[ciIdx].InstanceId = instanceId
				klog.V(3).Infof("The openstack vm for the allocation %s resource %s has been created at the host %v as instance %s", alloc.Name, resource.Name, host, instanceId)
			} else {
				klog.Warningf("The openstack vm for the allocation %s resource %s failed to create with the error %v", alloc.Name, resource.Name, err)
				return false
			}
		}
	}
	return true
}

func (p *Process) deleteAllocationResources(alloc *allocv1.Allocation) {
	var wg sync.WaitGroup
	for _, resource := range alloc.Spec.ResourceGroup.Resources {
		for _, clusterInstance := range resource.VirtualMachine.ClusterInstances {
			instanceid := clusterInstance.InstanceId
			clustername := clusterInstance.ClusterName
			if instanceid != "" {
				wg.Add(1)
				host, err := p.getHostIP(clustername)
				if err != nil {
					klog.Warningf("Failed to get host from the cluster %v", clustername)
					return
				}
				token, err := p.getToken(host)
				if err != nil {
					klog.Warningf("Failed to get token from host %v", host)
					return
				}

				go func() {
					defer wg.Done()
					err = openstack.DeleteInstance(host, token, instanceid)

					if err == nil {
						klog.V(3).Infof("The openstack vm for the allocation %s resource %s has been deleted at the host %v", alloc.Name, resource.Name, host)
					} else {
						klog.Warningf("The openstack vm for the allocation %s  resource %s failed to delete with the error %v", alloc.Name, resource.Name, err)
					}
				}()
			}
		}
	}
	wg.Wait()
}
