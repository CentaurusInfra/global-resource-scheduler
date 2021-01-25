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

function create_pod {
# Create multiple YAML objects from stdin
cat <<EOF | cluster/kubectl.sh apply -f -
apiVersion: globalscheduler.com/v1
kind: Cluster
metadata:
  name: $1
  labels:
    app: test
spec:
  ipaddress: "34.218.224.247"
  geolocation:
    city: "Renton"
    country: "USA"
  region:
    region: "NW"
    availabilityzone: "az"
  operator:
    operator: "None"
  flavors:
    - flavorid: "1"
      totalcapacity: 100
    - flavorid: "2"
      totalcapacity: 200
  storage:
    - typeid: "ssd"
      storagecapacity: 12
    - typeid: "sas"
      storagecapacity: 2200
  eipcapacity: 100
  cpucapacity: 200
  memcapacity: 300
  serverprice: 400
  homescheduler: "mockscheduler"
  homedispatcher: "mockdispatcher"
EOF
}

for ((i = 0 ; i < $(($1)) ; i++)); do
    name="cluster-$(($i))"
    create_pod $name
done
