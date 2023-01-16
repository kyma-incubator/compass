<p align="center">
 <img src="https://raw.githubusercontent.com/kyma-incubator/compass/main/logo.png" width="235">
</p>

## Overview

Compass is a central, multi-tenant system that allows you to connect Applications and manage them across multiple [Kyma Runtimes](./docs/compass/02-01-components.md#kyma-runtime). Using Compass, you can control and monitor your Application landscape in one central place. As an integral part of Kyma, Compass uses a set of features that Kyma provides, such as Istio, Prometheus, Monitoring, and Tracing. It also includes Compass UI Cockpit that exposes Compass APIs to users.
Compass allows you to:
- Connect and manage Applications and Kyma Runtimes in one central place
- Store Applications and Runtimes configurations
- Group Applications and Runtimes to enable integration
- Communicate the configuration changes to Applications and Runtimes
- Establish a trusted connection between Applications and Runtimes using various authentication methods

Compass by design does not participate in direct communication between Applications and Runtimes. It only sets up the connection. In case the cluster with Compass is down, the Applications and Runtimes cooperation still works.

For more information about the Compass architecture, technical details, and components, read the project [documentation](./docs).

## Prerequisites

- [Docker](https://www.docker.com/get-started)
- [k3d](https://github.com/k3d-io/k3d) v5.2.2+
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.23.0+
- [Kyma CLI](https://github.com/kyma-project/cli) 2.3.0
- [helm](https://github.com/helm/helm) v3.8.0+
- [yq](https://github.com/mikefarah/yq) v4+

## Installation

Install Compass locally or on a cluster. See the [installation document](https://github.com/kyma-incubator/compass/blob/main/docs/compass/04-01-installation.md) for details.

### Dependencies

Compass depends on [Kyma](https://github.com/kyma-project/kyma).
For installation and CI integration jobs, a fixed Kyma version is used, which can be checked at `./installation/resources/KYMA_VERSION`.

## Testing

Compass uses [Octopus](https://github.com/kyma-incubator/octopus/blob/master/README.md) for testing both locally and on a cluster. To run the Compass tests, use this script:

```bash
./installation/scripts/testing.sh
```

## Usage

Currently, the Compass Gateway is accessible under three different hosts secured with different authentication methods:

- `https://compass-gateway.{domain}` - secured with JWT token issued by an identity service
- `https://compass-gateway-mtls.{domain}` - secured with client certificates (mTLS)
- `https://compass-gateway-auth-oauth.{domain}` - secured with OAuth 2.0 access token issued by [Hydra](https://kyma-project.io/docs/components/security/#details-o-auth2-and-open-id-connect-server)

You can access Director GraphQL API under the `/director/graphql` endpoint, and Connector GraphQL API under `/connector/graphql`.

To access Connectivity Adapter, use the `https://adapter-gateway.{DOMAIN}` host secured with one-time tokens or `https://adapter-gateway-mtls.{DOMAIN}` secured with client certificates (mTLS).
