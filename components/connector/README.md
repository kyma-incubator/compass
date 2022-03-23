# Connector

You can use the Connector component to issue client certificates for applications and runtimes. For more information, see the [Compass](../../docs/compass/) and [Connector](../../docs/connector/) documentation.

## Development

> **NOTE:** To issue an one-time token (OTT) the Connector component requires the Director component. To learn how to run the Director component, see [Director](../director/README.md).

After you introduce changes in the GraphQL schema, run the `gqlgen.sh` script.
To run the Connector, use the following command:

```
go run cmd/main.go
```

The GraphQL API playground is available at `localhost:3000`.

## Configuration

To get a list of the configurable parameters of the Connector component, see [config.go](https://github.com/kyma-incubator/compass/blob/main/components/connector/config/config.go).
