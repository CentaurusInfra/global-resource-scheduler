package eipavailability

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"k8s.io/kubernetes/pkg/scheduler/client"
	"k8s.io/kubernetes/pkg/scheduler/client/cache"
	"k8s.io/kubernetes/pkg/scheduler/client/informers/internalinterfaces"
	"k8s.io/kubernetes/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/pkg/scheduler/types"
	"k8s.io/kubernetes/pkg/scheduler/utils"
)

// InformerEipAvailability provides access to a shared informer and lister for eip available.
type InformerEipAvailability interface {
	Informer() cache.SharedInformer
}

type informerEipAvailability struct {
	factory internalinterfaces.SharedInformerFactory
	name    string
	key     string
	period  time.Duration
}

// New initial the informerSite
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration) InformerEipAvailability {
	return &informerEipAvailability{factory: f, name: name, key: key, period: period}
}

// NewSiteInformer constructs a new informer for flavor.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewEipInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	return cache.NewSharedInformer(
		&cache.Lister{ListFunc: func(options interface{}) ([]interface{}, error) {
			// read eip init data
			confLocation := utils.GetConfigDirectory()
			confFilePath := filepath.Join(confLocation, "eip_pools.json")
			file, err := ioutil.ReadFile(confFilePath)
			if err != nil {
				return nil, fmt.Errorf("read eip pools data error:%v", err)
			}

			var eipPools typed.EipPools
			err = json.Unmarshal(file, &eipPools)
			if err != nil {
				return nil, fmt.Errorf("unmarshal eip pools data error:%v", err)
			}

			var interfaceSlice []interface{}
			for _, pool := range eipPools.Pools {
				interfaceSlice = append(interfaceSlice, pool)
			}

			return interfaceSlice, nil
		}}, resyncPeriod, name, key, types.ListSiteOpts{})
}

func (f *informerEipAvailability) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewEipInformer(client, resyncPeriod, name, key)
}

func (f *informerEipAvailability) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
