#!/usr/bin/env bash

# This script is responsible for running System Broker.

../director/run.sh
go run cmd/main.go
