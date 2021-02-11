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
locs=("NewYork NewYork NE-1 US" "Bellevue Washington NW-1 US" "Orlando Florida SE-1 US" "Austin Texas SW-1 US" "Chicago Illinois Central-1 US" "Boston Massachusettes NE-2 US" "SanFrancisco California NW-2 US" "Atlanta Georgia SE-2 US" "LasVegas Nevada SW-2 US" "Omaha Nebraska Central-2 US")

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
    name: vm$1
    image: "e756d63a-4c1d-4913-816d-6c9269ea6ff5"
    keyPairName: "demo-keypair"
    securityGroupId: "5c37cea2-934b-428a-93f5-b7b598cf3424"
    flavors:
      - flavorID: "42"
    resourceCommonInfo:
     count: 1
     selector:
       geoLocation:
         city: "$2"
         province: "$3"
         area: "$4"
         country: "$5"
       regions:
         - region: "$4"
           availablityZone:
           - "$4"

  nics:
    - name: "80e76da3-91c6-4b70-8fdb-d1ca1c2836cd"
EOF
}

locsLen=${#locs[@]}

for ((i = 0 ; i < $(($1)) ; i++)); do
    locsIdx=$(($i%locsLen))
    name="pod-$(($i))"
    create_pod $name ${locs[$locsIdx]}
    sleep $2
done
