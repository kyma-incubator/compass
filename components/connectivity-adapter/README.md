# Connectivity Adapter

Connectivity Adapter translates [Kyma Connector Service API](https://kyma-project.io/docs/master/components/application-connector/specifications/connectorapi/)
and [Kyma Application Registry API](https://kyma-project.io/docs/master/components/application-connector/specifications/metadataapi/)
to Compass Director and Compass Connnector GraphQL API.

## Development

> **NOTE:** Connectivity Adapter requires Director running. Read, how to run it in the [`README.md`](../director/README.md) document.

To launch Connectivity Adapter on local machine, run the following command:

```bash
go run cmd//main.go
```

## Configuration

The Connectivity Adapter binary allows to override some configuration parameters. You can specify following environment variables.

| ENV              | Default        | Description                                       |
| ---------------- | -------------- | ------------------------------------------------- |
| APP_ADDRESS      | 127.0.0.1:8080 | The address and port for the service to listen on |
| APP_DIRECTOR_URL | 127.0.0.1:3000 | The host of running Director                      |
 
