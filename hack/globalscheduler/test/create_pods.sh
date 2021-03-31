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
# locs=("NewYork NewYork NE-1 US" "Bellevue Washington NW-1 US" "Orlando Florida SE-1 US" "Austin Texas SW-1 US" "Chicago Illinois Central-1 US" "Boston Massachusettes NE-2 US" "SanFrancisco California NW-2 US" "Atlanta Georgia SE-2 US" "LasVegas Nevada SW-2 US" "Omaha Nebraska Central-2 US")

FILE="/home/ubuntu/go/src/k8s.io/arktos/globalscheduler/test/yaml/sample_1000_pods.yaml"

azs=("non-production-az" "production-az")

openstackimgs=("3bbfe0a3-3121-489d-a50f-11b02fa82e5f" "8b66d1b5-523d-4831-9260-547f8c0e12f5" "e2579c66-b27a-4cc0-80a2-660440a6c221" "a3e60807-629b-4bab-a19c-a8d28bb7ce25" "b33f48fb-cff6-46d1-a173-04a839cdfd09" "2d3d217d-d8ea-48f7-a556-99ae3dbe7b21" "d058d90f-640c-46ae-9a1f-03d04aa1dcf6" "ebc886b4-630d-46c1-89ff-3b15a54693cb" "b64e8222-af0b-4398-9312-6ff83a6d1dcb" "31e20e58-f80c-4950-ade7-bc9723626980" "5c19f56b-0db0-4ec6-b7c2-6540a4719a98" "f6464c02-50aa-4d5f-8f42-edff154c8a3a" "eefd04df-fc29-4f76-adbe-70353d7ef05c" "67760b43-f069-41f7-a20a-208fcbda7ac1" "6bcd91c5-7723-4ff4-b1f6-f9956bd2f8e2" "8ee97fa3-8bf5-4dd2-b6bf-d73bb3fe5dd2" "8117ce96-45eb-47a5-b575-5a8c055ab443" "7ddf20ab-e461-4901-acf1-75d5dd96894c" "893bec40-9202-4942-96af-e9b1c16f26f1" "e3ab6de2-12bd-4327-9366-8c258ad67e4a")

openstacksgs=("a3a9bfb6-6aac-4b87-b9db-623d7b2f898f" "8c64a7aa-f0b7-4a71-bbdf-2b97e5fecb53" "d7b2674e-7e89-4e9a-b1c6-e0e120e1cd3c" "8c82051d-3925-43cd-adf7-dc8c5d08c125" "b39b21d6-55a0-461e-a9c8-c8d9c0c18523" "83f6fbd3-5ef6-45fd-a7a8-6791c8a144fe" "a858b2de-2fdb-4802-a05b-08085655903c" "b4686a31-13d7-4c64-904b-01cb33d25605" "a6b955cf-17a9-4035-a322-ac61bc483dc2" "8094f8b6-0261-42a1-8460-f3d6d73161e6" "d3c5e284-3c27-40e1-91c5-51db524c4792" "aa41ff43-7722-428d-b4bf-94b6b94a0dd8" "3cfe8500-4b69-4f03-ae4d-aa98bb786697" "70df7ce5-a066-441b-9baa-5825ee6a364c" "f4d368e4-58fc-4f65-98cd-df3613599cbd" "1733f84f-38a9-4758-9515-f83cd1c4be97" "4a3dee15-af80-47ea-bbf7-1a5fcbea9f6d" "1f3dd06e-f1b2-48c1-b964-a7a5db3df2ab" "aef55a57-8ab9-465c-91aa-c315f911dc50" "0d723e90-b7bf-46e8-9056-0b67ccd3a17e")

openstacknwks=("ea45fb28-90c7-4ad1-a20d-7a33abedddc7" "c1b57043-751b-44fb-82e6-e535851888d3" "13ddc071-c82f-4901-8018-ce1d1c6b9838" "c6291f38-9922-49ad-867f-24a08919a727" "124e13c7-f777-4422-9f08-7d6370fa3d26" "38ae0668-6cf4-4407-8b18-e5b0c90b0d1b" "c4148970-20f4-4010-8c6c-425421a7e2dc" "a018bd9a-b078-4df1-ba39-6fead31ab96a" "22bdcc13-4803-4b27-9009-f6bf8c1e7495" "aba99295-8ef4-416d-9826-83d1005d38bd" "4367d94b-3565-49e5-8898-de76b724c41b" "dac608f2-465f-4856-8aef-9aa406a1139a" "7015395f-a89e-45c7-859a-11e185a08fff" "c95837fb-f90b-448f-9033-d0e12fce8235" "62bd22be-2a1a-46cc-9d7f-915010d9028a" "1ed0da3b-bbc3-4aaf-8056-305b4065d728" "e2682062-e88d-4da2-9a8d-2c10e7dc4453" "ae29deec-40f3-4555-9104-d78b5810c073" "ca2f6560-6595-4f89-adce-182029d5c96d" "47b65713-63bb-495e-a2f6-99eb7f00e812")

function create_pod {
# Create multiple YAML objects from stdin
cat <<EOM >> $FILE
apiVersion: v1
kind: Pod
metadata:
  name: $1
spec:
  resourceType: "vm"
  virtualMachine:
    name: vm$1
    image: "$7"
    keyPairName: "demo-keypair"
    securityGroupId: "$8"
    flavors:
      - flavorID: "42"
    resourceCommonInfo:
     count: 1
     selector:
       geoLocation:
         city: $3
         province: $4
         area: $2
         country: $5
       regions:
         - region: $2
           availablityZone:
           - "$6"
  nics:
    - name: "$9"
---
EOM
}

# locsLen=${#locs[@]}
azsIdx=0

for ((i = 0 ; i < $(($1)) ; i++)); do
    idx=$(($i%20))
    if [ $idx -eq 0 ]
    then
      azsIdx=$((azsIdx+1))
    fi
    name="pod-$(($i))"
    area="area-$(($idx))"
    city="city-$(($idx))"
    province="province-$(($idx))"
    country="US"
    az="az-$(($azsIdx))"
    create_pod $name $area $city $province $country $az ${openstackimgs[$idx]} ${openstacksgs[$idx]} ${openstacknwks[$idx]}
done