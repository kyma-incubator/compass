#!/usr/bin/env bash

make e2e-test
make e2e-test-application
make e2e-test-notification
make e2e-test-runtime
make e2e-test-formation

./copy-examples.sh