#!/usr/bin/env bash

set -eu

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "${CURRENT_DIR}/../utilities.sh" || { echo 'Cannot load utilities file.'; exit 1; }

assert_tiller_is_up() {
	LIMIT=60
  COUNTER=0

  while [ ${COUNTER} -lt ${LIMIT} ]; do
    if [[ $(kubectl get deployment -n kube-system tiller-deploy -ojson | jq '.status.availableReplicas') == 1 ]];then
      print_ok "Tiller is up"
      return 0
    else
      print_warning "Tiller is not ready"
    fi
    (( COUNTER++ ))
    echo -e "Tiller is not ready yet, retry (${COUNTER} attempt out of ${LIMIT})..."
    sleep 1
  done
}

assert_kyma_is_up() {
  while true; do \
    state=$(kubectl -n default get installation/kyma-installation -ojsonpath="{.status.state}")
    desc=$(kubectl -n default get installation/kyma-installation -ojsonpath="{.status.description}")
    echo "Status: ${state}, description: ${desc}"

    if [[ "${state}" == "Installed" && "${desc}" == "Kyma installed" ]];then
      print_ok "Kyma is installed"
      exit 0
    fi
    sleep 5;
  done
}

print_ok "Install tiller"
kubectl apply -f "${CURRENT_DIR}/tiller.yaml"
assert_tiller_is_up

print_ok "Install Kyma"
kubectl apply -f "${CURRENT_DIR}/kyma-installer.yaml"
assert_kyma_is_up
