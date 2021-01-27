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

package scheduler

import (
	fakeapiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/globalscheduler/controllers/util/consistenthashing"
	clusterfake "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned/fake"
	clusterinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions"
	clustercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	schedulerfake "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/clientset/versioned/fake"
	schedulerinformers "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/client/informers/externalversions"
	schedulercrdv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/scheduler/v1"
	"reflect"
	"testing"
	"time"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	schedulerClient     *schedulerfake.Clientset
	clusterClient       *clusterfake.Clientset
	kubeclient          *k8sfake.Clientset
	apiextensionsclient *fakeapiextensionsv1beta1.Clientset
	// Objects to put in the store.
	schedulerLister []*schedulercrdv1.Scheduler
	clusterLister   []*clustercrdv1.Cluster
	// Actions expected to happen on the client.
	kubeactions      []core.Action
	clusterActions   []core.Action
	schedulerActions []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects         []runtime.Object
	schedulerobjects    []runtime.Object
	clusterobjects      []runtime.Object
	apiextensionobjects []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.schedulerobjects = []runtime.Object{}
	f.clusterobjects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	return f
}

func newScheduler(name string) *schedulercrdv1.Scheduler {
	return &schedulercrdv1.Scheduler{
		TypeMeta: metav1.TypeMeta{APIVersion: schedulercrdv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
			Tenant:    "system",
		},
		Spec: schedulercrdv1.SchedulerSpec{
			Location: schedulercrdv1.GeolocationInfo{
				City:    "Bellevue",
				Area:    "WA",
				Country: "USA",
			},
			Tag: "1",
		},
	}
}

func newCluster(name string) *clustercrdv1.Cluster {
	return &clustercrdv1.Cluster{
		TypeMeta: metav1.TypeMeta{APIVersion: schedulercrdv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
			Tenant:    "system",
		},
		Spec: clustercrdv1.ClusterSpec{
			IpAddress:     "127.0.0.1",
			HomeScheduler: "scheduler1",
		},
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	var ret []core.Action
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "schedulers") ||
				action.Matches("watch", "schedulers") ||
				action.Matches("list", "clusters") ||
				action.Matches("watch", "clusters")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) newController() (*SchedulerController, schedulerinformers.SharedInformerFactory, clusterinformers.SharedInformerFactory) {
	f.schedulerClient = schedulerfake.NewSimpleClientset(f.schedulerobjects...)
	f.clusterClient = clusterfake.NewSimpleClientset(f.clusterobjects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	f.apiextensionsclient = fakeapiextensionsv1beta1.NewSimpleClientset(f.apiextensionobjects...)
	si := schedulerinformers.NewSharedInformerFactory(f.schedulerClient, noResyncPeriodFunc())
	ci := clusterinformers.NewSharedInformerFactory(f.clusterClient, noResyncPeriodFunc())

	p := NewSchedulerController(f.kubeclient, f.apiextensionsclient, f.schedulerClient, f.clusterClient, si.Globalscheduler().V1().Schedulers(), ci.Globalscheduler().V1().Clusters())
	p.schedulerSynced = alwaysReady
	p.recorder = &record.FakeRecorder{}

	for _, f := range f.schedulerLister {
		si.Globalscheduler().V1().Schedulers().Informer().GetIndexer().Add(f)
	}

	for _, f := range f.clusterLister {
		ci.Globalscheduler().V1().Clusters().Informer().GetIndexer().Add(f)
	}
	return p, si, ci
}

func (f *fixture) run(fooName string, eventType EventType) {
	f.runController(fooName, true, false, eventType)
}

func (f *fixture) runExpectError(fooName string, eventType EventType) {
	f.runController(fooName, true, true, eventType)
}

func (f *fixture) runController(fooName string, startInformer bool, expectError bool, eventType EventType) {
	p, si, ci := f.newController()
	p.consistentHash = consistenthashing.New()
	input := []string{"scheduler1"}
	p.consistentHash.Add(input)
	if startInformer {
		stopCh := make(chan struct{})
		defer close(stopCh)
		si.Start(stopCh)
		ci.Start(stopCh)
	}

	namespace, name, _ := cache.SplitMetaNamespaceKey(fooName)
	if !expectError {
		if eventType == EventTypeAddCluster {
			cl, _ := p.clusterInformer.Clusters(namespace).Get(name)
			ns := newScheduler(cl.Spec.HomeScheduler)
			p.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Create(ns)
			p.clusterclient.GlobalschedulerV1().Clusters(namespace).Create(cl)
		} else if eventType == EventTypeUpdateCluster {
			cl, _ := p.clusterInformer.Clusters(namespace).Get(name)
			ns := newScheduler(cl.Spec.HomeScheduler)
			ns.Spec.Cluster = []string{"scheduler1"}
			p.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Create(ns)
		}
	}

	keyWithEventType := &KeyWithEventType{Value: fooName, EventType: eventType}

	err := p.syncHandler(keyWithEventType)
	if !expectError && err != nil {
		f.t.Errorf("error syncing scheduler: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing scheduler, got nil")
	}

	if eventType == EventTypeAddCluster {
		clusterActions := filterInformersActions(f.clusterClient.Actions())
		for i, action := range clusterActions {
			if len(f.clusterActions) < i+1 {
				f.t.Errorf("%d unexpected actions: %+v", len(clusterActions)-len(f.clusterActions), clusterActions[i:])
				break
			}
			expectedAction := f.clusterActions[i]
			checkAction(expectedAction, action, f.t)
		}
		if len(f.clusterActions) > len(clusterActions) {
			f.t.Errorf("%d additional expected actions:%+v", len(f.clusterActions)-len(clusterActions), f.clusterActions[len(clusterActions):])
		}
	} else if eventType == EventTypeUpdateScheduler {
		schedulerActions := filterInformersActions(f.schedulerClient.Actions())
		for i, action := range schedulerActions {
			if len(f.schedulerActions) < i+1 {
				f.t.Errorf("%d unexpected actions: %+v", len(schedulerActions)-len(f.schedulerActions), schedulerActions[i:])
				break
			}
			expectedAction := f.schedulerActions[i]
			checkAction(expectedAction, action, f.t)
		}
		if len(f.schedulerActions) > len(schedulerActions) {
			f.t.Errorf("%d additional expected actions:%+v", len(f.schedulerActions)-len(schedulerActions), f.schedulerActions[len(schedulerActions):])
		}
	}
}

func checkAction(expected core.Action, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.GetActionImpl:
		e, _ := expected.(core.GetActionImpl)
		expObject := e.GetResource()
		object := a.GetResource()
		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.DeleteActionImpl:
		e, _ := expected.(core.DeleteActionImpl)
		expObject := e.GetResource()
		object := a.GetResource()
		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformersActions(actions []core.Action) []core.Action {
	var ret []core.Action
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "schedulers") ||
				action.Matches("watch", "schedulers") ||
				action.Matches("list", "clusters") ||
				action.Matches("watch", "clusters")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectCreateClusterAction(cluster *clustercrdv1.Cluster) {
	action := core.NewCreateAction(schema.GroupVersionResource{Resource: "clusters", Group: "globalscheduler.com", Version: "v1"}, cluster.Namespace, cluster)
	f.clusterActions = append(f.clusterActions, action)
}

func (f *fixture) expectUpdateClusterAction(cluster *clustercrdv1.Cluster) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "clusters", Group: "globalscheduler.com", Version: "v1"}, cluster.Namespace, cluster)
	f.clusterActions = append(f.clusterActions, action)
}

func (f *fixture) expectGetClusterAction(cluster *clustercrdv1.Cluster) {
	action := core.NewGetAction(schema.GroupVersionResource{Resource: "clusters", Group: "globalscheduler.com", Version: "v1"}, cluster.Namespace, cluster.Name)
	f.clusterActions = append(f.clusterActions, action)
}

func (f *fixture) expectCreateSchedulerAction(scheduler *schedulercrdv1.Scheduler) {
	action := core.NewCreateAction(schema.GroupVersionResource{Resource: "schedulers", Group: "globalscheduler.com", Version: "v1"}, scheduler.Namespace, scheduler)
	f.schedulerActions = append(f.schedulerActions, action)
}

func (f *fixture) expectGetSchedulerAction(scheduler *schedulercrdv1.Scheduler) {
	action := core.NewGetAction(schema.GroupVersionResource{Resource: "schedulers", Group: "globalscheduler.com", Version: "v1"}, scheduler.Namespace, scheduler.Name)
	f.schedulerActions = append(f.schedulerActions, action)
}

func (f *fixture) expectUpdateSchedulerAction(scheduler *schedulercrdv1.Scheduler) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "schedulers", Group: "globalscheduler.com", Version: "v1"}, scheduler.Namespace, scheduler)
	f.schedulerActions = append(f.schedulerActions, action)
}

func (f *fixture) expectDeleteSchedulerAction(scheduler *schedulercrdv1.Scheduler) {
	action := core.NewDeleteAction(schema.GroupVersionResource{Resource: "schedulers", Group: "globalscheduler.com", Version: "v1"}, scheduler.Namespace, scheduler.Name)
	f.schedulerActions = append(f.schedulerActions, action)
}

func getClusterKey(cluster *clustercrdv1.Cluster, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cluster)
	if err != nil {
		t.Errorf("Unexpected error getting key for cluster %v: %v", cluster.Name, err)
		return ""
	}
	return key
}

func getSchedulerKey(cluster *schedulercrdv1.Scheduler, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cluster)
	if err != nil {
		t.Errorf("Unexpected error getting key for cluster %v: %v", cluster.Name, err)
		return ""
	}
	return key
}

func TestCreatesCluster(t *testing.T) {
	f := newFixture(t)
	cluster := newCluster("cluster2")
	f.clusterLister = append(f.clusterLister, cluster)
	f.clusterobjects = append(f.clusterobjects, cluster)
	f.expectCreateClusterAction(cluster)
	f.run(getClusterKey(cluster, t), EventTypeAddCluster)
}

func TestDeleteCluster(t *testing.T) {
	f := newFixture(t)
	cluster := newCluster("cluster2")
	cluster.Status = "Delete"
	f.clusterLister = append(f.clusterLister, cluster)
	f.clusterobjects = append(f.clusterobjects, cluster)
	f.expectGetClusterAction(cluster)
	f.expectUpdateClusterAction(cluster)
	f.run(getClusterKey(cluster, t), EventTypeUpdateCluster)
}
