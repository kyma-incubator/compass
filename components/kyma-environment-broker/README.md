# Kyma Environment Broker

## Overview

Kyma Environment Broker provides a way to run Kyma as a Runtime on clusters provided by different cloud providers, such as AWS, GCP or Azure. The broker provides services that install Kyma with a separate plan for each provider. It uses Provisioner's API to install Kyma on the given cluster.

For more information, read the [documentation](../../docs/kyma-environment-broker).

## Development

Use the following environment variables to configure the Kyma Environment Broker:

| Name | Required | Default | Description |
|-----|:---------:|--------|------------|
| **APP_PORT** | No | `8080` | The port on which the HTTP server listens. |
| **APP_PROVISIONING_URL** | No |  | Specifies an URL to the provisioner API. |
| **APP_PROVISIONING_SECRET_NAME** | No | | Specifies the name of the Secret which holds credentials to the Provisioner API. |
| **APP_PROVISIONING_GARDENER_PROJECT_NAME** | No | `true` | Defines the used Gardener project name. |
| **APP_PROVISIONING_GCP_SECRET_NAME** | No | | Defines the name of the Secret which holds credentials to GCP. |
| **APP_PROVISIONING_AWS_SECRET_NAME** | No | | Defines the name of the Secret which holds credentials to AWS. |
| **APP_PROVISIONING_AZURE_SECRET_NAME** | No | | Defines the name of the Secret which holds credentials to Azure. |
| **APP_AUTH_USERNAME** | No | | Specifies the Kyma Environment Service Broker authentication username. |
| **APP_AUTH_PASSWORD** | No | | Specifies the Kyma Environment Service Broker authentication password. |
