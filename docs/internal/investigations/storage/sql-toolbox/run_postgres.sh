#!/usr/bin/env bash

docker rm -f my-postgres

docker run --name my-postgres \
    -v ${GOPATH}/src/github.com/kyma-incubator/compass/docs/investigations/storage/sql-toolbox/data:/docker-entrypoint-initdb.d \
    -e POSTGRES_PASSWORD=mysecretpassword \
    -p 5432:5432 \
     postgres:12

