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

// This document is based on marshalling (and optionally enhancing) the returned document from the fixture fixORDDocumentWithBaseURL located in: /compass/components/director/internal/open_resource_discovery/fixtures_test.go
// If any breaking/validation change is applied to the fixture's Document structure, it must be applied here and in the constants used in the e2e test (/compass/tests/ord-aggregator/tests/handler_test.go) as well. Otherwise, the aggregator e2e test will fail.
// describedSystemInstance.baseUrl should be the same as the url of external services mock in the cluster
const ordDocument = `{
	"$schema": "./spec/v1/generated/Document.schema.json",
	"apiResources": [{
		"apiProtocol": "odata-v2",
		"apiResourceLinks": [{
			"type": "console",
			"url": "https://example.com/shell/discover"
		},
		{
			"type": "console",
			"url": "/shell/discover/relative"
		}],
		"changelogEntries": [{
			"date": "2020-04-29",
			"description": "loremipsumdolorsitamet",
			"releaseStatus": "active",
			"url": "https://example.com/changelog/v1",
			"version": "1.0.0"
		}],
		"countries": ["BG", "US"],
		"customImplementationStandard": null,
		"customImplementationStandardDescription": null,
		"description": "lorem ipsum dolor sit amet",
		"disabled": true,
		"entryPoints": ["https://exmaple.com/test/v1", "https://exmaple.com/test/v2"],
		"extensible": {
			"description": "Please find the extensibility documentation",
			"supported": "automatic"
		},
		"implementationStandard": "cff:open-service-broker:v2",
		"industry": ["automotive","test"],
		"labels": {
			"label-key-1": ["label-value-1", "label-value-2"]
		},
		"lineOfBusiness": ["lineOfBusiness2"],
		"links": [{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "https://example.com/2018/04/11/testing/"
		},
		{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "/testing/relative"
		}],
		"ordId": "ns:apiResource:API_ID:v2",
		"partOfConsumptionBundles": [{
			"defaultEntryPoint": "https://exmaple.com/test/v1",
			"ordId": "ns:consumptionBundle:BUNDLE_ID:v1"
		},
		{
			"defaultEntryPoint": "https://exmaple.com/test/v1",
			"ordId": "ns:consumptionBundle:BUNDLE_ID:v2"
		}],
		"partOfPackage": "ns:package:PACKAGE_ID:v1",
		"partOfProducts": ["ns:product:id:"],
		"releaseStatus": "active",
		"resourceDefinitions": [{
			"accessStrategies": [{
				"customDescription": "",
				"customType": "",
				"type": "open"
			}],
			"customType": "",
			"mediaType": "application/json",
			"type": "openapi-v3",
			"url": "http://localhost:8080/odata/1.0/catalog.svc/$value?type=json"
		},
		{
			"accessStrategies": [{
				"customDescription": "",
				"customType": "",
				"type": "open"
			}],
			"customType": "",
			"mediaType": "text/yaml",
			"type": "openapi-v3",
			"url": "https://test.com/odata/1.0/catalog"
		},
		{
			"accessStrategies": [{
				"customDescription": "",
				"customType": "",
				"type": "open"
			}],
			"customType": "",
			"mediaType": "application/xml",
			"type": "edmx",
			"url": "https://TEST:443//odata/$metadata"
		}],
		"shortDescription": "lorem ipsum",
		"successor": null,
		"sunsetDate": null,
		"systemInstanceAware": true,
		"tags": ["apiTestTag"],
		"title": "API TITLE",
		"version": "2.1.2",
		"visibility": "public"
	},
	{
		"apiProtocol": "odata-v2",
		"apiResourceLinks": [{
			"type": "console",
			"url": "https://example.com/shell/discover"
		},
		{
			"type": "console",
			"url": "/shell/discover/relative"
		}],
		"changelogEntries": [{
			"date": "2020-04-29",
			"description": "loremipsumdolorsitamet",
			"releaseStatus": "active",
			"url": "https://example.com/changelog/v1",
			"version": "1.0.0"
		}],
		"countries": ["BR"],
		"customImplementationStandard": null,
		"customImplementationStandardDescription": null,
		"description": "lorem ipsum dolor sit amet",
		"disabled": null,
		"entryPoints": ["http://localhost:8080/some-api/v1"],
		"extensible": {
			"description": "Please find the extensibility documentation",
			"supported": "automatic"
		},
		"implementationStandard": "cff:open-service-broker:v2",
		"industry": ["automotive","test"],
		"labels": {
			"label-key-1": ["label-value-1", "label-value-2"]
		},
		"lineOfBusiness": ["lineOfBusiness2"],
		"links": [{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "https://example.com/2018/04/11/testing/"
		},
		{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "/testing/relative"
		}],
		"ordId": "ns:apiResource:API_ID2:v1",
		"partOfConsumptionBundles": [{
			"defaultEntryPoint": "",
			"ordId": "ns:consumptionBundle:BUNDLE_ID:v1"
		},
 		{
			"ordId": "ns:consumptionBundle:BUNDLE_ID:v2",
			"defaultEntryPoint": ""
		}],
		"partOfPackage": "ns:package:PACKAGE_ID:v1",
		"partOfProducts": ["ns:product:id:"],
		"releaseStatus": "deprecated",
		"resourceDefinitions": [{
			"accessStrategies": [{
				"customDescription": "",
				"customType": "",
				"type": "open"
			}],
			"customType": "",
			"mediaType": "application/xml",
			"type": "edmx",
			"url": "https://TEST:443//odata/$metadata"
		},
		{
			"accessStrategies": [{
				"customDescription": "",
				"customType": "",
				"type": "open"
			}],
			"customType": "",
			"mediaType": "application/json",
			"type": "openapi-v3",
			"url": "http://localhost:8080/odata/1.0/catalog.svc/$value?type=json"
		}],
		"shortDescription": "lorem ipsum",
		"successor": "ns:apiResource:API_ID:v2",
		"sunsetDate": "2020-12-08T15:47:04+0000",
		"systemInstanceAware": true,
		"tags": ["ZGWSAMPLE"],
		"title": "Gateway Sample Service",
		"version": "1.1.0",
		"visibility": "public"
	}],
	"consumptionBundles":[
        {
            "credentialExchangeStrategies":[
                {
                    "callbackUrl":"/credentials/relative",
                    "customType":"ns:credential-exchange:v1",
                    "type":"custom"
                },
                {
                    "callbackUrl":"http://example.com/credentials",
                    "customType":"ns:credential-exchange2:v3",
                    "type":"custom"
                }
            ],
            "description":"lorem ipsum dolor nsq sme",
            "labels":{
                "label-key-1":[
                    "label-value-1",
                    "label-value-2"
                ]
            },
            "links":[
                {
                    "description":"loremipsumdolornem",
                    "title":"LinkTitle",
                    "url":"https://example.com/2018/04/11/testing/"
                },
                {
                    "description":"loremipsumdolornem",
                    "title":"LinkTitle",
                    "url":"/testing/relative"
                }
            ],
            "ordId":"ns:consumptionBundle:BUNDLE_ID:v1",
            "shortDescription":"lorem ipsum",
            "title":"BUNDLE TITLE"
        },
        {
            "title":"BUNDLE TITLE 2",
            "description":"foo bar",
            "ordId":"ns:consumptionBundle:BUNDLE_ID:v2",
            "shortDescription":"foo",
            "links":[
                {
                    "description":"loremipsumdolornem",
                    "title":"LinkTitle",
                    "url":"https://example.com/2018/04/11/testing/"
                },
                {
                    "description":"loremipsumdolornem",
                    "title":"LinkTitle",
                    "url":"/testing/relative"
                }
            ],
            "labels":{
                "label-key-1":[
                    "label-value-1",
                    "label-value-2"
                ]
            },
            "credentialExchangeStrategies":[
                {
                    "callbackUrl":"/credentials/relative",
                    "customType":"ns:credential-exchange:v1",
                    "type":"custom"
                },
                {
                    "callbackUrl":"http://example.com/credentials",
                    "customType":"ns:credential-exchange2:v3",
                    "type":"custom"
                }
            ]
        }
    ],
	"describedSystemInstance": {
		"ApplicationTemplateID": null,
		"baseUrl": "http://compass-external-services-mock.compass-system.svc.cluster.local:8080",
		"Description": null,
		"HealthCheckURL": null,
		"IntegrationSystemID": null,
		"labels": {
			"label-key-1": ["label-value-1", "label-value-2"]
		},
		"Name": "",
		"ProviderName": null,
		"Status": null,
		"Tenant": ""
	},
	"description": "Test Document",
	"eventResources": [{
		"changelogEntries": [{
			"date": "2020-04-29",
			"description": "loremipsumdolorsitamet",
			"releaseStatus": "active",
			"url": "https://example.com/changelog/v1",
			"version": "1.0.0"
		}],
		"countries": ["BG", "US"],
		"description": "lorem ipsum dolor sit amet",
		"disabled": true,
		"extensible": {
			"description": "Please find the extensibility documentation",
			"supported": "automatic"
		},
		"industry": ["automotive", "test"],
		"labels": {
			"label-key-1": ["label-value-1", "label-value-2"]
		},
		"lineOfBusiness": ["lineOfBusiness2"],
		"links": [{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "https://example.com/2018/04/11/testing/"
		},
		{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "/testing/relative"
		}],
		"ordId": "ns:eventResource:EVENT_ID:v1",
		"partOfConsumptionBundles": [{
			"defaultEntryPoint": "",
			"ordId": "ns:consumptionBundle:BUNDLE_ID:v1"
		},
		{
			"defaultEntryPoint": "",
			"ordId": "ns:consumptionBundle:BUNDLE_ID:v2"
		}],
		"partOfPackage": "ns:package:PACKAGE_ID:v1",
		"partOfProducts": ["ns:product:id:"],
		"releaseStatus": "active",
		"resourceDefinitions": [{
			"accessStrategies": [{
				"customDescription": "",
				"customType": "",
				"type": "open"
			}],
			"customType": "",
			"mediaType": "application/json",
			"type": "asyncapi-v2",
			"url": "http://localhost:8080/asyncApi2.json"
		}],
		"shortDescription": "lorem ipsum",
		"successor": null,
		"sunsetDate": null,
		"systemInstanceAware": true,
		"tags": ["eventTestTag"],
		"title": "EVENT TITLE",
		"version": "2.1.2",
		"visibility": "public"
	},
	{
		"changelogEntries": [{
			"date": "2020-04-29",
			"description": "loremipsumdolorsitamet",
			"releaseStatus": "active",
			"url": "https://example.com/changelog/v1",
			"version": "1.0.0"
		}],
		"countries": ["BR"],
		"description": "lorem ipsum dolor sit amet",
		"disabled": null,
		"extensible": {
			"description": "Please find the extensibility documentation",
			"supported": "automatic"
		},
		"industry": ["automotive", "test"],
		"labels": {
			"label-key-1": ["label-value-1", "label-value-2"]
		},
		"lineOfBusiness": ["lineOfBusiness2"],
		"links": [{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "https://example.com/2018/04/11/testing/"
		},
		{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "/testing/relative"
		}],
		"ordId": "ns2:eventResource:EVENT_ID:v1",
		"partOfConsumptionBundles": [{
			"defaultEntryPoint": "",
			"ordId": "ns:consumptionBundle:BUNDLE_ID:v1"
		},
		{
			"defaultEntryPoint": "",
			"ordId": "ns:consumptionBundle:BUNDLE_ID:v2"
		}],
		"partOfPackage": "ns:package:PACKAGE_ID:v1",
		"partOfProducts": ["ns:product:id:"],
		"releaseStatus": "deprecated",
		"resourceDefinitions": [{
			"accessStrategies": [{
				"customDescription": "",
				"customType": "",
				"type": "open"
			}],
			"customType": "",
			"mediaType": "application/json",
			"type": "asyncapi-v2",
			"url": "http://localhost:8080/api/eventCatalog.json"
		}],
		"shortDescription": "lorem ipsum",
		"successor": "ns2:eventResource:EVENT_ID:v1",
		"sunsetDate": "2020-12-08T15:47:04+0000",
		"systemInstanceAware": true,
		"tags": ["eventTestTag2"],
		"title": "EVENT TITLE 2",
		"version": "1.1.0",
		"visibility": "public"
	}],
	"openResourceDiscovery": "1.0-rc.3",
	"packages": [{
		"countries": ["BG", "EN"],
		"customPolicyLevel": null,
		"description": "lorem ipsum dolor set",
		"industry": ["automotive", "finance"],
		"labels": {
			"label-key-1": ["label-val"],
			"pkg-label": ["label-val"]
		},
		"licenseType": "licence",
		"lineOfBusiness": ["lineOfBusiness"],
		"links": [{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "https://example.com/2018/04/11/testing/"
		},
		{
			"description": "loremipsumdolornem",
			"title": "LinkTitle",
			"url": "/testing/relative"
		}],
		"ordId": "ns:package:PACKAGE_ID:v1",
		"packageLinks": [{
			"type": "terms-of-service",
			"url": "https://example.com/en/legal/terms-of-use.html"
		},
		{
			"type": "client-registration",
			"url": "/ui/public/showRegisterForm"
		}],
		"partOfProducts": ["ns:product:id:"],
		"policyLevel": "sap",
		"shortDescription": "lorem ipsum",
		"tags": ["testTag"],
		"title": "PACKAGE 1 TITLE",
		"vendor": "ns:vendor:id:",
		"version": "1.1.2"
	}],
	"products": [{
		"correlationIds": [
			"foo.bar.baz:123456",
			"foo.bar.baz:654321"
		],
		"labels": {
			"label-key-1": ["label-value-1", "label-value-2"]
		},
		"ordId": "ns:product:id:",
		"parent": "ns:product:id2:",
		"shortDescription": "lorem ipsum",
		"title": "PRODUCT TITLE",
		"vendor": "ns:vendor:id:"
	}],
	"providerSystemInstance": null,
	"tombstones": [{
		"ordId": "ns:apiResource:API_ID2:v1",
		"removalDate": "2020-12-02T14:12:59Z"
	}],
	"vendors": [{
		"labels": {
			"label-key-1": ["label-value-1", "label-value-2"]
		},
		"ordId": "ns:vendor:id:",
		"partners": [
		"microsoft:vendor:Microsoft:"
		],
		"title": "SAP"
	}]
}`
