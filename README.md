<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-incubator/compass/master/logo.png" width="235">
</p>

# Compass

## Overview

Compass (also known as Management Plane Services) is a multi-tenant system which consists of components that provide a way to register, group, and manage your applications across multiple Kyma runtimes. Using Compass, you can control and monitor your application landscape in one central place.

Compass allows for registering different types of applications and runtimes.
These are the types of possible integration levels between an application and Compass:
- basic - administrator manually provides API/Events Metadata to Compass. This type of integration is used mainly for simple use-case scenarios and doesn't support all features.
- application - integration with Compass is built-in inside the application.
- proxy - a highly application-specific proxy component provides the integration.
- service -  a central service provides the integration for a class of applications. It manages multiple instances of these applications. You can integrate multiple services to support different types of applications.

You can register any runtime, providing that it fulfills a contract with Compass and implements its flow. First, your runtime must get a trusted connection to Compass. It must also allow for fetching application definitions and using these applications in a given tenant. The example runtimes are Kyma (Kubernetes), CloudFoundry, Serverless, etc.

Compass is a part of Kyma and it uses a set of Kyma features, such as Istio, Prometheus, Monitoring, or Tracing. This project also contains Compass UI Cockpit that exposes Compass APIs to users.

For more information about the Compass architecture and technical details, read the [documentation](./docs).

## Prerequisities

- [Docker](https://www.docker.com/get-started)
- [Minikube](https://github.com/kubernetes/minikube) 1.0.1
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.12.5
- [Helm](https://github.com/kubernetes/helm) 2.10.0
- [Kyma CLI](https://github.com/kyma-project/cli) master

## Installation

### Chart installation

If you already have a running Kyma 1.1.0 instance with created Secrets and Tiller client certificates, you can install the Compass Helm chart using this command:
```bash
helm install --name "compass" ./chart/compass --tls
```

### Local installation with Kyma

To install Compass along with the minimal Kyma installation from the `master` branch, run this script:
```bash
./installation/scripts/run.sh
```

You can also specify Kyma version, such as 1.2.2 or newer:
```bash
./installation/cmd/run.sh {version}
```

### Kyma installation with the Compass module enabled

To install Kyma with the experimental Compass module enabled, follow the instruction for [custom component installation](https://kyma-project.io/docs/root/kyma/#configuration-custom-component-installation) and enable the `compass` module.


### Testing

Compass, as a part of Kyma, uses [Octopus](https://github.com/kyma-incubator/octopus/blob/master/README.md) for testing. To run the Compass tests, run:

```bash
./installation/scripts/testing.sh
```

Read [this](https://kyma-project.io/docs/root/kyma#details-testing-kyma) document to learn more about testing in Kyma.

## Usage

Go to these URLs to see the documentation, GraphQL schemas, and to test some API operations:

- `https://compass-gateway.{domain}/director`
- `https://compass-gateway.{domain}/connector`
