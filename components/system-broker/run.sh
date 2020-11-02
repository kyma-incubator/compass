#!/usr/bin/env bash

# This script is responsible for running System Broker.

export HTTP_CLIENT_TIMEOUT="45s"
export HTTP_CLIENT_TLS_HANDSHAKE_TIMEOUT="30s"
export HTTP_CLIENT_IDLE_CONN_TIMEOUT="30s"
export HTTP_CLIENT_RESPONSE_HEADER_TIMEOUT="30s"
export HTTP_CLIENT_DIAL_TIMEOUT="30s"
export HTTP_CLIENT_TIMEOUT="30s"

export SERVER_REQUEST_TIMEOUT="60s"
export DIRECTOR_GQL_PAGE_CONCURRENCY=200

go run cmd/main.go
