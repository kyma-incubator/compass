# Kyma Control Plane

## Overview

> **:warning: WARNING:** This repo is in a very early stage. Use it at your own risk.

Kyma Control Plane is a central system to manage Kyma Runtimes.

## Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Minikube](https://github.com/kubernetes/minikube) 1.0.1
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.12.5
- [Kyma CLI](https://github.com/kyma-project/cli) stable

## Installation

### Local installation with Kyma

To install Kyma Control Plane with the minimal Kyma installation, Compass and Kyma Control Plane, run this script:
```bash
./installation/cmd/run.sh
```

You can also specify Kyma version, such as 1.6 or newer:
```bash
./installation/cmd/run.sh {version}
```

### Testing

Kyma Control Plane, as a part of Kyma, uses [Octopus](https://github.com/kyma-incubator/octopus/blob/master/README.md) for testing. To run the Kyma Control Plane tests, run:

```bash
./installation/scripts/testing.sh
```

Read [this](https://kyma-project.io/docs/root/kyma#details-testing-kyma) document to learn more about testing in Kyma.
