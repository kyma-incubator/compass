#!/usr/bin/env bash

source "$(dirname $0)/build.sh"

log_section "Starting Open Discovery Service..."
java -jar "$COMPONENT_DIR/target/od-service-$ARTIFACT_VERSION.jar"