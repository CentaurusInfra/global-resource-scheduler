## Global scheduler integration test steps
1. git clone global resource scheduler code 

2. In one terminal, run "global-resource-scheduler/hack/globalscheduler/globalscheduler-up.sh"

3. Go to another terminal, do the following:
--cd global-resource-scheduler/globalscheduler/test/yaml
--kubectl apply -f sample_1000_clusters.yaml to create 1000 clusters
--wait for resoruce collector's collectorCache completing update of cluster/site info from openstack, i.e., 
do "tail -f /tmp/resource_collector.log and make it sure that all the sites are added to the collector cache". 
Then do the following:
--kubectl apply -f sample_2_schedulers.yaml to create 2 schedulers
--kubectl apply -f sample_2_distributors.yaml to create 2 distrbutors
--kubectl apply -f sample_2_dispatchers.yaml to create 2 dispatchers

You can verify the successful creation of the scheduler/distributor/dispatcher processes via: 

"kubectl get ..." or "ps -ef | grep ..."

4. now you can do the following to create 100 PODs/s
global-resource-scheduler/hack/globalscheduler/test/create_pods.sh 100 0.01

5. you can use kubectl get pod to check the status of pod

6. The you can login to openstack http://34.218.224.247/dashboard/auth/login/?next=/dashboard/ to check that all the VM PODs are successfully created. The openstack takes quite a long time to complete the creation of 100 VMs

7. now you can do the following to delete 100 PODs/s 
global-resource-scheduler/hack/globalscheduler/test/delete_pods.sh 100 0.01

8. login to openstack http://34.218.224.247/dashboard/auth/login/?next=/dashboard/ to check that all the VM PODs are successfully deleted. 
