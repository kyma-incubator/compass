package ord_aggregator

// This document is created by simply marshalling the returned document from the fixture fixWellKnownConfig located in: /compass/components/director/internal/open_resource_discovery/fixtures_test.go
// If any breaking/validation change is applied to the fixture's WellKnownConfig structure, it must be applied here as well. Otherwise, the aggregator e2e test will fail.
const ordConfig = `{
    "$schema": "../spec/v1/generated/Configuration.schema.json",
    "openResourceDiscoveryV1": {
        "documents": [
            {
                "url": "/open-resource-discovery/v1/documents/example1",
                "systemInstanceAware": true,
                "accessStrategies": [
                    {
                        "type": "open",
                        "customType": "",
                        "customDescription": ""
                    }
                ]
            }
        ]
    }
}`

// This document is created by simply marshalling the returned document from the fixture fixORDDocumentWithBaseURL located in: /compass/components/director/internal/open_resource_discovery/fixtures_test.go
// If any breaking/validation change is applied to the fixture's Document structure, it must be applied here and in the constants used in the e2e test (/compass/tests/ord-aggregator/tests/handler_test.go) as well. Otherwise, the aggregator e2e test will fail.
const ordDocument = `{
    "$schema": "./spec/v1/generated/Document.schema.json",
    "openResourceDiscovery": "1.0-rc.1",
    "description": "Test Document",
    "describedSystemInstance": {
        "ProviderName": null,
        "Tenant": "",
        "Name": "",
        "Description": null,
        "Status": null,
        "HealthCheckURL": null,
        "IntegrationSystemID": null,
        "baseUrl": "http://compass-external-services-mock.compass-system.svc.cluster.local:8080",
        "labels": {
            "label-key-1": [
                "label-value-1",
                "label-value-2"
            ]
        }
    },
    "providerSystemInstance": null,
    "packages": [
        {
            "ordId": "ns:package:PACKAGE_ID:v1",
            "vendor": "sap",
            "title": "PACKAGE 1 TITLE",
            "shortDescription": "lorem ipsum",
            "description": "lorem ipsum dolor set",
            "version": "1.1.2",
            "packageLinks": [
                {
                    "type": "terms-of-service",
                    "url": "https://example.com/en/legal/terms-of-use.html"
                },
                {
                    "type": "client-registration",
                    "url": "/ui/public/showRegisterForm"
                }
            ],
            "links": [
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "https://example.com/2018/04/11/testing/"
                },
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "/testing/relative"
                }
            ],
            "licenseType": "licence",
            "tags": [
                "testTag"
            ],
            "countries": [
                "BG",
                "EN"
            ],
            "labels": {
                "label-key-1": [
                    "label-val"
                ],
                "pkg-label": [
                    "label-val"
                ]
            },
            "policyLevel": "sap",
            "customPolicyLevel": null,
            "partOfProducts": [
                "ns:PRODUCT_ID"
            ],
            "lineOfBusiness": [
                "lineOfBusiness"
            ],
            "industry": [
                "automotive",
                "finance"
            ]
        }
    ],
    "consumptionBundles": [
        {
            "title": "BUNDLE TITLE",
            "description": "lorem ipsum dolor nsq sme",
            "ordId": "ns:consumptionBundle:BUNDLE_ID:v1",
            "shortDescription": "lorem ipsum",
            "links": [
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "https://example.com/2018/04/11/testing/"
                },
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "/testing/relative"
                }
            ],
            "labels": {
                "label-key-1": [
                    "label-value-1",
                    "label-value-2"
                ]
            },
            "credentialExchangeStrategies": [
                {
                    "callbackUrl": "/credentials/relative",
                    "customType": "ns:credential-exchange:v1",
                    "type": "custom"
                },
                {
                    "callbackUrl": "http://example.com/credentials",
                    "customType": "ns:credential-exchange2:v3",
                    "type": "custom"
                }
            ]
        }
    ],
    "products": [
        {
            "id": "ns:PRODUCT_ID",
            "title": "PRODUCT TITLE",
            "shortDescription": "lorem ipsum",
            "vendor": "sap",
            "parent": "ns:PRODUCT_ID2",
            "sapPpmsObjectId": "12391293812",
            "labels": {
                "label-key-1": [
                    "label-value-1",
                    "label-value-2"
                ]
            }
        }
    ],
    "apiResources": [
        {
            "partOfConsumptionBundle": "ns:consumptionBundle:BUNDLE_ID:v1",
            "partOfPackage": "ns:package:PACKAGE_ID:v1",
            "title": "API TITLE",
            "description": "lorem ipsum dolor sit amet",
            "entryPoint": "https://exmaple.com/test/v1",
            "ordId": "ns:apiResource:API_ID:v2",
            "shortDescription": "lorem ipsum",
            "systemInstanceAware": true,
            "apiProtocol": "odata-v2",
            "tags": [
                "apiTestTag"
            ],
            "countries": [
                "BG",
                "US"
            ],
            "links": [
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "https://example.com/2018/04/11/testing/"
                },
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "/testing/relative"
                }
            ],
            "apiResourceLinks": [
                {
                    "type": "console",
                    "url": "https://example.com/shell/discover"
                },
                {
                    "type": "console",
                    "url": "/shell/discover/relative"
                }
            ],
            "releaseStatus": "active",
            "sunsetDate": null,
            "successor": null,
            "changelogEntries": [
                {
                    "date": "2020-04-29",
                    "description": "loremipsumdolorsitamet",
                    "releaseStatus": "active",
                    "url": "https://example.com/changelog/v1",
                    "version": "1.0.0"
                }
            ],
            "labels": {
                "label-key-1": [
                    "label-value-1",
                    "label-value-2"
                ]
            },
            "visibility": "public",
            "disabled": true,
            "partOfProducts": [
                "ns:PRODUCT_ID"
            ],
            "lineOfBusiness": [
                "lineOfBusiness2"
            ],
            "industry": [
                "automotive",
                "test"
            ],
            "resourceDefinitions": [
                {
                    "type": "openapi-v3",
                    "customType": "",
                    "mediaType": "application/json",
                    "url": "http://localhost:8080/odata/1.0/catalog.svc/$value?type=json",
                    "accessStrategies": [
                        {
                            "type": "open",
                            "customType": "",
                            "customDescription": ""
                        }
                    ]
                },
                {
                    "type": "openapi-v3",
                    "customType": "",
                    "mediaType": "text/yaml",
                    "url": "https://test.com/odata/1.0/catalog",
                    "accessStrategies": [
                        {
                            "type": "open",
                            "customType": "",
                            "customDescription": ""
                        }
                    ]
                }
            ],
            "version": "2.1.2"
        },
        {
            "partOfConsumptionBundle": "ns:consumptionBundle:BUNDLE_ID:v1",
            "partOfPackage": "ns:package:PACKAGE_ID:v1",
            "title": "Gateway Sample Service",
            "description": "lorem ipsum dolor sit amet",
            "entryPoint": "http://localhost:8080/some-api/v1",
            "ordId": "ns:apiResource:API_ID2:v1",
            "shortDescription": "lorem ipsum",
            "systemInstanceAware": true,
            "apiProtocol": "odata-v2",
            "tags": [
                "ZGWSAMPLE"
            ],
            "countries": [
                "BR"
            ],
            "links": [
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "https://example.com/2018/04/11/testing/"
                },
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "/testing/relative"
                }
            ],
            "apiResourceLinks": [
                {
                    "type": "console",
                    "url": "https://example.com/shell/discover"
                },
                {
                    "type": "console",
                    "url": "/shell/discover/relative"
                }
            ],
            "releaseStatus": "deprecated",
            "sunsetDate": "2020-12-08T15:47:04+0000",
            "successor": "ns:apiResource:API_ID:v2",
            "changelogEntries": [
                {
                    "date": "2020-04-29",
                    "description": "loremipsumdolorsitamet",
                    "releaseStatus": "active",
                    "url": "https://example.com/changelog/v1",
                    "version": "1.0.0"
                }
            ],
            "labels": {
                "label-key-1": [
                    "label-value-1",
                    "label-value-2"
                ]
            },
            "visibility": "public",
            "disabled": null,
            "partOfProducts": [
                "ns:PRODUCT_ID"
            ],
            "lineOfBusiness": [
                "lineOfBusiness2"
            ],
            "industry": [
                "automotive",
                "test"
            ],
            "resourceDefinitions": [
                {
                    "type": "edmx",
                    "customType": "",
                    "mediaType": "application/xml",
                    "url": "https://TEST:443//odata/$metadata",
                    "accessStrategies": [
                        {
                            "type": "open",
                            "customType": "",
                            "customDescription": ""
                        }
                    ]
                }
            ],
            "version": "1.1.0"
        }
    ],
    "eventResources": [
        {
            "partOfConsumptionBundle": "ns:consumptionBundle:BUNDLE_ID:v1",
            "partOfPackage": "ns:package:PACKAGE_ID:v1",
            "title": "EVENT TITLE",
            "description": "lorem ipsum dolor sit amet",
            "ordId": "ns:eventResource:EVENT_ID:v1",
            "shortDescription": "lorem ipsum",
            "systemInstanceAware": true,
            "changelogEntries": [
                {
                    "date": "2020-04-29",
                    "description": "loremipsumdolorsitamet",
                    "releaseStatus": "active",
                    "url": "https://example.com/changelog/v1",
                    "version": "1.0.0"
                }
            ],
            "links": [
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "https://example.com/2018/04/11/testing/"
                },
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "/testing/relative"
                }
            ],
            "tags": [
                "eventTestTag"
            ],
            "countries": [
                "BG",
                "US"
            ],
            "releaseStatus": "active",
            "sunsetDate": null,
            "successor": null,
            "labels": {
                "label-key-1": [
                    "label-value-1",
                    "label-value-2"
                ]
            },
            "visibility": "public",
            "disabled": true,
            "partOfProducts": [
                "ns:PRODUCT_ID"
            ],
            "lineOfBusiness": [
                "lineOfBusiness2"
            ],
            "industry": [
                "automotive",
                "test"
            ],
            "resourceDefinitions": [
                {
                    "type": "asyncapi-v2",
                    "customType": "",
                    "mediaType": "application/json",
                    "url": "http://localhost:8080/asyncApi2.json",
                    "accessStrategies": [
                        {
                            "type": "open",
                            "customType": "",
                            "customDescription": ""
                        }
                    ]
                }
            ],
            "version": "2.1.2"
        },
        {
            "partOfConsumptionBundle": "ns:consumptionBundle:BUNDLE_ID:v1",
            "partOfPackage": "ns:package:PACKAGE_ID:v1",
            "title": "EVENT TITLE 2",
            "description": "lorem ipsum dolor sit amet",
            "ordId": "ns2:eventResource:EVENT_ID:v1",
            "shortDescription": "lorem ipsum",
            "systemInstanceAware": true,
            "changelogEntries": [
                {
                    "date": "2020-04-29",
                    "description": "loremipsumdolorsitamet",
                    "releaseStatus": "active",
                    "url": "https://example.com/changelog/v1",
                    "version": "1.0.0"
                }
            ],
            "links": [
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "https://example.com/2018/04/11/testing/"
                },
                {
                    "description": "loremipsumdolornem",
                    "title": "LinkTitle",
                    "url": "/testing/relative"
                }
            ],
            "tags": [
                "eventTestTag2"
            ],
            "countries": [
                "BR"
            ],
            "releaseStatus": "deprecated",
            "sunsetDate": "2020-12-08T15:47:04+0000",
            "successor": "ns2:eventResource:EVENT_ID:v1",
            "labels": {
                "label-key-1": [
                    "label-value-1",
                    "label-value-2"
                ]
            },
            "visibility": "public",
            "disabled": null,
            "partOfProducts": [
                "ns:PRODUCT_ID"
            ],
            "lineOfBusiness": [
                "lineOfBusiness2"
            ],
            "industry": [
                "automotive",
                "test"
            ],
            "resourceDefinitions": [
                {
                    "type": "asyncapi-v2",
                    "customType": "",
                    "mediaType": "application/json",
                    "url": "http://localhost:8080/api/eventCatalog.json",
                    "accessStrategies": [
                        {
                            "type": "open",
                            "customType": "",
                            "customDescription": ""
                        }
                    ]
                }
            ],
            "version": "1.1.0"
        }
    ],
    "tombstones": [
        {
            "ordId": "ns:apiResource:API_ID2:v1",
            "removalDate": "2020-12-02T14:12:59Z"
        }
    ],
    "vendors": [
        {
            "id": "sap",
            "title": "SAP",
            "type": "sap",
            "labels": {
                "label-key-1": [
                    "label-value-1",
                    "label-value-2"
                ]
            }
        }
    ]
}`
