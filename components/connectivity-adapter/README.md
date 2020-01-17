# Connectivity Adapter

Connectivity Adapter translates [Kyma Connector Service API](https://kyma-project.io/docs/master/components/application-connector/specifications/connectorapi/)
and [Kyma Application Registry API](https://kyma-project.io/docs/master/components/application-connector/specifications/metadataapi/)
to Compass Director and Compass Connnector GraphQL API.

## Development

> **NOTE:** Connectivity Adapter requires the Director component. Read [this](../director/README.md) document to learn how to run it.

To launch Connectivity Adapter on local machine, run the following command:

```bash
go run cmd//main.go
```

## Configuration

The Connectivity Adapter binary allows you to override some configuration parameters. You can specify the following environment variables:

| Environment variable              | Default value      | Description                                       |
| ---------------- | -------------- | ------------------------------------------------- |
| **APP_ADDRESS**      | `127.0.0.1:8080` | Address and port for the service to listen on |
| **APP_DIRECTOR_URL** | `127.0.0.1:3000` | Host of the running Director component                      |
 
