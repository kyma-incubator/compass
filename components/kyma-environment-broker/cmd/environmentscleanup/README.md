# Environments Cleanup

## Overview

This application cleans up environments which do not meet requirements in given Gardener project.

## Prerequisites

Environments Cleanup requires access to:
1. Gardener Project of choice to filter Shoots without proper label
2. Compass Director to get Instance ID for each Runtime marked for deletion
3. Kyma Environment Broker to trigger Runtime deprovisioning

## Configuration

The Environments Cleanup binary allows you to override some configuration parameters. You can specify the following environment variables:

| Environment variable                       | Description                                                                                                                        | Default value                                                            |
|--------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| APP_MAX_AGE_HOURS                          | Defines the maximum time a Shoot can live without deletion in case the label is not specified. The Shoot age is provided in hours. | `24h`                                                                    |
| APP_GARDENER_PROJECT                       | Specifies Gardener project name.                                                                                                   | `kyma-dev`                                                               |
| APP_GARDENER_KUBECONFIG_PATH               | Specifies Gardener cluster's kubeconfig path.                                                                                      | `/gardener/kubeconfig/kubeconfig`                                        |
| APP_DIRECTOR_NAMESPACE                     | Specifies the Namespace in which Director is deployed.                                                                             | `compass-system`                                                         |
| APP_DIRECTOR_URL                           | Specifies the Director's URL.                                                                                                      | `https://compass-director.compass-system.svc.cluster.local:3000/graphql` |
| APP_DIRECTOR_OAUTH_CREDENTIALS_SECRET_NAME | Specifies the name of the Secret created by the Integration System.                                                                | `compass-kyma-environment-broker-credentials`                            |
| APP_DIRECTOR_OAUTH_CREDENTIALS_SECRET_NAME | Specifies whether TLS checks the presented certificates.                                                                           | `false`                                                                  |
| APP_BROKER_URL                             | Specifies the Kyma Environment Broker URL.                                                                                         | `https://kyma-env-broker.kyma.local`                                     |
| APP_BROKER_TOKEN_URL                       | Specifies the Kyma Environment Broker OAuth token endpoint.                                                                        | `https://oauth.2kyma.local/oauth2/token`                                 |
| APP_BROKER_CLIENT_ID                       | Specifies the username for the OAuth2 authentication in KEB.                                                                       | None                                                                     |
| APP_BROKER_CLIENT_SECRET                   | Specifies the password for the OAuth2 authentication in KEB.                                                                       | None                                                                     |
| APP_BROKER_SCOPE                           | Specifies the scope for the OAuth2 authentication in KEB.                                                                          | None                                                                     |