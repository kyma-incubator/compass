#!/usr/bin/env bash

echo "Generating code from GraphQL schema..."

cd "$(dirname "$0")"

cd ./pkg/graphql
go run ../../hack/gqlgen.go

