#!/usr/bin/env bash

source "$(dirname $0)/scripts/install_dependencies.sh"
source "$(dirname $0)/scripts/build.sh"

log_section "Starting Open Resource Discovery Service..."
java -jar "$COMPONENT_DIR/target/ord-service-$ARTIFACT_VERSION.jar"