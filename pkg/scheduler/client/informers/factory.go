package informers

import (
	"sync"
	"time"

	"k8s.io/kubernetes/pkg/scheduler/client"
	"k8s.io/kubernetes/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/pkg/scheduler/client/informers/eipavailability"
	"k8s.io/kubernetes/pkg/scheduler/client/informers/flavor"
	"k8s.io/kubernetes/pkg/scheduler/client/informers/internalinterfaces"
	"k8s.io/kubernetes/pkg/scheduler/client/informers/nodes"
	"k8s.io/kubernetes/pkg/scheduler/client/informers/sites"
	"k8s.io/kubernetes/pkg/scheduler/client/informers/volumepool"
	"k8s.io/kubernetes/pkg/scheduler/client/informers/volumetype"
	"k8s.io/kubernetes/pkg/scheduler/client/typed"
)

var (
	// InformerFac is the global param and produce the informers
	InformerFac SharedInformerFactory
)

const (
	// FLAVOR is the symbol of informer flavor
	FLAVOR = "flavor"

	// NODES is the symbol of informer nodes
	NODES = "nodes"

	// VOLUMETYPE is the symbol of informer volumeType
	VOLUMETYPE = "volumeType"

	// Sites is the symbol of informer site
	SITES = "sites"

	// EIPPOOLS eip pool
	EIPPOOLS = "eipools"

	// VOLUMEPOOLS volume pool
	VOLUMEPOOLS = "volumepools"
)

// SharedInformerFactory is used as a factory to produce informers
type SharedInformerFactory interface {
	internalinterfaces.SharedInformerFactory
	GetInformer(name string) cache.SharedInformer

	GetFlavor(flavorID string, region string) (typed.Flavor, bool)
	Flavor(name, key string, period time.Duration) flavor.InformerFlavor
	Sites(name, key string, period time.Duration) sites.InformerSite
	Nodes(name, key string, period time.Duration) nodes.InformerNodes
	VolumeType(name, key string, period time.Duration) volumetype.InformerVolumeType
	EipPools(name, key string, period time.Duration) eipavailability.InformerEipAvailability
	VolumePools(name, key string, period time.Duration) volumepool.InformerVolumePool
}

type sharedInformerFactory struct {
	client        client.Interface
	lock          sync.Mutex
	defaultResync time.Duration

	informers map[string]cache.SharedInformer
	// startedInformers is used for tracking which informers have been started.
	// This allows Start() to be called multiple times safely.
	startedInformers map[string]bool
}

// NewSharedInformerFactory constructs a new instance of sharedInformerFactory
func NewSharedInformerFactory(client client.Interface, defaultResync time.Duration) SharedInformerFactory {
	return NewFilteredSharedInformerFactory(client, defaultResync)
}

// NewFilteredSharedInformerFactory constructs a new instance of sharedInformerFactory.
// Listers obtained via this SharedInformerFactory will be subject to the same filters
// as specified here.
func NewFilteredSharedInformerFactory(client client.Interface, defaultResync time.Duration) SharedInformerFactory {
	return &sharedInformerFactory{
		client:           client,
		defaultResync:    defaultResync,
		informers:        make(map[string]cache.SharedInformer),
		startedInformers: make(map[string]bool),
	}
}

// Start initializes all requested informers.
func (f *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for informerType, informer := range f.informers {
		if !f.startedInformers[informerType] {
			go informer.Run(stopCh)
			f.startedInformers[informerType] = true
		}
	}
}

// InternalInformerFor returns the SharedIndexInformer for obj using an internal
// client.
func (f *sharedInformerFactory) GetInformer(name string) cache.SharedInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informer, exists := f.informers[name]
	if !exists {
		return cache.NewEmptyInformer()
	}

	if !informer.HasSynced() {
		informer.SyncOnce()
	}
	return informer

}

// InternalInformerFor returns the SharedIndexInformer for obj using an internal
// client.
func (f *sharedInformerFactory) InformerFor(name string, key string, newFunc internalinterfaces.NewInformerFunc) cache.SharedInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informer, exists := f.informers[name]
	if exists {
		return informer
	}
	informer = newFunc(f.client, f.defaultResync, name, key)
	f.informers[name] = informer

	return informer
}

//GetFlavor get flavor
func (f *sharedInformerFactory) GetFlavor(flavorID string, region string) (typed.Flavor, bool) {
	if region != "" {
		flvInter, exist := f.GetInformer(FLAVOR).GetStore().Get(region + "|" + flavorID)
		var ret = typed.Flavor{}
		if exist {
			ret = flvInter.(typed.RegionFlavor).Flavor
		}
		return ret, exist
	} else {
		flvInters := f.GetInformer(FLAVOR).GetStore().List()
		for _, flvInter := range flvInters {
			ret := flvInter.(typed.RegionFlavor).Flavor
			if ret.ID == flavorID {
				return ret, true
			}
		}
	}

	return typed.Flavor{}, false
}

//Flavor new Flavor informer
func (f *sharedInformerFactory) Flavor(name, key string, period time.Duration) flavor.InformerFlavor {
	return flavor.New(f, name, key, period)
}

//Nodes new nodes informer
func (f *sharedInformerFactory) Nodes(name, key string, period time.Duration) nodes.InformerNodes {
	return nodes.New(f, name, key, period)
}

//VolumeType new volume type informer
func (f *sharedInformerFactory) VolumeType(name, key string, period time.Duration) volumetype.InformerVolumeType {
	return volumetype.New(f, name, key, period)
}

//Sites new site informer
func (f *sharedInformerFactory) Sites(name, key string, period time.Duration) sites.InformerSite {
	return sites.New(f, name, key, period)
}

//EipPools new eip pool informer
func (f *sharedInformerFactory) EipPools(name, key string, period time.Duration) eipavailability.InformerEipAvailability {
	return eipavailability.New(f, name, key, period)
}

//VolumePools new volume pool informer
func (f *sharedInformerFactory) VolumePools(name, key string, period time.Duration) volumepool.InformerVolumePool {
	return volumepool.New(f, name, key, period)
}
