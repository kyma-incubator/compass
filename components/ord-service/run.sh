#!/usr/bin/env bash

COMPONENT_DIR="$(pwd)/$(dirname $0)"

source "$COMPONENT_DIR/scripts/install_dependencies.sh"

cd $COMPONENT_DIR
source "$COMPONENT_DIR/scripts/build.sh"

log_section "Starting Open Resource Discovery Service..."
java -jar "$COMPONENT_DIR/target/ord-service-$ARTIFACT_VERSION.jar"