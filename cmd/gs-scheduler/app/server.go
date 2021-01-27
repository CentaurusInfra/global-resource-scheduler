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
	"github.com/spf13/cobra"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/apiserver"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/options"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/router"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	utilflag "k8s.io/kubernetes/pkg/util/flag"
)

const (
	// GS Scheduler component name
	componentGSScheduler = "gs-scheduler"
)

// NewKubeletCommand creates a *cobra.Command object with default parameters
func NewGSSchedulerCommand() *cobra.Command {
	s := options.NewServerRunOptions()
	cmd := &cobra.Command{
		Use: componentGSScheduler,
		Long: `The gs-scheduler is the scheduler for the global resource scheduler. 
It can support run multiple process using one of: the schedulername; a flag to
specify the name of the scheduler.`,
		// The GSScheduler has special flag parsing requirements to enforce flag precedence rules,
		// so we do all our parsing manually in Run, below.
		// DisableFlagParsing=true provides the full set of flags passed to the gs-scheduler in the
		// `args` arg to Run, without Cobra's interference.
		RunE: func(cmd *cobra.Command, args []string) error {
			utilflag.PrintFlags(cmd.Flags())
			config := s.Config()

			// set up stopCh here
			stopCh := genericapiserver.SetupSignalHandler()

			return Run(config, stopCh)
		},
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}
	return cmd
}

// Run runt the service and other period monitor to make sure the consistency.
func Run(config *types.GSSchedulerConfiguration, stopCh <-chan struct{}) error {
	logger.Infof("Global Scheduler Running...")

	// init scheduler
	scheduler.InitScheduler(config, stopCh)

	sched := scheduler.GetScheduler()
	if sched == nil {
		return fmt.Errorf("get new scheduler failed")
	}

	// start scheduler pod informer and run
	//sched.StartInformersAndRun(stopCh)
	sched.StartPodInformerAndRun(stopCh)

	// start allocation API here
	// StartAPI(stopCh)
	select {
	case <-stopCh:
		break
	}

	return nil
}

// Start allocation API for gs-scheduler
func StartAPI(stopCh <-chan struct{}) error {
	router.Register()
	hs, err := apiserver.NewHTTPServer()
	if err != nil {
		logger.Errorf("new http server failed!, err: %s", err.Error())
		return err
	}

	return hs.BlockingRun(stopCh)
}
