#!/bin/sh

echo "Generating code from GraphQL schema..."

COMPONENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"


cd "$(dirname "$0")"

cd ${COMPONENT_DIR}/pkg/graphql/externalschema
go run ${COMPONENT_DIR}/hack/gqlgen.go --verbose --config ./config.yaml

cd ${COMPONENT_DIR}/pkg/graphql/internalschema
go run ${COMPONENT_DIR}/hack/gqlgen.go --verbose --config ./config.yaml
