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

### IF the user didn't supply an output/ for the build... Then we detect.
if [ "${GO_OUT}" == "" ]; then
  kube::common::detect_binary
fi

function start_scheduler {
  # Ensure CERT_DIR is created for auto-generated crt/key and kubeconfig
  mkdir -p "${CERT_DIR}" &>/dev/null || sudo mkdir -p "${CERT_DIR}"

  # install cni plugin based on env var CNIPLUGIN (bridge, alktron)
  kube::util::ensure-gnu-sed

  kube::common::set_service_accounts

  tag=$(($1))
  schedulerid=$2
  kube::common::start_gs_scheduler $tag $schedulerid
  
  echo "Done Starting Scheduler $tag"
}

start_scheduler $@