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

CONTROLPLANE_SUDO=$(test -w "${CERT_DIR}" || echo "sudo -E")
kubeconfigfilepaths="${CERT_DIR}/scheduler.kubeconfig"
if [[ $# -gt 1 ]] ; then
    kubeconfigfilepaths=$@
fi
SCHEDULER_LOG=${LOG_DIR}/kube-scheduler.log
${CONTROLPLANE_SUDO} "${GO_OUT}/hyperkube" kube-scheduler \
    --v="${LOG_LEVEL}" \
    --leader-elect=false \
    --kubeconfig "${kubeconfigfilepaths}" \
    --feature-gates="${FEATURE_GATES}" \
    --master="https://${API_HOST}:${API_SECURE_PORT}" >"${SCHEDULER_LOG}" 2>&1 &
SCHEDULER_PID=$!