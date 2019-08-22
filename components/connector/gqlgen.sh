#!/bin/sh

echo "Generating code from GraphQL schema..."

COMPONENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"


cd "$(dirname "$0")"

pushd ./pkg/graphql/externalschema
go run ${COMPONENT_DIR}/hack/gqlgen.go -v --config ./config.yaml
popd

pushd ./pkg/graphql/internalschema
go run ${COMPONENT_DIR}/hack/gqlgen.go -v --config ./config.yaml
popd