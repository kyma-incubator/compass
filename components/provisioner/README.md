# Provisioner

## Overview

The Runtime Provisioner is a Compass component responsible for provisioning, installing, and deprovisioning clusters with Kyma (Kyma Runtimes). The relationship between clusters and Runtimes is 1:1.

For more details, see the Runtime Provisioner [documentation](https://github.com/kyma-incubator/compass/tree/master/docs/provisioner).

## Prerequisites

Before you can run the Runtime Provisioner, you have to configure access to the PostgreSQL database. For development purposes, you can run a PostgreSQL instance in the Docker container executing the following command:

```bash
$ docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password postgres
```

The Runtime Provisioner also needs access to the cluster from which it fetches Secrets.  

## Development

### GraphQL schema

After you introduce changes in the GraphQL schema, run the `gqlgen.sh` script.

### Database schema

For tests to run properly, update the database schema in `./assets/database/provisioner.sql`. Provide the new migration in the Schema Migrator component in `migrations/provisioner`.

### Run Provisioner

To run the Runtime Provisioner, use the following command:
```bash
go run cmd/main.go
```

### Environment Variables

This table lists the environment variables, their descriptions, and default values:


| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **APP_ADDRESS** | Provisioner address with the port | `127.0.0.1:3000` |
| **APP_API_ENDPOINT** | Endpoint for the GraphQL API | `/graphql` |
| **APP_PLAYGROUND_API_ENDPOINT** | Endpoint for the API playground | `/graphql` |
| **APP_CREDENTIALS_NAMESPACE** | Namespace where Director credentials are stored | `compass-system` |
| **APP_DIRECTOR_URL** | Director URL | `https://compass-gateway-auth-oauth.kyma.local/director/graphql` |
| **APP_SKIP_DIRECTOR_CERT_VERIFICATION** | Flag to skip certificate verification for Director | `false` |
| **APP_OAUTH_CREDENTIALS_SECRET_NAME** | Runtime Provisioner credentials | `compass-provisioner-credentials` |
| **APP_DATABASE_USER** | Database username | `postgres` |
| **APP_DATABASE_PASSWORD** | Database user password | `password` |
| **APP_DATABASE_HOST** | Database host | `localhost` |
| **APP_DATABASE_PORT** | Database port | `5432` |
| **APP_DATABASE_NAME** | Database name | `provisioner` |
| **APP_DATABASE_SSL_MODE** | SSL Mode for PostgrSQL. See all the possible values [here](https://www.postgresql.org/docs/9.1/libpq-ssl.html)  | `disable`|
| **APP_INSTALLATION_TIMEOUT** | Kyma installation timeout | `30m`|
| **APP_INSTALLATION_ERRORS_COUNT_FAILURE_THRESHOLD** | Number of installation errors that cause installation to fail  | `5`|
| **APP_GARDENER_PROJECT** | Name of the Gardener project connected to the service account  | `gardenerProject`|
| **APP_GARDENER_KUBECONFIG_PATH** | Filepath for the Gardener kubeconfig  | `./dev/kubeconfig.yaml`|
| **APP_PROVISIONER** | Provisioning mechanism used by the Runtime Provisioner (Gardener or Hydroform) | `gardener`|
