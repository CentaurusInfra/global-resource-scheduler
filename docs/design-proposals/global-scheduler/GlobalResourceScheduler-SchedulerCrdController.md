# Global Resource Scheduler – Scheduler CRD

## 1. Description

The scheduler controller will list/watch the API server for scheduler creation/deletion/get/list API calls and then execute the controller logic documented in https://github.com/futurewei-cloud/global-resource-scheduler/blob/master/docs/design-proposals/global-scheduler/GlobalResourceScheduler-ClusterCrdController-ver0.1.md

-   The scheduler runs the scheduling algorithm to select the best cluster for the incoming VM/Container Request.

-   HA will be enabled on each active scheduler via endpoint/configmap object lock based leader election mechanism.

-   Each scheduler is responsible for a group/partition of clusters.

## 2. Requirements

-   [Scheduler CRD creation/deletion --- Create the CRD Controller code framework via code-generator](https://github.com/futurewei-cloud/global-resource-scheduler/issues/31)

-   [Scheduler CRD creation/Deletion---API implementation](https://github.com/futurewei-cloud/global-resource-scheduler/issues/34)

-   [Scheduler CRD creation/deletion --- implement the controller logic](https://github.com/futurewei-cloud/global-resource-scheduler/issues/33)

## 3. Scheduler CRD Yaml Definition

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// crd-scheduler.yaml

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: scheduler.globalscheduler.k8s.io
spec:
  group: globalscheduler.k8s.io
  version: v1
  names:
    kind: Scheduler
    plural: schedulers
  scope: Namespaced
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// example-scheduler.yaml

apiVersion: globalscheduler.k8s.io/v1
kind: Scheduler
metadata:
  name: example-scheduler
spec:
  id: "scheduler_id_1"
  name: "scheduler_1"
  cluster:
    - "cluster_id_1"
    - "cluster_id_2"
    - "cluster_id_3"
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## 4. Data Structure

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
type Scheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SchedulerSpec   `json:"spec"`
	Status SchedulerStatus `json:"status"`
}

type SchedulerSpec struct {
	// Name is the name of the scheduler
	Name string `json:"name"`

	// Location represent which geolocation the scheduler is responsible for
	Location GeolocationInfo `json:"location"`

	// Tag represents the Nth scheduler object
	Tag string `json:"tag"`

	// Cluster is an array that stores the name of clusters
	Cluster [][]*clustercrdv1.Cluster `json:"cluster"`

	// ClusterUnion is the union of all cluster spec
	Union ClusterUnion `json:"union"`
}

type ClusterUnion struct {
	IpAddress     []string                        `json:"ipaddress"`
	GeoLocation   []*clustercrdv1.GeolocationInfo `json:"geolocation"`
	Region        []*clustercrdv1.RegionInfo      `json:"region"`
	Operator      []*clustercrdv1.OperatorInfo    `json:"operator"`
	Flavors       [][]*clustercrdv1.FlavorInfo    `json:"flavors"`
	Storage       [][]*clustercrdv1.StorageSpec   `json:"storage"`
	EipCapacity   []int64                         `json:"eipcapacity"`
	CPUCapacity   []int64                         `json:"cpucapacity"`
	MemCapacity   []int64                         `json:"memcapacity"`
	ServerPrice   []int64                         `json:"serverprice"`
	HomeScheduler []string                        `json:"homescheduler"`
}

type SchedulerStatus struct {
	State string `json:"state"`
}

type SchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Scheduler `json:"items"`
}

type GeolocationInfo struct {
	City     string `json:"city"`
	Province string `json:"province"`
	Area     string `json:"area"`
	Country  string `json:"country"`
}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## 5. APIs Design

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create post an instance of CRD into Kubernetes.
func (sc *SchedulerClient) Create(obj *schedulerv1.Scheduler) (*schedulerv1.Scheduler, error) {
	return sc.clientset.GlobalschedulerV1().Schedulers(sc.namespace).Create(obj)
}

// Update puts new instance of CRD to replace the old one.
func (sc *SchedulerClient) Update(obj *schedulerv1.Scheduler) (*schedulerv1.Scheduler, error) {
	return sc.clientset.GlobalschedulerV1().Schedulers(sc.namespace).Update(obj)
}

// Delete removes the CRD instance by given name and delete options.
func (sc *SchedulerClient) Delete(name string, opts *metav1.DeleteOptions) error {
	return sc.clientset.GlobalschedulerV1().Schedulers(sc.namespace).Delete(name, opts)
}

// Get returns a pointer to the CRD instance.
func (sc *SchedulerClient) Get(name string, opts metav1.GetOptions) (*schedulerv1.Scheduler, error) {
	return sc.clientset.GlobalschedulerV1().Schedulers(sc.namespace).Get(name, opts)
}

// List returns a list of CRD instances by given list options.
func (sc *SchedulerClient) List(opts metav1.ListOptions) (*schedulerv1.SchedulerList, error) {
	return sc.clientset.GlobalschedulerV1().Schedulers(sc.namespace).List(opts)
}

// Get the number of how many schedulers have been created
func (sc *SchedulerClient) SchedulerNums() (int, error) {
	schedulerList, err := sc.List(metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	length := len(schedulerList.Items)

	return length, nil
}

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~