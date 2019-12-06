# Provisioner

## Overview

The Runtime Provisioner is a Compass component responsible for provisioning, installing, and deprovisioning clusters with Kyma (Kyma Runtimes). The relationship between clusters and Runtimes is 1:1.

> **NOTE:** Kyma installation is not implemented yet. 

For more details, see the Runtime Provisioner [documentation](https://github.com/kyma-incubator/compass/tree/master/docs/provisioner).

## Prerequisites

Before you can run the Runtime Provisioner, you have to configure access to the PostgreSQL database. For development purposes, you can run a PostgreSQL instance in the Docker container executing the following command:

```bash
$ docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password postgres
```

The Runtime Provisioner also needs access to the cluster from which it fetches Secrets.  

## Development

After you introduce changes in the GraphQL schema, run the `gqlgen.sh` script.
To run the Provisioner, use the following command:

```bash
go run cmd/main.go
```

### Environment Variables

This table lists the environment variables, their descriptions, and default values:


| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **APP_ADDRESS** | Provisioner address with the port | `127.0.0.1:3050` |
| **APP_API_ENDPOINT** | Endpoint for the GraphQL API | `/graphql` |
| **APP_PLAYGROUND_API_ENDPOINT** | Endpoint for the API playground | `/graphql` |
| **APP_SCHEMA_FILE_PATH** | Filepath for the database schema | `assets/database/provisioner.sql` |
| **APP_DATABASE_USE** | Database username | `postgres` |
| **APP_DATABASE_PASSWORD** | Database user password | `password` |
| **APP_DATABASE_HOST** | Database host | `localhost` |
| **APP_DATABASE_PORT** | Database port | `5432` |
| **APP_DATABASE_NAME** | Database name | `provisioner` |
| **APP_DATABASE_SSL_MODE** | SSL Mode for PostgrSQL. See all the possible values [here](https://www.postgresql.org/docs/9.1/libpq-ssl.html)  | `disable`|
| **APP_INSTALLATION_TIMEOUT** | Kyma installation timeout | `30m`|
| **APP_INSTALLATION_ERRORS_COUNT_FAILURE_THRESHOLD** | Amount of installation errors that cause installation to fail  | `5`|

