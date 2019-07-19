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
sleep 1

./director.test

kill ${DIRECTOR_PID}

#docker build -f prettier.Dockerfile -t prettier:latest .
#docker run -v ${GOPATH}/src/github.com/kyma-incubator/compass/examples:/prettier/examples  prettier:latest prettier ./examples/*.graphql