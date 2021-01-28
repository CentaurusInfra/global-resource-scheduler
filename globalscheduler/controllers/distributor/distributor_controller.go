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

package distributor

import (
	"bytes"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/util"
	distributortype "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor"
	distributorclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/clientset/versioned"
	distributorscheme "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/clientset/versioned/scheme"
	distributorinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/informers/externalversions/distributor/v1"
	distributorlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/listers/distributor/v1"
	distributorv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/v1"
	"k8s.io/kubernetes/pkg/controller"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const controllerAgentName = "distributor-controller"

const (
	SuccessSynced         = "Synced"
	ErrResourceExists     = "ErrResourceExists"
	MessageResourceExists = "Resource %q already exists and is not managed by distributor"
	MessageResourceSynced = "Distributors synced successfully"
)

// Controller is the controller implementation for Foo resources
type DistributorController struct {
	configfile             string
	kubeclientset          kubernetes.Interface
	apiextensionsclientset apiextensionsclientset.Interface
	clientset              distributorclientset.Interface
	informer               cache.SharedIndexInformer
	synced                 cache.InformerSynced
	lister                 distributorlisters.DistributorLister
	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	queue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new  controller
func NewController(
	configfile string,
	kubeclientset kubernetes.Interface,
	apiextensionsclientset apiextensionsclientset.Interface,
	clientset distributorclientset.Interface,
	distributorInformer distributorinformers.DistributorInformer) *DistributorController {

	utilruntime.Must(distributorscheme.AddToScheme(scheme.Scheme))
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(3).Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	que := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Distributor")

	distributorController := &DistributorController{
		configfile:             configfile,
		kubeclientset:          kubeclientset,
		clientset:              clientset,
		apiextensionsclientset: apiextensionsclientset,
		lister:                 distributorInformer.Lister(),
		synced:                 distributorInformer.Informer().HasSynced,
		queue:                  que,
		recorder:               recorder,
	}

	distributorInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := controller.KeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", obj, err))
				return
			}
			klog.V(3).Infof("Add a distributor: %s", key)
			que.Add(key)
		},
		DeleteFunc: distributorController.deleteDistributor,
	})

	return distributorController
}

func (c *DistributorController) RunController(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	c.Run(stopCh)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *DistributorController) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()
	if ok := cache.WaitForCacheSync(stopCh, c.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	go wait.Until(c.runWorker, time.Second, stopCh)

	klog.V(2).Info("Started workers")
	<-stopCh
	klog.V(2).Info("Shutting down distributor controller")
	return nil
}

// runWorker executes the loop to process new items added to the queue
func (c *DistributorController) runWorker() {
	// invoke processNextItem to fetch and consume the next change
	// to a watched or listed resource
	for c.processNextItem() {
		klog.V(2).Info("DistributorController.runWorker: processing next item")
	}
}

// processNextItem retrieves each queued item and takes the
// necessary handler action based off of if the item was
// created or deleted
func (c *DistributorController) processNextItem() bool {
	// fetch the next item (blocking) from the queue to process or
	// if a shutdown is requested then return out of this to stop
	// processing
	key, quit := c.queue.Get()
	// stop the worker loop from running as this indicates we
	// have sent a shutdown message that the queue has indicated
	// from the Get method
	if quit {
		return false
	}

	defer c.queue.Done(key)

	// assert the string out of the key (format `namespace/name`)
	keyRaw := key.(string)
	if err := c.syncHandlerAndUpdate(keyRaw); err != nil {
		c.queue.AddRateLimited(key)
		utilruntime.HandleError(fmt.Errorf("Handle %v of key %v failed with %v", "serivce", key, err))
	}
	c.queue.Forget(key)
	klog.Infof("Successfully synced '%s'", key)
	return true
}

func (c *DistributorController) syncHandlerAndUpdate(key string) error {
	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing  %q (%v)", key, time.Since(startTime))
	}()
	namespace, distributorName, err := cache.SplitMetaNamespaceKey(key)
	distributor, err := c.lister.Distributors(namespace).Get(distributorName)
	if err != nil || distributor == nil {
		return err
	}

	err = c.rebalance(namespace)
	if err != nil {
		klog.Fatalf("Failed to rebalance the pod hashkey range in namespace %vs", namespace)
	}
	args := strings.Split(fmt.Sprintf("-config %s -ns %s -n %s", c.configfile, namespace, distributorName), " ")

	//	Format the command
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		klog.Fatalf("Failed to get the path to the process with the err %v", err)
	}

	cmd := exec.Command(path.Join(dir, "distributor_process"), args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	//	Run the command
	go cmd.Run()

	//	Output our results
	klog.V(2).Infof("Running process with the result: %v / %v\n", out.String(), stderr.String())
	if err != nil {
		klog.Warningf("Failed to rebalance %v with the err %v", namespace, err)
	}

	c.recorder.Event(distributor, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)

	return nil
}

func (c *DistributorController) createCustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: distributortype.Name,
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   distributortype.GroupName,
			Version: distributortype.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     distributortype.Plural,
				Singular:   distributortype.Singluar,
				Kind:       distributortype.Kind,
				ShortNames: []string{distributortype.ShortName},
			},
			Validation: &apiextensions.CustomResourceValidation{
				OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextensions.JSONSchemaProps{
						"spec": {
							Type: "object",
							Properties: map[string]apiextensions.JSONSchemaProps{
								"range": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"start": {Type: "integer"},
										"end":   {Type: "integer"},
									},
								},
							},
						},
					},
				},
			},
			AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
				{
					Name:     "range",
					Type:     "string",
					JSONPath: ".spec.range",
				},
			},
		},
	}
}
func (c *DistributorController) EnsureCustomResourceDefinitionCreation() error {
	distributor := c.createCustomResourceDefinition()
	_, err := c.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(distributor)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	operation := func() (bool, error) {
		manifest, err := c.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(distributor.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for _, cond := range manifest.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, nil
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					return false, fmt.Errorf("name %s conflicts", distributor.Name)
				}
			}
		}

		return false, fmt.Errorf("%s not established", distributor.Name)
	}
	err = wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		return operation()
	})

	if err != nil {
		deleteErr := c.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(distributor.Name, nil)
		if deleteErr != nil {
			return deleteErr
		}

		return err
	}

	return nil
}

// Create a distributor for testing purpose
func (c *DistributorController) CreateDistributor(distributor *distributorv1.Distributor) error {
	_, err := c.clientset.GlobalschedulerV1().Distributors(corev1.NamespaceDefault).Create(distributor)

	if err != nil {
		klog.Fatalf("Failed to  create a distributor  %v with error %v", distributor, err)
	}
	klog.Infof("Created a distributor %s", distributor.Name)
	return err
}

func (c *DistributorController) rebalance(namespace string) error {
	distributors, err := c.lister.Distributors(namespace).List(labels.Everything())
	if err == nil {
		size := len(distributors)
		klog.V(3).Infof("Get the %d distributors to balance", size)
		if size < 1 {
			return nil
		}
		// hash function can only get uint32, uint64
		// k8s code base does not deal with uint32 properly
		// uint64 > MaxInt64 will have issue in converter. Need to map to 0 - maxInt64
		ranges := util.EvenlyDivide(len(distributors), math.MaxInt64)
		for i, d := range distributors {
			d.Spec.Range = distributorv1.DistributorRange{Start: ranges[i][0], End: ranges[i][1]}
			updated, err := c.clientset.GlobalschedulerV1().Distributors(namespace).Update(d)
			if err != nil {
				klog.Warningf("Failed to updated the distributor %v with the err %v", d, err)
				return err
			} else {
				klog.V(3).Infof("Updated the distributor %s with %v", updated.GetName(), updated.Spec.Range)
			}
		}
	}
	return err
}

func (c *DistributorController) deleteDistributor(object interface{}) {
	key, err := controller.KeyFunc(object)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", object, err))
		return
	}
	namespace, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get namespace for key %#v: %v", key, err))
		return
	}
	c.rebalance(namespace)
}
