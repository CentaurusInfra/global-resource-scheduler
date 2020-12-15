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

package app

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	clusterclient "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/kubernetes/globalscheduler/controllers/scheduler"
	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions"
	schedulerclient "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	schedulerinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions"
)

// client config
var (
	flagSet              = flag.NewFlagSet("crddemo", flag.ExitOnError)
	master               = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig           = flag.String("kubeconfig", "/var/run/kubernetes/controller.kubeconfig", "Path to a kubeconfig. Only required if out-of-cluster.")
	onlyOneSignalHandler = make(chan struct{})
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

const defaultNamespace = "default"

// setup stop signal
func setupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler)

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}

func StartSchedulerController() {
	flag.Parse()

	stopCh := setupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// Create kubeClientset
	kubeClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// Create schedulerClientset
	schedulerClientset, err := schedulerclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building scheduler clientset: %s", err.Error())
	}
	// Create scheduler client
	schedulerClient, err := schedulerclient.NewClient(schedulerClientset, defaultNamespace)
	if err != nil {
		klog.Fatalf("Error creating scheduler client: %s", err.Error())
	}

	// Create clusterClientset
	clusterClientset, err := clusterclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building clusterclientset: %s", err.Error())
	}
	// Create clusterClient
	clusterClient, err := clusterclient.NewClusterClient(clusterClientset, defaultNamespace)
	if err != nil {
		klog.Fatalf("Error creating cluster client: %s", err.Error())
	}

	schedulerInformerFactory := schedulerinformers.NewSharedInformerFactory(schedulerClientset, time.Second*30)
	schedulerInformer := schedulerInformerFactory.Globalscheduler().V1().Schedulers()

	clusterInformerFactory := clusterinformers.NewSharedInformerFactory(clusterClientset, time.Second*30)
	clusterInformer := clusterInformerFactory.Globalscheduler().V1().Clusters()

	schedulerController := scheduler.NewSchedulerController(kubeClientset, schedulerClient, clusterClient, schedulerInformer, clusterInformer)

	go schedulerInformerFactory.Start(stopCh)
	go clusterInformerFactory.Start(stopCh)

	if err = schedulerController.Run(1, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
