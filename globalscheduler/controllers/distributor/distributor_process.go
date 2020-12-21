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
package distributor

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	distributortype "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor"
	distributorv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/v1"
	"strconv"

	"errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	distributorclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/clientset/versioned"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	schedulerinformer "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions"
	schedulerlister "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/listers/scheduler/v1"
	schedulerv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
)

const distributorName = "distributor"

type predicateFunc func(scheduler *schedulerv1.Scheduler, pod *v1.Pod) bool
type priorityFunc func(scheduler *schedulerv1.Scheduler, pod *v1.Pod) int

type Process struct {
	namespace            string
	name                 string
	distributorClientset *distributorclientset.Clientset
	schedulerClientset   *schedulerclientset.Clientset
	clientset            *kubernetes.Clientset
	podQueue             chan *v1.Pod
	resetCh              chan struct{}
	schedulerLister      schedulerlister.SchedulerLister
	predicates           []predicateFunc
	priorities           []priorityFunc
	rangeStart           int64
	rangeEnd             int64
}

func NewProcess(config *rest.Config, namespace string, name string, quit chan struct{}, start int64, end int64) Process {
	podQueue := make(chan *v1.Pod, 300)
	defer close(podQueue)

	distributorClientset, err := distributorclientset.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	schedulerClientset, err := schedulerclientset.NewForConfig(config)
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
		namespace:            namespace,
		name:                 name,
		clientset:            clientset,
		distributorClientset: distributorClientset,
		schedulerClientset:   schedulerClientset,
		podQueue:             podQueue,
		resetCh:              resetCh,
		schedulerLister:      initSchedulerInformers(schedulerClientset, quit),
		predicates: []predicateFunc{
			GeoLocationPredicate,
		},
		priorities: []priorityFunc{
			GeoLocationPriority,
		},
		rangeStart: start,
		rangeEnd:   end,
	}
}

func initSchedulerInformers(schedulerClientset *schedulerclientset.Clientset, quit chan struct{}) schedulerlister.SchedulerLister {
	schedulerFactory := schedulerinformer.NewSharedInformerFactory(schedulerClientset, 0)
	schedulerInformer := schedulerFactory.Globalscheduler().V1().Schedulers()
	schedulerFactory.Start(quit)
	return schedulerInformer.Lister()
}

func (p *Process) Run(quit chan struct{}) {

	distributorSelector := fields.ParseSelectorOrDie(
		",metatdata.namespace" + p.namespace + ",metatdata.name=" + p.name)
	distributorLW := cache.NewListWatchFromClient(p.distributorClientset, string(distributortype.Plural), metav1.NamespaceAll, distributorSelector)

	distributorInformer := cache.NewSharedIndexInformer(distributorLW, &distributorv1.Distributor{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	distributorInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			quit <- struct{}{}
		},
		UpdateFunc: func(old, new interface{}) {
			oldDistributor, ok := old.(*distributorv1.DistributorRange)
			if !ok {
				klog.Warningf("Failed to convert a old object  %+v to a distributor", old)
				return
			}
			newDistributor, ok := new.(*distributorv1.DistributorRange)
			if !ok {
				klog.Warningf("Failed to convert a new object  %+v to a distributor", new)
				return
			}
			if oldDistributor.Start != newDistributor.Start || oldDistributor.End != newDistributor.End {
				podInformer := p.initPodInformers(p.resetCh, newDistributor.Start, newDistributor.End)
				podInformer.Run(p.resetCh)
			}
		},
	})

	podInformer := p.initPodInformers(p.resetCh, p.rangeStart, p.rangeEnd)
	go podInformer.Run(p.resetCh)
	go distributorInformer.Run(quit)
	wait.Until(p.ScheduleOne, 0, quit)
}

func (p *Process) initPodInformers(resetCh chan struct{}, start, end int64) cache.SharedIndexInformer {
	resetCh <- struct{}{}

	podSelector := fields.ParseSelectorOrDie(
		"status.phase=" + string(v1.PodPending) + ",status.phase!=" + string(v1.PodFailed) +
			",metatdata.hashkey=gte:" + strconv.FormatInt(start, 10) +
			",metatdata.hashkey=lte:" + strconv.FormatInt(end, 10))
	lw := cache.NewListWatchFromClient(p.clientset.CoreV1(), string(v1.ResourcePods), metav1.NamespaceAll, podSelector)

	podInformer := cache.NewSharedIndexInformer(lw, &v1.Pod{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert  object  %+v to a pod", obj)
				return
			}
			if pod.Spec.ResourceType != "vm" {
				return
			}
			p.podQueue <- pod
		},
	})
	return podInformer
}

func (p *Process) ScheduleOne() {

	pod := <-p.podQueue
	if pod != nil {
		klog.V(2).Infof("Found a pod %v-%v to schedule:", pod.Namespace, pod.Name)

		scheduler, err := p.findFit(pod)
		if err != nil {
			klog.Warningf("Failed to find  node that fits scheduler with the error %v", err)
			return
		}

		err = p.bindPod(pod, scheduler)
		if err != nil {
			klog.Warningf("Failed to bind cluster with the error %vs", err)
			return
		}

		klog.V(2).Infof("Placed pod [%s/%s] on %s\n", pod.Namespace, pod.Name, scheduler)
	}
}

func (p *Process) findFit(pod *v1.Pod) (string, error) {
	scheduler, err := p.schedulerLister.List(labels.Everything())
	if err != nil {
		return "", err
	}

	filteredClusters := p.runPredicates(scheduler, pod)
	if len(filteredClusters) == 0 {
		return "", errors.New("failed to find cluster that fits pod")
	}
	priorities := p.prioritize(filteredClusters, pod)
	return p.findBestScheduler(priorities), nil
}

func (p *Process) bindPod(pod *v1.Pod, scheduler string) error {

	return p.clientset.CoreV1().Pods(pod.Namespace).Bind(&v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Scheduler",
			Name:       scheduler,
		},
	})
}

func (p *Process) runPredicates(schedulers []*schedulerv1.Scheduler, pod *v1.Pod) []*schedulerv1.Scheduler {
	filteredSchedulers := make([]*schedulerv1.Scheduler, 0)
	for _, scheduler := range schedulers {
		if p.predicatesApply(scheduler, pod) {
			filteredSchedulers = append(filteredSchedulers, scheduler)
		}
	}
	klog.V(3).Infof("schedulers %v that fit the pod %v", filteredSchedulers, pod)
	return filteredSchedulers
}

func (p *Process) predicatesApply(scheduler *schedulerv1.Scheduler, pod *v1.Pod) bool {
	for _, predicate := range p.predicates {
		if !predicate(scheduler, pod) {
			return false
		}
	}
	return true
}

func GeoLocationPredicate(scheduler *schedulerv1.Scheduler, pod *v1.Pod) bool {
	schedulerLoc := scheduler.Spec.Location
	podLoc := pod.Spec.VirtualMachine.ResourceCommonInfo.Selector.GeoLocation
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" && schedulerLoc.City == "" && schedulerLoc.Area == "" {
		return true
	}
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" && schedulerLoc.City == "" {
		return schedulerLoc.Area == podLoc.Area
	}
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" {
		return schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City
	}
	if schedulerLoc.Country == "" {
		return schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City &&
			schedulerLoc.Province == podLoc.Province
	}
	return schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City &&
		schedulerLoc.Province == podLoc.Province && schedulerLoc.Country == podLoc.Country
}

func (p *Process) prioritize(schedulers []*schedulerv1.Scheduler, pod *v1.Pod) map[string]int {
	priorities := make(map[string]int)
	for _, scheduler := range schedulers {
		for _, priority := range p.priorities {
			priorities[scheduler.Name] += priority(scheduler, pod)
		}
	}
	klog.V(5).Infof("calculated priorities: %v", priorities)
	return priorities
}

func (p *Process) findBestScheduler(priorities map[string]int) string {
	var maxP int
	var bestschheduler string
	for scheduler, p := range priorities {
		if p > maxP {
			maxP = p
			bestschheduler = scheduler
		}
	}
	return bestschheduler
}

func GeoLocationPriority(scheduler *schedulerv1.Scheduler, pod *v1.Pod) int {
	schedulerLoc := scheduler.Spec.Location
	podLoc := pod.Spec.VirtualMachine.ResourceCommonInfo.Selector.GeoLocation
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" && schedulerLoc.City == "" && schedulerLoc.Area == "" {
		return 1
	}
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" && schedulerLoc.City == "" {
		if schedulerLoc.Area == podLoc.Area {
			return 10
		}
		return 0
	}
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" {
		if schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City {
			return 100
		}
		return 0
	}
	if schedulerLoc.Country == "" {
		if schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City &&
			schedulerLoc.Province == podLoc.Province {
			return 1000
		}
		return 0
	}
	if schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City &&
		schedulerLoc.Province == podLoc.Province && schedulerLoc.Country == podLoc.Country {
		return 10000
	}
	return 0
}
