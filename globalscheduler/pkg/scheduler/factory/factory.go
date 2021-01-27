/*
Copyright 2014 The Kubernetes Authors.
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

package factory

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type podInformer struct {
	informer cache.SharedIndexInformer
}

func (i *podInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *podInformer) Lister() corelisters.PodLister {
	return corelisters.NewPodLister(i.informer.GetIndexer())
}

// NewPodInformer creates a shared index informer that returns only non-terminal pods.
func NewPodInformer(schedulerName string, client clientset.Interface,
	resyncPeriod time.Duration) coreinformers.PodInformer {
	selector := fields.ParseSelectorOrDie(
		"status.phase=" + string(v1.PodAssigned) +
			",status.assignedScheduler.name=" + schedulerName)
	lw := cache.NewListWatchFromClient(client.CoreV1(), string(v1.ResourcePods), metav1.NamespaceAll, selector)
	return &podInformer{
		informer: cache.NewSharedIndexInformer(lw, &v1.Pod{}, resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}),
	}
}
