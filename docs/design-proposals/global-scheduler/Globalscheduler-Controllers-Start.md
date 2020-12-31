# Global Resource Scheduler – How to Run Controllers

Dec-30-2020

## 1. Build GlobalScheduler Controllers

This section describes how to build and run global scheduler's controllers such as scheduler, distributer, dispatcher and cluster controllers.

(1) ~/go/src/k8s.io/arktos$> make clean \break
(2) ~/go/src/k8s.io/arktos$> make WHAT=globalscheduler/cmd/gs-controllers
(3) ~/go/src/k8s.io/arktos$> make update
(4) ~/go/src/k8s.io/arktos$> bazel clean
(5) ~/go/src/k8s.io/arktos$> make bazel-test
(6) ~/go/src/k8s.io/arktos$> make bazel-test-integration

## 2. Run GlobalScheduler Controllers
(1) ~/go/src/k8s.io/arktos$> make WHAT=globalscheduler/cmd/gs-controllers
(2) ~/go/src/k8s.io/arktos$> cd _output/local/bin/linux/amd64
(3) ~/go/src/k8s.io/arktos/_output/local/bin/linux/amd64$>gs-controllers
The following step is optional. This starts cluster_grpc_server when global scheduler's ResourceCollector module is available.
(4) ~/go/src/k8s.io/kubernetes/globalscheduler/cmd$> go run cluster/cluster_grpc_server.go 

## 3. Check log files 
After running controllers, log files are generated to check the status of controllers.

(1) ~/go/src/k8s.io/arktos$> cd /tmp
(2) ~/go/src/k8s.io/arktos$> cat gs-controllers.INFO
(3) ~/go/src/k8s.io/arktos$> cat gs-controllers.ERROR
(4) ~/go/src/k8s.io/arktos$> cat gs-controllers.FATAL
(5) ~/go/src/k8s.io/arktos$> cat cluster_grpc_server.INFO
