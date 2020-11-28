# Scheduler works with API Server

Nov-28-2020, Wang jun

## 1. Description
This doc is used for understanding how the global scheduler works with API Server, with a usage instruction and implementation process description.

## 2. Usage Instruction
#### 1) Build and run API Server
```
# NOTE: Envionment must be configured. For details, please see the GlobalScheduler develop guide.
# go into the base repo directory
cd src/k8s.io/kubernetes

# get the latest code
git pull

# build and run API Server

```

#### 2) Build and run gs-scheduler
```
# go into the base repo directory
cd src/k8s.io/kubernetes

# build gs-scheduler
make WHAT=cmd/gs-scheduler

# export config base directory for gs-scheduler
export CONFIG_BASE=${PWD}/conf

# modify config file for your environment
# config file: ${PWD}/conf/conf.yaml, the following keys in this config file should be changed
# master: url of the API Server
# kubeconfig: config of the API Server
# scheduler_policy_file: the scheduler policy file path, the default file path is also in ${PWD}/conf/conf.yaml

# run gs-scheduler
./_output/bin/gs-scheduler
```

#### 3) Create VM and check the schedule result
```
# prepare the vm define yaml as follows
root@arktos-wj-dev:~/yaml# cat vm.yaml
apiVersion: v1
kind: Pod
metadata:
  name: demovm
spec:
  virtualMachine:
    keyPairName: "foobar"
    name: vm001
    image: download.cirros-cloud.net/0.3.5/cirros-0.3.5-x86_64-disk.img
    imagePullPolicy: IfNotPresent
    needEIP: true
    flavors:
      - flavorID: "c6.large.2"
    resourceCommonInfo:
      selector:
        regions:
          - region: "region1"
        strategy:
          localStrategy: "centralize"

# create vm
root@arktos-wj-dev:~/yaml# kubectl apply -f vm.yaml
pod/demovm created

# check the schedule result
root@arktos-wj-dev:~/yaml# kubectl describe pods demovm
Name:         demovm
Namespace:    default
Tenant:       system
Priority:     0
Node:         az1/
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"v1","kind":"Pod","metadata":{"annotations":{},"name":"demovm","namespace":"default","tenant":"system"},"spec":{"virtualMach...
Status:       Pending
IP:
Containers: <none>
Conditions:
  Type           Status
  PodScheduled   True
Volumes:         <none>
QoS Class:       BestEffort
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute for 300s
                 node.kubernetes.io/unreachable:NoExecute for 300s
Events:          <none>

# here we can see the schedule destination in the "Node" parameter az1
```

## 3. Implementation Description
![image.png](/images/schedule_work_with_api_server.png)

#### 1) gs-scheduler watches the unassigned pod
There is a pod informer in the gs-scheduler, and it will watch all the unassigned vm pod create events

#### 2) gs-scheduler puts the pod into the schedulingQueue
Because the gs-scheduler currently can just handle with the stack, which is a new define for the resource groups in the global scheduler scenario.

The resource group contains a group of resource, with the same deployment requirement and will be deploy to the same cluster. The resource conains some vms with same resource type(vm or container), storage requirement, network requirement.

The stack is the atomic scheduling unit in global scheduler. So the pod will be changed to the stack. Currently for the pod scheduling, there is only one resource in the stack, and only one vm in the resource, and the vm is the same configuration as the pod. The stack will be put into the schedulingQueue.

#### 3) gs-scheduler does the schedule process
The gs-scheduler will do the schedule pipeline, and schedule the clusters for the stack

#### 4) gs-scheduler binding the pod
The gs-scheduler will get the scheduled result of the stack, contains the selected region and az, and gs-scheduler will find the stack associated pod, and do the final binding operation.
The node name will be changed to the cluster.




