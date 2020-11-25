# Global Scheduler Internal API Demo

Nov-11-2020, Wang Jun

## 1. Description
This is used for understanding current scheduler process and api obj design. The demo will tell you how to build and config scheduler, and how to call the internal api.

## 2. Generate config data for your demo
We already prepare the demo mock data for you, and also you can change it. The path of the mock data file is src/k8s.io/kubernetes/conf.

- conf.yaml: internal api server config 
- edgecloud_policy.yaml: scheduler policy config
- eip_pools.json: eip pools mock data
- flavors.json: flavor mock data
- region.yaml: support region config data
- sample_allocation_req_body.json: internal API sample data
- site_nodes.json: site resource detail mock data
- sites.json: site info mock data
- volume_pools.json: volume pools mock data
- volume_types.json: support volume types mock data

For example, the sample volume_pools.json:
```
{
  "volume_pools": [{
      "region": "region1",
      "pools": [{
          "availability_zone": "az1",
          "name": "pool1",
          "capabilities": {
            "pool_name": "pool1",
            "provisioned_capacity_gb": 0,
            "allocated_capacity_gb": 10,
            "free_capacity_gb": 2000,
            "total_capacity_gb": 2000,
            "volume_type": "SAS",
            "max_over_subscription_ratio": 1.0,
            "thin_provisioning_support": true,
            "reserved_percentage": 0.3,
            "pool_model": 0
          }
        }
      ]
    }
  ]
}
```
Means there is one volume pool, and it is in region1 and az1, and available volume capacity is 2000 gb.

You should change scheduler_policy_file path as your environment in conf.yaml
```
root@arktos-wj-dev2:~/work/src/k8s.io/kubernetes# cat conf/conf.yaml
#scheduler info
address: 0.0.0.0
port: 13443

scheduler_policy: edgecloud
# ${PWD} should be replaced to real path
scheduler_policy_file: ${PWD}/conf/edgecloud_policy.yaml

spot_region_mapping:
  - region: region1
    spot_region: region1

```

## 3. Build and run gs-scheduler
```
# go into the base directory
cd src/k8s.io/kubernetes

# build gs-scheduler
make WHAT=cmd/gs-scheduler

# export config base directory for gs-scheduler
export CONFIG_BASE=${PWD}/conf

# run gs-scheduler
./_output/local/go/bin/gs-scheduler
```

## 4. Do the scheduler API call
```
root@arktos-wj-dev2:~/work/src/k8s.io/kubernetes# curl -H "Content-Type:application/json" -X POST -T conf/sample_allocation_req_body.json http://127.0.0.1:13443/v1/allocations
{
  "allocation": {
   "id": "7a41e0a0-bad9-49cc-9948-9531aa79c2b7",
   "resource_groups": [
    {
     "name": "RTC",
     "resources": [
      {
       "name": "rtc-name",
       "flavor_id": "c6.large.2",
       "count": 1
      }
     ],
     "selected": {
      "node_id": "site1",
      "region": "region1",
      "availability_zone": "az1"
     }
    }
   ]
  }
 }
```
