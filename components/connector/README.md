# Connector

The Connector component takes care of issuing client certificates for applications and runtimes. More details can be found in the [Compass](../../docs/compass/) and [Connector](../../docs/connector/) documentation.

## Development

> **NOTE:** Connector requires the Director component for One-Time Token generation. Read [this](../director/README.md) document to learn how to run it.

After you introduce changes in the GraphQL schema, run the `gqlgen.sh` script.
To run the Connector, use the following command:

```
go run cmd/main.go
```

The GraphQL API playground is available at `localhost:3000`.

## Configuration

Up-to-date list of the configurable parameters of Connector can be found [here](https://github.com/kyma-incubator/compass/blob/main/components/connector/config/config.go)
