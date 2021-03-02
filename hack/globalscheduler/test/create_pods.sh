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
    image: "9ecb51b6-f723-4654-bfcd-37459245c9dc"
    keyPairName: "demo-keypair"
    securityGroupId: "ca3a65fb-f304-438d-97bd-171d713f5aa5"
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
    - name: "211d3389-1c1a-4938-8686-c61ff81b7ff7"
EOF
}

locsLen=${#locs[@]}

starts=$(date +%s%N)
sleepseconds=$2
sleepms=$(echo "$sleepseconds*1000000000/1" | bc)

for ((i = 0 ; i < $(($1)) ; i++)); do
    locsIdx=$(($i%locsLen))
    name="pod-$(($i))"
    create_pod $name ${locs[$locsIdx]} &
    #sleep take more time than expected.
    tw=$((${tw} + ${sleepms}))
    #Pods are created in 10 units to save time
    if [ $(($i%10)) -eq 9 ] 
    then 
         while [ $(($(date +%s%N) - ${starts})) -lt ${tw} ]
         do
              sleep 0.0001
         done
    fi
done
#Wait the time left
while [ $(($(date +%s%N) - ${starts})) -lt ${tw} ]
do
    sleep 0.0001
done

echo "The shell script took $(echo "scale=3;($(date +%s%N) - ${starts})/(1*10^06)" | bc) milliseconds to run"
