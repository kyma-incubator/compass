#!/usr/bin/env bash

source "$(dirname $0)/build.sh"

log_section "Starting Open Resource Discovery Service..."
java -jar "$COMPONENT_DIR/target/ord-service-$ARTIFACT_VERSION.jar"