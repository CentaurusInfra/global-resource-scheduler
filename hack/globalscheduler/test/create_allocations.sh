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

FILE="/home/ubuntu/go/src/k8s.io/kubernetes/globalscheduler/test/yaml/sample_1000_allocations.yaml"

azs=("non-production-az" "production-az")

openstackimgs=("3bbfe0a3-3121-489d-a50f-11b02fa82e5f" "9ee2592a-ee7c-489d-8e5c-33e3b241f50b" "a302b941-523a-412b-b25b-a466d13e2484" "6bcd91c5-7723-4ff4-b1f6-f9956bd2f8e2" "23002b3b-da3e-4a84-bc51-013f7a6cf99f" "6c431086-1baf-4407-b791-0f5c58764a76" "7ddf20ab-e461-4901-acf1-75d5dd96894c" "60e8e418-3d3c-4258-ba1c-d6ba13c036c1" "8da60b24-fb74-41f5-ae2f-6fefd5ac85b7" "8b66d1b5-523d-4831-9260-547f8c0e12f5" "dcc20062-b7c8-4b5c-8375-b658dde14d7c" "04752559-4c4c-4a43-bcba-e06a5e0c61d9" "b33f48fb-cff6-46d1-a173-04a839cdfd09" "3fe943ef-4a6b-461e-aa1e-33fcccd74b1b" "d058d90f-640c-46ae-9a1f-03d04aa1dcf6" "ebc886b4-630d-46c1-89ff-3b15a54693cb" "e43d5857-763b-43f8-bd63-d34752cb2ace" "9f340c15-7159-468a-b214-f38a435cbda0" "5c19f56b-0db0-4ec6-b7c2-6540a4719a98" "eefd04df-fc29-4f76-adbe-70353d7ef05c")

openstacksgs=("a3a9bfb6-6aac-4b87-b9db-623d7b2f898f" "3e1c3fff-0dce-475e-bafc-f49d12eaf101" "bbcb1fa2-b621-4dd3-bfcb-ab45f73b4606" "62bd22be-2a1a-46cc-9d7f-915010d9028a" "04a436f5-bf8b-4935-b5ba-a31ed8393f8f" "2f20c888-5e25-4b0c-b8ef-5aa09615bd7c" "ae29deec-40f3-4555-9104-d78b5810c073" "e17ea9d1-7a18-45cd-b0ba-335230ab1d6c" "1c333d73-f08f-4b04-9be2-b5add540747e" "c1b57043-751b-44fb-82e6-e535851888d3" "f7b9d4ac-da9c-48d9-ae27-adf0826c2a33" "60746e12-febd-412a-8451-2a23cce72a41" "124e13c7-f777-4422-9f08-7d6370fa3d26" "683e06a4-6c5e-4841-8cdd-ba8fd90b20f2" "c4148970-20f4-4010-8c6c-425421a7e2dc" "a018bd9a-b078-4df1-ba39-6fead31ab96a" "8d152057-09a9-4c81-9c93-46e1bac062c0" "4e8130b5-4005-4e18-902f-22c7671bec50" "	6990938f-9278-43df-a628-efc7e54c6bd5" "7015395f-a89e-45c7-859a-11e185a08fff")

openstacknwks=("ea45fb28-90c7-4ad1-a20d-7a33abedddc7" "c8bb8227-4009-470a-b937-618ac106c4e6" "8673ecbf-17ea-4d5e-8641-f4e7abf7ceca" "3dc36059-4e7f-4953-9c4b-947e1e03bda5" "1c772308-bdb5-424b-bd0a-8174573d64d6" "4e9aac63-f1a8-46fc-96ac-3bea01835dd2" "3475ac6d-163d-4dab-a38b-1a3ad590031f" "	f2f6d4d6-7456-48e8-af09-59e3fb4bc1d9" "b556435f-c962-4abb-a2f8-6ea964c5e5d4" "2b95090f-1cad-4660-b8ba-bdf10340b52d" "10684223-99f4-4424-9c97-80be06bdc88b" "4bd153d9-8548-4463-957e-f62dea9e088c" "92f107a0-ca6b-472c-bea3-7c16fd48d6a7" "7d9fd83d-d8e0-467a-a370-81ef27ba3a89" "	64d1601d-cab8-46a0-9fe2-ff12d54091ff" "	40262c89-3ccb-46d9-86af-227479b2bcfa" "38865b0c-bdf8-4188-8d2e-ff143c10e899" "640fd30f-4e2b-4eda-9246-dc9a28f4af9c" "	6990938f-9278-43df-a628-efc7e54c6bd5" "	fdd58354-aece-4144-889c-cfd329f0c1d3")

function create_allocation {
# Create multiple YAML objects from stdin
cat <<EOM >> $FILE
apiVersion: globalscheduler.com/v1
kind: Allocation
metadata:
  name: $1
spec:
  resource_group:
    name: rg-$1
    resources:
      - name: resource-$1
        resource_type: vm
        flavors:
          - flavor_id: "42"
            spot:
              max_price: "1.5"
              spot_duration_hours: 2
              spot_duration_count: 3
              interruption_policy: immediate
        storage:
          sata: 50
          sas: 60
          ssd: 70
        need_eip: true
        virtual_machine:
          image: "$7"
          security_group_id: $8
          nic_name: "$9"
  selector:
    geo_location:
      city:  $3
      province: $4
      area: $2
      country: $5
    regions:
      - region: $2
        availability_zone: ["$6"]
    operator: globalscheduler
    strategy:
      location_strategy: discrete
  replicas: 1
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
    name="alloc-$(($i))"
    area="area-$(($idx))"
    city="city-$(($idx))"
    province="province-$(($idx))"
    country="US"
    az="az-$(($azsIdx))"
    create_allocation $name $area $city $province $country $az ${openstackimgs[$idx]} ${openstacksgs[$idx]} ${openstacknwks[$idx]}
done
