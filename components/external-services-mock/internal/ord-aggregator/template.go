package ord_aggregator

// This document is created by simply marshalling the returned document from the fixture fixWellKnownConfig located in: /compass/components/director/internal/open_resource_discovery/fixtures_test.go
// If any breaking/validation change is applied to the fixture's WellKnownConfig structure, it must be applied here as well. Otherwise, the aggregator e2e test will fail.
const ordConfig = `{
    "$schema": "../spec/v1/generated/Configuration.schema.json",
    %s
	"openResourceDiscoveryV1": {
        "documents": [
            {
                "url": "%s",
                "systemInstanceAware": true,
                "accessStrategies": [
                    {
                        "type": "%s",
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
   "$schema":"./spec/v1/generated/Document.schema.json",
   "openResourceDiscovery":"1.2",
   "policyLevel": "sap:core:v1",
   "description":"Test Document",
   "describedSystemInstance":{
      "baseUrl":"{{ .baseURL }}",
      "labels":{
         "label-key-1":[
            "label-value-1",
            "label-value-2"
         ]
      }
	  
   },
   "packages":[
      {
         "ordId":"ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
         "vendor":"sap:vendor:SAP:",
         "title":"PACKAGE 1 TITLE",
         "shortDescription":"Short description",
         "description":"lorem ipsum dolor set",
         "version":"1.1.2",
         "packageLinks":[
            {
               "type":"terms-of-service",
               "url":"https://example.com/en/legal/terms-of-use.html"
            },
            {
               "type":"client-registration",
               "url":"https://ui/public/showRegisterForm"
            }
         ],
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://testing/relative"
            }
         ],
         "licenseType":"AAL",
         "tags":[
            "testTag"
         ],
         "countries":[
            "BG",
            "US"
         ],
         "labels":{
            "label-key-1":[
               "label-val"
            ],
            "pkg-label":[
               "label-val"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "policyLevel":"sap:core:v1",
         "partOfProducts":[
            "sap:product:id{{ .randomSuffix }}:",
            "sap:product:SAPCloudPlatform:"
         ],
         "lineOfBusiness":[
            "Finance",
            "Sales"
         ],
         "industry":[
            "Automotive",
            "Banking",
            "Chemicals"
         ]
        
      },
      {
         "ordId":"ns:package:PACKAGE_ID_2{{ .randomSuffix }}:v2",
         "vendor":"sap:vendor:SAP:",
         "title":"PACKAGE 2 TITLE",
         "shortDescription":"Short description",
         "description":"lorem ipsum dolor set",
         "version":"2.1.2",
         "packageLinks":[
            {
               "type":"terms-of-service",
               "url":"https://example.com/en/legal/terms-of-use.html"
            },
            {
               "type":"client-registration",
               "url":"https://ui/public/showRegisterForm"
            }
         ],
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://testing/relative"
            }
         ],
         "licenseType":"AAL",
         "tags":[
            "testTag"
         ],
         "countries":[
            "BG",
            "US"
         ],
         "labels":{
            "label-key-1":[
               "label-val"
            ],
            "pkg-label":[
               "label-val"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "policyLevel":"sap:core:v1",
         "partOfProducts":[
            "sap:product:id{{ .randomSuffix }}:",
            "sap:product:SAPCloudPlatform:"
         ],
         "lineOfBusiness":[
            "Finance",
            "Sales"
         ],
         "industry":[
            "Automotive",
            "Banking",
            "Chemicals"
         ]
        
      }
   ],
   "entityTypes":[
      {
         "ordId":"ns:entityType:ENTITYTYPE_ID{{ .randomSuffix }}:v1",
         "localId":"BusinessPartner",
         "level":"aggregate",
         "title":"ENTITYTYPE 1 TITLE",
         "shortDescription":"short desc",
         "description":"lorem ipsum dolor set",
         "partOfPackage":"ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
         "visibility":"public",
         "version":"1.1.2",
         "releaseStatus":"active",
         "lastUpdate": "2023-01-26T15:47:04+00:00",
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://testing/relative"
            }
         ],
         "tags":[
            "testTag"
         ],
         "labels":{
            "label-key-1":[
               "label-val"
            ],
            "pkg-label":[
               "label-val"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "policyLevel":"sap:core:v1",
         "partOfProducts":[
            "sap:product:id{{ .randomSuffix }}:",
            "sap:product:SAPCloudPlatform:"
         ]
		 	 
      }
   ],
   "consumptionBundles":[
      {
         "title":"BUNDLE TITLE",
         "description":"Description for bundle",
         "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1",
         "shortDescription":"Short description bundle 1",
         "lastUpdate": "2023-01-26T15:47:04+00:00",
         "version": "1.0.0",
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://testing/relative"
            }
         ],
         "correlationIds":[
            "sap.s4:communicationScenario:SAP_COM_0001",
            "sap.s4:communicationScenario:SAP_COM_0002"
         ],
         "labels":{
            "label-key-1":[
               "label-value-1",
               "label-value-2"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "credentialExchangeStrategies":[
            {
               "callbackUrl":"https://credentials/relative",
               "customType":"ns:credential-exchange:v1",
               "customDescription":"custom description 1",
               "type":"custom"
            },
            {
               "callbackUrl":"https://example.com/credentials",
               "customType":"ns:credential-exchange2:v3",
               "customDescription":"custom description 2",
               "type":"custom"
            }
         ]
		 
      },
      {
         "title":"BUNDLE TITLE 2",
         "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v2",
         "lastUpdate": "2023-01-26T15:47:04+00:00",
         "version": "2.0.0",
         "credentialExchangeStrategies":[
            {
               "callbackUrl":"https://credentials/relative",
               "customType":"ns:credential-exchange:v1",
               "customDescription":"custom description 1",
               "type":"custom"
            },
            {
               "callbackUrl":"http://example.com/credentials",
               "customType":"ns:credential-exchange2:v3",
               "customDescription":"custom description 2",
               "type":"custom"
            }
         ],
         "correlationIds":[
            "sap.s4:communicationScenario:SAP_COM_0001",
            "sap.s4:communicationScenario:SAP_COM_0002"
         ],
         "labels":{
            "label-key-1":[
               "label-value-1",
               "label-value-2"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://testing/relative"
            }
         ]
		 
      }
   ],
   "products":[
      {
         "ordId":"sap:product:id{{ .randomSuffix }}:",
         "title":"PRODUCT TITLE",
         "description":"Description for product",
         "shortDescription":"Short description for product",
         "vendor":"sap:vendor:SAP:",
         "parent":"ns:product:id2:",
         "correlationIds":[
            "foo.bar.baz:foo:123456",
            "foo.bar.baz:bar:654321"
         ],
         "labels":{
            "label-key-1":[
               "label-value-1",
               "label-value-2"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         }
		 
      }
   ],
   "apiResources":[
      {
         "partOfPackage":"ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
         "title":"API TITLE",
         "description":"Description API 1",
         "entryPoints":[
            "https://exmaple.com/test/v1",
            "./example/test/v2"
         ],
         "ordId":"ns:apiResource:API_ID{{ .randomSuffix }}:v2",
         "shortDescription":"Short description for API",
         "systemInstanceAware":true,
         "apiProtocol":"odata-v2",
         "tags":[
            "apiTestTag"
         ],
         "countries":[
            "BG",
            "US"
         ],
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://testing/relative"
            }
         ],
         "apiResourceLinks":[
            {
               "type":"console",
               "url":"https://example.com/shell/discover"
            },
            {
               "type":"console",
               "url":"https://shell/discover/relative"
            }
         ],
         "releaseStatus":"active",
         "changelogEntries":[
            {
               "date":"2020-04-29",
               "description":"loremipsumdolorsitamet",
               "releaseStatus":"active",
               "url":"https://example.com/changelog/v1",
               "version":"1.0.0"
            }
         ],
         "labels":{
            "label-key-1":[
               "label-value-1",
               "label-value-2"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "visibility":"public",
         "disabled":true,
         "partOfProducts":[
            "sap:product:id{{ .randomSuffix }}:",
            "sap:product:SAPCloudPlatform:"
         ],
         "lineOfBusiness":[
            "Finance",
            "Sales"
         ],
         "industry":[
            "Automotive",
            "Banking",
            "Chemicals"
         ],
         "implementationStandard":"cff:open-service-broker:v2",
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "resourceDefinitions":[
            {
               "type":"openapi-v3",
               "mediaType":"application/json",
               "url":"/external-api/spec/flapping?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            },
            {
               "type":"openapi-v3",
               "mediaType":"text/yaml",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            },
            {
               "type":"edmx",
               "mediaType":"application/xml",
               "url":"/external-api/spec?format=yaml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            }
         ],
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1",
               "defaultEntryPoint":"https://exmaple.com/test/v1"
            },
            {
               "defaultEntryPoint":"https://exmaple.com/test/v1",
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v2"
            }
         ],
         "entityTypeMappings":[
            {
               "apiModelSelectors": [
                  {
                     "type": "odata",
                     "entitySetName": "A_OperationalAcctgDocItemCube"
                  }
               ],
               "entityTypeTargets": [
                  {
                     "ordId": "sap.odm:entityType:WorkforcePerson:v1"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:WorkForcePersonView_v1"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:sap.odm.JobDetails_v1"
                  }
               ]
            }
         ],
         "defaultConsumptionBundle":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1",
         "version":"2.1.2"
		 
      },
      {
         "partOfPackage":"ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
         "title":"Gateway Sample Service",
         "description":"Description API 1",
         "entryPoints":[
            "http://localhost:8080/some-api/v1"
         ],
         "ordId":"ns:apiResource:API_ID2{{ .randomSuffix }}:v1",
         "shortDescription":"Short description for API",
         "systemInstanceAware":true,
         "apiProtocol":"odata-v2",
         "tags":[
            "ZGWSAMPLE"
         ],
         "countries":[
            "BR"
         ],
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://testing/relative"
            }
         ],
         "apiResourceLinks":[
            {
               "type":"console",
               "url":"https://example.com/shell/discover"
            },
            {
               "type":"console",
               "url":"/shell/discover/relative"
            }
         ],
         "releaseStatus":"deprecated",
         "sunsetDate":"2020-12-08T15:47:04+0000",
         "deprecationDate": "2020-12-08T15:47:04+00:00",
         "successors":[
            "ns:apiResource:API_ID:v2"
         ],
         "changelogEntries":[
            {
               "date":"2020-04-29",
               "description":"loremipsumdolorsitamet",
               "releaseStatus":"active",
               "url":"https://example.com/changelog/v1",
               "version":"1.0.0"
            }
         ],
         "labels":{
            "label-key-1":[
               "label-value-1",
               "label-value-2"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "visibility":"public",
         "disabled":false,
         "partOfProducts":[
            "sap:product:id{{ .randomSuffix }}:",
            "sap:product:SAPCloudPlatform:"
         ],
         "lineOfBusiness":[
            "Finance",
            "Sales"
         ],
         "industry":[
            "Automotive",
            "Banking",
            "Chemicals"
         ],
         "implementationStandard":"cff:open-service-broker:v2",
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "resourceDefinitions":[
            {
               "type":"edmx",
               "mediaType":"application/xml",
               "url":"/external-api/spec?format=yaml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"	
                  }
               ]
            },
            {
               "type":"openapi-v3",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=yaml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            }
         ],
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1"
            },
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v2"
            }
         ],
         "version":"1.1.0"
		 
      },
      {
         "ordId":"ns:apiResource:API_ID{{ .randomSuffix }}:v3",
         "title":"API TITLE INTERNAL",
         "shortDescription":"Short description for API",
         "description":"Description for API internal",
         "entryPoints":[
            "https://exmaple.com/test/v1",
            "https://exmaple.com/test/v2"
         ],
         "version":"3.0.0",
         "visibility":"internal",
         "releaseStatus":"beta",
         "systemInstanceAware":true,
         "partOfPackage":"ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
         "apiProtocol":"odata-v2",
         "apiResourceLinks":[
            {
               "type":"console",
               "url":"https://example.com/discover"
            }
         ],
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1",
               "defaultEntryPoint":"https://exmaple.com/test/v1"
            }
         ],
         "entityTypeMappings":[
            {
               "apiModelSelectors": [
                  {
                     "type": "odata",
                     "entitySetName": "A_OperationalAcctgDocItemCube"
                  }
               ],
               "entityTypeTargets": [
                  {
                     "ordId": "sap.odm:entityType:WorkforcePerson:v1"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:WorkForcePersonView_v1"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:sap.odm.JobDetails_v1"
                  }
               ]
            },
            {
               "apiModelSelectors": [
                  {
                     "type": "odata",
                     "entitySetName": "B_OperationalAcctgDocItemCube"
                  }
               ],
               "entityTypeTargets": [
                  {
                     "ordId": "sap.odm:entityType:WorkforcePerson:v2"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:WorkForcePersonView_v2"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:sap.odm.JobDetails_v2"
                  }
               ]
            }
         ],
         "resourceDefinitions":[
            {
               "type":"edmx",
               "mediaType":"application/xml",
               "url":"/external-api/spec?format=xml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            },
            {
               "type":"openapi-v3",
               "mediaType":"text/yaml",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            }
         ]
		 
      },
      {
         "ordId":"ns:apiResource:API_ID{{ .randomSuffix }}:v4",
         "title":"API TITLE PRIVATE",
         "shortDescription":"Short description for API",
         "description":"Description for API private",
         "entryPoints":[
            "https://exmaple.com/test/v1",
            "https://exmaple.com/test/v2"
         ],
         "version":"4.0.0",
         "visibility":"private",
         "releaseStatus":"beta",
         "systemInstanceAware":true,
         "partOfPackage":"ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
         "apiProtocol":"odata-v2",
         "apiResourceLinks":[
            {
               "type":"console",
               "url":"https://example.com/discover"
            }
         ],
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "partOfConsumptionBundles":[
            {
               "defaultEntryPoint":"https://exmaple.com/test/v1",
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v2"
            }
         ],
         "resourceDefinitions":[
            {
               "type":"edmx",
               "mediaType":"application/xml",
               "url":"/external-api/spec?format=xml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            },
            {
               "type":"openapi-v3",
               "mediaType":"text/yaml",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            }
         ]
		 
      }
   ],
   "eventResources":[
      {
         "partOfPackage":"ns:package:PACKAGE_ID_2{{ .randomSuffix }}:v2",
         "title":"EVENT TITLE",
         "description":"Description Event 1",
         "ordId":"ns:eventResource:EVENT_ID{{ .randomSuffix }}:v2",
         "shortDescription":"Short description for Event",
         "systemInstanceAware":true,
         "lastUpdate": "2023-01-26T15:47:04+00:00",
         "changelogEntries":[
            {
               "date":"2020-04-29",
               "description":"loremipsumdolorsitamet",
               "releaseStatus":"active",
               "url":"https://example.com/changelog/v1",
               "version":"1.0.0"
            }
         ],
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://example.com/2018/04/11/testing/relative"
            }
         ],
         "tags":[
            "eventTestTag"
         ],
         "countries":[
            "BG",
            "US"
         ],
         "releaseStatus":"active",
         "labels":{
            "label-key-1":[
               "label-value-1",
               "label-value-2"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "visibility":"public",
         "disabled":true,
         "partOfProducts":[
            "sap:product:id{{ .randomSuffix }}:",
            "sap:product:SAPCloudPlatform:"
         ],
         "lineOfBusiness":[
            "Finance",
            "Sales"
         ],
         "industry":[
            "Automotive",
            "Banking",
            "Chemicals"
         ],
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "resourceDefinitions":[
            {
               "type":"asyncapi-v2",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=xml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            }
         ],
         "entityTypeMappings":[
            {
               "apiModelSelectors": [
                  {
                     "type": "json-pointer",
                     "jsonPointer": "#/components/messages/sap_odm_finance_costobject_CostCenter_Created_v1/payload"
                  }
               ],
               "entityTypeTargets": [
                  {
                     "ordId": "sap.odm:entityType:CostCenter:v1"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:CostCenter_v1"
                  }
               ]
            }
         ],         
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1"
            },
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v2"
            }
         ],
         "defaultConsumptionBundle":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1",
         "version":"2.1.2"
		 
      },
      {
         "partOfPackage":"ns:package:PACKAGE_ID_2{{ .randomSuffix }}:v2",
         "title":"EVENT TITLE 2",
         "description":"Description Event 2",
         "ordId":"ns2:eventResource:EVENT_ID{{ .randomSuffix }}:v1",
         "shortDescription":"Short description for Event",
         "systemInstanceAware":true,
         "lastUpdate": "2023-01-26T15:47:04+00:00",
         "changelogEntries":[
            {
               "date":"2020-04-29",
               "description":"loremipsumdolorsitamet",
               "releaseStatus":"active",
               "url":"https://example.com/changelog/v1",
               "version":"1.0.0"
            }
         ],
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"https://example.com/2018/04/11/testing/relative"
            }
         ],
         "tags":[
            "eventTestTag2"
         ],
         "countries":[
            "BR"
         ],
         "releaseStatus":"deprecated",
         "sunsetDate":"2020-12-08T15:47:04+0000",
         "deprecationDate": "2020-12-08T15:47:04+00:00",
         "successors":[
            "ns2:eventResource:EVENT_ID:v1"
         ],
         "labels":{
            "label-key-1":[
               "label-value-1",
               "label-value-2"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         },
         "visibility":"public",
         "disabled":false,
         "partOfProducts":[
            "sap:product:id{{ .randomSuffix }}:",
            "sap:product:SAPCloudPlatform:"
         ],
         "lineOfBusiness":[
            "Finance",
            "Sales"
         ],
         "industry":[
            "Automotive",
            "Banking",
            "Chemicals"
         ],
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "resourceDefinitions":[
            {
               "type":"asyncapi-v2",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            }
         ],
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1"
            },
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v2"
            }
         ],
         "version":"1.1.0"
		 
      },
      {
         "ordId":"ns3:eventResource:EVENT_ID{{ .randomSuffix }}:v1",
         "title":"EVENT TITLE INTERNAL",
         "shortDescription":"Short description for Event",
         "description":"Description for Event internal",
         "version":"1.1.0",
         "releaseStatus":"beta",
         "lastUpdate": "2023-01-26T15:47:04+00:00",
         "partOfPackage":"ns:package:PACKAGE_ID_2{{ .randomSuffix }}:v2",
         "visibility":"internal",
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v1"
            }
         ],
         "entityTypeMappings":[
            {
               "apiModelSelectors": [
                  {
                     "type": "json-pointer",
                     "jsonPointer": "#/components/messages/sap_odm_finance_costobject_CostCenter_Created_v1/payload"
                  }
               ],
               "entityTypeTargets": [
                  {
                     "ordId": "sap.odm:entityType:CostCenter:v2"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:CostCenter_v2"
                  }
               ]
            },
            {
               "apiModelSelectors": [
                  {
                     "type": "json-pointer",
                     "jsonPointer": "#/components/messages/sap_odm_finance_costobject_CostCenter_Created_v2/payload"
                  }
               ],
               "entityTypeTargets": [
                  {
                     "ordId": "sap.odm:entityType:CostCenter:v2"
                  },
                  {
                     "correlationId": "sap.s4:csnEntity:CostCenter_v2"
                  }
               ]
            }

         ],         
         "resourceDefinitions":[
            {
               "type":"asyncapi-v2",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            }
         ]
		 
      },
      {
         "ordId":"ns4:eventResource:EVENT_ID{{ .randomSuffix }}:v1",
         "title":"EVENT TITLE PRIVATE",
         "shortDescription":"Short description for Event",
         "description":"Description for Event private",
         "version":"1.1.0",
         "releaseStatus":"beta",
         "lastUpdate": "2023-01-26T15:47:04+00:00",
         "partOfPackage":"ns:package:PACKAGE_ID_2{{ .randomSuffix }}:v2",
         "visibility":"internal",
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID{{ .randomSuffix }}:v2"
            }
         ],
         "resourceDefinitions":[
            {
               "type":"asyncapi-v2",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}"
                  }
               ]
            }
         ]
		 
      }
   ],
	"capabilities":[
    {
      "ordId": "sap.s4:capability:{{ .randomSuffix }}:v1",
      "title": "CAPABILITY TITLE",
      "type": "sap.mdo:mdi-capability:v1",
      "shortDescription": "Short description of capability",
      "description": "Optional, longer description",
      "version": "1.0.0",
      "lastUpdate": "2023-01-26T15:47:04+00:00",
      "releaseStatus": "active",
      "visibility": "public",
      "partOfPackage": "ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
      "definitions": [
        {
          "type": "sap.mdo:mdi-capability-definition:v1",
          "mediaType": "application/json",
          "url": "/external-api/spec?format=json", 
          "accessStrategies": [
            {
                "type":"{{ .specsAccessStrategy }}"
			}
          ]
        }
      ]
    }
  ],
	"integrationDependencies": [
    {
      "ordId": "ns1:integrationDependency:INTEGRATION_DEPENDENCY_ID{{ .randomSuffix }}:v2",
      "version": "2.2.3",
      "title": "INTEGRATION DEPENDENCY TITLE",
      "shortDescription": "Short description of an integration dependency",
      "description": "longer description of an integration dependency",
      "partOfPackage": "ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
      "correlationIds": [
		 "sap.s4:communicationScenario:SAP_COM_123"
      ],
      "lastUpdate": "2023-08-03T10:14:26.941Z",
      "visibility": "public",
      "releaseStatus": "active",
	  "mandatory": true,
      "aspects": [
        {
          "title": "ASPECT TITLE",
		  "description": "Aspect desc",
          "mandatory": true,
          "eventResources": [
            {
              "ordId": "ns1:eventResource:ASPECT_EVENT_RESOURCE_ID{{ .randomSuffix }}:v1",
              "subset": [
                {
                  "eventType": "sap.billing.sb.Subscription.Created.v1"
                },
                {
                  "eventType": "sap.billing.sb.Subscription.Updated.v1"
                },
                {
                  "eventType": "sap.billing.sb.Subscription.Deleted.v1"
                }
              ]
            }
          ],
		  "apiResources": [
            {
              "ordId": "ns:apiResource:API_ID{{ .randomSuffix }}:v2",
              "minVersion": "2.3.0"
            }
          ]
        }
      ]
    },
	{
      "ordId": "ns2:integrationDependency:INTEGRATION_DEPENDENCY_ID{{ .randomSuffix }}:v2",
      "version": "2.2.3",
      "title": "INTEGRATION DEPENDENCY TITLE PRIVATE",
      "shortDescription": "Short description of a private integration dependency",
      "description": "longer description of a private integration dependency",
      "partOfPackage": "ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
      "correlationIds": [
		 "sap.s4:communicationScenario:SAP_COM_123"
      ],
      "lastUpdate": "2023-08-03T10:14:26.941Z",
      "visibility": "private",
      "releaseStatus": "active",
	  "mandatory": true,
      "aspects": [
        {
          "title": "ASPECT TITLE PRIVATE",
		  "description": "Aspect private desc",
          "mandatory": true,
          "eventResources": [
            {
              "ordId": "ns2:eventResource:ASPECT_EVENT_RESOURCE_ID{{ .randomSuffix }}:v1",
              "subset": [
                {
                  "eventType": "sap.billing.sb.Subscription.Created.v1"
                },
                {
                  "eventType": "sap.billing.sb.Subscription.Updated.v1"
                },
                {
                  "eventType": "sap.billing.sb.Subscription.Deleted.v1"
                }
              ]
            }
          ],
		  "apiResources": [
            {
              "ordId": "ns:apiResource:API_ID{{ .randomSuffix }}:v2",
              "minVersion": "2.3.0"
            }
          ]
        }
      ]
    },
	{
      "ordId": "ns3:integrationDependency:INTEGRATION_DEPENDENCY_ID{{ .randomSuffix }}:v2",
      "version": "2.2.3",
      "title": "INTEGRATION DEPENDENCY TITLE INTERNAL",
      "shortDescription": "Short description of an internal integration dependency",
      "description": "longer description of an internal integration dependency",
      "partOfPackage": "ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
      "correlationIds": [
		 "sap.s4:communicationScenario:SAP_COM_123"
      ],
      "lastUpdate": "2023-08-03T10:14:26.941Z",
      "visibility": "internal",
      "releaseStatus": "active",
	  "mandatory": true,
      "aspects": [
        {
          "title": "ASPECT TITLE INTERNAL",
		  "description": "Aspect internal desc",
          "mandatory": true,
          "eventResources": [
            {
              "ordId": "ns3:eventResource:ASPECT_EVENT_RESOURCE_ID{{ .randomSuffix }}:v1",
              "subset": [
                {
                  "eventType": "sap.billing.sb.Subscription.Created.v1"
                },
                {
                  "eventType": "sap.billing.sb.Subscription.Updated.v1"
                },
                {
                  "eventType": "sap.billing.sb.Subscription.Deleted.v1"
                }
              ]
            }
          ],
		  "apiResources": [
            {
              "ordId": "ns:apiResource:API_ID{{ .randomSuffix }}:v2",
              "minVersion": "2.3.0"
            }
          ]
        }
      ]
    }
  ],
   "dataProducts": [
      {
      "ordId": "ns:dataProduct:DATA_PRODUCT_ID{{ .randomSuffix }}:v1",
      "localId": "Customer",
      "correlationIds": [
        "sap.xref:foo:bar"
      ],
      "title": "DATA PRODUCT TITLE",
      "shortDescription": "Short description of Data Product",
      "description": "Long description for a public Data Product resource",
      "partOfPackage": "ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
      "version": "1.2.3",
      "visibility": "public",
      "releaseStatus": "deprecated",
      "disabled": false,
      "lastUpdate": "2020-12-08T15:47:04+00:00",
      "deprecationDate": "2020-12-08T15:47:04+00:00",
      "sunsetDate": "2022-01-08T15:47:04+00:00",
      "successors": [
        "sap.xref:dataProduct:Customer:v2"
      ],
      "type": "base",
      "category": "business-object",
      "entityTypes": ["ns:entityType:ENTITYTYPE_ID{{ .randomSuffix }}:v1"],
      "inputPorts": [
        {
          "ordId": "ns1:integrationDependency:INTEGRATION_DEPENDENCY_ID{{ .randomSuffix }}:v2"
        }
      ],
      "outputPorts": [
        {
          "ordId": "ns:apiResource:API_ID{{ .randomSuffix }}:v2"
        }
      ],
      "responsible": "sap:ach:CIC-DP-CO",
      "dataProductLinks": [
		{
  			"type": "support",
  			"url": "https://support.sap.com/CIC_DP_RT/issue/"
		}
	  ],
      "links": [
		{
		   "description":"loremipsumdolornem",
		   "title":"LinkTitle1",
		   "url":"https://example.com/2018/04/11/testing/"
		},
		{
		   "description":"loremipsumdolornem",
		   "title":"LinkTitle2",
		   "url":"https://example.com/2018/04/11/testing/relative"
		}
      ],
      "industry": [
		"Automotive",
		"Banking",
		"Chemicals"
	  ],
      "lineOfBusiness": [
		"Finance",
		"Sales"
	  ],
      "tags": [
		"testTag"
	  ],
      "labels": {
		"label-key-1": [
		   "label-val"
		],
		"pkg-label": [
		   "label-val"
		]
	 },
	 "documentationLabels": {
		"Documentation label key": [
		   "Markdown Documentation with links",
		   "With multiple values"
		]
	 },
     "policyLevel": "sap:core:v1",
     "systemInstanceAware": true
    },
    {
      "ordId": "ns:dataProduct:DATA_PRODUCT_ID_2{{ .randomSuffix }}:v2",
      "localId": "Customer",
      "correlationIds": [
        "sap.xref:foo:bar"
      ],
      "title": "DATA PRODUCT TITLE PRIVATE",
      "shortDescription": "Short description of Data Product",
      "description": "Long description for a private Data Product resource",
      "partOfPackage": "ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
      "visibility": "private",
      "releaseStatus": "active",
      "disabled": false,
      "version": "2.1.0",
      "lastUpdate": "2022-12-19T15:47:04+00:00",
      "type": "base",
      "category": "business-object",
      "entityTypes": ["ns:entityType:ENTITYTYPE_ID{{ .randomSuffix }}:v1"],
      "inputPorts": [
        {
          "ordId": "ns1:integrationDependency:INTEGRATION_DEPENDENCY_ID{{ .randomSuffix }}:v2"
        }
      ],
      "outputPorts": [
        {
          "ordId": "ns:apiResource:API_ID{{ .randomSuffix }}:v2"
        }
      ],
      "responsible": "sap:ach:CIC-DP-CO",
      "dataProductLinks": [
		{
  			"type": "support",
  			"url": "https://support.sap.com/CIC_DP_RT/issue/"
		}
	  ],
      "links": [
		{
		   "description": "loremipsumdolornem",
		   "title": "LinkTitle1",
		   "url": "https://example.com/2018/04/11/testing/"
		},
		{
		   "description": "loremipsumdolornem",
		   "title": "LinkTitle2",
		   "url": "https://example.com/2018/04/11/testing/relative"
		}
      ],
      "industry": [
		"Automotive",
		"Banking",
		"Chemicals"
	  ],
      "lineOfBusiness": [
		"Finance",
		"Sales"
	  ],
      "tags": [
		"testTag"
	  ],
      "labels": {
		"label-key-1": [
		   "label-val"
		],
		"pkg-label": [
		   "label-val"
		]
	 },
	 "documentationLabels": {
		"Documentation label key": [
		   "Markdown Documentation with links",
		   "With multiple values"
		]
	 },
     "policyLevel": "sap:core:v1",
     "systemInstanceAware": true
    },
    {
      "ordId": "ns:dataProduct:DATA_PRODUCT_ID_3{{ .randomSuffix }}:v3",
      "localId": "Customer",
      "correlationIds": [
        "sap.xref:foo:bar"
      ],
      "title": "DATA PRODUCT TITLE INTERNAL",
      "shortDescription": "Short description of Data Product",
      "description": "Long description for an internal Data Product resource",
      "partOfPackage": "ns:package:PACKAGE_ID{{ .randomSuffix }}:v1",
      "visibility": "internal",
      "releaseStatus": "active",
      "version": "3.1.0",
      "lastUpdate": "2022-12-19T15:47:04+00:00",
      "type": "base",
      "category": "business-object",
      "entityTypes": ["ns:entityType:ENTITYTYPE_ID{{ .randomSuffix }}:v1"],
      "inputPorts": [
        {
          "ordId": "ns1:integrationDependency:INTEGRATION_DEPENDENCY_ID{{ .randomSuffix }}:v2"
        }
      ],
      "outputPorts": [
        {
          "ordId": "ns:apiResource:API_ID{{ .randomSuffix }}:v2"
        }
      ],
      "responsible": "sap:ach:CIC-DP-CO",
      "dataProductLinks": [
		{
  			"type": "support",
  			"url": "https://support.sap.com/CIC_DP_RT/issue/"
		}
	  ],
      "links": [
		{
		   "description": "loremipsumdolornem",
		   "title": "LinkTitle1",
		   "url": "https://example.com/2018/04/11/testing/"
		},
		{
		   "description": "loremipsumdolornem",
		   "title": "LinkTitle2",
		   "url": "https://example.com/2018/04/11/testing/relative"
		}
      ],
      "industry": [
		"Automotive",
		"Banking",
		"Chemicals"
	  ],
      "lineOfBusiness": [
		"Finance",
		"Sales"
	  ],
      "tags": [
		"testTag"
	  ],
      "labels": {
		"label-key-1": [
		   "label-val"
		],
		"pkg-label": [
		   "label-val"
		]
	 },
	 "documentationLabels": {
		"Documentation label key": [
		   "Markdown Documentation with links",
		   "With multiple values"
		]
	 },
     "policyLevel": "sap:core:v1",
     "systemInstanceAware": true
    }
  ],
   "tombstones":[
      {
         "ordId":"ns:apiResource:API_ID2{{ .randomSuffix }}:v1",
         "removalDate":"2020-12-02T14:12:59Z"
		 
      }
   ],
   "vendors":[
      {
         "ordId":"partner:vendor:SAP{{ .randomSuffix }}:",
         "title":"SAP SE",
         "partners":[
            "microsoft:vendor:Microsoft:"
         ],
         "labels":{
            "label-key-1":[
               "label-value-1",
               "label-value-2"
            ]
         },
         "documentationLabels":{
            "Documentation label key":[
               "Markdown Documentation with links",
               "With multiple values"
            ]
         }
		 
      }
   ]
}`
