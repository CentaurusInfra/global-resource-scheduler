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

package options

import (
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// ServerRunOptions contains required server running options
type ServerRunOptions struct {
	KubeConfig string

	// Scheduler Name Differentiating Schedulers
	SchedulerName string

	// Resource Collector API URL
	ResourceCollectorApiUrl string
}

// NewServerRunOptions constructs a new ServerRunOptions if existed.
func NewServerRunOptions() *ServerRunOptions {
	return &ServerRunOptions{}
}

// Flags returns flags for a specific GSScheduler by section name
func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("gs-scheduler")
	fs.StringVar(&s.KubeConfig, "kubeconfig", s.KubeConfig, "Path to kubeconfig files, specifying how to connect to  API servers. Providing --kubeconfig enables API server mode, omitting --kubeconfig enables standalone mode.")
	fs.StringVar(&s.SchedulerName, "schedulername", s.SchedulerName, "Scheduler name of the gs-scheduler, specifying the name of the scheduler.")
	fs.StringVar(&s.ResourceCollectorApiUrl, "resourcecollectorapiurl", s.ResourceCollectorApiUrl, "Api url of the resource collector, specifying the api of the resource collector.")
	return fss
}

func (s *ServerRunOptions) Config() *types.GSSchedulerConfiguration {
	config := NewGSSchedulerConfiguration()
	config.SchedulerName = s.SchedulerName
	config.ResourceCollectorApiUrl = s.ResourceCollectorApiUrl
	if config.ResourceCollectorApiUrl == "" {
		config.ResourceCollectorApiUrl = constants.DefaultResourceCollectorAPIURL
	}

	return config
}

// NewGSSchedulerConfiguration will create a new KubeletConfiguration with default values
func NewGSSchedulerConfiguration() *types.GSSchedulerConfiguration {
	config := &types.GSSchedulerConfiguration{
		SchedulerName: "DefaultScheduler",
	}
	return config
}
