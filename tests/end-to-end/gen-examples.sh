#!/usr/bin/env bash

set -e
set -o errexit
set -o nounset
set -o pipefail

#
# Keep examples up-to-date
#
cd ./../../components/director/
dep ensure -v -vendor-only
go build -o directorBin ./cmd/main.go

./directorBin &
DIRECTOR_PID=$!

cd -
sleep 1 # wait for director to be up and running
./director.test

kill ${DIRECTOR_PID}
#remove binary
rm ./../../components/director/directorBin

img="prettier:latest"
docker build -t ${img} ./tools/prettier
docker run -v ${GOPATH}/src/github.com/kyma-incubator/compass/examples:/prettier/examples \
            ${img} prettier --write ./examples/*.graphql

cd ./tools/example-index-generator/
env EXAMPLES_DIRECTORY=${GOPATH}/src/github.com/kyma-incubator/compass/examples go run main.go

cd -