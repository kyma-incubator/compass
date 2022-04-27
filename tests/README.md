# Compass tests

Compass tests comprise end-to-end tests for the Compass project, as follows:
- The tests of each component are placed into a subdirectory with the corresponding title.
- All utilities directories are placed into the `pkg` directory.

## Prerequisites

To run the tests, it is required a running Minikube instance with installed Kyma and Compass.

## Usage

The global Dockerfile builds all component test directories and adds the binaries to the image. 

The tests are run via Octopus. The Octopus test definitions can be found in `kyma-incubator/compass/chart/compass/templates/tests`.

## Make targets

### Global Makefile
The global Makefile is located in the root tests directory `compass/tests/Makefile`, and supports the following commands:

- `deploy-tests-on-minikube` - Pushes a new image version of the tests into the Minikube cluster.
- `e2e-test` - Creates a new cluster-test-suite matching all test-definitions. Then, runs the tests and provides information about the current status. Finally, cleans up after the test is carried out.
- `e2e-test-clean` - In case of early termination of the `e2e-test`, the command cleans up the `cluster-test-suite` and `test-definition` created by the `e2e-test` after the test is carried out.

### Local Makefile
Each Compass component has its own test directory. It contains a local Makefile that supports the following commands:

- `e2e-test` - Creates a new cluster-test-suite matching the given component's test-definition. Then, runs the tests and provides information about the current status. Finally, cleans up after the test is carried out.
- `e2e-test-clean` - In case of early termination of the `e2e-test`, the command cleans up the `cluster-test-suite` and `test-definition` created by the `e2e-test` after the test is carried out.
- `sandbox-test` - Creates a copy of the component's test-definition with changed command and installs go into the pod which is created by Octopus.
- `run` - Runs the specified test. For example: `make testName=TestFullConnectorFlow run`.
- `bench-run` - Runs the specified benchmark test. For example: `make bench-run testName=BenchmarkSystemBundles`.
- `sandbox-deploy-test` - Creates new binary for the components tests and pushes it into the cluster.
- `sandbox-deploy-bench-test` - Creates new binary for the benchmark tests and pushes it into the cluster.
- `sandbox-test-clean` - Deletes the cluster-test-suite and test-definition created by `sandbox-test`.

### Execution steps for sandbox tests
The sandbox tests enable the run of a single E2E test case of a given Compass component. In case of an error, you can modify the failed test and rerun it again, without building a new image and waiting for the whole test suite to finish. This facilitates the iteration and testing of different scenarios. You can find a brief description for each command in the [Local Makefile](#local-makefile) section.

**Prerequisites:**
To run the sandbox tests, you can use the [`sponge`](https://rentes.github.io/unix/utilities/2015/07/27/moreutils-package/#sponge) CLI tool.
You can install it on Mac OS via Homebrew:
```bash
brew install sponge
```

**Execution steps:**
1. To set up the environment, carry out the following command once in the beginning:
    ```bash
    make sandbox-test
    ```
2. Once you have modified the existing tests, you should redeploy them by using the following command:
    ```bash
    make sandbox-deploy-test
    ```
3. To run a specific test, use:
    ```bash
    make run testName=<test-name>
    ```
    Note that, `make run testName=TestConsumerProviderFlow` runs the whole test suite. However, if there are any nested tests in the suite, you can specify the one you want to carry out. To set apart the entire test suite and run any nested tests, you can use forward slash `/`. If the test name contains any spaces, you must replace them with underscore characters `_`.
    
    **Example:**
    ```bash
    make run testName=TestConsumerProviderFlow/ConsumerProvider_flow:_calls_with_provider_certificate_and_consumer_token_are_successful_when_valid_subscription_exists
    ```

4. Repeat steps 2 and 3 as many times as needed.
5. Finally, to clean up the setup, carry out the following command: `make sandbox-test-clean`.
