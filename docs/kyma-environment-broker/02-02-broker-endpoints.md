# Kyma Environment Broker Endpoints

Kyma Environment Broker (KEB) implements the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/profile.md#service-metadata) (OSB API). All the OSB API endpoints are served with the following prefixes: 

| Prefix            | Description                                                                                                                                                                                                                                      |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `/`               | Defines a prefix for the legacy endpoint secured with the basic access authentication. EDP is configured with a region whose default value is specified under the **broker.defaultRequestRegion** parameter in the [`values.yaml`](./../../chart/compass/charts/kyma-environment-broker/values.yaml) file. |
| `/{region}`       | Defines a prefix for the legacy endpoint secured with the basic access authentication. EDP is configured with the region value specified in the request.                                                                                                                                |
| `/oauth`          | Defines a prefix for the endpoint secured with the OAuth2 authorization. EDP is configured with a region whose default value is specified under the **broker.defaultRequestRegion** parameter in the [`values.yaml`](./../../chart/compass/charts/kyma-environment-broker/values.yaml) file.               |
| `/oauth/{region}` | Defines a prefix for the endpoint secured with the OAuth2 authorization. EDP is configured with the region value specified in the request.                                                                                                                           |


> **NOTE:** KEB does not implement the OSB API update operation.

Besides OSB API endpoints, KEB exposes the REST `/info/runtimes` endpoint that provides information about all created Runtimes, both succeeded and failed. This endpoint is secured with the OAuth2 authorization.
