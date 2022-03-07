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

The GraphQL API playground is available at `localhost:3000`. In order to call the API, send the following headers:

```json
{
  "tenant": "380da7fb-767e-45cf-8fcc-829f97655d1b",
  "authorization": "Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OndyaXRlIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OnJlYWQgaGVhbHRoX2NoZWNrczpyZWFkIGFwcGxpY2F0aW9uOndyaXRlIHJ1bnRpbWU6d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjp3cml0ZSBsYWJlbF9kZWZpbml0aW9uOnJlYWQgcnVudGltZTpyZWFkIHRlbmFudDpyZWFkIiwidGVuYW50IjoiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlIn0."
}
```

where `tenant` is any valid UUID and `authorization` is JWT token with all scopes and tenant in payload. The token is not signed in development mode.

You can set `tenant` header as any UUID.

<h3 id="prerequisites">Prerequisites</h3>

> **NOTE:** Use script `run.sh` to perform these steps automatically. Check [Local Development](#local-development) section for more information.

Before you can run Director you have to configure access to PostgreSQL database. For development purpose you can run PostgreSQL instance in the docker container executing following command:

```bash
$ docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=pgsql@12345 postgres
```

When you have PostgreSQL instance running you must import the database schema running following command:

```bash
$ PGPASSWORD=pgsql@12345 psql -U postgres -W -h 127.0.0.1 -f <(cat components/schema-migrator/migrations/*.up.sql)
```

### Configuration

The Director binary allows overriding of some configuration parameters. Up-to-date list of the configurable parameters can be found [here](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/director/cmd/director/main.go#L90).

Director also depends on a configuration file, containing the required scopes for each GraphQL query and mutation. The file used for local development is located under [hack/config-local.yaml](./hack/config-local.yaml), and the one used for in-cluster setup is located in the Director subchart in [chart/compass/charts/director/config.yaml](../../chart/compass/charts/director/config.yaml).

## Local Development

<h3 id="local-prerequisites">Prerequisites</h3>

- You should install `kubectl` version 1.18 or higher.
- To use `--debug` flag, first you must install `delve`.

<h3 id="local-run">Run</h3>

There is a `./run.sh` script that automatically runs director locally with the necessary configuration and environment variables. There are several flags that can be used:
- `--skip-db-cleanup` - Does not delete the DB on script termination.
- `--reuse-db` - Can be used in combination with `--skip-db-cleanup` to reuse an already existing DB.
- `--dump-db` - Starts director with DB, populated with data from CMP development environment.
- `--debug` - Starts director in debugging mode on default port `40000`.
- `--async-enabled` - Enables asynchronous operations scheduling. A prerequisite for this option is a running [Operations Controller](../operations-controller/) component.

> **NOTE**: Director component has certificate cache, which is populated with an external certificate through Kubernetes secret. Locally, you can override the secret data with certificate and key that you need for testing or debugging. Check the table below for environment variables.

| Environment variable                         | Default value                   | Description                                                        |
| -------------------------------------------- | ------------------------------- | ------------------------------------------------------------------ |
| **APP_EXTERNAL_CLIENT_CERT_VALUE**           | `certValue`                     | External client certificate, which is used to populate the certificate cache   | 
| **APP_EXTERNAL_CLIENT_KEY_VALUE**            | `keyValue`                      | External client certificate key, which is added into certificate cache   | 

## Usage

Find examples of GraphQL calls [here](examples/README.md).

## Other Binaries

As the source code required by a few other Compass components is the same as Director, they are just different binaries located in the `cmd` directory. You can check their own documentations in order to see how they can be configured and ran locally:
- [ORD Aggregator](./cmd/ordaggregator/README.md)
- [Tenant Fetcher (Job)](./cmd/tenantfetcher-job/README.md)
- [Tenant Fetcher (Deployment)](./cmd/tenantfetcher-svc/README.md)
- [Tenant Loader](./cmd/tenantloader/README.md)
- [System Fetcher](./cmd/systemfetcher/README.md)
- [Scopes Synchronizer Job](./cmd/scopessynchronizer/README.md)
