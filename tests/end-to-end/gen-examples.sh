#!/usr/bin/env bash

#
# Keep examples up-to-date
#
cd ./../../components/director/
dep ensure -v -vendor-only
go build -o director ./cmd/main.go

./director &
DIRECTOR_PID=$!

cd -
sleep 1 # wait for director to be up and running
./director.test

kill ${DIRECTOR_PID}

img="prettier:latest"
docker build -t ${img} ./prettier
docker run -v ${GOPATH}/src/github.com/kyma-incubator/compass/examples:/prettier/examples \
            ${img} prettier --write ./examples/*.graphql