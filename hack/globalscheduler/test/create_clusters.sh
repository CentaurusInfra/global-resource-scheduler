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
openstackips=("18.236.217.191" "34.219.215.27" "34.216.194.0" "18.237.132.54" "34.219.98.84" "18.237.126.169" "34.217.135.219" "34.219.85.67" "35.163.74.249" "54.185.103.249" "35.165.187.214" "34.211.111.114" "35.160.19.193" "54.191.208.66" "54.149.37.227" "34.211.4.11" "18.237.203.189" "18.237.63.170" "54.202.231.12" "54.244.161.7")
azs=("non-production-az" "production-az")
FILE="/home/ubuntu/go/src/k8s.io/kubernetes/globalscheduler/test/yaml/sample_1000_clusters_new.yaml"

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