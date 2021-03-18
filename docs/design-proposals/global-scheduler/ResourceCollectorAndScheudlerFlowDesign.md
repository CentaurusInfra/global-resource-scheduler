Mar-8-2021, Authored by Cathy Hong Zhang, Jun Shao, Eunju Kim

This document describes the coordinataion among the four components: Openstack Cluster, Cluster Controller, Resource Collector, and Scheduler as well as the cluster resource collection flow design. 

In current code base, when the scheduler gets a POD request, it will send a HTTP request to the Resource Collector to get the the cluster resource information and the Resource Collector may loop through a linked list of all the siteinfocaches. This whole operation introduces a long latency to the POD creation path. 
To reduce the POD creation latency, instead of a pull model, we will use a push model in our new code design. That is, the Resource Collector should proactively send to the Scheduler the cluster resource snapshot of those clusters when their resources change. Then in the POD scheduling flow path, the Scheduler only needs to query its local cache for all the clusters resource. 

A cluster information consists of two parts. One part is static info. The other is dynamic info. The scheduling platform gets the cluster's static info through cluster CRD creation/registration. The scheduler should list/watch for the cluster's crd object creation/update/delete in ETCD, and should save the cluster's static info into its local cache. 

The scheduler should get each cluster's dynamic resource info, such as available cpu, memory, disk storage, eip, etc. from the resource collector. 

The following data structures show the cluster's dynamic info fields whose values will be dynamically obtained from the Openstack by the resource collector and then sent to the scheduler.

|  Cluster Resource Dynamic Info Fields |  Description                                          | 
|:-------------------------------------:|:-----------------------------------------------------:|
|  Available CPUAndMemory               | currently available CPUAndMemory of a cluster         |
|  Available DiskStorage                | currently available disk space of a cluster           |
|  Available EIPs                       | currently available EIP count of a cluster            |

When the resource collector process first starts, it will query the Openstack for the total allocable CPU, Memory, DiskStorage, and EIP information and send these info to the scheduler. 

TotalAllocableCPU/cluster= sum of (Total empty host's CPU capacity * the host's CPU allocation ratio)
TotalAllocableMemory/cluster= sum of (Total empty host's memory capacity * the host's memory allocation ratio)
TotalAllocableStorage/cluster= sum of (Total empty host's storage capacity * the host's storage allocation ratio)
TotalAllocableEIP/cluster= sum of (Total empty host's eip capacity * the host's eip allocation ratio)

Note that the resource collector only needs to send this info to the scheduler once. 

## Coordination Between Cluster CRD Spec and the region-AZ configuration in Openstack Clusters
When a cluster CRD is registered with the scheduling platform, the cluster controller will send the static cluster information, such as the cluster's region and AZ, to the resource collector. The resource collector should save this information into its local cache. Later when the resource collector queries Openstack for a cluster's dynamic info, it should use the cluster's region and AZ from its local cache as the unique ideintifier of the cluster. Since the resource collector can get each cluster's region and AZ from its local cache, it does not need to send another query to Openstack for cluster's region and AZ info. 

Care should be taken to make sure that each cluster's region and AZ in the cluster CRD matches those configured on the corresponding Openstack cluster. 

## Resource Collection Flow from Openstack to the ResourceCollector
The resource collector queries Openstack for all clusters' dynamic info in per-region batch request every "query interval". The "query interval" should be configurable with default value being 60 seconds. It should use each cluster's region and AZ from its local cache as the unique identifier of that cluster. 

The following requests are pulled every 60 seconds per region

This API will get all the hosts' resource info for a region
hypervisorsPages, err := hypervisors.List(client).AllPages()

This API will extract all the hosts' resource info for the region
hs, err := hypervisors.ExtractHypervisors(hypervisorsPages)

This is the host data structure
type Host struct {
    HostID       string
	UsedVCPUs    int
    UsedMem      int
	TotalVCPUs   int
	TotalMem     int
    ...
}

This API will get all the aggegates' AZ-Host mapping info for a region
aggregatePages, err := aggregates.List(client).AllPages()

This API will extract all the AZ-Host mapping info for the region
aggregates, err:= aggregates.ExtractAggregates(aggregatePages)

This is the aggregate data structure
type Aggregate struct {
       AvailabilityZone string
       Hosts []string //HostID array
}

This API will get all the volume info for the region
allPages, err := schedulerstats.List(client, listOpts).AllPages()

This API will extract all the volume info for the region
allStats, err := schedulerstats.ExtractStoragePools(allPages)
This is the shared volume data structure
type VolumeCapabilities struct {
	ProvisionedCapacityGb    float64
	FreeCapacityGb           float64
	TotalCapacityGb          float64
	VolumeType               string
    ...
}

The following information are only pulled once

This API will get all flavor type info for the region
flasPages, err := flavors.ListDetail(client, flavors.ListOpts{}).AllPages()

This API will extract all the flavor type info for the region
flas, err := flavors.ExtractFlavors(flasPages)

This is the flavor type data structure
type FlavorType struct {
	ID string
	Name string
	Vcpus string
	Ram int64
}

This API will get all volume type info for the region
allPages, err := volumetypes.List(client, volumetypes.ListOpts{}).AllPages()

This API will extract all the volume type info for the region
volumeTypes, err := volumetypes.ExtractVolumeTypes(allPages)

This is the volume type data structure
type VolumeType struct {
       ID string
       Name string
       VolumeBackendName string
}

## Cluster-level Resource Info Construction and Caching Flow Inside the ResourceCollector
After the resource collector gets the response from the batch query, it will construct each cluster's available cpu, memory, storage info and save them into its local cache. Since the available resource info in the Openstack response is per host, the resource collector has to construct per cluster resource info by adding all the hosts' resource info associated with a cluster. In the future we will incoporate an external middleware component that will construct per-AZ granular resource info in the response. Therefore, the resource collector does not need to loop through all the hosts to construct per-AZ resource info, which will improve the resource collector's performance. 

Then the resource collector should send all clusters' updated resource info in batch message to the scheduler. 

Each time the resource collector gets a new response from the Openstack, the resource collector should compare the latest resource info of a cluster with what was saved in its local cache. If there is a change, it will include that cluster in its batch resource update message to the scheduler. If there is no change, it should not include that cluster in the batch update message. 

The following are the cluster resource info data structure

type RegionResource struct {
         RegionName string
         CPUAndMemoryResources map[string]CPUAndMemoryResource //map key is Az
         VolumeReource VolumeReource 
}   
 
type CPUAndMemoryResource struct {
	UsedVCPUs    int
    UsedMem      int
	TotalVCPUs   int
	TotalMem     int
    ResourceType string //hardcoded value is "default"
}

type VolumeReource struct {
	ProvisionedCapacityGb    float64
	FreeCapacityGb           float64
	TotalCapacityGb          float64
	VolumeType               string
    MaxOverSubscriptionRatio float64 //hardcoded value is 1
}

## Cluster-level Resource Info Flow from the ResourceCollector to the Scheduler
The following are the http API and message structure. 

### Message structure
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
type SiteResourcesSpec struct {
        	SiteResources	[]SiteResourceSpec
}

type SiteResourceSpec struct {
        	SiteID: string  // region+AZ 
        	CPUCapacity   	int64   
    	    MemCapacity   	int64   
        	Storage 	    []StorageSpec 
}
type StorageSpec struct {
    	TypeID          	string   // (sata, sas, ssd)
    	StorageCapacity 	float64
}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

### HTTP Rest API 
| **Group**   | **API Name**            | **Method** | **Request**                               |
|-------------|-------------------------|------------|-------------------------------------------|
|SiteResource | updateSiteResource      | PATCH      | /globalscheduler/v1/siteresources/update  |

(1) Update SiteResources 
-   Method: PATCH
 
-   Request Path: /globalscheduler/v1/regionresources/{RegionName}
 
-   Request Parameter:
 
-   Response:
 
    -   Normal response codes: 200
 
    -   Error response codes: 204(No Content) 404(Not Found). Please refer to https://tools.ietf.org/html/rfc5789#section-2.2 for more status codes
 
-   Example
 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Request: 
    http://127.0.0.1:8665/globalscheduler/v1/regionresources/{RegionName}
Body: 
    {
        "CPUMemResources": [
            {
                "AvailabilityZone": "West-az1",
                "CpuCapacity": 800, 
                "MemCapacity": 128,   
            },
            {
                "AvailabilityZone": "West-az2",
                "CpuCapacity": 600, 
                "MemCapacity": 256,   
            },
            ...
        ],
        "VolumeResources": [
            {
                "TypeId": "sata",
                "StorageCapacity": 1000.00, 
            },
            {
                "TypeId": "sas",
                "StorageCapacity": 2000.00, 
            },
            {
                "TypeId": "ssd",
                "StorageCapacity": 3000.00, 
            },
            ...
        ], 

 

    } 
 
Response:
{
    HTTP/1.1 200 OK
}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## Cluster-level Resource Info Caching Flow Inside the Scheduler
The scheduler should decode the http/gRPC message to get updated clusters' resource information. Note that the http/gRPC message only includes new available resource info for those clusters whose resource has changed from last sync point. It will reconstruct per cluster's flavor count information based on each cluster's latest available cpu/memory. Then it will update its local cache with updated cluster dynamic info. 

The scheduler should use a separate thread to handle the interaction with the resource collector. That is, the the thread to handle the interaction with the resoruce collector should be different from the thread of handling the pod request. 

The following are the API flow and data structures

### Scheduler APIs 
| **Group** | **API Name**            | **Method** | **Request**                                      |
|-----------|-------------------------|------------|--------------------------------------------------|
| Cluster   | registerCluster         | POST       | /globalscheduler/v1/clusters                     |
|           | unregisterClusterById   | DELETE     | /globalscheduler/v1/clusters/id/{cluster_id}     |
|           | unregisterClusterByName | DELETE     | /globalscheduler/v1/clusters/name/{cluster_name} |
|           | listCluster             | GET        | /globalscheduler/v1/clusters                     |
|           | getClusterById          | GET        | /globalscheduler/v1/clusters/id/{cluster_id}     |
|           | getClusterByName        | GET        | /globalscheduler/v1/clusters/name/{cluster_name} |
|-----------|-------------------------|------------|--------------------------------------------------|
| Pod       | addPod                  | POST       | /globalscheduler/v1/pods                         |
|           | deletePodById           | DELETE     | /globalscheduler/v1/pods/id/{pod_id}             |
|           | listPods                | GET        | /globalscheduler/v1/pods/name/{pod_name}         |
|           | getPodById              | GET        | /globalscheduler/v1/pods                         |

### Scheduler Data Structure
(1) Resource’s data structures
•	SiteCacheInfoSnapshot – SiteCacheInfo of all sites
•	SiteCacheInfo – one site’s cache info such as available resources and allocation status
•	SiteTree –This is a map (not a tree structure in spite of the name) of zone (key) sites of a zone

![image.png](/images/SiteCacheInfo.png)

## POD Scheduling Flow Inside the Scheduler 
The POD scheduling flow should be as follows. Each step may need to query the scheduler's local cache for each cluster's resource info. 
The scheduler should not query the resource collector during this process.
1. run pre-filter plugins
2. run filter plugins
3. run pre-score plugins
4. run score plugins, the scoring algorithm and plugins will be adjusted according to scheduling strategy. 
5. run bind plugin
6. update ETCD with the scheduling result for the POD. 
