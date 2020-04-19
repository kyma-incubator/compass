# Kyma Environment Broker Endpoints

Kyma Environment Broker (KEB) implements the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/profile.md#service-metadata) (OSB API). All the OSB API endpoints are served with the following prefixes: 

| Prefix            | Description                                                                                                                                                                                                                                      |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `/`               | Defines prefix for legacy endpoint secured by basic access authentication. EDP is configured with region which defaults to **broker.defaultRequestRegion** value specified in [values.yaml](./../../chart/compass/charts/kyma-environment-broker/values.yaml) file. |
| `/{region}`       | Defines prefix for legacy endpoint secured by basic access authentication. EDP is configured with region value specified in request.                                                                                                                                |
| `/oauth`          | Defines prefix for endpoint secured by OAuth2 authorization. EDP is configured with region which defaults to **broker.defaultRequestRegion** value specified in [values.yaml](./../../chart/compass/charts/kyma-environment-broker/values.yaml) file.               |
| `/oauth/{region}` | Defines prefix for endpoint secured by OAuth2 authorization. EDP is configured with region value specified in request.                                                                                                                           |


> **NOTE:** The KEB does not implement the OSB API update operation.

Beside OSB API endpoints KEB exposes the REST `/info/runtimes` endpoint that provides information about all created Runtimes, both succeeded and failed. This endpoint is secured with the OAuth2 authorization.
