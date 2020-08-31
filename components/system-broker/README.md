# System Broker

The System Broker exposes OSB API as well as an endpoint for fetching specifications.

## Development

In order to run System Broker locally without minikube replace the `oauthTokenProvider` in `main.go` with the following:

```golang
oauthTokenProvider := oauth.NewTokenProviderFromValue("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OndyaXRlIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OnJlYWQgaGVhbHRoX2NoZWNrczpyZWFkIGFwcGxpY2F0aW9uOndyaXRlIHJ1bnRpbWU6d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjp3cml0ZSBsYWJlbF9kZWZpbml0aW9uOnJlYWQgcnVudGltZTpyZWFkIHRlbmFudDpyZWFkIiwidGVuYW50IjoiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlIn0.")
``` 

Now run the system broker:

```bash
./run.sh
```

By default, the system broker API is accessible at `locahost:8080/broker`.

In order to run full Kyma on minikube and Compass with system broker:

1. Checkout `system-broker`, navigate to `compass/installation/cmd` and execute `./run.sh --kyma-installation full`. Specify `--kyma-installation minimal` if you want to run just Compass with minimal Kyma dependencies.
2. When the installer prints out `Status: InProgress, description: install component compass` navigate to the `system-broker` directory and execute `make deploy-on-minikube`
3. Checkout `global_applications` branch, navigate to the `director` directory and execute `make deploy-on-minikube` 

By default, the system broker API is accessible at `https://compass-gateway.kyma.local/broker`.

## Configuration

The System Broker binary allows to override some configuration parameters. You can specify following as command line flags or as environment variables.

[Config struct](https://github.com/kyma-incubator/compass/blob/0f0eeb38e7a5d8db655b6870138e5add257ebb1d/components/system-broker/internal/config/config.go#L30) is self descriptive. Default values can be found in `DefaultConfig` method. 
Example for using env variables can be found in the [helm charts](https://github.com/kyma-incubator/compass/blob/4b49dae2cce65f0efa98d0a9e664ae65c0f059f8/chart/compass/charts/system-broker/templates/deployment.yaml#L53).

## Usage

Provision and deprovision are async and require `accepts_incomplete=true`. Bind and unbind are sync.

### Specifications

URLs pointing to the specs API are included in the OSB plan metadata as part of the catalog response.
Specs API returns a single JSON, XML or YAML document containing the specification of the api or event definition defined by the specified query parameters.

Example link to a specification file:
 `https://compass-gateway.kyma.local/broker/specifications?app_id=53acc071-42ec-4561-962d-bf3dbc286cb7&package_id=b2bb4664-930b-491f-a922-8ac586ec84f9&definition_id=0c17b77e-530b-47d8-a23f-ad462ed4ee0a`

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
                                "specification_url": "https://compass-gateway.kyma.local/broker/specifications?app_id=53acc071-42ec-4561-962d-bf3dbc286cb7&package_id=b2bb4664-930b-491f-a922-8ac586ec84f9&definition_id=0c17b77e-530b-47d8-a23f-ad462ed4ee0a"
                            },
                            {
                                "definition_id": "240031e6-4f15-40f0-b93e-948040020f70",
                                "definition_name": "Commerce Webservices",
                                "specification_category": "api_definition",
                                "specification_format": "application/json",
                                "specification_type": "OPEN_API",
                                "specification_url": "https://compass-gateway.kyma.local/broker/specifications?app_id=53acc071-42ec-4561-962d-bf3dbc286cb7&package_id=b2bb4664-930b-491f-a922-8ac586ec84f9&definition_id=240031e6-4f15-40f0-b93e-948040020f70"
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

For the example above, the package instance auth credentials were set with the [following mutation](https://github.com/kyma-incubator/compass/blob/1c4490318bfd39cbab5e6b2b1c9a78f3ec0ce10d/components/director/examples/set-package-instance-auth/set-package-instance-auth.graphql).
