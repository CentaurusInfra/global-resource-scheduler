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
	"fmt"
	"reflect"
	"testing"
	"time"

	fakeapiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/clientset/versioned/fake"
	informers "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/client/informers/externalversions"
	distributorv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor/v1"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	client              *fake.Clientset
	kubeclient          *k8sfake.Clientset
	apiextensionsClient *fakeapiextensionsv1beta1.Clientset
	// Objects to put in the store.
	distributorLister []*distributorv1.Distributor
	// Actions expected to happen on the client.
	kubeactions []core.Action
	actions     []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects         []runtime.Object
	objects             []runtime.Object
	apiextensionobjects []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	f.apiextensionobjects = []runtime.Object{}
	return f
}

func newDistributor(name string) *distributorv1.Distributor {
	return &distributorv1.Distributor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: distributorv1.DistributorSpec{
			Range: distributorv1.DistributorRange{Start: 0, End: 10},
		},
	}
}

func (f *fixture) newController() (*DistributorController, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	f.apiextensionsClient = fakeapiextensionsv1beta1.NewSimpleClientset(f.apiextensionobjects...)
	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())
	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())
	c := NewController("", f.kubeclient, f.apiextensionsClient, f.client, i.Globalscheduler().V1().Distributors(), "")
	c.synced = alwaysReady
	c.recorder = &record.FakeRecorder{}

	for _, f := range f.distributorLister {
		i.Globalscheduler().V1().Distributors().Informer().GetIndexer().Add(f)
	}
	c.EnsureCustomResourceDefinitionCreation()
	return c, i, k8sI
}

func (f *fixture) run(key string) {
	f.runController(key, true, false)
}

func (f *fixture) runExpectError(key string) {
	f.runController(key, true, true)
}

func (f *fixture) runController(key string, startInformers bool, expectError bool) {
	c, i, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Errorf("distributor name is not correct - %v", key)
		return
	}

	distributor := newDistributor(name)
	c.CreateDistributor(distributor)

	err = c.syncHandlerAndUpdate(key)
	if expectError && err != nil {
		fmt.Printf("expected error syncing distributor & got %v", err)
	} else if !expectError && err != nil {
		f.t.Errorf("error syncing distributor: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing distributor, got nil")
	}

	actions := filterInformerActions(f.client.Actions())
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
	k8sActions := filterInformerActions(f.kubeclient.Actions())
	for i, action := range k8sActions {
		if len(f.kubeactions) < i+1 {
			f.t.Errorf("%d unexpected k8s actions: %+v", len(k8sActions)-len(f.kubeactions), k8sActions[i:])
			break
		}
		expectedAction := f.kubeactions[i]
		checkAction(expectedAction, action, f.t)
	}
	if len(f.kubeactions) > len(k8sActions) {
		f.t.Errorf("%d additional k8s expected actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
	}
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
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

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "distributors") ||
				action.Matches("update", "distributors") ||
				action.Matches("watch", "distributors")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectCreateDistributorStatusAction(distributor *distributorv1.Distributor) {
	action := core.NewCreateAction(schema.GroupVersionResource{Resource: "distributors"}, distributor.Namespace, distributor)
	f.actions = f.actions[:0]
	f.actions = append(f.actions, action)
	fmt.Printf("expect actions: %v, %v\n", action, f.actions)
}

func (f *fixture) expectUpdateDistributorStatusAction(distributor *distributorv1.Distributor) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "distributors"}, distributor.Namespace, distributor)
	f.actions = f.actions[:0]
	f.actions = append(f.actions, action)
	fmt.Printf("expect actions: %v, %v\n", action, f.actions)
}

func (f *fixture) expectDeleteDistributorStatusAction(distributor *distributorv1.Distributor) {
	action := core.NewDeleteAction(schema.GroupVersionResource{Resource: "distributors"}, distributor.Namespace, distributor.Name)
	f.actions = f.actions[:0]
	f.actions = append(f.actions, action)
	fmt.Printf("expect actions: %v, %v\n", action, f.actions)
}

func (f *fixture) expectErrorDistributorStatusAction(distributor *distributorv1.Distributor) {
	f.actions = f.actions[:0]
	fmt.Printf("expectDistributorStatusAction2= %v\n", f.actions)
}

func getKey(distributor *distributorv1.Distributor, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(distributor)
	if err != nil {
		t.Errorf("Unexpected error getting key for distributor %v: %v", distributor.Name, err)
		return ""
	}
	return key
}

func TestCreateDistributor(t *testing.T) {
	f := newFixture(t)
	distributor := newDistributor("d1")
	f.distributorLister = append(f.distributorLister, distributor)
	f.objects = append(f.objects, distributor)
	f.expectCreateDistributorStatusAction(distributor)
	f.run(getKey(distributor, t))
}
