#!/usr/bin/env bash

echo "Generating code from GraphQL schema..."

cd "$(dirname "$0")"

cd ./pkg/graphql
GO111MODULE=on go run ../../hack/gqlgen.go

