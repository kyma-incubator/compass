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
  "tenant":"380da7fb-767e-45cf-8fcc-829f97655d1b",
  "authorization":"Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGhlYWx0aF9jaGVja3M6cmVhZCBhcHBsaWNhdGlvbjp3cml0ZSBydW50aW1lOndyaXRlIGxhYmVsX2RlZmluaXRpb246d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjpyZWFkIHJ1bnRpbWU6cmVhZCIsInRlbmFudCI6ImI4MmVhN2NkLTc0NjYtNGFkZS1iNmU5LTIwZmI3N2EwOGNlNiJ9."
}
```
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

| ENV                         | Default        | Description                                       |
|-----------------------------|----------------|---------------------------------------------------|
| APP_ADDRESS                 | 127.0.0.1:3000 | The address and port for the service to listen on |
| APP_DB_USER                 | postgres       | Database username                                 |
| APP_DB_PASSWORD             | pgsql@12345    | Database password                                 |
| APP_DB_HOST                 | localhost      | Database host                                     |
| APP_DB_PORT                 | 5432           | Database port                                     |
| APP_DB_NAME                 | postgres       | Database name                                     |
| APP_DB_SSL                  | disable        | Database SSL mode (disable / enable)              |
| APP_API_ENDPOINT            | /graphql       | The endpoint for GraphQL API                      |
| APP_PLAYGROUND_API_ENDPOINT | /graphql       | The endpoint of GraphQL API for the Playground    |
