#!/usr/bin/env bash

COMPONENT_DIR="$GOPATH/src/github.com/kyma-incubator/compass/components/od-service"
OLINGO_JPA_LIB_DIR="$COMPONENT_DIR/olingo-jpa-processor-v4"
OLINGO_VERSION_TAG=0.3.7-a

source "$COMPONENT_DIR/formatting.sh"

if [[ -d "$OLINGO_JPA_LIB_DIR" ]]
then
    log_section "Olingo JPA library already exists locally. Will attempt to sync it with remote..."
    cd "$OLINGO_JPA_LIB_DIR"
    git checkout 0.3.7-a
    git pull
    cd "$COMPONENT_DIR"
else
    log_section "Pulling Olingo JPA library..."
    git clone --single-branch --branch "$OLINGO_VERSION_TAG" https://github.com/SAP/olingo-jpa-processor-v4.git
fi

cd "$OLINGO_JPA_LIB_DIR/jpa/"

log_section "Installing Olingo JPA Library..."
mvn clean install -DskipTests

cd "$COMPONENT_DIR"

log_section "Installing Open Discovery Service..."
mvn clean install -DskipTests

log_section "Starting Open Discovery Service..."
java -jar "$COMPONENT_DIR/target/od-service-0.0.1-SNAPSHOT.jar"
