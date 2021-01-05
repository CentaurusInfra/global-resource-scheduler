# gRPC Communication between ClusterController and ResourceCollector 

Nov-20-2020, Cathy Hong Zhang, Eunju Kim, Created
Jan-5-2021, Updated

Both Cluster Controller and Resource Collector should implement and support gRPC server and client interface. 

## Protocol 1. ClusterController to ResourceCollector  
- Issue #74 - API for handling requests from ClusterController to ResourceCollector, and corresponding responses from ResourceCollector.
- When a cluster is registered/unregistered, the ClusterController should relay/remove the cluster's registration information, such as the cluster's geolocation, to/from the resource collector through this API. The ResourceCollector, which implements this API, will save/remove the cluster's static information in its local cache so that the scheduling algorithm can use this info. If the response is failure, should retry maximum times (5 times). If it still fails, set the cluster status to unvailable and detach the cluster from the Home scheduler. 
- ClusterController triggers ResourceCollector to collect status of a registered cluster. To illustrate, when ClusterController registers a cluster and send the cluster information to ResourceCollector using this API,  ResourceCollects starts to collect status information of the cluster.
- Cluster2ResourceCollector.Proto (Version: Proto3)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// service
service ClusterProtocol { 
    rpc SendClusterProfile(ClusterProfile) returns (ReturnMessageClusterProfile) {}
}

//message from ClusterController to ResourceCollector
message ClusterProfile {
    string ClusterNameSpace = 1;
    string ClusterName = 2;
    message ClusterSpecInfo {
        string ClusterIpAddress = 1;
        message GeoLocationInfo{
            string City = 1;
            string Province = 2;
            string Area = 3;
            string Country = 4;             
        }
        message RegionInfo {
            string Region = 1;
            string AvailabilityZone = 2; 
        }
        message OperatorInfo {
            string Operator = 1;
        }
        message FlavorInfo {
            string FlavorID = 1;  //Small, Medium, Large, Xlarge, 2xLarge 
            int64 TotalCapacity = 2;
        }
        message StorageInfo {
            string  TypeID = 1;     //"SATA", "SAS", "SSD"
            int64   StorageCapacity =2;
        }
        GeoLocationInfo GeoLocation = 2;
        RegionInfo Region = 3;
        OperatorInfo Operator = 4;
        repeated FlavorInfo Flavor = 5;
        repeated StorageInfo Storage = 6;
        int64 EipCapacity = 7;  
        int64 CPUCapacity = 8;  
        int64 MemCapacity = 9;   
        int64 ServerPrice = 10;    
        string HomeScheduler = 11;
        string HomeDispatcher = 12;
   }
   ClusterSpecInfo ClusterSpec = 3;
   string ClusterStatus = 4;
}

//Message from ResourceCollector, Cluster Controller should get response from ResourceCollector.
message ReturnMessageClusterProfile {
    enum CodeType{
        Error = 0;
        OK = 1;
    }               
    string ClusterNameSpace = 1;
    string ClusterName = 2;
    CodeType  ReturnCode = 3;
    string Message = 4;
}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## Cluster2ResourceCollector.Proto (Version: Proto2)
-	If Golang version 1.12.9 and Proto2 does not support nested structure, json format is used.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// service
service ClusterProtocol { 
rpc SendClusterProfile(ClusterProfile) returns (ReturnCode) {ReturnMessage}
}

//message from ClusterController to ResourceCollector
message ClusterProfile {
        string ClusterNameSpace = 1;
        string ClusterName = 2;
        string ClusterProfile = 3;	//json format 
}

//message from ResourceCollector
message ReturnMessage {
        string ClusterNameSpace = 1;
        string ClusterName = 2;
        int32 ReturnCode = 3;	   //0: Error, 1: OK 
}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## ClusterProfile example
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
"clusterprofile": { 
        "clusternamespace": "default",
        "clustername": "cluster1",
        "clusterspec": 
            “ipaddress”: “10.0.0.3”, 
            “geolocation”: { 
                “city”: “Bellevue”, 
                “province”: “Washington”, 
                “area”: “West”, 
                “country”: “US”
            }, 
            “region”: { 
                “region”: “us-west”,
                “availabilityzone”: “us-west-1” 
            }, 
            “operator”: { 
                “operator”: “globalscheduler” }, 
            “flavors”: [ {“flavorid”: 1, “TotalCapacity”: 5}, 
                     {“flavorid”: 2, “TotalCapacity”: 10}, 
                     {“flavorid”: 3, “TotalCapacity”: 20}, 
                     {“flavorid”: 4, “TotalCapacity”: 10}, 
                     {“flavorid”: 5, “TotalCapacity”: 5 }
                     ], 
            “storage”: [ { “typeid”: “sata”, “StorageCapacity”: 2000}, 
                      { “typeid”: “sas”, “StorageCapacity”: 1000}, 
                      { “typeid”: “ssd”, “StorageCapacity”: 3000}
                    ], 
            “eipcapacity”: 3, 
            “cpucapacity”: 8, 
            “memcapacity”: 256, 
            “serverprice”: 10, 
            “homescheduler”: “scheduler1”,
            “homedispatcher”: “dispatcher1” 
        }
        "clusterstatus": "ready"   
    } 
} 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## Cluster Data Structure
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
    NameSpace     string          `json:"namespace"`
    Name          string          `json:"name"`
    IpAddress     string          `json:"ipaddress"`
    GeoLocation   GeolocationInfo `json:"geolocation"`
    Region        RegionInfo      `json:"region"`
    Operator      OperatorInfo    `json:"operator"`
    Flavors       []FlavorInfo    `json:"flavors"`
    Storage       []StorageSpec   `json:"storage"`
    EipCapacity   int64           `json:"eipcapacity"`
    CPUCapacity   int64           `json:"cpucapacity"`
    MemCapacity   int64           `json:"memcapacity"`
    ServerPrice   int64           `json:"serverprice"`
    HomeScheduler string          `json:"homescheduler"`
    HomeDispatcher string         `json:"homedispatcher"`
    Status         string         `json:"status"`

type FlavorInfo struct {
    FlavorID      string `json:"flavorid"`
    TotalCapacity int64  `json:"totalcapacity"`
}

type StorageSpec struct {
    TypeID          string `json:"typeid"` //(sata, sas, ssd)
    StorageCapacity int64  `json:"storagecapacity"`
}

type GeolocationInfo struct {
    City     string `json:"city"`
    Province string `json:"province"`
    Area     string `json:"area"`
    Country  string `json:"country"`
}
type RegionInfo struct {
    Region           string   `json:"region"`
    AvailabilityZone []string `json:"availabilityzone"`
}

type OperatorInfo struct {
    Operator string `json:"operator"`
}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

## Protocol 2. ResourceCollector to ClusterController 
- Issue #75 - Cluster status update APIs from Resource Collector to Cluster Controller
- When there is a change in a cluster status (e.g., crash), ResourceCollector will detect the change. Then ResourceCollector will pass the cluster's status to Cluster Controller. The Cluster Controller should update the cluster object in the ETCD with the latest cluster status. If the status is "down" or "unreachable" (this depends on how we define the status enum values), the cluster controller should detach the association of the cluster with its scheduler (i.e., remove the cluster from the scheduler's partition pool) so that the scheduler will not schedule any POD to this cluster. Note that registration and unregistration of a cluster will only come from one entry, i.e. through the API server by an admin.
- ResourceCollector2ClusterController.Proto 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// service
service ResourceCollectorProtocol { 
    rpc UpdateClusterStatus(ClusterState) returns (ReturnMessageClusterState) {}  
}                                                                                               
message ClusterState {
    string NameSpace = 1;  
    string Name = 2;
    int64 State = 3; 	     //1: ready, 2: down, 3: unreachable…
}

//message from ClusterController
message ReturnMessageClusterState {
    string NameSpace = 1;  
    string Name = 2;
    int64 ReturnCode = 3;	 //0: Error, 1: OK 
}
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

