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
	"sync"
	"time"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
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

	//2. config
	config, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to load config %v with errors %v", *kubeconfig, err)
	}

	//3. stop channel
	stopCh := make(chan struct{})
	defer close(stopCh)
	quit := make(chan struct{})
	defer close(quit)

	//4. kubecluent
	kubeClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("error building Kubernetes client: %s", err.Error())
	}

	//5. apiextensions clientset to create crd programmatically
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("error - building global scheduler cluster apiextensions client: %s", err.Error())
	}

	//6. distributer_server - clientset, controller
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

	//7. scheduler
	// Create schedulerClientset
	schedulerClientset, err := schedulerclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building scheduler clientset: %s", err.Error())
	}
	// Create clusterClientset
	clusterClientset, err := clusterclientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building clusterclientset: %s", err.Error())
	}
	schedulerInformerFactory := schedulerinformer.NewSharedInformerFactory(schedulerClientset, time.Second*30)
	schedulerInformer := schedulerInformerFactory.Globalscheduler().V1().Schedulers()
	clusterInformerFactory := clusterinformer.NewSharedInformerFactory(clusterClientset, time.Second*30)
	clusterInformer := clusterInformerFactory.Globalscheduler().V1().Clusters()
	schedulerController := scheduler.NewSchedulerController(kubeClientset, schedulerClientset, clusterClientset, schedulerInformer, clusterInformer)

	//8. cluster clientset
	clusterController := cluster.NewClusterController(kubeClientset, apiextensionsClient, clusterClientset, clusterInformer, *grpcHost)
	err = clusterController.CreateCRD()
	if err != nil {
		klog.Fatalf("error - register cluster crd: %s", err.Error())
	}

	//9. start controllers in the order of less dependent one first
	var wg sync.WaitGroup
	klog.Infof("Start controllers")
	klog.Infof("Start cluster controller")
	wg.Add(1)
	go clusterController.RunController(*workers, stopCh, &wg)
	klog.Infof("Start scheduler controller")
	wg.Add(1)
	go schedulerController.RunController(*workers, stopCh, &wg)
	klog.Infof("Start distributor controller")
	wg.Add(1)
	go distributorController.RunController(quit, &wg)
	klog.Info("Main: Waiting for controllers to start")
	wg.Wait()
	klog.Info("Main: Completed to start all controllers")
}
