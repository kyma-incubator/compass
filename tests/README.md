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

### Execution steps for sandbox tests
The sandbox tests give us a possibility to run a single e2e test and in case of error we can modify that test and rerun it again without waiting the new image to be build. That way we can easily and fast iterate and test different scenarios. In the [Local Makefile](#local-makefile) section you can find a brief description for each command. Execution steps:
1. `make sandbox-test`, executed once in the beginning to setup the environment
2. `make run testName=<test-name>` to run a specific test. For example `make run testName=TestConsumerProviderFlow` will run the whole test suite but if there are any nested tests inside the suite you can specify which one of them you want to execute. The test suite is separated from the inner tests with `/`. If the test run name containts spaces they need to be replaces with `_`. Example: `make run testName=TestConsumerProviderFlow/ConsumerProvider_flow:_calls_with_provider_certificate_and_consumer_token_are_successful_when_valid_subscription_exists`
3. Optionally, you can modify the existing test and redeploy it using the following command: `make sandbox-deploy-test`. Step 2 and 3 can repeat as much as needed.
4. Finally, to tear down the setup execute `make sandbox-test-clean`