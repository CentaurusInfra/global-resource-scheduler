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

for ((i = 0 ; i < $(($1)) ; i++)); do
echo "apiVersion: globalscheduler.com/v1
kind: Allocation
metadata:
  name: allocation-$(($i))
spec:
  resource_group:
    name: resource-group-$(($i))
    resources:
      - name: allocation-resource-$(($i))
        resource_type: vm
        flavors:
          - flavor_id: \"42\"
            spot:
              max_price: \"1.5\"
              spot_duration_hours: 2
              spot_duration_count: 3
              interruption_policy: immediate
        storage:
          sata: 50
          sas: 60
          ssd: 70
        need_eip: true
        virtual_machine:
          image: \"0df07567-87a8-4d01-b7d9-c70f91c86427\"
          security_group_id: \"58b9fbbf-6f04-49ca-8c7e-ac797c6d236c\"
          nic_name: \"9e29aa2d-6943-4109-bc5c-1a882b086122\"
  selector:
    geo_location:
      city: Bellevue
      province: Washington
      area: NW-1
      country: US
    regions:
      - region: \"NW-1\"
        availability_zone: [\"production-az\"]
    operator: globalscheduler
    strategy:
      location_strategy: discrete
  replicas: 2
---"
done  > sample_allocations.yaml
