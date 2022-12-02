#!/usr/bin/env bash

#set -e

CURRENT_PATH=$(cd `dirname $0` && pwd)
SOURCE_PATH=${GOPATH}/src/github.com/kyma-incubator/compass
COMMIT_ID=${1}

pushd () {
    command pushd "$@" > /dev/null
}

popd () {
    command popd "$@" > /dev/null
}

process_project(){
    local PROJECT_NAME=${1}

    echo "[1]--------------  Processing: "${PROJECT_NAME}"  --------------"
    if [ -d "${SOURCE_PATH}/${PROJECT_NAME}" ]; then
        pushd "${SOURCE_PATH}/${PROJECT_NAME}"
        declare -a dependencies=$(cat go.mod | grep "github.com/kyma-incubator/compass" | grep -v module | cut -d ' ' -f 1 | xargs)
        for i in "${dependencies[@]}"; do
            DEPENDENCY=$i
            go get -u ${DEPENDENCY}${COMMIT_ID:+"@$COMMIT_ID"}
        done
        popd
    else
        echo "[2]---------------  No such project "${PROJECT_NAME}"...  ---------------"
    fi
    echo "[1]--------------  Finished: "${PROJECT_NAME}"  --------------"
}

if [[ "null" == "${COMMIT_ID}" ]] || [[ -z "${COMMIT_ID}" ]]; then
    echo "Commit ID was not provided. Updating to latest commit from main branch in repo."
    COMMIT_ID=""
else
    echo "Commit ID was provided. Updating to commit ${COMMIT_ID}."
fi
declare -a projects=("gateway" "pairing-adapter" "system-broker" "connector" "external-services-mock" "operations-controller" "connectivity-adapter" "hydrator" "director")
for i in "${projects[@]}"; do
    PROJECT_NAME=$i
    process_project "components/${PROJECT_NAME}"
done

process_project "tests"
