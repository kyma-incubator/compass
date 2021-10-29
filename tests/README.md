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
The sandbox tests enable the run of a single e2e test, and in case of an error, allow you to modify that test and rerun it again, without waiting and building a new image. This facilitates the iteration and testing of different scenarios. You can find a brief description for each command in the [Local Makefile](#local-makefile) section.

Execution steps:
1. To set up the environment, carry out the following command once in the beginning: `make sandbox-test`.
2. To run a specific test, use the following command: `make run testName=<test-name>`.
<br>Note that, `make run testName=TestConsumerProviderFlow` runs the whole test suite. However, if there are any nested tests in the suite, you can specify the one you want to carry out. To separate the test suite and the inner tests, you can use forward slash `/`. If the test name contains any spaces, you must replace them with underscore characters `_`.
<br>Example:
<br>`make run testName=TestConsumerProviderFlow/ConsumerProvider_flow:_calls_with_provider_certificate_and_consumer_token_are_successful_when_valid_subscription_exists`

3. Optionally, you can modify the existing test and redeploy it using the following command: `make sandbox-deploy-test`.
4. Repeat steps 2 and 3 as many times as needed.
5. Finally, to clean up the setup, carry out the following command: `make sandbox-test-clean`.
