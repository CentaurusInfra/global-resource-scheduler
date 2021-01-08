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

package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"log"
	"math/rand"
	"reflect"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	schedulerv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
)

const schedulerName = "mock-scheduler"

type Scheduler struct {
	name               string
	schedulerclientset *schedulerclientset.Clientset
	clientset          *kubernetes.Clientset
	podQueue           chan *v1.Pod
	clusters           map[string][]string
}

func NewScheduler(config *rest.Config, podQueue chan *v1.Pod, name string, quit chan struct{}) Scheduler {
	schedulerclientset, err := schedulerclientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	clusterMap := make(map[string][]string)
	if len(name) > 0 {
		scheduler, err := schedulerclientset.GlobalschedulerV1().Schedulers("default").Get(name, metav1.GetOptions{})
		if err != nil {
			klog.Fatal(err)
		}
		clusterMap[name] = scheduler.Spec.Cluster
	} else {
		schedulerList, err := schedulerclientset.GlobalschedulerV1().Schedulers("default").List(metav1.ListOptions{})
		if err != nil {
			klog.Fatal(err)
		}
		for _, scheduler := range schedulerList.Items {
			clusterMap[scheduler.Name] = scheduler.Spec.Cluster
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return Scheduler{
		name:               name,
		clientset:          clientset,
		schedulerclientset: schedulerclientset,
		podQueue:           podQueue,
		clusters:           clusterMap,
	}
}

func (s *Scheduler) initInformers(quit chan struct{}) {
	schedulerSelector := fields.Everything()
	if len(s.name) > 0 {
		schedulerSelector = fields.ParseSelectorOrDie("metadata.name=" + s.name)
	}
	schedulerLw := cache.NewListWatchFromClient(s.schedulerclientset.GlobalschedulerV1(), "schedulers", metav1.NamespaceAll, schedulerSelector)

	schedulerInformer := cache.NewSharedIndexInformer(schedulerLw, &schedulerv1.Scheduler{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	schedulerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldScheduler, ok := old.(*schedulerv1.Scheduler)
			if !ok {
				klog.Warningf("Failed to convert  object  %+v to a scheduler", old)
				return
			}
			newScheduler, ok := new.(*schedulerv1.Scheduler)
			if !ok {
				klog.Warningf("Failed to convert  object  %+v to a scheduler", new)
				return
			}
			if !reflect.DeepEqual(oldScheduler.Spec.Cluster, newScheduler.Spec.Cluster) {
				s.clusters[newScheduler.Name] = newScheduler.Spec.Cluster
			}
		},
	})

	podSelector := fields.ParseSelectorOrDie("status.phase=assigned,status.assignedScheduler.name!=''")
	if len(s.name) > 0 {
		podSelector = fields.ParseSelectorOrDie("status.phase=assigned,status.assignedScheduler.name=" + s.name)
	}

	lw := cache.NewListWatchFromClient(s.clientset.CoreV1(), string(v1.ResourcePods), metav1.NamespaceAll, podSelector)

	podInformer := cache.NewSharedIndexInformer(lw, &v1.Pod{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			pod, ok := obj.(*v1.Pod)

			if !ok {
				klog.Warningf("Failed to convert  object  %+v to a pod", obj)
				return
			}
			fmt.Printf("A new pod %s has been added\n", pod.Name)
			s.podQueue <- pod
		},
		UpdateFunc: func(old, new interface{}) {

			oldPod, ok := old.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert  object  %+v to a pod", old)
				return
			}
			newPod, ok := new.(*v1.Pod)
			if !ok {
				klog.Warningf("Failed to convert  object  %+v to a pod", new)
				return
			}
			fmt.Printf("A  pod %s has been updated\n", newPod.Name)
			if oldPod.Status.Phase != "assigned" && newPod.Status.Phase == "assigned" {
				s.podQueue <- newPod
			}
		},
	})

	podInformer.Run(quit)
	schedulerInformer.Run(quit)
}

func main() {
	configFile := flag.String("config", "/var/run/kubernetes/admin.kubeconfig", "Path to a kubeconfig. Only required if out-of-cluster.")
	name := flag.String("n", "mockscheduler", "The name of the scheduler name")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *configFile)

	if err != nil {
		klog.Fatal("Failed to load config %v with errors %v", *configFile, err)
	}
	rand.Seed(time.Now().Unix())

	podQueue := make(chan *v1.Pod, 300)
	defer close(podQueue)

	quit := make(chan struct{})
	defer close(quit)

	scheduler := NewScheduler(config, podQueue, *name, quit)
	go scheduler.initInformers(quit)
	go scheduler.Run(quit)
	<-quit
}

func (s *Scheduler) Run(quit chan struct{}) {
	wait.Until(s.ScheduleOne, 0, quit)
}

func (s *Scheduler) ScheduleOne() {

	p := <-s.podQueue
	fmt.Println("found a pod to schedule:", p.Namespace, "/", p.Name)

	idx := rand.Intn(len(s.clusters[p.Status.AssignedScheduler.Name]))
	fmt.Println("The current cluster to bind is: ", s.clusters[p.Status.AssignedScheduler.Name][idx])
	err := s.bindPod(p, s.clusters[p.Status.AssignedScheduler.Name][idx])
	if err != nil {
		log.Println("failed to bind cluster", err.Error())
		return
	}

	message := fmt.Sprintf("Placed pod [%s/%s] on %s\n", p.Namespace, p.Name, s.clusters[p.Status.AssignedScheduler.Name][idx])

	fmt.Println(message)
}

func (s *Scheduler) bindPod(p *v1.Pod, cluster string) error {

	return s.clientset.CoreV1().Pods(p.Namespace).Bind(&v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Cluster",
			Name:       cluster,
		},
	})
}
