#!/usr/bin/env bash

OLINGO_JPA_LIB_DIR="$COMPONENT_DIR/olingo-jpa-processor-v4"
OLINGO_VERSION_TAG="0.3.8"

source "$COMPONENT_DIR/scripts/commons.sh"

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
