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
apiVersion: v1
kind: Pod
metadata:
  name: $1
spec:
  resourceType: "vm"
  virtualMachine:
    name: openstack$1
    image: "9e2b5eb9-3bb0-469b-b901-a75d87e4d958"
    keyPairName: "demo-keypair"
    securityGroupId: "5e9513a8-b779-4cb9-825b-a1e994f28ddc"
    flavors:
      - flavorID: "42"
    resourceCommonInfo:
     count: 1
     selector:
       geoLocation:
         city: ""
         province: ""
         country: ""
       regions:
         - region: "NE"
  nics:
    - name: "c32fa6da-ee3f-4c65-80a7-c920c0a9ce7e"
EOF
}

for ((i = 0 ; i < $(($1)) ; i++)); do
    name="pod-$(($i))"
    create_pod $name
done
