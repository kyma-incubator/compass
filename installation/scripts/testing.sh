#!/usr/bin/env bash
CURRENT_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
# Currently not used but kept for easier transition to enable db dump
DUMP_DB=false

source "${CURRENT_PATH}"/kyma-scripts/testing-common.sh
source "${CURRENT_PATH}"/utils.sh

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
        # Currently not used but kept for easier transition to enable db dump
        --dump-db)
            DUMP_DB=true
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

# Benchmark tests are executed in a GCP environment not k3d
# All other tests should be executed facing k3d kyma
KUBECTL="kubectl_k3d_kyma"
if [[ "${BENCHMARK}" == "true" ]]
then
  benchmarkLabelSelector='benchmark'
  KUBECTL="kubectl"
fi

echo "${1:-All tests}"
echo "----------------------------"
echo "- Testing Compass..."
echo "----------------------------"

"$KUBECTL" get clustertestsuites.testing.kyma-project.io > /dev/null 2>&1
if [[ $? -eq 1 ]]
then
   echo "ERROR: script requires ClusterTestSuite CRD"
   exit 1
fi

"$KUBECTL" get cts  ${suiteName} -ojsonpath="{.metadata.name}" > /dev/null 2>&1
if [[ $? -eq 0 ]]
then
   echo "ERROR: Another ClusterTestSuite CRD is currently running in this cluster."
   exit 1
fi

# match all tests
if [ -z "$testDefinitionName" ]
then
  labelSelector=''
  if [ "$DUMP_DB" = true ]
  then
    labelSelector=',!disable-db-dump'
  fi
      cat <<EOF | "$KUBECTL" apply -f -
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
      cat <<EOF | "$KUBECTL" apply -f -
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
    statusSucceeded=$("$KUBECTL" get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Succeeded')]}")
    statusFailed=$("$KUBECTL" get cts ${suiteName}  -ojsonpath="{.status.conditions[?(@.type=='Failed')]}")
    statusError=$("$KUBECTL" get cts  ${suiteName} -ojsonpath="{.status.conditions[?(@.type=='Error')]}" )

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
        running_test=$("$KUBECTL" get cts ${suiteName} -o yaml | grep "status: Running" -B5 | head -n 1 | cut -d ':' -f 2 | tr -d " ")
        director_pod=$("$KUBECTL" get pods -n compass-system | grep compass-director | head -n 1 | cut -d ' ' -f 1)
        if [ -z "$running_test" ]; then
          running_test="none"
        fi
        echo "Running test is ${running_test}"
        if [[ ! "${running_test}" == "none" ]]; then
          logs_from_last_min=$("$KUBECTL" logs -n kyma-system --since=1m ${running_test})
          echo "Logs from test execution:"
          echo "${logs_from_last_min}"
          echo "----------------------------"
          echo "Director:"
          echo "----------------------------"
          "$KUBECTL" logs -n compass-system --since=1m ${director_pod}
          echo "----------------------------"
          if echo "${logs_from_last_min}" | grep -q "FAIL:"; then
            echo "----------------------------"
            echo "A test has failed in the last minute."
            echo "Logs from test execution:"
            echo "${logs_from_last_min}"
            echo "----------------------------"
            echo "Director:"
            echo "----------------------------"
            echo "${logs_from_director_last_min}"
            echo "----------------------------"
          fi
        fi
        previousPrintTime=${min}
    fi
    sleep 3
done

echo "Test summary"
"$KUBECTL" get cts  ${suiteName} -o=go-template --template='{{range .status.results}}{{printf "Test status: %s - %s" .name .status }}{{ if gt (len .executions) 1 }}{{ print " (Retried)" }}{{end}}{{print "\n"}}{{end}}'

waitForTerminationAndPrintLogs ${suiteName} ${KUBECTL}
cleanupExitCode=$?

echo "ClusterTestSuite details:"
"$KUBECTL" get cts ${suiteName} -oyaml

echo "Pod execution time details:"
podInfo=$("$KUBECTL" get cts ${suiteName} -o=go-template --template='{{range .status.results}}{{range .executions }}{{printf "%s %s %s\n" .id .startTime .completionTime }}{{end}}{{end}}')

if [ "$(uname)" == "Darwin" ]; then
  extra_flags="-j -f %Y-%m-%dT%H:%M:%SZ"
else
  extra_flags="-D %Y-%m-%dT%H:%M:%SZ -d"
fi

while read -r podName startTime endTime;
do
  startTimeTimestamp=$(date $extra_flags "$startTime" +%s)
  endTimeTimestamp=$(date $extra_flags "$endTime" +%s)
  duration=$((endTimeTimestamp - startTimeTimestamp))

  min=$((duration/60))
  sec=$((duration%60))
  minString=""
  if ((min > 0)); then
    minString="${min}m"
  fi
  echo "$podName execution time: ${minString}${sec}s"
done <<< "$podInfo"

if [[ ! "${BENCHMARK}" == "true" ]]
then
  "$KUBECTL" delete cts ${suiteName}
fi

printImagesWithLatestTag "$KUBECTL"
latestTagExitCode=$?

exit $((${testExitCode} + ${cleanupExitCode} + ${latestTagExitCode}))
