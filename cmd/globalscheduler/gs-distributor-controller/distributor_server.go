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
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	distributorinformer "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/informers/externalversions"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/logs"
	"k8s.io/klog"
	controller "k8s.io/kubernetes/globalscheduler/controllers/distributor"
	clientset "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/clientset/versioned"
)

func main() {
	configFile := flag.String("config", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.Parse()
	logs.InitLogs()
	defer logs.FlushLogs()
	config, err := clientcmd.BuildConfigFromFlags("", *configFile)
	if err != nil {
		klog.Fatal("Failed to load config %v with errors %v", *configFile, err)
	}

	quit := make(chan struct{})
	defer close(quit)

	kubeClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %v", err)
	}

	distributorClientset, err := clientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building distributor clientset: %v", err)
	}
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(config)
	distributorFactory := distributorinformer.NewSharedInformerFactory(distributorClientset, 0)
	distributorInformer := distributorFactory.Globalscheduler().V1().Distributors()

	controller := controller.NewController(*configFile, kubeClientset, apiextensionsClient, distributorClientset, distributorInformer)
	err = controller.EnsureCustomResourceDefinitionCreation()
	if err != nil {
		klog.Fatalf("Error registering distributor: %v", err)
	}
	distributorFactory.Start(quit)
	controller.Run(quit)
	klog.V(2).Infof("global scheduler distributor controller exited")
}
