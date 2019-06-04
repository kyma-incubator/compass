<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-incubator/compass/master/logo.png" width="235">
</p>

# Compass

## Overview
A flexible and easy way to register, manage and group your applications.

## Prerequisities

- [Docker](https://www.docker.com/get-started)
- [Minikube](https://github.com/kubernetes/minikube) 1.0.1
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.12.5
- [Helm](https://github.com/kubernetes/helm) 2.10.0
- [Kyma CLI](https://github.com/kyma-project/cli) master

## Installation

### Chart installation  
If you have already running Kyma 1.1.0 instance with created secrets which contains Tiller client certificates, run:
```bash
helm install --name "compass" ./chart/compass --tls
```

### Local installation with Kyma
To install the Compass along with Kyma 1.1.0, run:
```bash
./installation/scripts/run.sh
```

## Tests

To run tests, install the Compass and run:
```bash
./installation/scripts/testing.sh
```

## Deep dive

To learn more about how the installation and testing are performed, check [this document](./installation/README.md)