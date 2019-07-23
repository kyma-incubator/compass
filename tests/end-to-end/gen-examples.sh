#!/usr/bin/env bash

set -e
set -o errexit
set -o nounset
set -o pipefail

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function cleanup {
   #remove created binaries
   rm "${GOPATH}/src/github.com/kyma-incubator/compass/components/director/directorBin"
   kill ${DIRECTOR_PID}
}

trap cleanup EXIT

##
## Keep examples up-to-date
##
cd "${SCRIPT_DIR}/../../components/director/"
dep ensure -v -vendor-only
go build -o directorBin ./cmd/main.go

./directorBin &
DIRECTOR_PID=$!

cd "${SCRIPT_DIR}"

# wait for Director to be up and running

echo "Checking if Director is up"
directorIsUp=false
set +e
for i in {1..5}; do
    curl --fail  'http://localhost:3000/graphql'  -H 'Content-Type: application/json'  -H 'tenant: any' --data-binary '{"query":"{\n  __schema {\n    queryType {\n      name\n    }\n  }\n}"}'
    res=$?

    if [[ ${res} = 0 ]]
	then
	    directorIsUp=true
	    break
	fi
    sleep 1
done
set -e

if [[ "$directorIsUp" = false ]]; then
    echo "Cannot access Director API"
    exit -1
fi

# remove previous files
rm -f "${GOPATH}/src/github.com/kyma-incubator/compass/examples/"*

go test -c "${SCRIPT_DIR}/director/"
./director.test

img="prettier:latest"
docker build -t ${img} ./tools/prettier
docker run -v "${GOPATH}/src/github.com/kyma-incubator/compass/examples":/prettier/examples \
            ${img} prettier --write ./examples/*.graphql

cd "${SCRIPT_DIR}/tools/example-index-generator/"
EXAMPLES_DIRECTORY="${GOPATH}/src/github.com/kyma-incubator/compass/examples" go run main.go
