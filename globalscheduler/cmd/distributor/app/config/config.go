package config

import (
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	v1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions/scheduler/v1"
)

// Config has all the context to run a distributor
type Config struct {
	Client            clientset.Interface
	InformerFactory   informers.SharedInformerFactory
	PodInformer       coreinformers.PodInformer
	SchedulerInformer v1.SchedulerInformer
}

type completedConfig struct {
	*Config
}

// CompletedConfig same as Config, just to swap private object.
type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() CompletedConfig {
	cc := completedConfig{c}
	return CompletedConfig{&cc}
}
