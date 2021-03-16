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

package dispatcher

import (
	"bytes"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/util"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1"
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	dispatcherclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned"
	dispatcherscheme "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/clientset/versioned/scheme"
	dispatcherinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/client/informers/externalversions/dispatcher/v1"
	dispatchercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/dispatcher/v1"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/runtime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type DispatcherController struct {
	kubeconfigfile string
	// kubeclientset is a standard kubernetes clientset
	kubeclientset          kubernetes.Interface
	apiextensionsclientset apiextensionsclientset.Interface
	dispatcherclient       dispatcherclientset.Interface
	clusterclient          clusterclientset.Interface

	dispatcherInformer cache.SharedIndexInformer
	clusterInformer    cache.SharedIndexInformer
	// mutex for updating dispatchers
	mu sync.Mutex
	// stop channel
	stopCh <-chan struct{}
	// cluster mapping
	clusters []string
	// dispatcher list
	dispatchers []*dispatchercrdv1.Dispatcher
}

// NewDispatcherController returns a new dispatcher controller
func NewDispatcherController(
	kubeconfigfile string,
	kubeclientset kubernetes.Interface,
	apiextensionsclientset apiextensionsclientset.Interface,
	dispatcherclient dispatcherclientset.Interface,
	clusterclient clusterclientset.Interface,
	dispatcherInformer dispatcherinformers.DispatcherInformer,
	clusterInformer clusterinformers.ClusterInformer) *DispatcherController {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(dispatcherscheme.AddToScheme(scheme.Scheme))

	controller := &DispatcherController{
		kubeconfigfile:         kubeconfigfile,
		kubeclientset:          kubeclientset,
		apiextensionsclientset: apiextensionsclientset,
		dispatcherclient:       dispatcherclient,
		clusterclient:          clusterclient,
		dispatcherInformer:     dispatcherInformer.Informer(),
		clusterInformer:        clusterInformer.Informer(),
		clusters:               make([]string, 0),
		dispatchers:            make([]*dispatchercrdv1.Dispatcher, 0),
	}

	return controller
}

func (dc *DispatcherController) addDispatcher(obj interface{}) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if dispatcher, ok := obj.(*dispatchercrdv1.Dispatcher); ok {
		dc.dispatchers = append(dc.dispatchers, dispatcher)
		if err := dc.balance(); err != nil {
			klog.Fatalf("Failed to balance the clusters among dispatchers with error %v", err)
		} else {
			go func() {
				args := strings.Split(fmt.Sprintf("-config %s -ns %s -n %s", dc.kubeconfigfile, dispatcher.Namespace, dispatcher.Name), " ")

				//	Format the command
				dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
				if err != nil {
					klog.Fatalf("Failed to get the path to the process with the err %v", err)
				} else {
					cmd := exec.Command(path.Join(dir, "dispatcher_process"), args...)
					var out bytes.Buffer
					var stderr bytes.Buffer
					cmd.Stdout = &out
					cmd.Stderr = &stderr

					//	Run the command
					go cmd.Run()

					//	Output our results
					klog.V(2).Infof("Running process with the result: %v / %v\n", out.String(), stderr.String())

					if err != nil {
						klog.Warningf("Failed to run dispatcher process %v - %v with the err %v", dispatcher.Namespace, dispatcher.Name, err)
					}
				}
			}()
		}
	} else {
		klog.Warningf("Failed to convert the added object %v to a dispatcher", obj)
	}
}

func (dc *DispatcherController) deleteDispatcher(obj interface{}) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if dispatcher, ok := obj.(*dispatchercrdv1.Dispatcher); ok {
		dispLen := len(dc.dispatchers)
		for idx, disp := range dc.dispatchers {
			if disp.Namespace == dispatcher.Namespace && disp.Name == dispatcher.Name {
				dc.dispatchers[idx] = dc.dispatchers[dispLen-1]
				dc.dispatchers = dc.dispatchers[0 : dispLen-1]
				break
			}
		}
		if err := dc.balance(); err != nil {
			klog.Fatalf("Failed to balance the clusters among dispatchers with error %v", err)
		}
	}
}

func (dc *DispatcherController) addCluster(obj interface{}) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if cluster, ok := obj.(*clustercrdv1.Cluster); ok {
		dc.clusters = util.InsertIntoSortedArray(dc.clusters, cluster.Name)
		if err := dc.balance(); err != nil {
			klog.Fatalf("Failed to balance the clusters among dispatchers with error %v", err)
		}
	} else {
		klog.Warningf("Failed to convert the added object %v to a cluster", obj)
	}
}

func (dc *DispatcherController) deleteCluster(obj interface{}) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if cluster, ok := obj.(*clustercrdv1.Cluster); ok {
		dc.clusters = util.RemoveFromSortedArray(dc.clusters, cluster.Name)
		if err := dc.balance(); err != nil {
			klog.Fatalf("Failed to balance the clusters among dispatchers with error %v", err)
		}
	} else {
		klog.Warningf("Failed to convert the deleted object %v to a cluster", obj)
	}
}

func (dc *DispatcherController) RunController(workers int, stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer runtime.HandleCrash()
	klog.Info("Setting up dispatcher event handlers")
	dc.dispatcherInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dc.addDispatcher,
		DeleteFunc: dc.deleteDispatcher,
	})

	klog.Info("Setting up cluster event handlers")
	dc.clusterInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dc.addCluster,
		DeleteFunc: dc.deleteCluster,
	})

	go dc.dispatcherInformer.Run(stopCh)
	go dc.clusterInformer.Run(stopCh)

	<-stopCh
}

func (dc *DispatcherController) balance() error {
	if len(dc.dispatchers) == 0 {
		return nil
	}

	if len(dc.clusters) > 0 {
		ranges := util.EvenlyDivide(len(dc.dispatchers), int64(len(dc.clusters)-1))
		for idx, dispatcher := range dc.dispatchers {
			if idx < len(ranges) {
				dispatcher.Spec.ClusterRange = dispatchercrdv1.DispatcherRange{Start: dc.clusters[ranges[idx][0]], End: dc.clusters[ranges[idx][1]]}
			} else {
				dispatcher.Spec.ClusterRange = dispatchercrdv1.DispatcherRange{}
			}

			if updatedDispatcher, err := dc.dispatcherclient.GlobalschedulerV1().Dispatchers(corev1.NamespaceDefault).Update(dispatcher); err != nil {
				return fmt.Errorf("Failed to update dispatcher with the error error %v", err)
			} else {
				dc.dispatchers[idx] = updatedDispatcher
				klog.V(3).Infof("The dispatcher %s has updated to the new cluster range %v", dispatcher.GetName(), dispatcher.Spec.ClusterRange)
			}
		}
	}
	return nil
}
