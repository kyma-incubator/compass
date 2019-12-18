# Kyma Environment Service Broker

## Overview

The Kyma Environment Service Broker provide a way to run Kyma as a service on different providers like: AWS, GCP or Azure.

The broker provides a services which installs Kyma with a separate plan for each provider.

It uses the Compass Provisioner API to install Kyma on the given cluster.

## Configuration

Use the following environment variables to configure the `Kyma Environment Service Broker`:

| Name | Required | Default | Description |
|-----|:---------:|--------|------------|
| **APP_PORT** | No | `8080` | The port on which the HTTP server listens. |
| **APP_PROVISIONING_URL** | No |  | Specifies an URL to the provisioner API. |
| **APP_PROVISIONING_SECRET_NAME** | No | | Specifies the name of the Secret which holds credentials to the provisioner API. |
| **APP_PROVISIONING_GARDENER_PROJECT_NAME** | No | `true` | Defines the used gardener project name. |
| **APP_PROVISIONING_GCP_SECRET_NAME** | No | | Defines the name of the Secret which holds credentials to GCP. |
| **APP_PROVISIONING_AWS_SECRET_NAME** | No | | Defines the name of the Secret which holds credentials to AWS. |
| **APP_PROVISIONING_AZURE_SECRET_NAME** | No | | Defines the name of the Secret which holds credentials to AZURE. |
| **APP_AUTH_USERNAME** | No | | Specifies the Kyma Environment Service Broker authentication username. |
| **APP_AUTH_PASSWORD** | No | | Specifies the Kyma Environment Service Broker authentication password. |
