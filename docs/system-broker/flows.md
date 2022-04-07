# System Broker Flows

The System Broker is an [Open Service Broker API](https://www.openservicebrokerapi.org/) (OSB) compliant component, which provides access to applications (systems) credentials for OSB platforms.

## Broker Catalog
The OSB catalog contains services and their plans. Compass represents applications as services and their bundles as plans. This way, you can requests credentials for a specific bundle.

### Specifications
URLs that point to the specifications API are included in the OSB plan metadata as part of the catalog response. Specifications API returns a single JSON, XML, or YAML document that contains the specification of the API, or event definition defined by the specified query parameters.

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

## Provision and Deprovision
Provision and deprovision are asynchronous operations that require the `accepts_incomplete=true` parameter. These operations either create or delete credentials for the chosen application and bundle.

## Bind and Unbind
Both bind and unbind operations are synchronous and do not trigger any external calls. They just retrieve or delete the credentials in the Director DB, which were created during the instance provisioning.

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
