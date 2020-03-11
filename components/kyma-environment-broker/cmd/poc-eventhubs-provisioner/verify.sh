#!/usr/bin/env bash
go build -o provisioner provisioner.go
./provisioner
rm -f ./provisioner
