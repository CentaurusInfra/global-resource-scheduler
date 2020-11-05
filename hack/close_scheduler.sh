#!/usr/bin/env bash

# Copyright 2020 Authors of Arktos.
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

function kill_process {
  if [[ $# -eq 0 ]] ; then
     echo "The process to kill is not specified."
  else
     for process_name in "$@"; do
        process=$(ps aux|grep " ${process_name} "| wc -l)

        if [[ $process > 1 ]]; then
           process=$(ps aux|grep " ${process_name} "| grep arktos |awk '{print $2}')
           for i in ${process[@]}; do
               sudo kill -9 $i
           done
        fi
     done
  fi

  echo "Scheduler process has been closed"
}

kill_process kube-scheduler