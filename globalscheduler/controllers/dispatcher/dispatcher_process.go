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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	dispatcherv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/v1"
	"reflect"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	clusterinformer "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions"
	clusterlister "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/listers/cluster/v1"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	dispatcherclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
)

const dispatcherName = "dispatcher"

type Process struct {
	namespace           string
	name                string
	dispatcherClientset *dispatcherclientset.Clientset
	schedulerClientset  *schedulerclientset.Clientset
	clusterClientset    *clusterclientset.Clientset
	clientset           *kubernetes.Clientset
	podQueue            chan *v1.Pod
	resetCh             chan struct{}
	clusterLister       clusterlister.ClusterLister
	clusterIds          []string
}

func NewProcess(config *rest.Config, namespace string, name string, quit chan struct{}, start int64, end int64) Process {
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

	schedulerClientset, err := schedulerclientset.NewForConfig(config)
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

	resetCh := make(chan struct{})
	defer close(resetCh)
	go func() {
		for {
			select {
			case <-quit:
				resetCh <- struct{}{}
				return
			}
		}
	}()

	return Process{
		namespace:           namespace,
		name:                name,
		clientset:           clientset,
		dispatcherClientset: dispatcherClientset,
		clusterClientset:    clusterClientset,
		schedulerClientset:  schedulerClientset,
		podQueue:            podQueue,
		resetCh:             resetCh,
		clusterLister:       initClusterInformers(clusterClientset, quit),
		//TODO: It should be changed to cluster id list
		clusterIds: dispatcher.Spec.POD,
	}
}

func initClusterInformers(clusterClientset *clusterclientset.Clientset, quit chan struct{}) clusterlister.ClusterLister {
	clusterFactory := clusterinformer.NewSharedInformerFactory(clusterClientset, 0)
	clusterInformer := clusterFactory.Globalscheduler().V1().Clusters()
	clusterFactory.Start(quit)
	return clusterInformer.Lister()
}

func (p *Process) Run(quit chan struct{}) {

	dispatcherSelector := fields.ParseSelectorOrDie(
		",metatdata.namespace" + p.namespace + ",metatdata.name=" + p.name)
	dispatcherLW := cache.NewListWatchFromClient(p.dispatcherClientset, "Dispatchers", metav1.NamespaceAll, dispatcherSelector)

	dispatcherInformer := cache.NewSharedIndexInformer(dispatcherLW, &dispatcherv1.Dispatcher{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	dispatcherInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			quit <- struct{}{}
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
			if reflect.DeepEqual(oldDispatcher.Spec.POD, newDispatcher.Spec.POD) {
				podInformer := p.initClusterInformers(p.resetCh, newDispatcher.Spec.POD)
				podInformer.Run(p.resetCh)
			}
		},
	})

	clusterInformer := p.initClusterInformers(p.resetCh, p.clusterIds)
	podInformer := p.initPodInformers(p.resetCh, p.clusterIds)
	go clusterInformer.Run(p.resetCh)
	go podInformer.Run(p.resetCh)
	go dispatcherInformer.Run(quit)
	wait.Until(p.ScheduleOne, 0, quit)
}

func (p *Process) initClusterInformers(resetCh chan struct{}, clusterIds []string) cache.SharedIndexInformer {
	resetCh <- struct{}{}
	conditions := "metedata.namespace=" + p.namespace + ","
	for _, clusterId := range clusterIds {
		conditions = conditions + "metadata.name=" + clusterId + ";"
	}
	clusterSelector := fields.ParseSelectorOrDie(conditions)

	lw := cache.NewListWatchFromClient(p.clusterClientset.GlobalschedulerV1(), "Dispatchers", metav1.NamespaceAll, clusterSelector)

	clusterInformer := cache.NewSharedIndexInformer(lw, &clusterv1.Cluster{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	return clusterInformer
}

func (p *Process) initPodInformers(resetCh chan struct{}, clusterIds []string) cache.SharedIndexInformer {
	resetCh <- struct{}{}
	conditions := "metedata.namespace=" + p.namespace + ","
	for _, clusterId := range clusterIds {
		conditions = conditions + "spec.clusterName=" + clusterId + ";"
	}
	clusterSelector := fields.ParseSelectorOrDie(conditions)

	lw := cache.NewListWatchFromClient(p.clientset.CoreV1(), "Pods", metav1.NamespaceAll, clusterSelector)

	podInformer := cache.NewSharedIndexInformer(lw, &v1.Pod{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		//TODO
		DeleteFunc: func(obj interface{}) {
			//Here we can get all the pods information and compose openstack request
		},
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
			//TODO need to add a new status condition "SchedulerAssigned"
			if !contains(oldPod.Status.Conditions, "SchedulerAssigned") && contains(newPod.Status.Conditions, "SchedulerAssigned") {
				p.podQueue <- newPod
			}
		},
	})
	return podInformer
}
func contains(list []v1.PodCondition, value string) bool {
	for _, item := range list {
		if item.String() == value {
			return true
		}
	}
	return false
}

func (p *Process) ScheduleOne() {

	pod := <-p.podQueue
	if pod != nil {
		klog.V(2).Infof("Found a pod %v-%v to compose a request:", pod.Namespace, pod.Name)
		//TODO compose a request sent to openstack
	}
}
