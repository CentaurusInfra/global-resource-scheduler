# Global Resource Scheduler – How to Run Controllers

Dec-30-2020

## 1. Build GlobalScheduler Controllers

This section describes how to build and run scheduler, distributor, dispatcher and cluster controllers of global scheduler concurrently using a single main() function.

(1) ~/go/src/k8s.io/arktos$> make clean \
(2) ~/go/src/k8s.io/arktos$> make update \
(3) ~/go/src/k8s.io/arktos$> make all WHAT=globalscheduler/cmd/gs-controllers \
(4) ~/go/src/k8s.io/arktos$> make all WHAT=globalscheduler/cmd/grpc-server \
(5) ~/go/src/k8s.io/arktos$> bazel clean \
(6) ~/go/src/k8s.io/arktos$> make bazel-test \
(7) ~/go/src/k8s.io/arktos$> make bazel-test-integration

## 2. Run GlobalScheduler Controllers
(1) ~/go/src/k8s.io/arktos$> ./hack/arktos-up.sh \
    Open another terminal. \
(2) ~/go/src/k8s.io/arktos$> cd _output/local/bin/linux/amd64 \
(3) ~/go/src/k8s.io/arktos/_output/local/bin/linux/amd64$>./gs-controllers \
The following step is optional. This starts cluster_grpc_server when global scheduler's ResourceCollector module is available. \
(4) ~/go/src/k8s.io/arktos/_output/local/bin/linux/amd64$>./grpc-server \
Testing yaml files are located in /globalscheduler/test/yaml
(5) ~/go/src/k8s.io/arktos/globalscheduler/test/yaml$>kubectl apply -f tested_cluster_object2.yaml

## 3. Check log files 
After running controllers, log files are generated to check the status of controllers.

(1) ~/go/src/k8s.io/arktos$> cd /tmp \
(2) ~/go/src/k8s.io/arktos$> cat gs-controllers.INFO \
(3) ~/go/src/k8s.io/arktos$> cat gs-controllers.ERROR \
(4) ~/go/src/k8s.io/arktos$> cat gs-controllers.FATAL \
(5) ~/go/src/k8s.io/arktos$> cat cluster_grpc_server.INFO

