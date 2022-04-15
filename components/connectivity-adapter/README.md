# Connectivity Adapter

Connectivity Adapter translates [Kyma Connector Service API](https://kyma-project.io/docs/main/components/application-connector/specifications/connectorapi/)
and [Kyma Application Registry API](https://kyma-project.io/docs/main/components/application-connector/specifications/metadataapi/)
to Compass Director and Compass Connnector GraphQL API.

## Development

> **NOTE:** Connectivity Adapter requires the Director component. To learn how to run it, see [Director](../director/README.md).

To launch Connectivity Adapter on local machine, run the following command:

```bash
go run cmd/main.go
```

## Configuration

The Connectivity Adapter binary allows you to override some configuration parameters. 
To get a list of the supported parameters, open: [main.go](https://github.com/kyma-incubator/compass/blob/75aff5226d4a105f4f04608416c8fa9a722d3534/components/connectivity-adapter/cmd/main.go#L24)

You can specify the following environment variables:

| Environment variable                    | Default value                                                                    | Description                                                                 |                                                                             
| ----------------------------------------| ---------------------------------------------------------------------------------| --------------------------------------------------------------------------- |
| **APP_ADDRESS**                         | `127.0.0.1:8080`                                                                 | Address and port for the service to listen on                               |                                                                             |
| **APP_SERVER_TIMEOUT**                  | `119s`                                                                           | The timeout used for incoming calls to the connectivity adapter server      |
| **APP_APP_REGISTRY_DIRECTOR_ENDPOINT**  | `127.0.0.1:3000/graphql`                                                         | GraphQL endpoint of the running Director component                          |                      
| **APP_APP_REGISTRY_CLIENT_TIMEOUT**     | `115s`                                                                           | Client timeout for calls to the running Director component                  |
| **APP_CONNECTOR_CONNECTOR_ENDPOINT**    | `http://compass-connector.compass-system.svc.cluster.local:3000/graphql`         | GraphQL endpoint of the running Connector component                         |
| **APP_CONNECTOR_CLIENT_TIMEOUT**        | `115s`                                                                           | Client timeout for calls to the running Connector component                 |
| **APP_CONNECTOR_ADAPTER_BASE_URL**      | `https://adapter-gateway.kyma.local`                                             | Token secured endpoint of the Connectivity Adapter component                |
| **APP_CONNECTOR_ADAPTER_MTLS_BASE_URL** | `https://adapter-gateway-mtls.kyma.local`                                        | Certificate secured endpoint of the Connectivity Adapter component          |
| **APP_LOG_LEVEL**                       | `info`                                                                           | Log level                                                                   |
| **APP_LOG_FORMAT**                      | `text`                                                                           | Format of the written logs. Supported values are `text` and `kibana`.        |
| **APP_LOG_OUTPUT**                      | `/dev/stdout`                                                                    | Log output location                                                         |
| **APP_LOG_FORMAT**                      | `text`                                                                           | Format of the written logs. Supported values are `text` and `kibana`.        |
