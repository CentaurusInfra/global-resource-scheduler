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

package process

import (
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/controllers/util/union"
	clusterclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned"
	schedulerclientset "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned"
	schedulerscheme "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned/scheme"
	schedulerinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions"
	"k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions/scheduler/v1"
	schedulerlisters "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/listers/scheduler/v1"
	schedulercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// process config
var (
	flagSet         = flag.NewFlagSet("scheduler_process", flag.ExitOnError)
	master          = flag.String("process", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	processconfig   = flag.String("processconfig", "/var/run/kubernetes/controller.kubeconfig", "Path to a kubeconfig. Only required if out-of-cluster.")
	shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT}
)

const (
	defaultNamespace      = "default"
	controllerProcessName = "scheduler-process"
	SuccessSynced         = "Synced"
	MessageResourceSynced = "Scheduler synced successfully"
)

var closeProcessName string

type SchedulerProcess struct {
	kubeclientset   kubernetes.Interface
	schedulerclient schedulerclientset.Interface
	clusterClient   clusterclientset.Interface
	processName     string

	schedulerInformer schedulerlisters.SchedulerLister
	schedulerSynced   cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func (sp *SchedulerProcess) Run(threadiness int, stopCh <-chan struct{}, name string) error {
	defer runtime.HandleCrash()
	defer sp.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting scheduler process loop")
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, sp.schedulerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	sp.processName = name

	klog.Infof("Starting scheduler %s process", sp.processName)
	for i := 0; i < threadiness; i++ {
		go wait.Until(sp.runWorker, time.Second, stopCh)
	}

	klog.Infof("Started scheduler %s process", sp.processName)
	<-stopCh
	klog.Infof("Shutting down %s process", sp.processName)

	return nil
}

func (sp *SchedulerProcess) enqueueScheduler(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	sp.workqueue.AddRateLimited(key)
}

func (sp *SchedulerProcess) runWorker() {
	for sp.processNextWorkItem() {
	}
}

func (sp *SchedulerProcess) processNextWorkItem() bool {
	obj, shutdown := sp.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer sp.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			sp.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Scheduler resource to be synced.
		if err := sp.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		sp.workqueue.Forget(obj)
		klog.Infof("[ Scheduler Process %s ]Successfully synced '%s'", sp.processName, key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

func (sp *SchedulerProcess) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Scheduler resource with this namespace/name
	scheduler, err := sp.schedulerInformer.Schedulers(namespace).Get(name)
	if err != nil {
		// The Scheduler resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			klog.Warningf("DemoCRD: %s/%s does not exist in local cache, will delete it from Scheduler ...",
				namespace, name)

			klog.Infof("[SchedulerCRD] Deleting scheduler: %s/%s ...", namespace, name)

			return nil
		}

		runtime.HandleError(fmt.Errorf("failed to list scheduler by: %s/%s", namespace, name))
		return err
	}

	if scheduler.Name != sp.processName {
		return fmt.Errorf("process name doesn't match the scheduler name")
	} else {
		if scheduler.Status == schedulercrdv1.SchedulerDelete {
			closeProcessName = sp.processName
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return nil
		} else {
			schedulerCopy := scheduler.DeepCopy()
			// Update Union
			newUnion := schedulercrdv1.ClusterUnion{}
			for _, v := range schedulerCopy.Spec.Cluster {
				clusterObj, err := sp.clusterClient.GlobalschedulerV1().Clusters(namespace).Get(v, metav1.GetOptions{})
				if err != nil {
					klog.Infof("Error getting cluster object")
					return err
				}
				newUnion = union.UpdateUnion(newUnion, clusterObj)
			}
			schedulerCopy.Spec.Union = newUnion
			_, err = sp.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Update(schedulerCopy)
			if err != nil {
				klog.Infof("Fail to update scheduler object")
				return err
			}
		}
	}

	//sp.recorder.Event(mydemo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (sp *SchedulerProcess) enqueueSchedulerForDelete(obj interface{}) {
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	sp.workqueue.AddRateLimited(key)
}

func setupSignalHandler(name string) (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		if closeProcessName == name {
			close(stop)
		}
		<-c
		os.Exit(1)
	}()

	return stop
}

func StartSchedulerProcess(name string) {
	flag.Parse()

	stopCh := setupSignalHandler(name)

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *processconfig)
	if err != nil {
		klog.Fatalf("Error building processconfig: %s", err.Error())
	}

	// Create kubeClientset
	kubeClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// Create schedulerClientset
	schedulerClientset, err := schedulerclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building scheduler clientset: %s", err.Error())
	}

	// Create clusterClientset
	clusterClientset, err := clusterclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building clusterclientset: %s", err.Error())
	}

	schedulerInformerFactory := schedulerinformers.NewSharedInformerFactory(schedulerClientset, time.Second*30)
	schedulerInformer := schedulerInformerFactory.Globalscheduler().V1().Schedulers()

	schedulerProcess := newSchedulerProcess(kubeClientset, schedulerClientset, clusterClientset, schedulerInformer)

	go schedulerInformerFactory.Start(stopCh)

	if err = schedulerProcess.Run(1, stopCh, name); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func newSchedulerProcess(kubeclientset kubernetes.Interface, schedulerClient schedulerclientset.Interface, clusterClient clusterclientset.Interface, schedulerInformer v1.SchedulerInformer) *SchedulerProcess {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(schedulerscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerProcessName})

	process := &SchedulerProcess{
		kubeclientset:     kubeclientset,
		schedulerclient:   schedulerClient,
		clusterClient:     clusterClient,
		schedulerInformer: schedulerInformer.Lister(),
		schedulerSynced:   schedulerInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Scheduler"),
		recorder:          recorder,
	}

	klog.Info("Setting up scheduler event handlers")
	schedulerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			oldScheduler := old.(*schedulercrdv1.Scheduler)
			newScheduler := new.(*schedulercrdv1.Scheduler)
			if oldScheduler.ResourceVersion == newScheduler.ResourceVersion {
				return
			} else if !clusterEqual(oldScheduler.Spec.Cluster, newScheduler.Spec.Cluster) {
				process.enqueueScheduler(new)
			}
		},
		DeleteFunc: process.enqueueSchedulerForDelete,
	})

	return process
}

func clusterEqual(oldClusters []string, newClusters []string) bool {
	if len(oldClusters) != len(newClusters) {
		return false
	}
	resMap := make(map[string]bool)
	for _, v := range oldClusters {
		resMap[v] = true
	}

	for _, v := range newClusters {
		if _, exist := resMap[v]; !exist {
			return false
		}
	}

	return true
}
