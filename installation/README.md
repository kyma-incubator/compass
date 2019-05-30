# Compass installation and testing

## Overview

This document presents the deep dive into Compass installation and testing.

## Installation

To install the Compass, run:
```bash
./installation/scripts/run.sh
```

The `run.sh` scripts do the following:
1. Provision local Kubernetes cluster on Minikube adjusted for Kyma installation via `Kyma CLI`.
2. Install Kyma on the cluster with hardened list of components provided in `./installation/resources/installer-cr.yaml` file.  
3. Run `tiller-tls.sh` script to download and use Helm Client Certs created with Kyma installation.
4. Perform installation of `compass` chart.

## Testing

To run the Compass tests, run:

```bash
./installation/scripts/testing.sh
```

In Compass, [Octopus](https://github.com/kyma-incubator/octopus/) is used as a testing framework. \
To learn how to work with Octopus, check [that document](https://github.com/kyma-project/kyma/blob/master/docs/kyma/03-03-testing.md).
 
> **NOTE:** After adding a new test, remember to add it also to `ClusterTestSuite` inside `testing.sh` script.