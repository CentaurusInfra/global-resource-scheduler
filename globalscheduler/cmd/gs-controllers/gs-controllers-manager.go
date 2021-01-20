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
	"sync"
	"time"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	//dispatcher
	"k8s.io/kubernetes/globalscheduler/controllers/dispatcher"
	dispatcherclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned"
	dispatcherinformer "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/informers/externalversions"
	//distributer server & worker
	"k8s.io/kubernetes/globalscheduler/controllers/distributor"
	distributorclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/clientset/versioned"
	distributorinformer "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/informers/externalversions"
	//scheduler
	"k8s.io/kubernetes/globalscheduler/controllers/scheduler"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	schedulerinformer "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions"
	//cluster
	"k8s.io/kubernetes/globalscheduler/controllers/cluster"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	clusterinformer "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions"
)

const (
	defaultNamespace = "default"
	defaultWorkers   = 1
)

func main() {
	//1. flag & log
	kubeconfig := flag.String("kubeconfig", "/var/run/kubernetes/admin.kubeconfig", "Path to a kubeconfig. Only required if out-of-cluster.")
	masterURL := flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	workers := flag.Int("concurrent-workers", 0, "The number of workers that are allowed to process concurrently.")
	grpcHost := flag.String("grpchost", "", "IP address of resource collector.")
	flag.Parse()
	if *workers <= 0 {
		*workers = defaultWorkers
	}
	defer klog.Flush()
	klog.Infof("gs-controllers-manager configuration: kubeconfig %s, masterURL %s, workers %v, grpcHost %s", *kubeconfig, *masterURL, *workers, *grpcHost)

	//2. config
	config, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to load config %v with errors %v", *kubeconfig, err)
	}

	//3. stop channel
	//cluster, scheduler, dispatcher
	stopCh := make(chan struct{})
	defer close(stopCh)
	//distributer
	quit := make(chan struct{})
	defer close(quit)

	//4. kubecluent
	klog.Infof("configure kubernetes cluent")
	kubeClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("error building Kubernetes client: %s", err.Error())
	}

	//5. apiextensions clientset to create crd programmatically
	klog.Infof("configure apiextensions client")
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("error - building global scheduler cluster apiextensions client: %s", err.Error())
	}

	//6. distributer_server
	klog.Infof("configure distributor controller")
	distributorClientset, err := distributorclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building distributor clientset: %v", err)
	}
	distributorFactory := distributorinformer.NewSharedInformerFactory(distributorClientset, 0)
	distributorInformer := distributorFactory.Globalscheduler().V1().Distributors()
	distributorController := distributor.NewController(*kubeconfig, kubeClientset, apiextensionsClient, distributorClientset, distributorInformer)
	err = distributorController.EnsureCustomResourceDefinitionCreation()
	if err != nil {
		klog.Fatalf("Error registering distributor: %v", err)
	}

	//7. cluster
	klog.Infof("configure cluster controller")
	clusterClientset, err := clusterclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building clusterclientset: %s", err.Error())
	}
	clusterInformerFactory := clusterinformer.NewSharedInformerFactory(clusterClientset, time.Second*30)
	clusterInformer := clusterInformerFactory.Globalscheduler().V1().Clusters()
	clusterController := cluster.NewClusterController(kubeClientset, apiextensionsClient, clusterClientset, clusterInformer, *grpcHost)
	err = clusterController.CreateClusterCRD()
	if err != nil {
		klog.Fatalf("error - register cluster crd: %s", err.Error())
	}

	//8. dispatcher
	klog.Infof("configure dispatcher controller")
	dispatcherClientset, err := dispatcherclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building dispatcher clientset: %s", err.Error())
	}
	dispatcherInformerFactory := dispatcherinformer.NewSharedInformerFactory(dispatcherClientset, time.Second*30)
	dispatcherInformer := dispatcherInformerFactory.Globalscheduler().V1().Dispatchers()
	dispatcherController := dispatcher.NewDispatcherController(*kubeconfig, kubeClientset, apiextensionsClient, dispatcherClientset, clusterClientset, dispatcherInformer, clusterInformer)
	err = dispatcherController.CreateDispatcherCRD()
	if err != nil {
		klog.Fatalf("error - register dispatcher crd: %s", err.Error())
	}

	//9. scheduler
	klog.Infof("configure scheduler controller")
	schedulerClientset, err := schedulerclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building scheduler clientset: %s", err.Error())
	}
	schedulerInformerFactory := schedulerinformer.NewSharedInformerFactory(schedulerClientset, time.Second*30)
	schedulerInformer := schedulerInformerFactory.Globalscheduler().V1().Schedulers()
	schedulerController := scheduler.NewSchedulerController(kubeClientset, apiextensionsClient, schedulerClientset, clusterClientset, schedulerInformer, clusterInformer)
	if err = schedulerController.CreateSchedulerCRD(); err != nil {
		klog.Fatalf("error - register scheduler crd: %s", err.Error())
	}

	//10. start controllers, independent controller first
	var wg sync.WaitGroup
	klog.Infof("Start controllers")
	fmt.Println("Start controllers")

	//distributor
	klog.Infof("Start distributor controller")
	wg.Add(1)
	distributorFactory.Start(quit)
	go distributorController.RunController(quit, &wg)
	fmt.Println("distributor controller started")

	//cluster
	klog.Infof("Start cluster controller")
	wg.Add(1)
	clusterInformerFactory.Start(stopCh)
	go clusterController.RunController(*workers, stopCh, &wg)
	fmt.Println("cluster controller started")

	//dispatcher
	klog.Infof("Start dispatcher controller")
	wg.Add(1)
	dispatcherInformerFactory.Start(stopCh)
	go dispatcherController.RunController(*workers, stopCh, &wg)
	fmt.Println("dispatcher controller started")

	//scheduler
	klog.Infof("Start scheduler controller")
	wg.Add(1)
	schedulerInformerFactory.Start(stopCh)
	go schedulerController.RunController(*workers, stopCh, &wg)
	fmt.Println("scheduler controller started")

	klog.Info("Main: Waiting for controllers to start")
	wg.Wait()
	klog.Info("Main: Completed to start all controllers")
}
