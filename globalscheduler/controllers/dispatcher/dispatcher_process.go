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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/util/openstack"
	dispatcherclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned"
	dispatcherv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/v1"
	"os"
	"reflect"
	"strings"
	"syscall"
)

const dispatcherName = "dispatcher"

type Process struct {
	namespace           string
	name                string
	dispatcherClientset *dispatcherclientset.Clientset
	clientset           *kubernetes.Clientset
	podQueue            chan *v1.Pod
	resetCh             chan struct{}
	clusterIdList       []string
	clusterIpMap        map[string]string
	tokenMap            map[string]string
	pid                 int
}

func NewProcess(config *rest.Config, namespace string, name string, quit chan struct{}) Process {
	podQueue := make(chan *v1.Pod, 300)
	defer close(podQueue)

	dispatcherClientset, err := dispatcherclientset.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	dispatcher, err := dispatcherClientset.GlobalschedulerV1().Dispatchers(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		klog.Fatal(err)
	}

	clusterIdList, clusterIpMap := convertClustersToMap(dispatcher.Spec.Cluster)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	resetCh := make(chan struct{})

	return Process{
		namespace:           namespace,
		name:                name,
		clientset:           clientset,
		dispatcherClientset: dispatcherClientset,
		podQueue:            podQueue,
		resetCh:             resetCh,
		clusterIdList:       clusterIdList,
		clusterIpMap:        clusterIpMap,
		pid:                 os.Getgid(),
	}
}

func (p *Process) Run(quit chan struct{}) {
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
			oldDispatcher, ok := old.(*dispatcherv1.Dispatcher)
			if !ok {
				klog.Warningf("Failed to convert a old object  %+v to a dispatcher", old)
				return
			}
			newDispatcher, ok := new.(*dispatcherv1.Dispatcher)
			if !ok {
				klog.Warningf("Failed to convert a new object  %+v to a dispatcher", new)
				return
			}
			if !reflect.DeepEqual(oldDispatcher.Spec.Cluster, newDispatcher.Spec.Cluster) {
				clusterIdList, clusterIpMap := convertClustersToMap(newDispatcher.Spec.Cluster)
				p.clusterIdList = clusterIdList
				p.clusterIpMap = clusterIpMap
				p.addBoundedPodsToQueue(p.resetCh, p.clusterIdList)
				p.addDeletedPodsToQueue(p.resetCh, p.clusterIdList)
			}
		},
	})

	p.addBoundedPodsToQueue(p.resetCh, p.clusterIdList)
	p.addDeletedPodsToQueue(p.resetCh, p.clusterIdList)
	go dispatcherInformer.Run(quit)
	wait.Until(p.SendPodToCluster, 0, quit)
}

func (p *Process) initPodInformer(clusterIds []string, statusPhase v1.PodPhase) cache.SharedIndexInformer {
	close(p.resetCh)
	p.resetCh = make(chan struct{})
	conditions := "status.phase=" + string(statusPhase) + ","

	for _, clusterId := range clusterIds {
		conditions = conditions + "spec.clusterName=" + clusterId + ";"
	}
	clusterSelector := fields.ParseSelectorOrDie(conditions)
	lw := cache.NewListWatchFromClient(p.clientset.CoreV1(), string(v1.ResourcePods), metav1.NamespaceAll, clusterSelector)

	return cache.NewSharedIndexInformer(lw, &v1.Pod{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (p *Process) addDeletedPodsToQueue(resetCh chan struct{}, clusterIds []string) {
	//Since we did not set up scheduler, we don't know its actual status, using Running for now
	podInformer := p.initPodInformer(clusterIds, v1.PodRunning)
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert a deleted object  %+v to a pod", obj)
				return
			}
			p.podQueue <- pod
		},
	})
	podInformer.Run(resetCh)
}

func (p *Process) addBoundedPodsToQueue(resetCh chan struct{}, clusterIds []string) {
	podInformer := p.initPodInformer(clusterIds, v1.PodBound)
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldPod, ok := old.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert a old object  %+v to a pod", old)
				return
			}
			newPod, ok := new.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert a new object  %+v to a pod", new)
				return
			}
			if oldPod.ClusterName == "" && newPod.ClusterName != "" {
				p.podQueue <- newPod
			}
		},
	})
	podInformer.Run(resetCh)
}

func (p *Process) SendPodToCluster() {

	pod := <-p.podQueue
	if pod != nil {
		klog.V(3).Infof("Processing the item %v", pod)
		host := p.clusterIpMap[pod.Spec.ClusterName]
		token, err := p.getToken(host)
		if err != nil {
			klog.Warningf("Failed to get token from host %v", host)
			return
		}
		if pod.ObjectMeta.DeletionTimestamp != nil {
			err = openstack.DeleteInstance(host, token, pod.Spec.Hostname)
			if err == nil {
				klog.V(3).Infof("Deleting request for pod %v has been sent to %v", pod.ObjectMeta.Name, host)
			} else {
				klog.Warningf("Failed to delete the pod %v with error %v", pod.ObjectMeta.Name, err)
			}
		} else {
			instanceId, err := openstack.ServerCreate(host, token, &pod.Spec)
			if err == nil {
				klog.V(3).Infof("Creating request for pod %v has been sent to %v", pod.ObjectMeta.Name, host)
				pod.Spec.Hostname = instanceId
				pod.Status.Phase = v1.PodRunning
				updatedPod, err := p.clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).Update(pod)
				if err == nil {
					klog.V(3).Infof("Creating request for pod %v returned successfully with %v", updatedPod, instanceId)
				} else {
					klog.Warningf("Failed to update the pod %v with error %v", pod.ObjectMeta.Name, err)
				}
			} else {
				klog.Warningf("Failed to create the pod %v with error %v", pod.ObjectMeta.Name, err)
			}
		}
	}
}

func convertClustersToMap(clusters []string) ([]string, map[string]string) {
	clusterIdList := make([]string, len(clusters))
	clusterIpMap := make(map[string]string)
	for idx, cluster := range clusters {
		clusterIdIp := strings.Split(cluster, "&")
		clusterIdList[idx] = clusterIdIp[0]
		if len(clusterIdIp) != 2 {
			klog.Warningf("The input has a bad formatted cluster item %v", clusterIdIp)
		} else {
			clusterIpMap[clusterIdIp[0]] = clusterIdIp[1]
		}
	}
	return clusterIdList, clusterIpMap
}

func (p *Process) getToken(ip string) (string, error) {
	if token, ok := p.tokenMap[ip]; ok {
		if !openstack.TokenExpired(token, ip) {
			return token, nil
		}
	}
	token, err := openstack.RequestToken(ip)
	if err != nil {
		return "", err
	}
	p.tokenMap[ip] = token
	return token, nil
	return "", nil
}
