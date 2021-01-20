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
	"errors"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	distributortype "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor"
	distributorclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/clientset/versioned"
	distributorv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/v1"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	schedulerv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
	"os"
	"reflect"
	"strconv"
	"syscall"
)

const distributorName = "distributor"

type predicateFunc func(scheduler schedulerv1.Scheduler, pod *v1.Pod) bool
type priorityFunc func(scheduler schedulerv1.Scheduler, pod *v1.Pod) int

type Process struct {
	namespace            string
	name                 string
	distributorClientset *distributorclientset.Clientset
	schedulerClientset   *schedulerclientset.Clientset
	schedulers           []schedulerv1.Scheduler
	clientset            *kubernetes.Clientset
	podQueue             chan *v1.Pod
	predicates           []predicateFunc
	priorities           []priorityFunc
	rangeStart           int64
	rangeEnd             int64
	pid                  int
}

// NewProcess creates a process to handle pods whose status phase is pending and resource type is vm
func NewProcess(config *rest.Config, namespace string, name string, quit chan struct{}) Process {
	podQueue := make(chan *v1.Pod, 300)

	distributorClientset, err := distributorclientset.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	distributor, err := distributorClientset.GlobalschedulerV1().Distributors(namespace).Get(name, metav1.GetOptions{})
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


	if err != nil {
		klog.Warningf("Failed to run distributor process %v - %v with the err %v", namespace, distributorName, err)
	}

	return Process{
		namespace:            namespace,
		name:                 name,
		clientset:            clientset,
		distributorClientset: distributorClientset,
		schedulerClientset:   schedulerClientset,
		podQueue:             podQueue,
		predicates: []predicateFunc{
			GeoLocationPredicate,
		},
		priorities: []priorityFunc{
			GeoLocationPriority,
		},
		rangeStart: distributor.Spec.Range.Start,
		rangeEnd:   distributor.Spec.Range.End,
		pid:        os.Getgid(),
		schedulers: make([]schedulerv1.Scheduler, 0),
	}
}

// Run gets all the distributor, scheduler, and pod information and assign schedulers to pods
func (p *Process) Run(quit chan struct{}) {

	distributorSelector := fields.ParseSelectorOrDie("metadata.name=" + p.name)
	distributorLW := cache.NewListWatchFromClient(p.distributorClientset.GlobalschedulerV1(), string(distributortype.Plural), p.namespace, distributorSelector)
	distributorInformer := cache.NewSharedIndexInformer(distributorLW, &distributorv1.Distributor{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	distributorInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			if err := syscall.Kill(-p.pid, 15); err != nil {
				klog.Fatalf("Fail to exit the current process %v\n", err)
			}
		},
		UpdateFunc: func(old, new interface{}) {

			newDistributor, ok := new.(*distributorv1.Distributor)
			if !ok {
				klog.Warningf("Failed to convert a new object  %+v to a distributor", new)
				return
			}
			if newDistributor.Spec.Range.Start != p.rangeStart || newDistributor.Spec.Range.End != p.rangeEnd  {
				p.rangeStart = newDistributor.Spec.Range.Start
				p.rangeEnd = newDistributor.Spec.Range.End
				if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
					klog.Fatal(err)
				}
			}
		},
	})
	schedulerLW := cache.NewListWatchFromClient(p.schedulerClientset.GlobalschedulerV1(), "schedulers", p.namespace, fields.Everything())
	schedulerInformer := cache.NewSharedIndexInformer(schedulerLW, &schedulerv1.Scheduler{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	schedulerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			scheduler, ok := obj.(*schedulerv1.Scheduler)
			if !ok {
				klog.Warningf("Failed to convert the object  %+v to a scheduler", obj)
				return
			}
			p.schedulers = append(p.schedulers, *scheduler)
		},
		DeleteFunc: func(obj interface{}) {
			scheduler, ok := obj.(*schedulerv1.Scheduler)
			if !ok {
				klog.Warningf("Failed to convert the object  %+v to a scheduler", obj)
				return
			}
			for _, item := range p.schedulers {
				if reflect.DeepEqual(item, scheduler) {
					_, p.schedulers = p.schedulers[len(p.schedulers) - 1], p.schedulers[:len(p.schedulers) - 1]
					return
				}
			}
		},
	})
	go schedulerInformer.Run(quit)
	podInformer := p.initPodInformers(p.rangeStart, p.rangeEnd)
	go podInformer.Run(quit)
	go distributorInformer.Run(quit)

	wait.Until(p.ScheduleOne, 0, quit)
}

func (p *Process) initPodInformers(start, end int64) cache.SharedIndexInformer {
	podSelector := fields.ParseSelectorOrDie(fmt.Sprintf("status.phase=%s,metadata.hashkey=gte:%s,metadata.hashkey=lte:%s",
		string(v1.PodPending), strconv.FormatInt(start, 10),  strconv.FormatInt(end, 10)))

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
			go func() {
				p.podQueue <- pod
			}()
		},
	})
	return podInformer
}

// ScheduleOne is to process pods by assign a scheduler to it
func (p *Process) ScheduleOne() {
	pod := <-p.podQueue
	if pod != nil {
		klog.V(2).Infof("Found a pod %v-%v to schedule:", pod.Namespace, pod.Name)

		scheduler, err := p.findFit(pod)
		if err != nil {
			klog.Warningf("Failed to find scheduler that fits scheduler with the error %v", err)
			return
		}

		err = p.bindPod(pod, scheduler)
		if err != nil {
			klog.Warningf("Failed to assign scheduler %v to pod %v with the error %vs", scheduler, pod, err)
			pod.Status.Phase = v1.PodFailed
			if _, err = p.clientset.CoreV1().Pods(p.namespace).UpdateStatus(pod); err != nil {
				klog.Warningf("Failed to update pod %v - %v status to failed", pod.Namespace, pod.Name)
			}
			return
		}

		klog.V(3).Infof("Assigned pod [%s/%s] on %s\n", pod.Namespace, pod.Name, scheduler)
	}
}

func (p *Process) findFit(pod *v1.Pod) (string, error) {
	filteredSchedulers := p.runPredicates(p.schedulers, pod)
	if len(filteredSchedulers) == 0 {
		return "", errors.New("failed to find a scheduler that fits pod")
	}
	priorities := p.prioritize(filteredSchedulers, pod)
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

func (p *Process) runPredicates(schedulers []schedulerv1.Scheduler, pod *v1.Pod) []schedulerv1.Scheduler {
	filteredSchedulers := make([]schedulerv1.Scheduler, 0)
	for _, scheduler := range schedulers {
		if p.predicatesApply(scheduler, pod) {
			filteredSchedulers = append(filteredSchedulers, scheduler)
		}
	}
	klog.V(3).Infof("schedulers %v that fit the pod %v", filteredSchedulers, pod)
	return filteredSchedulers
}

func (p *Process) predicatesApply(scheduler schedulerv1.Scheduler, pod *v1.Pod) bool {
	for _, predicate := range p.predicates {
		if !predicate(scheduler, pod) {
			return false
		}
	}
	return true
}

// GeoLocationPredicate is to find if schedulers match a pod based on their geoLocations
func GeoLocationPredicate(scheduler schedulerv1.Scheduler, pod *v1.Pod) bool {
	schedulerLoc := scheduler.Spec.Location
	podLoc := pod.Spec.VirtualMachine.ResourceCommonInfo.Selector.GeoLocation
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" && schedulerLoc.City == "" && schedulerLoc.Area == "" {
		return true
	}
	if podLoc.Area == "" && podLoc.City == "" && podLoc.Province == "" && podLoc.Country == "" {
		return true
	}
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" && schedulerLoc.City == "" {
		return schedulerLoc.Area == podLoc.Area
	}
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" {
		return schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City
	}
	if schedulerLoc.Country == "" {
		return schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City && schedulerLoc.Province == podLoc.Province
	}
	return schedulerLoc.Area == podLoc.Area && schedulerLoc.City == podLoc.City &&
		schedulerLoc.Province == podLoc.Province && schedulerLoc.Country == podLoc.Country
}

func (p *Process) prioritize(schedulers []schedulerv1.Scheduler, pod *v1.Pod) map[string]int {
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

// GeoLocationPriority is to prioritize schedulers on their geoLocations
func GeoLocationPriority(scheduler schedulerv1.Scheduler, pod *v1.Pod) int {
	schedulerLoc := scheduler.Spec.Location
	podLoc := pod.Spec.VirtualMachine.ResourceCommonInfo.Selector.GeoLocation
	if schedulerLoc.Country == "" && schedulerLoc.Province == "" && schedulerLoc.City == "" && schedulerLoc.Area == "" {
		return 1
	}
	if podLoc.Area == "" && podLoc.City == "" && podLoc.Province == "" && podLoc.Country == "" {
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
