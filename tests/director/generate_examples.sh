#!/usr/bin/env bash

make e2e-test
docker cp k3d-kyma-agent-0:/examples/ ../../components/director/

../../components/director/hack/prettify-examples.sh