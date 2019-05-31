# Compass installation and testing

## Overview

This document presents the deep dive into Compass installation and testing.

## Installation

If you have already running Kyma instance with created secrets which contains Tiller client certificates, run:
```bash
helm install --name "compass" "${ROOT_PATH}"/chart/compass --tls
```

To install the Compass, run:
```bash
./installation/scripts/run.sh
```

The `run.sh` scripts do the following:
1. Provision local Kubernetes cluster on Minikube adjusted for Kyma installation via `Kyma CLI`.
2. Install Kyma on the cluster with hardened list of components provided in `./installation/resources/installer-cr.yaml` file.  
3. Download Helm client certificates created with Kyma installation.
4. Perform installation of `compass` Helm chart.

## Testing

To run the Compass tests, run:

```bash
./installation/scripts/testing.sh
```

Compass, as a part of Kyma, uses [Octopus](https://github.com/kyma-incubator/octopus/) for testing. Learn more in [Testing Kyma](https://kyma-project.io/docs/root/kyma#details-testing-kyma) section.

> **NOTE:** After adding a new test, remember to add it also to `ClusterTestSuite` inside `testing.sh` script.