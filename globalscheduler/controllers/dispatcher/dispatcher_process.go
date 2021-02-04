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
	"k8s.io/kubernetes/globalscheduler/controllers/util"
	"k8s.io/kubernetes/globalscheduler/controllers/util/openstack"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	dispatcherclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned"
	dispatcherv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/v1"
	"os"
	"reflect"
	"syscall"
	"time"
)

const dispatcherName = "dispatcher"

type Process struct {
	namespace           string
	name                string
	dispatcherClientset *dispatcherclientset.Clientset
	clusterClientset    *clusterclientset.Clientset
	clientset           *kubernetes.Clientset
	podQueue            chan *v1.Pod
	clusterIpMap        map[string]string
	tokenMap            map[string]string
	clusterRange        dispatcherv1.DispatcherRange
	pid                 int
	totalCreateLatency  int64
	totalDeleteLatency  int64
	totalPodCreateNum   int
	totalPodDeleteNum   int
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

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	return Process{
		namespace:           namespace,
		name:                name,
		clientset:           clientset,
		clusterClientset:    clusterClientset,
		dispatcherClientset: dispatcherClientset,
		podQueue:            podQueue,
		clusterIpMap:        make(map[string]string),
		tokenMap:            make(map[string]string),
		pid:                 os.Getgid(),
		clusterRange:        dispatcher.Spec.ClusterRange,
		totalCreateLatency:  0,
		totalDeleteLatency:  0,
		totalPodCreateNum:   0,
		totalPodDeleteNum:   0,
	}
}

func (p *Process) Run(quit chan struct{}) {
	p.init()

	dispatcherSelector := fields.ParseSelectorOrDie("metadata.name=" + p.name)
	dispatcherLW := cache.NewListWatchFromClient(p.dispatcherClientset.GlobalschedulerV1(), "dispatchers", p.namespace, dispatcherSelector)

	dispatcherInformer := cache.NewSharedIndexInformer(dispatcherLW, &dispatcherv1.Dispatcher{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	dispatcherInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			if err := syscall.Kill(-p.pid, 15); err != nil {
				klog.Fatalf("Fail to exit the current process %v\n", err)
			}
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
	boundPodnformer := p.initPodInformer(v1.PodBound, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert an added object  %+v to a pod", obj)
				return
			}
			klog.V(4).Infof("Pod %s with cluster %s has been added", pod.Name, pod.Spec.ClusterName)
			util.CheckTime(pod.Name, "dispatcher", "Run-AddFunc-gofunc", 1)
			go func() {
				util.CheckTime(pod.Name, "dispatcher", "Run-AddFunc-gofunc-SendPodToCluster", 1)
				p.SendPodToCluster(pod)
				util.CheckTime(pod.Name, "dispatcher", "Run-AddFunc-gofunc-SendPodToCluster", 2)
			}()
			util.CheckTime(pod.Name, "dispatcher", "Run-AddFunc-gofunc", 2)
		},
	})
	go boundPodnformer.Run(quit)
	scheduledPodnformer := p.initPodInformer(v1.ClusterScheduled, cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert an deleted object  %+v to a pod", obj)
				return
			}
			klog.V(4).Infof("Pod %s with cluster %s has been deleted", pod.Name, pod.Spec.ClusterName)
			util.CheckTime(pod.Name, "dispatcher", "Run-DeleteFunc-gofunc", 1)
			go func() {
				util.CheckTime(pod.Name, "dispatcher", "Run-DeleteFunc-gofunc-SendPodToCluster", 1)
				p.SendPodToCluster(pod)
				util.CheckTime(pod.Name, "dispatcher", "Run-DeleteFunc-gofunc-SendPodToCluster", 2)
			}()
			util.CheckTime(pod.Name, "dispatcher", "Run-DeleteFunc-gofunc", 2)
		},
	})
	go scheduledPodnformer.Run(quit)
	util.CheckTime("local", "dispatcher", "Run-refreshToken", 1)
	go p.refreshToken()
	util.CheckTime("local", "dispatcher", "Run-refreshToken", 2)
	<-quit
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
		util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster", 1)
		klog.V(3).Infof("Processing the item %v", pod)
		util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-getHostIP", 1)
		host, err := p.getHostIP(pod.Spec.ClusterName)
		util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-getHostIP", 2)
		if err != nil {
			klog.Warningf("Failed to get host from the cluster %v", pod.Spec.ClusterName)
			return
		}
		util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-getToken", 1)
		token, err := p.getToken(host)
		util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-getToken", 2)
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
			util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-deletePod", 1)
			go func() {
				err = openstack.DeleteInstance(host, token, pod.Status.ClusterInstanceId)
				if err == nil {
					klog.V(3).Infof("The openstack vm for the pod %v has been deleted at the host %v", pod.ObjectMeta.Name, host)
				} else {
					klog.Warningf("The openstack vm for the pod %v failed to delete with the error %v", pod.ObjectMeta.Name, err)
				}
			}()
			util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-deletePod", 2)
		} else {
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
			util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-createPod", 1)
			go func() {
				instanceId, err := openstack.ServerCreate(host, token, &pod.Spec)
				if err == nil {
					klog.V(3).Infof("The openstack vm for the pod %v has been created at the host %v", pod.ObjectMeta.Name, host)
					pod.Status.ClusterInstanceId = instanceId
					pod.Status.Phase = v1.ClusterScheduled
					util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-createPod-UpdateStatus", 1)
					updatedPod, err := p.clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).UpdateStatus(pod)
					util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-createPod-UpdateStatus", 2)
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
			util.CheckTime(pod.Name, "dispatcher", "SendPodToCluster-createPod", 2)
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
	util.CheckTime("local", "dispatcher", "refreshToken", 1)
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
	util.CheckTime("local", "dispatcher", "refreshToken", 2)
}

func (p *Process) init() {
	fieldSelector := fmt.Sprintf("metadata.name=gte:%s,metadata.name=lte:%s", p.clusterRange.Start, p.clusterRange.End)
	clusters, err := p.clusterClientset.GlobalschedulerV1().Clusters(metav1.NamespaceDefault).List(metav1.ListOptions{FieldSelector: fieldSelector})
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
