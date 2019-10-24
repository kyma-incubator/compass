<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-incubator/compass/master/logo.png" width="235">
</p>

# Compass

## Overview

Compass is a central, multi-tenant system that allows you to connect Applications and manage them across multiple [Kyma Runtimes](#architecture-components-kyma-runtime). Using Compass, you can control and monitor your Application landscape in one central place. As an integral part of Kyma, Compass uses a set of features that Kyma provides, such as Istio, Prometheus, Monitoring, and Tracing. It also includes Compass UI Cockpit that exposes Compass APIs to users.
Compass allows you to:
- Connect and manage Applications and Kyma Runtimes in one central place
- Store Applications and Runtimes configurations
- Group Applications and Runtimes to enable integration
- Communicate the configuration changes to Applications and Runtimes
- Establish a trusted connection between Applications and Runtimes using various authentication methods

Compass by design does not participate in direct communication between Applications and Runtimes. It only sets up the connection. In case the cluster with Compass is down, the Applications and Runtimes cooperation still works.

For more information about the Compass architecture and technical details, read the project [documentation](./docs).

## Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Minikube](https://github.com/kubernetes/minikube) 1.0.1
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.12.5
- [Helm](https://github.com/kubernetes/helm) 2.10.0
- [Kyma CLI](https://github.com/kyma-project/cli) master

## Installation

### Enable Compass in Kyma

Read [this](https://kyma-project.io/docs/master/components/compass/#installation-installation) document to learn about different modes in which you can enable Compass in Kyma.

### Chart installation

If you already have a running Kyma instance in at least 1.6 version, with created Secrets and Tiller client certificates, you can install the Compass Helm chart using this command:
```bash
helm install --name "compass" ./chart/compass --tls
```

For local installation, specify additional parameters:
```bash
helm install --set=global.isLocalEnv=true --set=global.minikubeIP=$(eval minikube ip) --name "compass" ./chart/compass --tls
```

### Local installation with Kyma

To install Compass along with the minimal Kyma installation from the `master` branch, run this script:
```bash
./installation/cmd/run.sh
```

You can also specify Kyma version, such as 1.6 or newer:
```bash
./installation/cmd/run.sh {version}
```

### Testing

Compass, as a part of Kyma, uses [Octopus](https://github.com/kyma-incubator/octopus/blob/master/README.md) for testing. To run the Compass tests, run:

```bash
./installation/scripts/testing.sh
```

Read [this](https://kyma-project.io/docs/root/kyma#details-testing-kyma) document to learn more about testing in Kyma.

## Usage

Currently, the Compass Gateway is accessible under three different hosts secured with different authentication methods:

- `https://compass-gateway.{domain}` - secured with JWT token issued by Identity Service, such as Dex
- `https://compass-gateway-mtls.{domain}` - secured with client certificates (mTLS)
- `https://compass-gateway-auth-oauth.{domain}` - secured with OAuth 2.0 access token issued by [Hydra](https://kyma-project.io/docs/components/security/#details-o-auth2-and-open-id-connect-server)

You can access Director GraphQL API under the `/director/graphql` endpoint, and Connector GraphQL API under `/connector/graphql`.
