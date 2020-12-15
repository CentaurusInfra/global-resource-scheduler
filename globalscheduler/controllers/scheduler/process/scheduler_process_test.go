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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	clusterfake "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned/fake"
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

	schedulerClient *schedulerfake.Clientset
	clusterClient   *clusterfake.Clientset
	kubeclient      *k8sfake.Clientset
	// Objects to put in the store.
	schedulerLister []*schedulercrdv1.Scheduler
	// Actions expected to happen on the client.
	kubeactions []core.Action
	actions     []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	return f
}

func newScheduler(name string) *schedulercrdv1.Scheduler {
	return &schedulercrdv1.Scheduler{
		TypeMeta: metav1.TypeMeta{APIVersion: schedulercrdv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
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

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	var ret []core.Action
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "schedulers") ||
				action.Matches("watch", "schedulers")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) newProcess() (*SchedulerProcess, schedulerinformers.SharedInformerFactory) {
	f.schedulerClient = schedulerfake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	i := schedulerinformers.NewSharedInformerFactory(f.schedulerClient, noResyncPeriodFunc())

	p := newSchedulerProcess(f.kubeclient, f.schedulerClient, f.clusterClient, i.Globalscheduler().V1().Schedulers())
	p.schedulerSynced = alwaysReady
	p.recorder = &record.FakeRecorder{}
	p.processName = "test-1"

	for _, f := range f.schedulerLister {
		i.Globalscheduler().V1().Schedulers().Informer().GetIndexer().Add(f)
	}
	return p, i
}

func (f *fixture) run(schedulerName string) {
	f.runController(schedulerName, true, false)
}

func (f *fixture) runExpectError(schedulerName string) {
	f.runController(schedulerName, true, true)
}

func (f *fixture) runController(schedulerName string, startInformer bool, expectError bool) {
	p, i := f.newProcess()
	if startInformer {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
	}

	namespace, name, _ := cache.SplitMetaNamespaceKey(schedulerName)
	if !expectError {
		sc, _ := p.schedulerInformer.Schedulers(namespace).Get(name)
		p.schedulerclient.GlobalschedulerV1().Schedulers(namespace).Create(sc)
	}

	err := p.syncHandler(schedulerName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing scheduler: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing scheduler, got nil")
	}

	actions := filterInformerActions(f.schedulerClient.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}
}

func (f *fixture) expectCreateSchedulerAction(sc *schedulercrdv1.Scheduler) {
	action := core.NewCreateAction(schema.GroupVersionResource{Resource: "schedulers", Group: "globalscheduler.com", Version: "v1"}, sc.Namespace, sc)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectUpdateSchedulerAction(sc *schedulercrdv1.Scheduler) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "schedulers", Group: "globalscheduler.com", Version: "v1"}, sc.Namespace, sc)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectErrorAction(sc *schedulercrdv1.Scheduler) {
	f.actions = f.actions[:0]
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

func getKey(scheduler *schedulercrdv1.Scheduler, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(scheduler)
	if err != nil {
		t.Errorf("Unexpected error getting key for scheduler %v: %v", scheduler.Name, err)
		return ""
	}
	return key
}

func TestUpdateScheduler(t *testing.T) {
	f := newFixture(t)
	sc := newScheduler("test-1")

	sc.Spec.Tag = "3"

	f.schedulerLister = append(f.schedulerLister, sc)
	f.objects = append(f.objects, sc)

	f.expectCreateSchedulerAction(sc)
	f.expectUpdateSchedulerAction(sc)

	f.run(getKey(sc, t))
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)
	sc := newScheduler("test-1")

	f.schedulerLister = append(f.schedulerLister, sc)
	f.objects = append(f.objects, sc)

	f.expectCreateSchedulerAction(sc)
	f.expectUpdateSchedulerAction(sc)

	f.run(getKey(sc, t))
}

func TestSchedulerNotMacth(t *testing.T) {
	f := newFixture(t)
	sc := newScheduler("test-2")

	f.schedulerLister = append(f.schedulerLister, sc)
	f.objects = append(f.objects, sc)

	f.expectErrorAction(sc)

	f.runExpectError(getKey(sc, t))
}
