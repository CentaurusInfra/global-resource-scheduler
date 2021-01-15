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

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/options"
	utilflag "k8s.io/kubernetes/pkg/util/flag"
	"k8s.io/kubernetes/resourcecollector/pkg/collector"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/common/apiserver"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/router"
	"k8s.io/kubernetes/resourcecollector/pkg/collector/rpc"

	"github.com/spf13/cobra"
)

const (
	// Resource Collector component name
	componentResourceCollector = "resource-collector"
)

// NewResourceCollectorCommand creates a *cobra.Command object with default parameters
func NewResourceCollectorCommand() *cobra.Command {
	s := options.NewServerRunOptions()
	cmd := &cobra.Command{
		Use:  componentResourceCollector,
		Long: `The resource-collector is site resource collection service.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			utilflag.PrintFlags(cmd.Flags())

			// set up stopCh here
			stopCh := genericapiserver.SetupSignalHandler()

			return Run(stopCh)
		},
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}
	return cmd
}

func Run(stopCh <-chan struct{}) error {
	logger.Infof("Resource Collector Running...")

	// init collector
	collector.InitCollector(stopCh)
	col := collector.GetCollector()
	if col == nil {
		return fmt.Errorf("get new collector failed")
	}

	// start scheduler resource cache informer and run
	col.StartInformersAndRun(stopCh)

	// start the gRPC service to get cluster static info from ClusterController
	go rpc.NewRpcServer()

	// todo  wait until all informer synced
	// start http server to provide resource information to the Scheduler
	err := startHttpServer(stopCh)
	if err != nil {
		return err
	}

	select {
	case <-stopCh:
		break
	}

	return nil
}

// Start snapshot API for resource-collector
func startHttpServer(stopCh <-chan struct{}) error {
	router.Register()
	hs, err := apiserver.NewHTTPServer()
	if err != nil {
		logger.Errorf("new http server failed!, err: %s", err.Error())
		return err
	}

	return hs.BlockingRun(stopCh)
}
