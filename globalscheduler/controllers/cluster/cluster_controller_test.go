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

package cluster

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
	"k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned/fake"
	informers "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	client              *fake.Clientset
	kubeclient          *k8sfake.Clientset
	apiextensionsclient *fakeapiextensionsv1beta1.Clientset
	// Objects to put in the store.
	clusterLister []*clusterv1.Cluster
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

func newCluster(name string) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{Kind: clusterv1.Kind, APIVersion: clusterv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: clusterv1.ClusterSpec{
			IpAddress: "10.0.0.1",
			GeoLocation: clusterv1.GeolocationInfo{
				City:     "Bellevue",
				Province: "Washington",
				Area:     "West",
				Country:  "US",
			},
			Region: clusterv1.RegionInfo{
				Region:           "us-west",
				AvailabilityZone: []string{"us-west-1", "us-west-2"},
			},
			Operator: clusterv1.OperatorInfo{
				Operator: "globalscheduler",
			},
			Flavors: []clusterv1.FlavorInfo{
				{FlavorID: "small", TotalCapacity: 10},
				{FlavorID: "medium", TotalCapacity: 20},
			},
			Storage: []clusterv1.StorageSpec{
				{TypeID: "sata", StorageCapacity: 2000},
				{TypeID: "sas", StorageCapacity: 1000},
				{TypeID: "ssd", StorageCapacity: 3000},
			},
			EipCapacity:   3,
			CPUCapacity:   8,
			MemCapacity:   256,
			ServerPrice:   10,
			HomeScheduler: "scheduler1",
		},
	}
}

func (f *fixture) newController() (*ClusterController, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)
	f.apiextensionsclient = fakeapiextensionsv1beta1.NewSimpleClientset(f.apiextensionobjects...)
	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())
	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())
	c := NewClusterController(f.kubeclient, f.apiextensionsclient, f.client, i.Globalscheduler().V1().Clusters())

	c.clusterSynced = alwaysReady
	c.recorder = &record.FakeRecorder{}

	for _, f := range f.clusterLister {
		i.Globalscheduler().V1().Clusters().Informer().GetIndexer().Add(f)
	}
	c.CreateCRD()
	//c.CreateObject(newCluster("cluster1"))
	return c, i, k8sI
}

func (f *fixture) run(clusterName string, eventType EventType) {
	f.runController(clusterName, true, false, eventType)
}

func (f *fixture) runExpectError(clusterName string, event EventType) {
	f.runController(clusterName, true, true, event)
}

func (f *fixture) runController(clusterName string, startInformers bool, expectError bool, eventType EventType) {
	c, i, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}
	_, name, err := cache.SplitMetaNamespaceKey(clusterName)
	if eventType == EventType_Create {
		cluster := newCluster(name)
		c.CreateObject(cluster)
	}
	if eventType == EventType_Update {
		c.UpdateClusterStatus(clusterName)
	}
	if eventType == EventType_Delete {
		c.DeleteObject(name)
	}
	keyWithEventType := KeyWithEventType{Key: clusterName, EventType: eventType}
	fmt.Println("keyWithEventType=%v", keyWithEventType)
	err = c.syncHandler(keyWithEventType)
	if expectError && err != nil {
		fmt.Println("expected error syncing cluster & got %v", err)
	} else if !expectError && err != nil {
		f.t.Errorf("error syncing cluster: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing cluster, got nil")
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
	fmt.Println("actions = %d, %v", len(actions), actions)
	fmt.Println("f.actions = %d, %v", len(f.actions), f.actions)
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
	fmt.Println("k8sActions = %d, %v", len(k8sActions), k8sActions)
	fmt.Println("f.kubeactions = %d, %v", len(f.kubeactions), f.kubeactions)
	if len(f.kubeactions) > len(k8sActions) {
		f.t.Errorf("%d additional k8s expected actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
	}
	fmt.Println("Done Test = %v", eventType)
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
	case core.DeleteActionImpl:
		e, _ := expected.(core.DeleteActionImpl)
		expName := e.GetName()
		name := a.GetName()

		if expName != name {
			t.Errorf("Action %s %s has wrong cluster\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expName, name))
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
			(action.Matches("list", "clusters") ||
				action.Matches("watch", "clusters")) {
			continue
		}
		ret = append(ret, action)
	}
	return ret
}

func (f *fixture) expectCreateClusterStatusAction(cluster *clusterv1.Cluster) {
	f.actions = f.actions[:0]
	f.actions = append(f.actions, core.NewCreateAction(schema.GroupVersionResource{Resource: "clusters"}, cluster.Namespace, cluster))

}

func (f *fixture) expectUpdateClusterStatusAction(cluster *clusterv1.Cluster) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "clusters"}, cluster.Namespace, cluster)
	f.actions = f.actions[:0]
	f.actions = append(f.actions, action)
	fmt.Println("expectUpdateClusterStatusAction1= %v", action)
	fmt.Println("expectUpdateClusterStatusAction2= %v", f.actions)
}

func (f *fixture) expectDeleteClusterStatusAction(cluster *clusterv1.Cluster) {
	action := core.NewDeleteAction(schema.GroupVersionResource{Resource: "clusters"}, cluster.Namespace, cluster.Name)
	f.actions = f.actions[:0]
	f.actions = append(f.actions, action)
	fmt.Println("expectDeleteClusterStatusAction1= %v", action)
	fmt.Println("expectDeleteClusterStatusAction2= %v", f.actions)
}

func (f *fixture) expectErrorClusterStatusAction(cluster *clusterv1.Cluster) {
	f.actions = f.actions[:0]
	fmt.Println("expectClusterStatusAction2= %v", f.actions)
}

func getKey(cluster *clusterv1.Cluster, t *testing.T) string {
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
	f.objects = append(f.objects, cluster)
	f.expectCreateClusterStatusAction(cluster)
	f.run(getKey(cluster, t), EventType_Create)
}

func TestUpdateCluster(t *testing.T) {
	f := newFixture(t)
	cluster := newCluster("cluster2")

	f.clusterLister = append(f.clusterLister, cluster)
	f.objects = append(f.objects, cluster)
	f.expectUpdateClusterStatusAction(cluster)
	f.run(getKey(cluster, t), EventType_Update)
}

func TestDeleteCluster(t *testing.T) {
	f := newFixture(t)
	cluster := newCluster("cluster2")
	f.clusterLister = append(f.clusterLister, cluster)
	f.objects = append(f.objects, cluster)
	f.expectDeleteClusterStatusAction(cluster)
	f.run(getKey(cluster, t), EventType_Delete)
}

func TestClusterUnknownEvent(t *testing.T) {
	f := newFixture(t)
	cluster := newCluster("cluster3")
	f.clusterLister = append(f.clusterLister, cluster)
	f.objects = append(f.objects, cluster)
	f.expectErrorClusterStatusAction(cluster)
	f.runExpectError(getKey(cluster, t), -1)
}
