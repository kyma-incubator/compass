#!/bin/sh

echo "Generating code from GraphQL schema..."

cd "$(dirname "$0")"

cd ./internal/graphql
go run ../../hack/gqlgen.go -v --config ./config.yaml
