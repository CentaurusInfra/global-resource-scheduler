# Scheduler works with API Server

Jan-8-2021, Wang jun

## 1. Description
This doc is used for understanding how the global scheduler works with API Server, with a usage instruction and implementation process description.

## 2. Usage Instruction
#### 1) Build and run API Server and gs-scheduler
```
# NOTE: Envionment must be configured. For details, please see the GlobalScheduler develop guide.
# go into the base repo directory
cd src/k8s.io/kubernetes

# export config base directory for gs-scheduler
export CONFIG_BASE=${PWD}/conf

# modify config file for your environment
# config file: ${PWD}/conf/conf.yaml, the following keys in this config file should be changed
# master: url of the API Server
# kubeconfig: config of the API Server
# scheduler_policy_file: the scheduler policy file path, the default file path is also in ${PWD}/conf/conf.yaml

# build and run
hack/gs-scheduler-up.sh
```

#### 2) Create VM and check the schedule result
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

# binding pod with default schedulername "scheduler1"
root@arktos-wj-dev:~/yaml# cat binding.json
{
"apiVersion":"v1",
"kind":"Binding",
"metadata": {
"name":"demovm01"
},
"target": {
"apiVersion":"v1",
"kind":"Scheduler",
"name":"scheduler1"
}
}

root@arktos-wj-dev:~/yaml# curl -X POST http://127.0.0.1:8080/api/v1/namespaces/default/pods/demovm/binding -H "Content-Type: application/json" -d @binding.json
{
  "kind": "Status",
  "apiVersion": "v1",
  "metadata": {

  },
  "status": "Success",
  "code": 201
}

# check the schedule result
root@arktos-wj-dev:~/yaml# kubectl get pods
NAME       HASHKEY               READY   STATUS   RESTARTS   AGE
demovm01   8970278610605413590   0/1     Bound   0          4m47s

# check the clustername
root@arktos-wj-dev:~/yaml# kubectl get pods  --field-selector spec.clusterName="az1"
NAME       HASHKEY               READY   STATUS   RESTARTS   AGE
demovm01   8970278610605413590   0/1     Bound   0          5m52s

# here we can see the schedule destination in the spec.clusterName parameter az1
```





