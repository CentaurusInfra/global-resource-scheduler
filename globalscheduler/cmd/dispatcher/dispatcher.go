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
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	"os"
	"syscall"
	"time"

	"k8s.io/kubernetes/globalscheduler/controllers/dispatcher"
	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions"
	dispatcherclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned"
	dispatcherinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/informers/externalversions"
)

// client config
var (
	flagSet              = flag.NewFlagSet("dispatcher_controller", flag.ExitOnError)
	master               = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig           = flag.String("kubeconfig", "/var/run/kubernetes/controller.kubeconfig", "Path to a kubeconfig. Only required if out-of-cluster.")
	onlyOneSignalHandler = make(chan struct{})
	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

func StartDispatcherController() {
	flag.Parse()

	//stopCh := setupSignalHandler()
	stopCh := make(chan struct{})
	defer close(stopCh)

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// Create kubeClientset
	kubeClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// Create Dispatcher Clientset
	dispatcherClientset, err := dispatcherclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building dispatcher clientset: %s", err.Error())
	}

	// Create clusterClientset
	clusterClientset, err := clusterclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building clusterclientset: %s", err.Error())
	}

	dispatcherInformerFactory := dispatcherinformers.NewSharedInformerFactory(dispatcherClientset, time.Second*30)
	dispatcherInformer := dispatcherInformerFactory.Globalscheduler().V1().Dispatchers()

	clusterInformerFactory := clusterinformers.NewSharedInformerFactory(clusterClientset, time.Second*30)
	clusterInformer := clusterInformerFactory.Globalscheduler().V1().Clusters()

	dispatcherController := dispatcher.NewDispatcherController(kubeClientset, dispatcherClientset, clusterClientset, dispatcherInformer, clusterInformer)

	go dispatcherInformerFactory.Start(stopCh)
	go clusterInformerFactory.Start(stopCh)

	if err = dispatcherController.Run(1, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
