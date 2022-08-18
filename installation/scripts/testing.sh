#!/usr/bin/env bash
ROOT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

source ${ROOT_PATH}/kyma-scripts/testing-common.sh

readonly TMP_DIR=$(mktemp -d)

echo "ARTIFACTS: ${ARTIFACTS}"
readonly JUNIT_REPORT_PATH="${ARTIFACTS:-${TMP_DIR}}/junit_compass_octopus-test-suite.xml"

function checkInputParameterValue() {
    if [ -z "${1}" ] || [ "${1:0:2}" == "--" ]; then
        echo "Wrong parameter value"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --dump-db)
            checkInputParameterValue "${2}"
            DUMP_DB="$2"
            shift
            shift
        ;;
        --benchmark)
            checkInputParameterValue "${2}"
            BENCHMARK="$2"
            shift
            shift
        ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

suiteName="compass-e2e-tests"
testDefinitionName=$1
benchmarkLabelSelector="!benchmark"
echo "${1:-All tests}"
echo "----------------------------"
echo "- Testing Compass..."
echo "----------------------------"

kc="kubectl $(context_arg)"

${kc} get clustertestsuites.testing.kyma-project.io > /dev/null 2>&1
if [[ $? -eq 1 ]]
then
   echo "ERROR: script requires ClusterTestSuite CRD"
   exit 1
fi

# match all tests
if [ -z "$testDefinitionName" ]
then
  if [[ "${DUMP_DB}" == "true" ]]
  then
    labelSelector=',!disable-db-dump'
  elif [[ "${DUMP_DB}" == "false" ]]
  then
    labelSelector=',disable-db-dump'
  fi
  if [[ "${BENCHMARK}" == "true" ]]
  then
    benchmarkLabelSelector='benchmark'
  fi
      cat <<EOF | ${kc} apply -f -
      apiVersion: testing.kyma-project.io/v1alpha1
      kind: ClusterTestSuite
      metadata:
        labels:
          controller-tools.k8s.io: "1.0"
        name: ${suiteName}
      spec:
        maxRetries: 1
        concurrency: 1
        selectors:
          matchLabelExpressions:
            - "${benchmarkLabelSelector}${labelSelector}"
EOF

else
      cat <<EOF | ${kc} apply -f -
      apiVersion: testing.kyma-project.io/v1alpha1
      kind: ClusterTestSuite
      metadata:
        labels:
          controller-tools.k8s.io: "1.0"
        name: ${suiteName}
      spec:
        maxRetries: 1
        concurrency: 1
        selectors:
          matchNames:
            - name: compass-e2e-${testDefinitionName}
              namespace: kyma-system
EOF

fi

startTime=$(date +%s)

testExitCode=0
previousPrintTime=-1

while true
do
    currTime=$(date +%s)
    statusSucceeded=$(${kc} get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    statusFailed=$(${kc} get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Failed')]}")
    statusError=$(${kc} get cts  ${suiteName} -ojsonpath="{.status.conditions[?(@.type=='Error')]}" )

    if [[ "${statusSucceeded}" == *"True"* ]]; then
       echo "Test suite '${suiteName}' succeeded."
       break
    fi

    if [[ "${statusFailed}" == *"True"* ]]; then
        echo "Test suite '${suiteName}' failed."
        testExitCode=1
        break
    fi

    if [[ "${statusError}" == *"True"* ]]; then
        echo "Test suite '${suiteName}' errored."
        testExitCode=1
        break
    fi

    sec=$((currTime-startTime))
    min=$((sec/60))
    if (( min > 120 )); then
        echo "Timeout for test suite '${suiteName}' occurred."
        testExitCode=1
        break
    fi
    if (( ${previousPrintTime} != ${min} )); then
        echo "ClusterTestSuite not finished. Waiting..."
        previousPrintTime=${min}
    fi
    sleep 3
done

echo "Test summary"
kubectl get cts  ${suiteName} -o=go-template --template='{{range .status.results}}{{printf "Test status: %s - %s" .name .status }}{{ if gt (len .executions) 1 }}{{ print " (Retried)" }}{{end}}{{print "\n"}}{{end}}'

waitForTerminationAndPrintLogs ${suiteName}
cleanupExitCode=$?

echo "ClusterTestSuite details:"
kubectl get cts ${suiteName} -oyaml

echo "Generate JUnit test summary (${JUNIT_REPORT_PATH})"
kyma test status "${suiteName}" -ojunit | sed 's/ (executions: [0-9]*)"/"/g' > "${JUNIT_REPORT_PATH}"

if [[ ! "${BENCHMARK}" == "true" ]]
then
  kubectl delete cts ${suiteName}
fi

printImagesWithLatestTag
latestTagExitCode=$?

exit $((${testExitCode} + ${cleanupExitCode} + ${latestTagExitCode}))
