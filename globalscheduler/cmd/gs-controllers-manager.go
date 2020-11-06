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
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/cluster"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	"k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions"
)

const defaultWorkers = 4

var (
	masterURL  string
	kubeconfig string
	domainName string
	workers    int
)

func StartClusterController() {
	//klog.InitFlags(nil)
	flag.Parse()
	if workers <= 0 {
		workers = defaultWorkers
	}

	defer klog.Flush()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("error getting client config: %s", err.Error())
	}

	clusterClient, err := clusterclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("error building global scheduler cluster client: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("error building Kubernetes client: %s", err.Error())
	}

	informerFactory := externalversions.NewSharedInformerFactory(clusterClient, 10*time.Minute)
	stopCh := make(chan struct{})
	defer close(stopCh)

	clusterInformer := informerFactory.Globalscheduler().V1().Clusters()
	controller := cluster.NewClusterController(kubeClient, clusterClient, clusterInformer)
	err := controller.CreateCRD()
	if err != nil {
		klog.Fatalf("error register cluster crd: %s", err.Error())
	}

	informerFactory.Start(stopCh)
	controller.Run(workers, stopCh)
	klog.Infof("global scheduler cluster controller exited")
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.IntVar(&workers, "concurrent-workers", defaultWorkers, "The number of workers that are allowed to process concurrently.")
	flag.StringVar(&domainName, "cluster-domain", "cluster.local", "the cluster-internal domain name for Services.")
}
