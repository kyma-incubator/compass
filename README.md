<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-incubator/compass/master/logo.png" width="235">
</p>

# Compass

## Overview

Compass (also known as Management Plane Services) is a multi-tenant system which consists of components that provide a way to register, group, and manage your applications across multiple Kyma runtimes. Using Compass, you can control and monitor your application landscape in one central place.

Compass allows for registering different types of applications and runtimes.
These are the types of applications due to the way of integration with Compass:
- basic - administrator manually provides API/Events Metadata to Compass. This type of integration is used mainly for simple use-case scenarios and doesn't support all features.
- application - integration with Compass is built-in inside the application.
- proxy - a highly application-specific proxy component provides the integration.
- service -  a central service provides the integration for a class of applications. It manages multiple instances of these applications. You can integrate multiple services to support different types of applications.

You can also register different types of runtimes providing that they allow for integration with Compass. Your runtimes must allow for fetching application definitions and using applications in these runtimes. The example runtimes are Kyma, CP CloudFoundry, Serverless, etc.

The example usage is Wordpress integration with Kyma Runtime. (?)

This project also contains Compass UI Cockpit that exposes Compass APIs to users. (?)

For more information about the Compass architecture and technical details, read the [documentation](./docs).


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

### Tests

To run tests, install the Compass and run:
```bash
./installation/scripts/testing.sh
```

To learn more about how the installation and testing are performed, check [this document](./installation/README.md)

## Usage

## Development
