# Environments Cleanup

## Overview

This application cleans up environments which do not meet requirements in given Gardener project.

## Prerequisites

Environments Cleanup requires access to:
1. Gardener Project of choice to filter Shoots without proper label
2. Database to get Instance ID for each Runtime marked for deletion
3. Kyma Environment Broker to trigger Runtime deprovisioning

## Configuration

The Environments Cleanup binary allows you to override some configuration parameters. You can specify the following environment variables:

| Environment variable                       | Description                                                                                                                        | Default value                                                            |
|--------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| APP_DB_IN_MEMORY                          | Defines if KEB is using embedded database. | `false`                                                                    |
| APP_MAX_AGE_HOURS                          | Defines the maximum time a Shoot can live without deletion in case the label is not specified. The Shoot age is provided in hours. | `24h`                                                                    |
| APP_LABEL_SELECTOR                          | Defines the label selector to filter out Shoots for deletion. | `owner.do-not-delete!=true`                                                                    |
| APP_GARDENER_PROJECT                       | Specifies Gardener project name.                                                                                                   | `kyma-dev`                                                               |
| APP_GARDENER_KUBECONFIG_PATH               | Specifies Gardener cluster's kubeconfig path.                                                                                      | `/gardener/kubeconfig/kubeconfig`                                        |
| APP_DATABASE_USER | Database username | `postgres` |
| APP_DATABASE_PASSWORD | Database user password | `password` |
| APP_DATABASE_HOST | Database host | `localhost` |
| APP_DATABASE_PORT | Database port | `5432` |
| APP_DATABASE_NAME | Database name | `provisioner` |
| APP_DATABASE_SSL_MODE | SSL Mode for PostgrSQL. See all the possible values [here](https://www.postgresql.org/docs/9.1/libpq-ssl.html)  | `disable`|
| APP_BROKER_URL                             | Specifies the Kyma Environment Broker URL.                                                                                         | `https://kyma-env-broker.kyma.local`                                     |
| APP_BROKER_TOKEN_URL                       | Specifies the Kyma Environment Broker OAuth token endpoint.                                                                        | `https://oauth.2kyma.local/oauth2/token`                                 |
| APP_BROKER_CLIENT_ID                       | Specifies the username for the OAuth2 authentication in KEB.                                                                       | None                                                                     |
| APP_BROKER_CLIENT_SECRET                   | Specifies the password for the OAuth2 authentication in KEB.                                                                       | None                                                                     |
| APP_BROKER_SCOPE                           | Specifies the scope for the OAuth2 authentication in KEB.                                                                          | None                                                                     |