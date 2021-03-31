#!/usr/bin/env bash

# Copyright 2020 Authors of Arktos - file created.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
openstackips=("18.236.217.191" "54.185.103.249" "54.149.22.83" "18.237.167.235" "34.220.133.183" "54.212.231.62" "18.236.245.207" "34.211.110.194" "54.189.196.126" "35.165.106.249" "34.210.104.15" "34.220.41.107" "52.34.63.99" "54.184.175.142" "34.221.41.172" "54.189.234.9" "18.237.126.53" "18.236.244.165" "35.166.177.124" "18.236.217.216")
azs=("non-production-az" "production-az")
FILE="/home/ubuntu/go/src/k8s.io/arktos/globalscheduler/test/yaml/sample_1000_clusters_new.yaml"

function create_cluster {
# Create multiple YAML objects from stdin
cat <<EOM >> $FILE
apiVersion: globalscheduler.com/v1
kind: Cluster
metadata:
  name: $1
  namespace: default
spec:
  cpucapacity: 8
  eipcapacity: 3
  flavors:
  - flavorid: "42"
    totalcapacity: 1000
  geolocation:
    area: $2
    city: $3
    country: $5
    province: $4
  ipaddress: $6
  memcapacity: 256
  operator:
    operator: globalscheduler
  region:
    availabilityzone: $7
    region: $2
  serverprice: 10
  storage:
  - storagecapacity: 256
    typeid: sas
---
EOM
}

ipsLen=${#openstackips[@]}
azsIdx=0

for ((i = 0 ; i < $(($1)) ; i++)); do
    ipsIdx=$(($i%ipsLen))
    if [ $ipsIdx -eq 0 ]
    then
      azsIdx=$((azsIdx+1))
    fi
    name="cluster-$(($i))"
    area="area-$(($ipsIdx))"
    city="city-$(($ipsIdx))"
    province="province-$(($ipsIdx))"
    country="US"
    az="az-$(($azsIdx))"
    create_cluster $name $area $city $province $country ${openstackips[$ipsIdx]} $az
done