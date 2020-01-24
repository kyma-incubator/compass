# Kyma Environment Broker

## Overview

Kyma Environment Broker is a component that allows you to run Kyma as a Runtime on clusters provided by third-party providers. It uses Provisioner's API to install Kyma on a given cluster.

For more information, read the [documentation](../../docs/kyma-environment-broker).


## Development

This table lists the environment variables, their descriptions, and default values:

| Name | Description | Default value |
|-----|---------|:--------:|
| **APP_PORT** | Specifies the port on which the HTTP server listens. | `8080` |
| **APP_PROVISIONING_URL** | Specifies a URL to the Provisioner's API. | None |
| **APP_PROVISIONING_SECRET_NAME** | Specifies the name of the Secret which holds credentials to the Provisioner API. | None |
| **APP_PROVISIONING_GARDENER_PROJECT_NAME** | Defines the Gardener project name. | `true` |
| **APP_PROVISIONING_GCP_SECRET_NAME** | Defines the name of the Secret which holds credentials to GCP. | None |
| **APP_PROVISIONING_AWS_SECRET_NAME** | Defines the name of the Secret which holds credentials to AWS. | None |
| **APP_PROVISIONING_AZURE_SECRET_NAME** | Defines the name of the Secret which holds credentials to Azure. | None |
| **APP_AUTH_USERNAME** | Specifies the Kyma Environment Service Broker authentication username. | None |
| **APP_AUTH_PASSWORD** | Specifies the Kyma Environment Service Broker authentication password. | None |
| **APP_DIRECTOR_NAMESPACE** | Specifies the Namespace in which Director is deployed. | `compass-system` |
| **APP_DIRECTOR_URL** | Specifies the Director's URL. | `http://compass-director.compass-system.svc.cluster.local:3000/graphql` |
| **APP_DIRECTOR_OAUTH_CREDENTIALS_SECRET_NAME** | Specifies the name of the Secret created by the Integration System. | `compass-kyma-environment-broker-credentials` |
| **APP_DIRECTOR_SKIP_CERT_VERIFICATION** | Specifies whether TLS checks the presented certificates. | `false` |
| **APP_DATABASE_USER** | Defines database username. | `postgres` |
| **APP_DATABASE_PASSWORD** | Defines database user password | `password` |
| **APP_DATABASE_HOST** | Defines database host | `localhost` |
| **APP_DATABASE_PORT** | Defines database port | `5432` |
| **APP_DATABASE_NAME** | Defines database name | `broker` |
| **APP_DATABASE_SSL** | Specifies SSL Mode for PostgrSQL. See all the possible values [here](https://www.postgresql.org/docs/9.1/libpq-ssl.html).  | `disable`|
