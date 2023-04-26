# Director

The Director exposes a GraphQL API for managing applications and runtimes.

- [Development](#development)
    - [Prerequisites](#prerequisites)
    - [Configuration](#configuration)
- [Local Development](#local-development)
    - [Prerequisites](#local-prerequisites)
    - [Run](#local-run)
- [Usage](#usage)

## Development

After you introduce changes in the GraphQL schema, run the `gqlgen.sh` script.

To run Director with PostgreSQL container on local machine with latest DB schema, run the following command:

```bash
./run.sh
```

The GraphQL API playground is available at `localhost:3000`. To call the API, send the following headers:

```json
{
  "tenant": "380da7fb-767e-45cf-8fcc-829f97655d1b",
  "authorization": "Bearer eyAiYWxnIjogIm5vbmUiLCAidHlwIjogIkpXVCIgfQo.eyAic2NvcGVzIjogIndlYmhvb2s6d3JpdGUgZm9ybWF0aW9uX3RlbXBsYXRlLndlYmhvb2tzOnJlYWQgcnVudGltZS53ZWJob29rczpyZWFkIGFwcGxpY2F0aW9uLmxvY2FsX3RlbmFudF9pZDp3cml0ZSB0ZW5hbnRfc3Vic2NyaXB0aW9uOndyaXRlIHRlbmFudDp3cml0ZSBmZXRjaC1yZXF1ZXN0LmF1dGg6cmVhZCB3ZWJob29rcy5hdXRoOnJlYWQgYXBwbGljYXRpb24uYXV0aHM6cmVhZCBhcHBsaWNhdGlvbi53ZWJob29rczpyZWFkIGFwcGxpY2F0aW9uLmFwcGxpY2F0aW9uX3RlbXBsYXRlOnJlYWQgYXBwbGljYXRpb24udGVuYW50X2J1c2luZXNzX3R5cGU6cmVhZCBhcHBsaWNhdGlvbl90ZW1wbGF0ZTp3cml0ZSBhcHBsaWNhdGlvbl90ZW1wbGF0ZTpyZWFkIGFwcGxpY2F0aW9uX3RlbXBsYXRlLndlYmhvb2tzOnJlYWQgZG9jdW1lbnQuZmV0Y2hfcmVxdWVzdDpyZWFkIGV2ZW50X3NwZWMuZmV0Y2hfcmVxdWVzdDpyZWFkIGFwaV9zcGVjLmZldGNoX3JlcXVlc3Q6cmVhZCBydW50aW1lLmF1dGhzOnJlYWQgaW50ZWdyYXRpb25fc3lzdGVtLmF1dGhzOnJlYWQgYnVuZGxlLmluc3RhbmNlX2F1dGhzOnJlYWQgYnVuZGxlLmluc3RhbmNlX2F1dGhzOnJlYWQgYXBwbGljYXRpb246cmVhZCBhdXRvbWF0aWNfc2NlbmFyaW9fYXNzaWdubWVudDpyZWFkIGhlYWx0aF9jaGVja3M6cmVhZCBhcHBsaWNhdGlvbjp3cml0ZSBydW50aW1lOndyaXRlIGxhYmVsX2RlZmluaXRpb246d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjpyZWFkIHJ1bnRpbWU6cmVhZCB0ZW5hbnQ6cmVhZCBmb3JtYXRpb246cmVhZCBmb3JtYXRpb246d3JpdGUgaW50ZXJuYWxfdmlzaWJpbGl0eTpyZWFkIGZvcm1hdGlvbl90ZW1wbGF0ZTpyZWFkIGZvcm1hdGlvbl90ZW1wbGF0ZTp3cml0ZSBmb3JtYXRpb25fY29uc3RyYWludDpyZWFkIGZvcm1hdGlvbl9jb25zdHJhaW50OndyaXRlIGNlcnRpZmljYXRlX3N1YmplY3RfbWFwcGluZzpyZWFkIGNlcnRpZmljYXRlX3N1YmplY3RfbWFwcGluZzp3cml0ZSBmb3JtYXRpb24uc3RhdGU6d3JpdGUgdGVuYW50X2FjY2Vzczp3cml0ZSIsICJ0ZW5hbnQiOiJ7XCJjb25zdW1lclRlbmFudFwiOlwiXCIsXCJleHRlcm5hbFRlbmFudFwiOlwiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlXCJ9IiB9Cg."
}
```

where `tenant` is any valid UUID and `authorization` is JWT token with all scopes and tenant in payload. The token is not signed in development mode.

You can set `tenant` header as any UUID.

<h3 id="prerequisites">Prerequisites</h3>

> **NOTE:** To perform the following steps automatically, you can use the `run.sh` script. For more information, see the section [Local Development](#local-development).

To run the Director, first you must configure access to PostgreSQL database. For development purposes, you can run the PostgreSQL instance in the docker container by running the following command:

```bash
$ docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=pgsql@12345 postgres
```

When you have the PostgreSQL instance running, you must import the database schema by running the following command:

```bash
$ PGPASSWORD=pgsql@12345 psql -U postgres -W -h 127.0.0.1 -f <(cat components/schema-migrator/migrations/director/*.up.sql)
```

### Configuration

The Director binary allows you to override some configuration parameters. To get a list of the configurable parameters, see [main.go](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/director/cmd/director/main.go#L90).

The Director also depends on a configuration file that contains the required scopes for each GraphQL query and mutation. For local development you can use the file at [hack/config-local.yaml](./hack/config-local.yaml). For in-cluster setup you can use the file that is located in the Director subchart at [chart/compass/charts/director/config.yaml](../../chart/compass/charts/director/config.yaml).

## Local Development

<h3 id="local-prerequisites">Prerequisites</h3>

- Install `kubectl` version 1.18 or higher.
- To use `--debug` flag, first you must install `delve`.

<h3 id="local-run">Run</h3>

There is a `./run.sh` script that automatically runs director locally with the necessary configuration and environment variables. There are several flags that can be used:
- `--skip-db-cleanup` - Does not delete the DB on script termination.
- `--reuse-db` - Can be used in combination with `--skip-db-cleanup` to reuse an already existing DB.
- `--dump-db` - Starts director with DB, populated with data from CMP development environment.
- `--debug` - Starts director in debugging mode on default port `40000`.
- `--async-enabled` - Enables scheduling of asynchronous operations. To use this option, make sure that the [Operations Controller](../operations-controller/) component is running.

> **NOTE**: Director component has certificate cache, which is populated with an external certificate through Kubernetes secret. Locally, you can override the secret data with certificate and key that you need for testing or debugging. Check the table below for environment variables.

| Environment variable                         | Default value                   | Description                                                        |
| -------------------------------------------- | ------------------------------- | ------------------------------------------------------------------ |
| **APP_EXTERNAL_CLIENT_CERT_VALUE**           | `certValue`                     | External client certificate, which is used to populate the certificate cache   | 
| **APP_EXTERNAL_CLIENT_KEY_VALUE**            | `keyValue`                      | External client certificate key, which is added into certificate cache   | 

## Usage

You can find examples of GraphQL calls at: [Examples](examples/README.md).

## Other Binaries

The Director's source code is also used by other Compass's components. For this reason, the code comprises different binaries, located in the `cmd` directory. To configure it and run it locally, you can see the following documentation sources:
- [ORD Aggregator](./cmd/ordaggregator/README.md)
- [Tenant Fetcher (Deployment)](./cmd/tenantfetcher-svc/README.md)
- [Tenant Loader](./cmd/tenantloader/README.md)
- [System Fetcher](./cmd/systemfetcher/README.md)
- [Scopes Synchronizer Job](./cmd/scopessynchronizer/README.md)
