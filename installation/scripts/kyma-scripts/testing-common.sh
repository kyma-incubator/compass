#Script downloaded from https://raw.githubusercontent.com/kyma-project/kyma/1.1.0/installation/scripts/utils.sh

ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${ROOT_PATH}/utils.sh
# Compass' utils
source ${ROOT_PATH}/../utils.sh

function cmdGetPodsForSuite() {
    local suiteName=$1
    KUBECTL="$2"

    cmd="${KUBECTL} get pods -l testing.kyma-project.io/suite-name=${suiteName} \
            --all-namespaces \
            --no-headers=true \
            -o=custom-columns=name:metadata.name,ns:metadata.namespace"
    echo $cmd
}

function printLogsFromFailedTests() {
    local suiteName=$1
    KUBECTL="$2"
    cmd=$(cmdGetPodsForSuite "$suiteName" "$KUBECTL")

    pod=""
    namespace=""
    idx=0

    for podOrNs in $($cmd)
    do
        n=$((idx%2))
         if [[ "$n" == 0 ]];then
            pod=${podOrNs}
            idx=$((${idx}+1))
            continue
        fi
        namespace=${podOrNs}
        idx=$((${idx}+1))

        log "Testing '${pod}' from namespace '${namespace}'" nc bold

        phase=$("$KUBECTL"  get pod ${pod} -n ${namespace} -o jsonpath="{ .status.phase }")

        case ${phase} in
        "Failed")
            log "'${pod}' has Failed status" red
            printLogsFromPod ${namespace} ${pod} ${KUBECTL}
        ;;
        "Running")
            log "'${pod}' failed due to too long Running status" red
            printLogsFromPod ${namespace} ${pod} ${KUBECTL}
        ;;
        "Pending")
            log "'${pod}' failed due to too long Pending status" red
            printf "Fetching events from '${pod}':\n"
            "$KUBECTL"  describe po ${pod} -n ${namespace} | awk 'x==1 {print} /Events:/ {x=1}'
        ;;
        "Unknown")
            log "'${pod}' failed with Unknown status" red
            printLogsFromPod ${namespace} ${pod} ${KUBECTL}
        ;;
        "Succeeded")
            # do nothing
        ;;
        *)
            log "Unknown status of '${pod}' - ${phase}" red
            printLogsFromPod ${namespace} ${pod} ${KUBECTL}
        ;;
        esac
        log "End of testing '${pod}'\n" nc bold

    done
}

function getContainerFromPod() {
    local namespace="$1"
    local pod="$2"
    KUBECTL="$3"
    local containers2ignore="istio-init istio-proxy manager"
    containersInPod=$("$KUBECTL" get pods ${pod} -o jsonpath='{.spec.containers[*].name}' -n ${namespace})
    for container in $containersInPod; do
        if [[ ! ${containers2ignore[*]} =~ "${container}" ]]; then
            echo "${container}"
        fi
    done
}

function printLogsFromPod() {
    local namespace=$1 pod=$2
    KUBECTL="$3"
    log "Fetching logs from '${pod}'" nc bold
    testPod=$(getContainerFromPod ${namespace} ${pod} ${KUBECTL})
    result=$("$KUBECTL" logs -n ${namespace} -c ${testPod} ${pod})
    if [ "${#result}" -eq 0 ]; then
        log "FAILED" red
        return 1
    fi
    echo "${result}"
}

function checkTestPodTerminated() {
    local suiteName=$1
    KUBECTL="$2"
    runningPods=false

    pod=""
    namespace=""
    idx=0

    cmd=$(cmdGetPodsForSuite "$suiteName" "$KUBECTL")
    for podOrNs in $($cmd)
    do
       n=$((idx%2))
       if [[ "$n" == 0 ]];then
         pod=${podOrNs}
         idx=$((${idx}+1))
         continue
       fi
        namespace=${podOrNs}
        idx=$((${idx}+1))

        phase=$("$KUBECTL"  get pod "$pod" -n ${namespace} -o jsonpath="{ .status.phase }")
        # A Pod's phase  Failed or Succeeded means pod has terminated.
        # see: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase
        if [ "${phase}" !=  "Succeeded" ] && [ "${phase}" != "Failed" ]
        then
          log "Test pod '${pod}' has not terminated, pod phase: ${phase}" red
          runningPods=true
        fi
    done

    if [ ${runningPods} = true ];
    then
        return 1
    fi
}

function waitForTestPodsTermination() {
    local retry=0
    local suiteName=$1
    KUBECTL="$2"

    log "All test pods should be terminated. Checking..." nc bold
    while [ ${retry} -lt 3 ]; do
        checkTestPodTerminated ${suiteName} ${KUBECTL}
        checkTestPodTerminatedErr=$?
        if [ ${checkTestPodTerminatedErr} -ne 0 ]; then
            echo "Waiting for test pods to terminate..."
            sleep 1
        else
            log "OK" green bold
            return 0
        fi
        retry=$[retry + 1]
    done
    log "FAILED" red
    return 1
}

function waitForTerminationAndPrintLogs() {
    local suiteName=$1
    KUBECTL="$2"

    waitForTestPodsTermination ${suiteName} ${KUBECTL}
    checkTestPodTerminatedErr=$?

    printLogsFromFailedTests ${suiteName} ${KUBECTL}
    if [ ${checkTestPodTerminatedErr} -ne 0 ]
    then
        return 1
    fi
}

function printImagesWithLatestTag() {
    KUBECTL="$1"
    local images=$("$KUBECTL"  get pods --all-namespaces -o jsonpath="{..image}" |\
    tr -s '[[:space:]]' '\n' |\
    grep ":latest")

    log "Images with tag latest are not allowed. Checking..." nc bold
    if [ ${#images} -ne 0 ]; then
        log "${images}" red
        log "FAILED" red
        return 1
    fi
    log "OK" green bold
    return 0
}