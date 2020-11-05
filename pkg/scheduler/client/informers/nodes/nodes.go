package nodes

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

// InformerNodes provides access to a shared informer and lister for Nodes.
type InformerNodes interface {
	Informer() cache.SharedInformer
}

type informerNodes struct {
	factory internalinterfaces.SharedInformerFactory
	name    string
	key     string
	period  time.Duration
}

// New initial the informerNodes
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration) InformerNodes {
	return &informerNodes{factory: f, name: name, key: key, period: period}
}

// NewNodesInformer constructs a new informer for nodes.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNodesInformer(client client.Interface, reSyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	return cache.NewSharedInformer(
		&cache.Lister{ListFunc: func(options interface{}) ([]interface{}, error) {
			// read node infos init data
			confLocation := utils.GetConfigDirectory()
			confFilePath := filepath.Join(confLocation, "site_nodes.json")
			file, err := ioutil.ReadFile(confFilePath)
			if err != nil {
				return nil, fmt.Errorf("read node infos data error:%v", err)
			}

			var siteNodes typed.SiteNodes
			err = json.Unmarshal(file, &siteNodes)
			if err != nil {
				return nil, fmt.Errorf("unmarshal node infos data error:%v", err)
			}

			var interfaceSlice []interface{}
			for _, siteNode := range siteNodes.SiteNodes {
				interfaceSlice = append(interfaceSlice, siteNode)
			}

			return interfaceSlice, nil
		}}, reSyncPeriod, name, key, nil)
}

func (f *informerNodes) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewNodesInformer(client, resyncPeriod, name, key)
}

func (f *informerNodes) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
