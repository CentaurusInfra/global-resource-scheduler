package flavor

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
	"k8s.io/kubernetes/pkg/scheduler/utils"
)

// InformerFlavor provides access to a shared informer and lister for Flavors.
type InformerFlavor interface {
	Informer() cache.SharedInformer
}

type informerFlavor struct {
	factory internalinterfaces.SharedInformerFactory
	name    string
	key     string
	period  time.Duration
}

// New initial the informerFlavor
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration) InformerFlavor {
	return &informerFlavor{factory: f, name: name, key: key, period: period}
}

// NewFlavorInformer constructs a new informer for flavor.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFlavorInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	return cache.NewSharedInformer(
		&cache.Lister{ListFunc: func(options interface{}) ([]interface{}, error) {
			// read flavor init data
			confLocation := utils.GetConfigDirectory()
			confFilePath := filepath.Join(confLocation, "flavors.json")
			file, err := ioutil.ReadFile(confFilePath)
			if err != nil {
				return nil, fmt.Errorf("read flavor data error:%v", err)
			}

			var flavors typed.RegionFlavors
			err = json.Unmarshal(file, &flavors)
			if err != nil {
				return nil, fmt.Errorf("unmarshal flavor data error:%v", err)
			}

			var interfaceSlice []interface{}
			for _, flavor := range flavors.Flavors {
				interfaceSlice = append(interfaceSlice, flavor)
			}

			return interfaceSlice, nil
		}}, resyncPeriod, name, key, typed.ListOpts{})
}

func (f *informerFlavor) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewFlavorInformer(client, resyncPeriod, name, key)
}

func (f *informerFlavor) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
