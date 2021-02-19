#!/usr/bin/env bash

echo "Generating code from GraphQL schema..."

cd "$(dirname "$0")"

cd ./pkg/graphql
GO111MODULE=on go run ../../hack/gqlgen.go

cd internalschema
go run ../../../hack/gqlgen_internal.go -v --config ./config.yaml