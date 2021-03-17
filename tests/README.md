# Compass tests

Compass tests consists of end-to-end tests for Compass project:
- each component's tests are placed into subdirectory named after the component 
- all util directories are placed into `pkg`

## Usage

The global Dockerfile builds all component test directories and adds the binaries to the image. 
To run the tests running minikube with installed kyma and Compass is needed.

The tests are run using Octopus. The Octopus test definitions can be found in `kyma-incubator/compass/chart/compass/templates/tests`.

## Make targets

### Global Makefile
- `deploy-tests-on-minikube` - push a new version of the tests into the Minikube cluster
- `e2e-test` - create a new cluster-test-suite matching all test-definitions, runs the tests and provides information about the current status, finally cleans up after the test's execution is finished
- `e2e-test-clean` - clean up after the test's execution is finished in case of abnormal termination of `e2e-test`

### Local Makefile
Each component's test directory contains local Makefile.

- `e2e-test` - create a new cluster-test-suite matching the component's test-definition, runs the tests and provides information about the current status, finally cleans up after the test's execution is finished
- `e2e-test-clean` - clean up after the test's execution is finished in case of abnormal termination of `e2e-test`
- `sandbox-test` - creates copy of the component's test-definition with changed command and installs go into the pod which is created by Octopus
- `run` - runs the specified test. Example `make testName=TestFullConnectorFlow run`
- `sandbox-deploy-test` - creates new binary for the components tests and pushes it into the cluster
- `sandbox-test-clean` - deletes the cluster-test-suite and test-definition created by `sandbox-test`
