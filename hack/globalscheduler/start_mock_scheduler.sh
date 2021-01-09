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

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..

source "${KUBE_ROOT}/hack/lib/common-var-init.sh"

BINARY_DIR=${BINARY_DIR:-}

source ${KUBE_ROOT}/hack/arktos-cni.rc

source "${KUBE_ROOT}/hack/lib/init.sh"

source "${KUBE_ROOT}/hack/lib/common.sh"

### Allow user to supply the source directory.
GO_OUT=${GO_OUT:-}
while getopts "ho:O" OPTION
do
    case ${OPTION} in
        o)
            echo "skipping build"
            GO_OUT="${OPTARG}"
            echo "using source ${GO_OUT}"
            ;;
        O)
            GO_OUT=$(kube::common::guess_built_binary_path)
            if [ "${GO_OUT}" == "" ]; then
                echo "Could not guess the correct output directory to use."
                exit 1
            fi
            ;;
        h)
            usage
            exit
            ;;
        ?)
            usage
            exit
            ;;
    esac
done

if [ "x${GO_OUT}" == "x" ]; then
  make -C "${KUBE_ROOT}" WHAT="cmd/hyperkube globalscheduler/cmd/scheduler_process/mock_scheduler"
else
  echo "skipped the build."
fi

### IF the user didn't supply an output/ for the build... Then we detect.
if [ "${GO_OUT}" == "" ]; then
  kube::common::detect_binary
fi

kube::common::set_service_accounts

### to replace gs-controllers for missing service account
### kube::common::start_controller_manager

### The following commented ones are to set up a pod to bind to  a cluster for quick testing only

#${KUBECTL} --kubeconfig "${CERT_DIR}/admin.kubeconfig" apply -f ${KUBE_ROOT}/globalscheduler/test/mock/scheduler_def.yaml

#${KUBECTL} --kubeconfig "${CERT_DIR}/admin.kubeconfig" apply -f ${KUBE_ROOT}/globalscheduler/test/mock/pod.json

### it takes time for scheduler CRD to get registered
#sleep 5

${KUBECTL} --kubeconfig "${CERT_DIR}/admin.kubeconfig" apply -f ${KUBE_ROOT}/globalscheduler/test/mock/mock_scheduler.yaml

#curl -X POST http://127.0.0.1:8080/api/v1/namespaces/default/pods/nginx/binding -H "Content-Type: application/json" -d @${KUBE_ROOT}/globalscheduler/test/mock/assign_scheduler.json

CONTROLPLANE_SUDO=$(test -w "${CERT_DIR}" || echo "sudo -E")

${CONTROLPLANE_SUDO} ${GO_OUT}/mock_scheduler

