#!/usr/bin/env bash

# This script is responsible for running System Broker.

export HTTP_CLIENT_TIMEOUT="45s"
export HTTP_CLIENT_TLS_HANDSHAKE_TIMEOUT="30s"
export HTTP_CLIENT_IDLE_CONN_TIMEOUT="30s"
export HTTP_CLIENT_RESPONSE_HEADER_TIMEOUT="30s"
export HTTP_CLIENT_DIAL_TIMEOUT="30s"
export HTTP_CLIENT_TIMEOUT="30s"
export HTTP_CLIENT_FORWARD_HEADERS="Authorization"
export HTTP_CLIENT_UNAUTHORIZED_STRING="insufficient scopes provided"

export OAUTH_PROVIDER_WAIT_KUBE_MAPPER_TIMEOUT="1s"

export SERVER_REQUEST_TIMEOUT="60s"
export DIRECTOR_GQL_PAGE_CONCURRENCY=10

go run cmd/main.go
