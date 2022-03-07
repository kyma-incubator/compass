# System Broker

The System Broker exposes OSB API as well as an endpoint for fetching specifications.

## Development

### Local

The System Broker component can be run locally using the `run.sh` script.

```bash
./run.sh
```

By default, the System Broker API is accessible at: `locahost:8080/broker`. 

Calls made to the API should provide **Authorization** and **X-Broker-API-Version** headers.  The following token 
can be used for authorization:

`eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OndyaXRlIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OnJlYWQgaGVhbHRoX2NoZWNrczpyZWFkIGFwcGxpY2F0aW9uOndyaXRlIHJ1bnRpbWU6d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjp3cml0ZSBsYWJlbF9kZWZpbml0aW9uOnJlYWQgcnVudGltZTpyZWFkIHRlbmFudDpyZWFkIiwidGVuYW50IjoiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlIn0.`

Thus, an example call to the local broker looks like the following:

`curl -H "X-Broker-API-Version: 2.15" -H "Authorization: Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OndyaXRlIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OnJlYWQgaGVhbHRoX2NoZWNrczpyZWFkIGFwcGxpY2F0aW9uOndyaXRlIHJ1bnRpbWU6d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjp3cml0ZSBsYWJlbF9kZWZpbml0aW9uOnJlYWQgcnVudGltZTpyZWFkIHRlbmFudDpyZWFkIiwidGVuYW50IjoiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlIn0." localhost:8080/broker/v2/catalog`

#### Note
The System Broker makes API calls to the Director component. Therefore, to use the System Broker successfully, the 
Director component must be started, too. To do this, go to the `compass/components/director/` directory and run the `run.sh` 
script.

### Minikube

The System Broker is a component that is part of the Compass installation, so it is enough to start Compass on `minikube`.

To do this, perform the following procedure:

1. Navigate to `compass/installation/cmd`.
2. Execute `./run.sh --kyma-installation full`.
This script starts Compass on top of a fully-featured Kyma cluster. Alternativey, if you want to run Compass with minimal Kyma dependencies, you can use the `--kyma-installation minimal` argument.
3. Access the System Broker API at the following URL: `https://compass-gateway-mtls.kyma.local/broker`.

## Configuration

The System Broker binary allows to override some configuration parameters. You can specify the following as command line arguments or as environment variables.

The configuration structure ([config.go](https://github.com/kyma-incubator/compass/blob/0f0eeb38e7a5d8db655b6870138e5add257ebb1d/components/system-broker/internal/config/config.go#L30)) is self descriptive. Default values can be found in the `DefaultConfig` method. 
You can find an example for using environment variables in the following Helm charts at: [deployment.yaml](https://github.com/kyma-incubator/compass/blob/4b49dae2cce65f0efa98d0a9e664ae65c0f059f8/chart/compass/charts/system-broker/templates/deployment.yaml#L53).

## Usage

Provision and deprovision are asynchronous and require `accepts_incomplete=true`. Bind and unbind are synchronous.

### Specifications

URLs pointing to the specifications API are included in the OSB plan metadata as part of the catalog response.
Specifications API returns a single JSON, XML, or YAML document containing the specification of the API, or event definition defined by the specified query parameters.

Example link to a specifications file:
 `https://compass-gateway.kyma.local/broker/specifications?app_id=53acc071-42ec-4561-962d-bf3dbc286cb7&bundle_id=b2bb4664-930b-491f-a922-8ac586ec84f9&definition_id=0c17b77e-530b-47d8-a23f-ad462ed4ee0a`

Example catalog containing specifications metadata:
```json
{
    "services": [
        {
            "id": "53acc071-42ec-4561-962d-bf3dbc286cb7",
            "name": "commerce",
            "description": "commerce",
            "bindable": true,
            "plan_updateable": false,
            "plans": [
                {
                    "id": "b2bb4664-930b-491f-a922-8ac586ec84f9",
                    "name": "SAP Commerce Cloud",
                    "description": "SAP Commerce Cloud",
                    "bindable": true,
                    "metadata": {
                        "specifications": [
                            {
                                "definition_id": "0c17b77e-530b-47d8-a23f-ad462ed4ee0a",
                                "definition_name": "Inbound OMM OrderEntry",
                                "specification_category": "api_definition",
                                "specification_format": "application/xml",
                                "specification_type": "ODATA",
                                "specification_url": "https://compass-gateway-mtls.kyma.local/open-resource-discovery-static/v0/api/0c17b77e-530b-47d8-a23f-ad462ed4ee0a/specification/b2bb4664-930b-491f-a922-8ac586ec84f9"
                            },
                            {
                                "definition_id": "240031e6-4f15-40f0-b93e-948040020f70",
                                "definition_name": "Commerce Webservices",
                                "specification_category": "api_definition",
                                "specification_format": "application/json",
                                "specification_type": "OPEN_API",
                                "specification_url": "https://compass-gateway-mtls.kyma.local/open-resource-discovery-static/v0/api/53acc071-42ec-4561-962d-bf3dbc286cb7/specification/240031e6-4f15-40f0-b93e-948040020f70"
                            }
                        ]
                    }
                }
            ],
            "metadata": {
                "displayName": "commerce",
                "integrationSystemID": "",
                "name": "commerce",
                "providerDisplayName": "commerce",
                "scenarios": [
                    "DEFAULT"
                ]
            }
        }
    ]
}
```

### Bindings Format

Example OSB bind response:

```json
{
    "credentials": {
        "id": "52ae3582-0052-4114-87ea-617b6854fb7f",
        "credentials_type": "basic_auth",
        "target_urls": {
            "comments-v1": "http://mywordpress.com/comments",
            "reviews-v1": "http://mywordpress.com/reviews",
            "xml": "http://mywordpress.com/xml"
        },
        "auth_details": {
            "request_parameters": {
                "headers": {
                    "header-A": [
                        "ha1",
                        "ha2"
                    ],
                    "header-B": [
                        "hb1",
                        "hb2"
                    ]
                },
                "query_parameters": {
                    "qA": [
                        "qa1",
                        "qa2"
                    ],
                    "qB": [
                        "qb1",
                        "qb2"
                    ]
                }
            },
            "auth": {
                "username": "admin",
                "password": "secret"
            }
        }
    }
}
```

For the example above, the bundle instance auth credentials were set with the following mutation: [set-bundle-instance-auth.graphql](https://github.com/kyma-incubator/compass/blob/v1.27.0/components/director/examples/set-bundle-instance-auth/set-bundle-instance-auth.graphql).
