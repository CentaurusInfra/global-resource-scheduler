package volumepool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/internalinterfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

// InformerVolumePool provides access to a shared informer and lister for eip available.
type InformerVolumePool interface {
	Informer() cache.SharedInformer
}

type informerVolumePool struct {
	factory internalinterfaces.SharedInformerFactory
	name    string
	key     string
	period  time.Duration
}

// New initial the informerSite
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration) InformerVolumePool {
	return &informerVolumePool{factory: f, name: name, key: key, period: period}
}

// NewVolumePoolInformer constructs a new informer for volumepool.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewVolumePoolInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	return cache.NewSharedInformer(
		&cache.Lister{ListFunc: func(options interface{}) ([]interface{}, error) {
			// read volume type init data
			confLocation := utils.GetConfigDirectory()
			confFilePath := filepath.Join(confLocation, "volume_pools.json")
			file, err := ioutil.ReadFile(confFilePath)
			if err != nil {
				return nil, fmt.Errorf("read volume pool data error:%v", err)
			}

			var volumePools typed.RegionVolumePools
			err = json.Unmarshal(file, &volumePools)
			if err != nil {
				return nil, fmt.Errorf("unmarshal volume pool data error:%v", err)
			}
			var interfaceSlice []interface{}
			for _, volumePools := range volumePools.VolumePools{
				interfaceSlice = append(interfaceSlice, volumePools)
			}

			return interfaceSlice, nil
		}}, resyncPeriod, name, key, types.ListSiteOpts{})
}

func (f *informerVolumePool) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewVolumePoolInformer(client, resyncPeriod, name, key)
}

func (f *informerVolumePool) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
