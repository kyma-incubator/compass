# Provisioner

## Development

After you introduce changes in the GraphQL schema, run the `gqlgen.sh` script.
To run the Provisioner, use the following command:

```
go run cmd/main.go
```

## Environment Variables

This table lists the environment variables, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **Address** | Provisioner address with the port | `127.0.0.1:3050` |
| **APIEdnpoint** | Endpoint for the GraphQL API | `/graphql` |
| **PlaygroundAPIEndpoint** | Endpoint for the API Playground | `/graphql` |
| **SchemaFilePath** | Filepath for the database schema | `assets/database/provisioner.sql` |
| **Database.User** | Database username | `postgres` |
| **Database.Password** | Database user password | `password` |
| **Database.Host** | Database host | `localhost` |
| **Database.Port** | Database port | `5432` |
| **Database.Name** | Database name | `provisioner` |
| **Database.SSLMode** | SSL Mode for PostgrSQL. See all the possible values [here](https://www.postgresql.org/docs/9.1/libpq-ssl.html).  | `disable`|

## Prerequisites

Before you can run the Runtime Provisioner, you have to configure access to PostgreSQL database. For development purposes, you can run a PostgreSQL instance in the docker container executing the following command:

```bash
$ docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password postgres
```