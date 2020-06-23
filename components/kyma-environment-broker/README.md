# Kyma Environment Broker

## Overview

Kyma Environment Broker (KEB) is a component that allows you to provision Kyma as a Runtime on clusters provided by third-party providers. It uses Provisioner's API to install Kyma on a given cluster.

For more information, read the [documentation](../../docs/kyma-environment-broker) where you can find information on:

- [Architecture](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/02-01-architecture.md)
- [Service description](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/03-01-service-description.md)
- [Runtime components](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/03-02-runtime-components.md)
- [Runtime provisioning and deprovisioning](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/03-03-runtime-provisioning-and-deprovisioning.md)
- [Hyperscaler account pool](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/03-04-hyperscaler-account-pool.md)
- [Authorization](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/03-05-authorization.md)
- [Runtime overrides](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/03-06-runtime-overrides.md)
- [Provisioning Kyma environment](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/08-01-provisioning-kyma-environment.md)
- [Deprovisioning Kyma environment](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/08-02-deprovisioning-kyma-environment.md)
- [Operation status](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/08-03-operation-status.md)
- [Instance details](https://github.com/kyma-incubator/compass/blob/master/docs/kyma-environment-broker/08-04-instance-details.md)

## Configuration

KEB binary allows you to override some configuration parameters. You can specify the following environment variables:

| Name | Description | Default value |
|-----|---------|:--------:|
| **APP_PORT** | Specifies the port on which the HTTP server listens. | `8080` |
| **APP_BROKER_DEFAULT_GARDENER_SHOOT_PURPOSE** | Specifies the purpose of the created cluster. The possible values are: `development`, `evaluation`, `production`, `testing`. | `development` |
| **APP_PROVISIONING_URL** | Specifies a URL to the Provisioner's API. | None |
| **APP_PROVISIONING_SECRET_NAME** | Specifies the name of the Secret which holds credentials to the Provisioner's API. | None |
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
| **APP_DATABASE_USER** | Defines the database username. | `postgres` |
| **APP_DATABASE_PASSWORD** | Defines the database user password. | `password` |
| **APP_DATABASE_HOST** | Defines the database host. | `localhost` |
| **APP_DATABASE_PORT** | Defines the database port. | `5432` |
| **APP_DATABASE_NAME** | Defines the database name. | `broker` |
| **APP_DATABASE_SSL** | Specifies the SSL Mode for PostgrSQL. See all the possible values [here](https://www.postgresql.org/docs/9.1/libpq-ssl.html).  | `disable`|
