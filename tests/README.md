# Compass tests

Compass tests comprise end-to-end tests for the Compass project, as follows:
- The tests of each component are placed into a subdirectory with the corresponding title.
- All utilities directories are placed into the `pkg` directory.

## Prerequisites

To run the tests, it is required a running minikube instance with installed Kyma and Compass.

## Usage

The global Dockerfile builds all component test directories and adds the binaries to the image. 

The tests are run via Octopus. The Octopus test definitions can be found in `kyma-incubator/compass/chart/compass/templates/tests`.

## Make targets

### Global Makefile
The global Makefile comprises the following commands:

- `deploy-tests-on-minikube` - Pushes a new version of the tests into the Minikube cluster.
- `e2e-test` - Creates a new cluster-test-suite matching all test-definitions. Then, runs the tests and provides information about the current status. Finally, cleans up after the test is carried out.
- `e2e-test-clean` - In case of early termination of the `e2e-test`, the command cleans up the `cluster-test-suite` and `test-definition` created by the `e2e-test` after the test is carried out.

### Local Makefile
Each component test directory contains local Makefile that comprises the following commands:

- `e2e-test` - Creates a new cluster-test-suite matching the component's test-definition. Then, runs the tests and provides information about the current status. Finally, cleans up after the test is carried out.
- `e2e-test-clean` - In case of early termination of the `e2e-test`, the command cleans up the `cluster-test-suite` and `test-definition` created by the `e2e-test` after the test is carried out.
- `sandbox-test` - Creates a copy of the component's test-definition with changed command and installs go into the pod which is created by Octopus.
- `run` - Runs the specified test. For example: `make testName=TestFullConnectorFlow run`.
- `sandbox-deploy-test` - Creates new binary for the components tests and pushes it into the cluster.
- `sandbox-test-clean` - Deletes the cluster-test-suite and test-definition created by `sandbox-test`.
