/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

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

package siteresources

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
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

// InformerSiteResource provides access to a shared informer and lister for Hosts.
type InformerSiteResource interface {
	Informer() cache.SharedInformer
}

type informerSiteResources struct {
	factory internalinterfaces.SharedInformerFactory
	name    string
	key     string
	period  time.Duration
}

// New initial the informerSiteResources
func New(f internalinterfaces.SharedInformerFactory, name string, key string, period time.Duration) InformerSiteResource {
	return &informerSiteResources{factory: f, name: name, key: key, period: period}
}

// NewSiteResourcesInformer constructs a new informer for sites.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSiteResourcesInformer(client client.Interface, reSyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	return cache.NewSharedInformer(
		&cache.Lister{ListFunc: func(options interface{}) ([]interface{}, error) {
			// read site infos init data
			confLocation := utils.GetConfigDirectory()
			confFilePath := filepath.Join(confLocation, "site_resources.json")
			file, err := ioutil.ReadFile(confFilePath)
			if err != nil {
				return nil, fmt.Errorf("read site infos data error:%v", err)
			}

			var siteResources typed.SiteResources
			err = json.Unmarshal(file, &siteResources)
			if err != nil {
				return nil, fmt.Errorf("unmarshal site infos data error:%v", err)
			}

			var interfaceSlice []interface{}
			for _, siteResource := range siteResources.SiteResources {
				interfaceSlice = append(interfaceSlice, siteResource)
			}

			return interfaceSlice, nil
		}}, reSyncPeriod, name, key, nil)
}

func (f *informerSiteResources) defaultInformer(client client.Interface, resyncPeriod time.Duration, name string, key string) cache.SharedInformer {
	if f.period > 0 {
		resyncPeriod = f.period
	}
	return NewSiteResourcesInformer(client, resyncPeriod, name, key)
}

func (f *informerSiteResources) Informer() cache.SharedInformer {
	return f.factory.InformerFor(f.name, f.key, f.defaultInformer)
}
