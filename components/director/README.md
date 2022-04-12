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
  "authorization": "Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGFwcGxpY2F0aW9uOndyaXRlIGFwcGxpY2F0aW9uX3RlbXBsYXRlOnJlYWQgYXBwbGljYXRpb25fdGVtcGxhdGU6d3JpdGUgaW50ZWdyYXRpb25fc3lzdGVtOnJlYWQgaW50ZWdyYXRpb25fc3lzdGVtOndyaXRlIHJ1bnRpbWU6cmVhZCBydW50aW1lOndyaXRlIGxhYmVsX2RlZmluaXRpb246cmVhZCBsYWJlbF9kZWZpbml0aW9uOndyaXRlIGV2ZW50aW5nOm1hbmFnZSB0ZW5hbnQ6cmVhZCBhdXRvbWF0aWNfc2NlbmFyaW9fYXNzaWdubWVudDpyZWFkIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OndyaXRlIGFwcGxpY2F0aW9uLmF1dGhzOnJlYWQgYXBwbGljYXRpb24ud2ViaG9va3M6cmVhZCBhcHBsaWNhdGlvbl90ZW1wbGF0ZS53ZWJob29rczpyZWFkIGJ1bmRsZS5pbnN0YW5jZV9hdXRoczpyZWFkIGRvY3VtZW50LmZldGNoX3JlcXVlc3Q6cmVhZCBldmVudF9zcGVjLmZldGNoX3JlcXVlc3Q6cmVhZCBhcGlfc3BlYy5mZXRjaF9yZXF1ZXN0OnJlYWQgaW50ZWdyYXRpb25fc3lzdGVtLmF1dGhzOnJlYWQgcnVudGltZS5hdXRoczpyZWFkIGZldGNoLXJlcXVlc3QuYXV0aDpyZWFkIHdlYmhvb2tzLmF1dGg6cmVhZCBmb3JtYXRpb246d3JpdGUgaW50ZXJuYWxfdmlzaWJpbGl0eTpyZWFkIiwidGVuYW50Ijoie1wiZXh0ZXJuYWxUZW5hbnRcIjogXCIzZTY0ZWJhZS0zOGI1LTQ2YTAtYjFlZC05Y2NlZTE1M2EwYWVcIixcImNvbnN1bWVyVGVuYW50XCI6IFwiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlXCJ9In0."
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
- [Tenant Fetcher (Job)](./cmd/tenantfetcher-job/README.md)
- [Tenant Fetcher (Deployment)](./cmd/tenantfetcher-svc/README.md)
- [Tenant Loader](./cmd/tenantloader/README.md)
- [System Fetcher](./cmd/systemfetcher/README.md)
- [Scopes Synchronizer Job](./cmd/scopessynchronizer/README.md)
