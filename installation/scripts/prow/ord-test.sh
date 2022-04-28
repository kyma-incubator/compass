#!/bin/bash

###
# Following script installs necessary tooling for Debian, starts Director and Ord service, and runs the smoke tests.
#

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INSTALLATION_DIR=${CURRENT_DIR}/../../

export ARTIFACTS="/var/log/prow_artifacts"
sudo mkdir -p "${ARTIFACTS}"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
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


#sudo ${INSTALLATION_DIR}/cmd/run.sh
echo "Install java"
export JAVA_HOME="/usr/local/openjdk-11"
mkdir -p "$JAVA_HOME"
export PATH="$JAVA_HOME/bin:$PATH"
curl -fLSs -o adoptopenjdk11.tgz "https://github.com/AdoptOpenJDK/openjdk11-binaries/releases/download/jdk-11.0.9.1%2B1/OpenJDK11U-jdk_x64_linux_hotspot_11.0.9.1_1.tar.gz"


tar --extract --file adoptopenjdk11.tgz --directory "$JAVA_HOME" --strip-components 1 --no-same-owner
rm adoptopenjdk11.tgz* 
echo "-----------------------------------"
java -version
echo "-----------------------------------"
echo "pwd:"
pwd
echo "-----------------------------------"
echo "tree ."
tree .
echo "-----------------------------------"
echo "tree /"
tree /
echo "-----------------------------------"
echo "CURRENT_DIR=${CURRENT_DIR}"
echo "INSTALLATION_DIR=${INSTALLATION_DIR}"
echo "ARTIFACTS=${ARTIFACTS}"
echo "ord-test end reached!"
