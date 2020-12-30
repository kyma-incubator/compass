#!/usr/bin/env bash

OPTS=$1

source "$COMPONENT_DIR/scripts/commons.sh"

log_section "Installing Open Resource Discovery Service..."
mvn clean install $OPTS

ARTIFACT_VERSION=$(mvn org.apache.maven.plugins:maven-help-plugin:3.2.0:evaluate -Dexpression=project.version -DforceStdout -q)
log_section "Installed Open Resource Discovery Service version: $ARTIFACT_VERSION"