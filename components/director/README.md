# Director

The Director exposes GraphQL API.

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
  "authorization": "Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGhlYWx0aF9jaGVja3M6cmVhZCBhcHBsaWNhdGlvbjp3cml0ZSBydW50aW1lOndyaXRlIGxhYmVsX2RlZmluaXRpb246d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjpyZWFkIHJ1bnRpbWU6cmVhZCIsInRlbmFudCI6ImI4MmVhN2NkLTc0NjYtNGFkZS1iNmU5LTIwZmI3N2EwOGNlNiJ9."
}
```

where `tenant` is any valid UUID and `authorization` is JWT token with all scopes and tenant in payload. The token is not signed in development mode.

You can set `tenant` header as any UUID.

### Prerequisites

> **NOTE:** Use script `run.sh` to perform these steps automatically.

Before you can run Director you have to configure access to PostgreSQL database. For development purpose you can run PostgreSQL instance in the docker container executing following command:

```bash
$ docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=pgsql@12345 postgres
```

When you have PostgreSQL instance running you must import the database schema running following command:

```bash
$ PGPASSWORD=pgsql@12345 psql -U postgres -W -h 127.0.0.1 -f <(cat components/schema-migrator/migrations/*.up.sql)
```

## Configuration

The Director binary allows to override some configuration parameters. You can specify following environment variables.

| ENV                                      | Default                         | Description                                               |
| ---------------------------------------- | ------------------------------- | --------------------------------------------------------- |
| APP_ADDRESS                              | 127.0.0.1:3000                  | The address and port for the service to listen on         |
| APP_DB_USER                              | postgres                        | Database username                                         |
| APP_DB_PASSWORD                          | pgsql@12345                     | Database password                                         |
| APP_DB_HOST                              | localhost                       | Database host                                             |
| APP_DB_PORT                              | 5432                            | Database port                                             |
| APP_DB_NAME                              | postgres                        | Database name                                             |
| APP_DB_SSL                               | disable                         | Database SSL mode (disable / enable)                      |
| APP_API_ENDPOINT                         | /graphql                        | The endpoint for GraphQL API                              |
| APP_PLAYGROUND_API_ENDPOINT              | /graphql                        | The endpoint of GraphQL API for the Playground            |
| APP_TENANT_MAPPING_ENDPOINT              | /tenant-mapping                 | The endpoint of Tenant Mapping Service                    |
| APP_SCOPES_CONFIGURATION_FILE            |                                 | The path for scopes configuration file                    |
| APP_SCOPES_CONFIGURATION_FILE_RELOAD     | `1m`                            | The period when the scopes configuration file is reloaded |
| APP_JWKS_ENDPOINT                        | `file://hack/default-jwks.json` | The path for JWKS                                         |
| APP_JWKS_SYNC_PERIOD                     | `5m`                            | The period when the JWKS is synced                        |
| APP_ONE_TIME_TOKEN_URL                   |                                 | The endpoint for fetching one time token                  |
| APP_CONNECTOR_URL                        |                                 | The endpoint of Connector                                 |
| APP_OAUTH20_CLIENT_ENDPOINT              |                                 | The endpoint for managing OAuth 2.0 clients               |
| APP_OAUTH20_PUBLIC_ACCESS_TOKEN_ENDPOINT |                                 | The public endpoint for fetching OAuth 2.0 access token   |
| APP_STATIC_USERS_SRC                     |                                 | The path for static users configuration file              |
| APP_EVENT_DEFAULT_EVENT_URL              |                                 | The default Event URL                                     |

## Usage

Example GraphQL calls can be found [here](examples/README.md)