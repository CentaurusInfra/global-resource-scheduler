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
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"

	distributorconfig "k8s.io/kubernetes/globalscheduler/cmd/distributor/app/config"

	schedulerclient "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	"k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions"
	"time"
)

// Options has all the params needed to run a Scheduler
type Options struct {
	// ConfigFile is the location of the scheduler server's configuration file.
	ConfigFile string

	Master string
}

// NewOptions returns default scheduler app options.
func NewOptions() (*Options, error) {
	o := &Options{}
	return o, nil
}

// Flags returns flags for a specific scheduler by section name
func (o *Options) Flags() (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet("misc")
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "The path to the configuration file. Flags override values in this file.")
	fs.StringVar(&o.Master, "master", o.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")

	return nfs
}

// ApplyTo applies the scheduler options to the given scheduler app configuration.
func (o *Options) ApplyTo(c *distributorconfig.Config) error {
	cfg, err := clientcmd.BuildConfigFromFlags(o.Master, o.ConfigFile)
	if err != nil {
		klog.Fatalf("error getting client onfig: %s", err.Error())
	}
	schedulerclient, err := schedulerclient.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("error building scheduler client: %s", err.Error())
	}
	schedulerInformerFactory := externalversions.NewSharedInformerFactory(schedulerclient, 10*time.Minute)
	c.SchedulerInformer = schedulerInformerFactory.Globalscheduler().V1().Schedulers()

	podclient, err := kubernetes.NewForConfig(cfg)

	c.Client = podclient

	podInformerFactory := informers.NewSharedInformerFactory(podclient, 10*time.Minute)
	c.PodInformer = podInformerFactory.Core().V1().Pods()

	return nil
}

// Validate validates all the required options.
func (o *Options) Validate() []error {
	var errs []error

	return errs
}

// Config return a distributor  config object
func (o *Options) Config() (*distributorconfig.Config, error) {

	c := &distributorconfig.Config{}
	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}
	return c, nil
}
