#!/usr/bin/env bash

COMPONENT_DIR="$(pwd)/$(dirname $0)"
OLINGO_JPA_LIB_DIR="$COMPONENT_DIR/olingo-jpa-processor-v4"
OLINGO_VERSION_TAG="0.3.7-a"

source "$COMPONENT_DIR/commons.sh"

if [[ -d "$OLINGO_JPA_LIB_DIR" ]]
then
    log_section "Olingo JPA library already exists locally. Will attempt to sync it with remote..."
    cd "$OLINGO_JPA_LIB_DIR"
    git checkout "$OLINGO_VERSION_TAG"
    git pull
    cd "$COMPONENT_DIR"
else
    log_section "Pulling Olingo JPA library..."
    git clone --single-branch --branch "$OLINGO_VERSION_TAG" https://github.com/SAP/olingo-jpa-processor-v4.git "$OLINGO_JPA_LIB_DIR"
fi

cd "$OLINGO_JPA_LIB_DIR/jpa/"

log_section "Installing Olingo JPA Library..."
mvn clean install -DskipTests

cd "$COMPONENT_DIR"

log_section "Installing Open Discovery Service..."
mvn clean install -DskipTests

ARTIFACT_VERSION=$(mvn org.apache.maven.plugins:maven-help-plugin:3.2.0:evaluate -Dexpression=project.version -DforceStdout -q)
log_section "Installed Open Discovery Service version: $ARTIFACT_VERSION"
