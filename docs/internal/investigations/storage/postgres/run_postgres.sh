#!/usr/bin/env bash

go run gen.go
docker rm -f my-postgres

docker run --name my-postgres \
    -v ${GOPATH}/src/github.com/kyma-incubator/compass/docs/investigations/storage/postgres/sql:/docker-entrypoint-initdb.d \
    -e POSTGRES_PASSWORD=mysecretpassword \
    -p 5432:5432 \
     postgres

