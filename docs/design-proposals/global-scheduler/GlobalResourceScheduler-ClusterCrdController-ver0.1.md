# Global Resource Scheduler – Cluster CRD

Oct-20-2020, Hong Zhang, Eunju Kim

## 1. Module Description

This module allows an admin user to register, unregister, list, get an existing cluster
with the global resource scheduling platform..

-   Cluster consists of a group of nodes and nodes are host machines for
    hosting/running VM/Container PODs.

-   The cluster, whether it is an openstack cluster or a kubernetes cluster,
    should have been deployed before this cluster object is created.
    But in order for the global scheduling platform to know that such a cluster exists or
    is removed, the cluster needs to be registered or unregistered with the platform.

-   This can be done through this cluster CRD object registration/unregistration API.
    Note that this cluster registration/unregistration API does not actually deploy
    a cluster or undeploy a cluster. It just registers and unregisters a cluster
    object with the global scheduling platform. When a cluster CRD object is created/deleted,
    the cluster CRD object controller should execute the logic listed in section 2.3.

-   The cluster CRD object should be created/deleted by an admin user, not an end user.
    Permission must be set properly when implementing the API.

-   The cluster object holds cluster static information, such as the cluster's geolocation
    and resource profile attributes. These information will be used in the partition code
    flow and could also be used in the scheduling algorithm
    

## 2. Requirements

-   In milestone1, The global-scheduler can interface with different types of
    clusters (openstack cluster, kubernetes cluster, etc.) through a generic
    southbound API. They share one single cluster resoruce data strcuture.

-   The cluster resource controller does not create a cluster. it just registers
    existing clusters which are already deployed through other approaches. The
    deployment of clusters themselves is out of the scope of this project.

![](/images/global-scheduler-cluster-diagram.png)

[Picture1] Flow of registration of a cluster

**2.1** [API
Implementation](https://github.com/futurewei-cloud/global-resource-scheduler/issues/23)

(1) CLI: command line APIs like kubectl

-   Register cluster, Unregister cluster, List cluster, Get cluster

(2) REST APIs: REST WEB APIs

-   Register cluster, Unregister cluster, List cluster, Get cluster

**2.2** [Create the CRD Controller code framework via
code-generator](https://github.com/futurewei-cloud/global-resource-scheduler/issues/25)

-   Global resource scheduler defines a cluster CRD definition.

**2.3** [Implement the controller
logic](https://github.com/futurewei-cloud/global-resource-scheduler/issues/27)
(1) cluster creation: 

-   List/watch the cluster object creation through Informer.

-   Run consistent hashing algorithm to select the home scheduler for this cluster.

-   Save the cluster-scheduler binding to ETCD. That is, update the selected scheduler object to add the binding of this cluster to the scheduler.  

-   Add the cluster's geolocation and resource profile information to the selected scheduler object. 

(2) cluster deletion:

-   List/watch the cluster object deletion through Informer.

-   Run consistent hashing algorithm to retrieve the home scheduler for this cluster.

-   Remove the cluster-scheduler binding in ETCD. That is, update the selected scheduler object to remove the binding of this cluster to the scheduler.   

-   Remove the cluster's geolocation and resource profile information from the selected scheduler object. Since a scheduler could be tagged with multiple geolocations and many resoruce attributes, care must be taken not to remove a geolocation or resource attribute that is shared by another cluster bound to that scheduler object.

(3) scheduler process creation: 

-   List/watch the completion status of the scheduler object, which means the scheduler process creation is completed, through Informer. Note that the scheduler controller should list/watch the "scheduler CRD object creation" and run the scheduler process and then set the scheduler object status to completion. 

-   Update the consistent hashing ring range assignment to assign a range of clusters to this new scheduler.

-   reset the binding of this range of clusters, whose hash keys fall into this new scheduler's range, to this new scheduler object, i.e. this new scheduler becomes the new home scheduler for this range of clusters. 

-   Add the geolocation and resource profile attributes of this range of clusters to the new scheduler object in ETCD

-   Update the geolocation and resource profile attributes of the other scheduler which was previously bound to this range of clusters.


(4) Scheduler CRD deletion: 

-   List/watch the "scheduler CRD deletion" through Informer. It is this scheduler-cluster CRD controller, not the scheduler controller, that list/watch the "scheduler CRD deletion"

-   Update the consistent hashing ring range assignment to reassign this scheduler's range of clusters to the other schedulers. 

-   Update the binding of the range of clusters to the other schedulers accordingly. Add the geolocation and resource profile attributes of the range of clusters to their new Home scheduler objects in ETCD. Once this is done, the distributor will not forward any more POD requests to this scheduler. 

-   Remove the scheduler object from ETCD.  

-   Note that the scheduler controller should list/watch the "scheduler object removal from ETCD". WHen this happens, the scheduler controller should wait until all POD requests in its scheduling queue is scheduled and then delete the scheduler process. 


## 3. Cluster CRD Definition & Data Structure

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
type Cluster struct {  
    apiversion:         string          //v1
    kind:               string          //Cluster
    Name:               string          
    Spec                ClusterSpec     // Spec is the custom resource spec 
} 
// MyResourceSpec is the spec for a MyResource resource. 
//This is where you would put your custom resource data
type ClusterSpec struct { 
    ipAdrress           string 
    GeoLocation         GeolocationInfo 
    Region              RegionInfo 
    Operator            OperatorInfo 
    flavors             []FlavorInfo 
    storage             []StorageSpec 
    EipCapacity         int64 
    CPUCapacity         int64 
    MemCapacity         int64 
    ServerPrice         int64 
    HomeScheduler       string 
} 
type FlavorInfo struct { 
    FlavorID            string 
    TotalCapacity       int64 
} 
type StorageSpec struct { 
    TypeID              string      //(sata, sas, ssd) 
    StorageCapacity     int64 
} 
type GeolocationInfo struct { 
    city                string 
    province            string 
    area                string 
    country             string 
} 
type RegionInfo { 
    region              string 
    AvailabilityZone    string 
} 
type OperatorInfo { 
    operator            string 
}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## 4. APIs Design

**4.1 CLI APIs**

-   register cluster -filename FILENAME

    `Example: register -filename ./cluster1.yaml`

-   unregister cluster -name CLUSTERNAME

    `Example: unregister -name cluster1`

-   unregister cluster -id CLUSTERID

    `Example: unregister -id 3dda2801-d675-4688-a63f-dcda8d327f51`

-   list clusters

    `Example: list clusters`

-   get cluster -name CLUSTERNAME

    `Example: get cluster -name cluster1`

-   get cluster -id CLUSTERID

    `Example: get cluster -id 3dda2801-d675-4688-a63f-dcda8d327f51`

**4.2 REST APIs & Error Codes Design**

| **Group** | **API Name**            | **Method** | **Request**                                      |
|-----------|-------------------------|------------|--------------------------------------------------|
| Cluster   | registerCluster         | POST       | /globalscheduler/v1/clusters                     |
|           | unregisterClusterById   | DELETE     | /globalscheduler/v1/clusters/id/{cluster_id}     |
|           | unregisterClusterByName | DELETE     | /globalscheduler/v1/clusters/name/{cluster_name} |
|           | listCluster             | GET        | /globalscheduler/v1/clusters                     |
|           | getClusterById          | GET        | /globalscheduler/v1/clusters/id/{cluster_id}     |
|           | getClusterByName        | GET        | /globalscheduler/v1/clusters/name/{cluster_name} |

(1) Register Cluster

-   Method: POST

-   Request: /globalscheduler/v1/clusters

-   Request Parameter:

-   Response: cluster profile

    -   Normal response codes: 201

    -   Error response codes: 400, 409, 412, 500, 503

-   Example

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Request: 
    http://127.0.0.1:8080/globalscheduler/v1/clusters 
Body: 
    {   
        "cluster_profile": { 
            "cluster_name": "cluster1", 
            “cluster_spec”: { 
                “ipAdrress”: “10.0.0.3”, 
                “GeoLocation”: { “city”: “Bellevue”, “province”: “Washington”, “area”: “West”, “country”: “US” }, 
                “Region”: { “region”: “us-west”, “AvailabilityZone”: “us-west-1” }, 
                “Operator”: { “operator”: “globalscheduler”, }, 
                “flavors”: [ {“FlavorID”: “small”, “TotalCapacity”: 5}, { “FlavorID”: “medium”, “TotalCapacity”: 10}, { “FlavorID”: “large”, “TotalCapacity”: 20}, { “FlavorID”: “xlarge”,“TotalCapacity”: 10}, { “FlavorID”: “2xlarge”, “TotalCapacity”: 5 }], 
                “storage”: [ {“TypeID”: “sata”, “StorageCapacity”: 2000}, { “TypeID”: “sas”, “StorageCapacity”: 1000}, { “TypeID”: “ssd”, “StorageCapacity”: 3000}, }], 
                “EipCapacity”: 3, 
                “CPUCapacity”: 8, 
                “MemCapacity”: 256, 
                “ServerPrice”: 10, 
                “HomeScheduler”: “scheduler1” 
            } 
        } 
    } 

Response: 
    { "cluster_profile": { 
        "cluster_id": "3dda2801-d675-4688-a63f-dcda8d327f51", 
        "cluster_name": "cluster1", 
        “cluster_spec”: { 
            “ipAdrress”: “10.0.0.3”, 
            “GeoLocation”: { “city”: “Bellevue”, “province”: “Washington”, “area”: “West”, “country”: “US” }, 
            “Region”: { “region”: “us-west”, “AvailabilityZone”: “us-west-1” }, 
            “Operator”: { “operator”: “globalscheduler”, }, 
            “flavors”: [ {“FlavorID”: “small”, “TotalCapacity”: 5}, { “FlavorID”: “medium”, “TotalCapacity”: 10}, { “FlavorID”: “large”, “TotalCapacity”: 20}, { “FlavorID”: “xlarge”,“TotalCapacity”: 10}, { “FlavorID”: “2xlarge”, “TotalCapacity”: 5 }], 
            “storage”: [ {“TypeID”: “sata”, “StorageCapacity”: 2000}, { “TypeID”: “sas”, “StorageCapacity”: 1000}, { “TypeID”: “ssd”, “StorageCapacity”: 3000}, }], 
            “EipCapacity”: 3, 
            “CPUCapacity”: 8, 
            “MemCapacity”: 256, 
            “ServerPrice”: 10, 
            “HomeScheduler”: “scheduler1” 
            } 
        } 
    } 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

(2) Unregister Cluster By Id

-   Method: DELETE

-   Request: /globalscheduler/v1/clusters/id/{cluster_id}

-   Request Parameter: \@PathVariable String cluster_id

-   Response: cluster_id

    -   Normal response codes: 200

    -   Error response codes: 400, 412, 500

-   Example

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Request: 
    http://127.0.0.1:8080/globalscheduler/v1/clusters/3dda2801-d675-4688-a63f-dcda8d327f51

Response: 
    deleted: 3dda2801-d675-4688-a63f-dcda8d327f50 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

(3) Unregister Cluster By Name

-   Method: DELETE

-   Request: /globalscheduler/v1/clusters/name/{cluster_name}

-   Request Parameter: \@PathVariable String cluster_name

-   Response: cluster_name

    -   Normal response codes: 200

    -   Error response codes: 400, 412, 500

-   Example

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Request: 
    http://127.0.0.1:8080/globalscheduler/v1/clutsers/name/cluster1

Response: 
    Deleted: cluster1
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

(4) List Clusters

-   Method: GET

-   Request: v1/clusters

-   Request Parameter:

-   Response: clusters list

    -   Normal response codes: 200

    -   Error response codes: 400, 409, 412, 500, 503

-   Example

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Request: 
    http://127.0.0.1:8080/globalscheduler/v1/clusters

Response: 
    { 
        [ 
            { 
                "cluster_id": "3dda2801-d675-4688-a63f-dcda8d327f51", 
                "cluster_name": "cluster1", “cluster_spec”: {…} 
            }, 
            { "cluster_id": "3dda2801-d675-4688-a63f-dcda8d327f52", 
                "cluster_name": "cluster2", “cluster_spec”: {…} 
            },
            …
        ] 
    } 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

(5) Get ClusterById

-   Method: GET

-   Request: v1/clusters/id/{cluster_id}

-   Request Parameter: \@PathVariable String cluster_id

-   Response: cluster profile

    -   Normal response codes: 200

    -   Error response codes: 400, 412, 500

-   Example

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Request: 
    http://127.0.0.1:8080/globalscheduler/v1/clusters/3dda2801-d675-4688-a63f-dcda8d327f50 

Response: 
    { "cluster_profile": { 
        "cluster_id": "3dda2801-d675-4688-a63f-dcda8d327f51", 
        "cluster_name": "cluster1", 
        “cluster_spec”: { 
            “ipAdrress”: “10.0.0.3”, 
            “GeoLocation”: { “city”: “Bellevue”, “province”: “Washington”, “area”: “West”, “country”: “US” }, 
            “Region”: { “region”: “us-west”, “AvailabilityZone”: “us-west-1” }, 
            “Operator”: { “operator”: “globalscheduler”, }, 
            “flavors”: [ {“FlavorID”: “small”, “TotalCapacity”: 5}, { “FlavorID”: “medium”, “TotalCapacity”: 10}, { “FlavorID”: “large”, “TotalCapacity”: 20}, { “FlavorID”: “xlarge”,“TotalCapacity”: 10}, { “FlavorID”: “2xlarge”, “TotalCapacity”: 5 }], 
            “storage”: [ {“TypeID”: “sata”, “StorageCapacity”: 2000}, { “TypeID”: “sas”, “StorageCapacity”: 1000}, { “TypeID”: “ssd”, “StorageCapacity”: 3000}, }], 
            “EipCapacity”: 3, 
            “CPUCapacity”: 8, 
            “MemCapacity”: 256, 
            “ServerPrice”: 10, 
            “HomeScheduler”: “scheduler1” 
            } 
        } 
    } 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

(6) Get Cluster By Name

-   Method: GET

-   Request: /clusters/name/{cluster_name}

-   Request Parameter: \@PathVariable String cluster_name

-   Response: cluster profile

    -   Normal response codes: 200

    -   Error response codes: 400, 412, 500

-   Example

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Request: 
    http://127.0.0.1:8080/globalscheduler/v1/clusters/cluster1 
Response: 
    { "cluster_profile": { 
        "cluster_id": "3dda2801-d675-4688-a63f-dcda8d327f51", 
        "cluster_name": "cluster1", 
        “cluster_spec”: { 
            “ipAdrress”: “10.0.0.3”, 
            “GeoLocation”: { “city”: “Bellevue”, “province”: “Washington”, “area”: “West”, “country”: “US” }, 
            “Region”: { “region”: “us-west”, “AvailabilityZone”: “us-west-1” }, 
            “Operator”: { “operator”: “globalscheduler”, }, 
            “flavors”: [ {“FlavorID”: “small”, “TotalCapacity”: 5}, { “FlavorID”: “medium”, “TotalCapacity”: 10}, { “FlavorID”: “large”, “TotalCapacity”: 20}, { “FlavorID”: “xlarge”,“TotalCapacity”: 10}, { “FlavorID”: “2xlarge”, “TotalCapacity”: 5 }], 
            “storage”: [ {“TypeID”: “sata”, “StorageCapacity”: 2000}, { “TypeID”: “sas”, “StorageCapacity”: 1000}, { “TypeID”: “ssd”, “StorageCapacity”: 3000}, }], 
            “EipCapacity”: 3, 
            “CPUCapacity”: 8, 
            “MemCapacity”: 256, 
            “ServerPrice”: 10, 
            “HomeScheduler”: “scheduler1” 
            } 
        } 
    } 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## References

[1] Kubernetes Cluster, https://kubernetes.io/docs/tutorials/kubernetes

[2] Openstack Cluster,
https://docs.openstack.org/senlin/latest/user/clusters.html\#creating-a-cluster-basics/create-cluster/cluster-intro
