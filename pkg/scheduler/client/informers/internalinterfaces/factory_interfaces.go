package internalinterfaces

import (
	"time"

	"k8s.io/kubernetes/pkg/scheduler/client"
	"k8s.io/kubernetes/pkg/scheduler/client/cache"
)

// NewInformerFunc provide the function template that informer could set as param
type NewInformerFunc func(client.Interface, time.Duration, string, string) cache.SharedInformer

// SharedInformerFactory a small interface to allow for adding an informer without an import cycle
type SharedInformerFactory interface {
	Start(stopCh <-chan struct{})
	InformerFor(name string, key string, newFunc NewInformerFunc) cache.SharedInformer
}
