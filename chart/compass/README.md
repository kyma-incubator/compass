# Compass

## Overview

The Compass consists of the following sub-charts:

- `connector` 
- `director` 
- `gateway` 
- `healthchecker`
- `postgresql`

## Details

To learn more about the Compass, see the [Overview](https://github.com/kyma-incubator/compass#Overview) document.

## Configuration

| Parameter | Description | Values | Default |
| --- | --- | --- | --- |
| `database.useEmbedded` | Specifies whether `postgresql` chart should be installed | true/false | `true` |

## Install the Compass with managed GCP PostgreSQL database

To install the Compass with GCP managed Postgres database, set `database.useEmbedded` value to `false`, and fill those values:

| Parameter | Description | Values | Default |
| --- | --- | --- | --- |
| `database.gcpKey` | Specifies base64 encoded key with Google credentials | base64 encoded string | "" |
| `database.instanceConnectionName` | Specifies instance connection name to GCP PostgreSQL database | string | "" |
| `database.dbUser` | Specifies database username | base64 encoded string | "" |
| `database.dbPassword` | Specifies password for database user | base64 encoded string | "" |
| `database.dbName` | Specifies database name | string | "" |
| `database.host` | Specifies cloudsql-proxy host (usually `localhost`) | string | "" |
| `database.hostPort` | Specifies cloudsql-proxy port (usually `5432`) | string | "" |
| `database.sslMode` | Specifies SSL connection mode | string | "" |

To connect to managed database, we use [cloudsql-proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy) provided by Google, which consumes `gcpKey` and `instanceConnectionName` values.
The rest of values are consumed by applications which want to connect to database.
