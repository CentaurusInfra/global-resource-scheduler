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
	"fmt"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/apiserver"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/options"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/router"
)

// Run runt the service and other period monitor to make sure the consistency.
func Run(serverOptions *options.ServerRunOptions, stopCh <-chan struct{}) error {
	logger.Infof("Global Scheduler Running...")

	// init scheduler
	sched := scheduler.GetScheduler(stopCh)
	if sched == nil {
		return fmt.Errorf("get new scheduler failed")
	}

	// start scheduler resource cache informer and run
	sched.StartInformersAndRun(stopCh)

	router.Register()
	hs, err := apiserver.NewHTTPServer()
	if err != nil {
		logger.Errorf("new http server failed!, err: %s", err.Error())
		return err
	}

	return hs.BlockingRun(stopCh)
}
