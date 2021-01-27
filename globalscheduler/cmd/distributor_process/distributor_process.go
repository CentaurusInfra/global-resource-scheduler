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
	"k8s.io/kubernetes/globalscheduler/controllers/util"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	process "k8s.io/kubernetes/globalscheduler/controllers/distributor"
)

func main() {
	configFile := flag.String("config", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	namespace := flag.String("ns", "", "The namespace of the distributor process")
	name := flag.String("n", "", "The name of the distributor process")
	logFile := flag.String("logfile", "/tmp/gs_distributor_process.log", "The log file of the distributor process")
	logLevel := flag.String("loglevel", "2", "The log level of the distributor process")

	flag.Parse()
	util.InitKlog(*namespace, *name, *logFile, *logLevel)
	defer util.FlushKlog()

	config, err := clientcmd.BuildConfigFromFlags("", *configFile)
	if err != nil {
		klog.Fatal("Failed to load config %v with errors %v", *configFile, err)
	}

	quit := make(chan struct{})
	defer close(quit)

	klog.V(2).Infof("Starting distributor process namespace %s name %s", *namespace, *name)
	process := process.NewProcess(config, *namespace, *name, quit)
	process.Run(quit)
}
