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
	"k8s.io/kubernetes/globalscheduler/controllers/util"
	"log"
	"math/rand"
	"os"
	"reflect"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	clusters           []string
}

func NewScheduler(config *rest.Config, podQueue chan *v1.Pod, name string, quit chan struct{}) Scheduler {
	schedulerclientset, err := schedulerclientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	scheduler, err := schedulerclientset.GlobalschedulerV1().Schedulers("default").Get(name, metav1.GetOptions{})
	if err != nil {
		klog.Fatal(err)
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
		clusters:           scheduler.Spec.Cluster,
	}
}

func (s *Scheduler) initInformers(quit chan struct{}) {
	schedulerSelector := fields.ParseSelectorOrDie("metadata.name=" + s.name)
	schedulerLw := cache.NewListWatchFromClient(s.schedulerclientset.GlobalschedulerV1(), "schedulers", metav1.NamespaceAll, schedulerSelector)

	schedulerInformer := cache.NewSharedIndexInformer(schedulerLw, &schedulerv1.Scheduler{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	schedulerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			newScheduler, ok := new.(*schedulerv1.Scheduler)
			if !ok {
				klog.Warningf("Failed to convert  object  %+v to a scheduler", new)
				return
			}
			if !reflect.DeepEqual(s.clusters, newScheduler.Spec.Cluster) {
				s.clusters = newScheduler.Spec.Cluster
			}
		},
		DeleteFunc: func(obj interface{}) {
			klog.V(2).Infof("The current mock scheduler %s is existing...", s.name)
			os.Exit(3)
		},
	})

	podSelector := fields.ParseSelectorOrDie("status.phase=" + string(v1.PodAssigned) + ",status.assignedScheduler.name=" + s.name)
	lw := cache.NewListWatchFromClient(s.clientset.CoreV1(), string(v1.ResourcePods), metav1.NamespaceAll, podSelector)
	podInformer := cache.NewSharedIndexInformer(lw, &v1.Pod{}, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			pod, ok := obj.(*v1.Pod)

			if !ok {
				klog.Warningf("Failed to convert  object  %+v to a pod", obj)
				return
			}
			go func() {
				s.ScheduleOne(pod)
			}()
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
			if oldPod.Status.Phase != v1.PodAssigned && newPod.Status.Phase == v1.PodAssigned {
				go func() {
					s.ScheduleOne(newPod)
				}()
			}
		},
	})

	go podInformer.Run(quit)
	go schedulerInformer.Run(quit)
}

func main() {
	configFile := flag.String("config", "/var/run/kubernetes/admin.kubeconfig", "Path to a kubeconfig. Only required if out-of-cluster.")
	name := flag.String("n", "mockscheduler", "The name of the scheduler name")
	logFile := flag.String("logfile", "/tmp/gs_scheduler.log", "The log file of the distributor process")
	logLevel := flag.String("loglevel", "2", "The log level of the distributor process")
	flag.Parse()
	util.InitKlog("default", *name, *logFile, *logLevel)
	defer util.FlushKlog()
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
	scheduler.initInformers(quit)
	<-quit
}

func (s *Scheduler) ScheduleOne(p *v1.Pod) {

	idx := 0
	if p.Status.AssignedScheduler.Name == s.name {
		if clusterLen := len(s.clusters); clusterLen > 0 {
			idx = rand.Intn(clusterLen)
			err := s.bindPod(p, s.clusters[idx])
			if err != nil {
				klog.Warningf("Failed to bind cluster with the error %v \n", err.Error())
			}
			klog.V(2).Infof("The pod %s has been bound to cluster %s", s.name, s.clusters[idx])
		} else {
			klog.V(3).Info("there are no clusters")
		}
	} else {
		klog.Errorf("The assigned name %s is not valid", s.name)
	}
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
			FieldPath:  s.name,
		},
	})
}
