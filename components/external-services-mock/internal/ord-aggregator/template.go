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
   "description":"Test Document",
   "describedSystemInstance":{
      "ProviderName":null,
      "Tenant":"",
      "Name":"",
      "Description":null,
      "Status":null,
      "HealthCheckURL":null,
      "IntegrationSystemID":null,
      "ApplicationTemplateID":null,
      "baseUrl":"{{ .baseURL }}",
      "labels":{
         "label-key-1":[
            "label-value-1",
            "label-value-2"
         ]
      }
	  {{ .additionalProperties }}
   },
   "packages":[
      {
         "ordId":"ns:package:PACKAGE_ID:v1",
         "vendor":"sap:vendor:SAP:",
         "title":"PACKAGE 1 TITLE",
         "shortDescription":"short desc",
         "description":"lorem ipsum dolor set",
         "version":"1.1.2",
         "packageLinks":[
            {
               "type":"terms-of-service",
               "url":"https://example.com/en/legal/terms-of-use.html"
            },
            {
               "type":"client-registration",
               "url":"/ui/public/showRegisterForm"
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
               "url":"/testing/relative"
            }
         ],
         "licenseType":"licence",
         "tags":[
            "testTag"
         ],
         "countries":[
            "BG",
            "EN"
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
         "customPolicyLevel":null,
         "partOfProducts":[
            "sap:product:id:",
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
		 {{ .additionalProperties }}	 
      }
   ],
   "entityTypes":[
      {
         "ordId":"ns:entityType:ENTITYTYPE_ID:v1",
         "localId":"BusinessPartner",
         "level":"aggregate",
         "title":"ENTITYTYPE 1 TITLE",
         "shortDescription":"short desc",
         "description":"lorem ipsum dolor set",
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "visibility":"public",
         "version":"1.1.2",
         "releaseStatus":"active",
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"/testing/relative"
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
         "customPolicyLevel":null,
         "partOfProducts":[
            "sap:product:id:",
            "sap:product:SAPCloudPlatform:"
         ]
		 {{ .additionalProperties }}	 
      }
   ],
   "consumptionBundles":[
      {
         "title":"BUNDLE TITLE",
         "description":"lorem ipsum dolor nsq sme",
         "ordId":"ns:consumptionBundle:BUNDLE_ID:v1",
         "shortDescription":"lorem ipsum",
         "links":[
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle1",
               "url":"https://example.com/2018/04/11/testing/"
            },
            {
               "description":"loremipsumdolornem",
               "title":"LinkTitle2",
               "url":"/testing/relative"
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
		 {{ .additionalProperties }}
      },
      {
         "title":"BUNDLE TITLE 2",
         "ordId":"ns:consumptionBundle:BUNDLE_ID:v2",
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
               "url":"/testing/relative"
            }
         ]
		 {{ .additionalProperties }}
      }
   ],
   "products":[
      {
         "ordId":"sap:product:id:",
         "title":"PRODUCT TITLE",
         "description":"lorem ipsum dolor nsq sme",
         "shortDescription":"lorem ipsum",
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
		 {{ .additionalProperties }}
      }
   ],
   "apiResources":[
      {
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "title":"API TITLE",
         "description":"lorem ipsum dolor sit amet",
         "entryPoints":[
            "https://exmaple.com/test/v1",
            "https://exmaple.com/test/v2"
         ],
         "ordId":"ns:apiResource:API_ID:v2",
         "shortDescription":"lorem ipsum",
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
               "url":"/testing/relative"
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
         "releaseStatus":"active",
         "sunsetDate":null,
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
            "sap:product:id:",
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
         "customImplementationStandard":null,
         "customImplementationStandardDescription":null,
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "resourceDefinitions":[
            {
               "type":"openapi-v3",
               "customType":"",
               "mediaType":"application/json",
               "url":"/external-api/spec/flapping?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            },
            {
               "type":"openapi-v3",
               "customType":"",
               "mediaType":"text/yaml",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            },
            {
               "type":"edmx",
               "customType":"",
               "mediaType":"application/xml",
               "url":"/external-api/spec?format=yaml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            }
         ],
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v1",
               "defaultEntryPoint":"https://exmaple.com/test/v1"
            },
            {
               "defaultEntryPoint":"https://exmaple.com/test/v1",
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v2"
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
         "defaultConsumptionBundle":"ns:consumptionBundle:BUNDLE_ID:v1",
         "version":"2.1.2"
		 {{ .additionalProperties }}
      },
      {
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "title":"Gateway Sample Service",
         "description":"lorem ipsum dolor sit amet",
         "entryPoints":[
            "http://localhost:8080/some-api/v1"
         ],
         "ordId":"ns:apiResource:API_ID2:v1",
         "shortDescription":"lorem ipsum",
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
               "url":"/testing/relative"
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
         "disabled":null,
         "partOfProducts":[
            "sap:product:id:",
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
         "customImplementationStandard":null,
         "customImplementationStandardDescription":null,
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "resourceDefinitions":[
            {
               "type":"edmx",
               "customType":"",
               "mediaType":"application/xml",
               "url":"/external-api/spec?format=yaml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            },
            {
               "type":"openapi-v3",
               "customType":"",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=yaml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            }
         ],
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v1"
            },
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v2"
            }
         ],
         "version":"1.1.0"
		 {{ .additionalProperties }}
      },
      {
         "ordId":"ns:apiResource:API_ID:v3",
         "title":"API TITLE INTERNAL",
         "shortDescription":"Test",
         "description":"Test description internal",
         "entryPoints":[
            "https://exmaple.com/test/v1",
            "https://exmaple.com/test/v2"
         ],
         "version":"1.0.0",
         "visibility":"internal",
         "releaseStatus":"beta",
         "systemInstanceAware":true,
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "apiProtocol":"rest",
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
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v1",
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
               "customType":"",
               "mediaType":"application/xml",
               "url":"/external-api/spec?format=xml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            },
            {
               "type":"openapi-v3",
               "customType":"",
               "mediaType":"text/yaml",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            }
         ]
		 {{ .additionalProperties }}
      },
      {
         "ordId":"ns:apiResource:API_ID:v4",
         "title":"API TITLE PRIVATE",
         "shortDescription":"Test",
         "description":"Test description private",
         "entryPoints":[
            "https://exmaple.com/test/v1",
            "https://exmaple.com/test/v2"
         ],
         "version":"1.0.0",
         "visibility":"private",
         "releaseStatus":"beta",
         "systemInstanceAware":true,
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "apiProtocol":"rest",
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
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v2"
            }
         ],
         "resourceDefinitions":[
            {
               "type":"edmx",
               "customType":"",
               "mediaType":"application/xml",
               "url":"/external-api/spec?format=xml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            },
            {
               "type":"openapi-v3",
               "customType":"",
               "mediaType":"text/yaml",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            }
         ]
		 {{ .additionalProperties }}
      }
   ],
   "eventResources":[
      {
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "title":"EVENT TITLE",
         "description":"lorem ipsum dolor sit amet",
         "ordId":"ns:eventResource:EVENT_ID:v1",
         "shortDescription":"lorem ipsum",
         "systemInstanceAware":true,
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
               "url":"/testing/relative"
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
         "sunsetDate":null,
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
            "sap:product:id:",
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
               "customType":"",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=xml",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
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
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v1"
            },
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v2"
            }
         ],
         "defaultConsumptionBundle":"ns:consumptionBundle:BUNDLE_ID:v1",
         "version":"2.1.2"
		 {{ .additionalProperties }}
      },
      {
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "title":"EVENT TITLE 2",
         "description":"lorem ipsum dolor sit amet",
         "ordId":"ns2:eventResource:EVENT_ID:v1",
         "shortDescription":"lorem ipsum",
         "systemInstanceAware":true,
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
               "url":"/testing/relative"
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
         "disabled":null,
         "partOfProducts":[
            "sap:product:id:",
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
               "customType":"",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            }
         ],
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v1"
            },
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v2"
            }
         ],
         "version":"1.1.0"
		 {{ .additionalProperties }}
      },
      {
         "ordId":"ns3:eventResource:EVENT_ID:v1",
         "title":"EVENT TITLE INTERNAL",
         "shortDescription":"Test",
         "description":"Test description internal",
         "version":"0.1.0",
         "releaseStatus":"beta",
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "visibility":"internal",
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v1"
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
               "customType":"",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            }
         ]
		 {{ .additionalProperties }}
      },
      {
         "ordId":"ns4:eventResource:EVENT_ID:v1",
         "title":"EVENT TITLE PRIVATE",
         "shortDescription":"Test",
         "description":"Test description private",
         "version":"0.1.0",
         "releaseStatus":"beta",
         "partOfPackage":"ns:package:PACKAGE_ID:v1",
         "visibility":"internal",
         "extensible":{
            "supported":"automatic",
            "description":"Please find the extensibility documentation"
         },
         "partOfConsumptionBundles":[
            {
               "ordId":"ns:consumptionBundle:BUNDLE_ID:v2"
            }
         ],
         "resourceDefinitions":[
            {
               "type":"asyncapi-v2",
               "customType":"",
               "mediaType":"application/json",
               "url":"/external-api/spec?format=json",
               "accessStrategies":[
                  {
                     "type":"{{ .specsAccessStrategy }}",
                     "customType":"",
                     "customDescription":""
                  }
               ]
            }
         ]
		 {{ .additionalProperties }}
      }
   ],
	"capabilities":[
    {
      "ordId": "sap.foo.bar:capability::v1",
      "title": "CAPABILITY TITLE",
      "type": "sap.mdo:mdi-capability:v1",
      "shortDescription": "Short description of capability",
      "description": "Optional, longer description",
      "version": "1.0.0",
      "lastUpdate": "2023-01-26T15:47:04+00:00",
      "releaseStatus": "active",
      "visibility": "public",
      "partOfPackage": "ns:package:PACKAGE_ID:v1",
      "definitions": [
        {
          "type": "sap.mdo:mdi-capability-definition:v1",
          "mediaType": "application/json",
          "url": "/external-api/spec?format=json", 
          "accessStrategies": [
            {
                "type":"{{ .specsAccessStrategy }}",
                "customType":"",
				"customDescription":""
			}
          ]
        }
      ]
    }
  ],
	"integrationDependencies": [
    {
      "ordId": "ns1:integrationDependency:INTEGRATION_DEPENDENCY_ID:v2",
      "version": "2.2.3",
      "title": "INTEGRATION DEPENDENCY TITLE",
      "shortDescription": "Short description of an integration dependency",
      "description": "longer description of an integration dependency",
      "partOfPackage": "ns:package:PACKAGE_ID:v1",
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
              "ordId": "ns1:eventResource:ASPECT_EVENT_RESOURCE_ID:v1",
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
              "ordId": "ns:apiResource:API_ID:v2",
              "minVersion": "2.3.0"
            }
          ]
        }
      ]
    },
	{
      "ordId": "ns2:integrationDependency:INTEGRATION_DEPENDENCY_ID:v2",
      "version": "2.2.3",
      "title": "INTEGRATION DEPENDENCY TITLE PRIVATE",
      "shortDescription": "Short description of a private integration dependency",
      "description": "longer description of a private integration dependency",
      "partOfPackage": "ns:package:PACKAGE_ID:v1",
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
              "ordId": "ns2:eventResource:ASPECT_EVENT_RESOURCE_ID:v1",
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
              "ordId": "ns:apiResource:API_ID:v2",
              "minVersion": "2.3.0"
            }
          ]
        }
      ]
    },
	{
      "ordId": "ns3:integrationDependency:INTEGRATION_DEPENDENCY_ID:v2",
      "version": "2.2.3",
      "title": "INTEGRATION DEPENDENCY TITLE INTERNAL",
      "shortDescription": "Short description of an internal integration dependency",
      "description": "longer description of an internal integration dependency",
      "partOfPackage": "ns:package:PACKAGE_ID:v1",
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
              "ordId": "ns3:eventResource:ASPECT_EVENT_RESOURCE_ID:v1",
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
              "ordId": "ns:apiResource:API_ID:v2",
              "minVersion": "2.3.0"
            }
          ]
        }
      ]
    }
  ],
   "dataProducts": [
      {
      "ordId": "ns:dataProduct:DATA_PRODUCT_ID:v1",
      "localId": "Customer",
      "correlationIds": [
        "sap.xref:foo:bar"
      ],
      "title": "DATA PRODUCT TITLE",
      "shortDescription": "Short description of Data Product",
      "description": "Long description for a public Data Product resource",
      "partOfPackage": "ns:package:PACKAGE_ID:v1",
      "visibility": "public",
      "releaseStatus": "deprecated",
      "version": "1.9.2",
      "lastUpdate": "2020-12-08T15:47:04+00:00",
      "deprecationDate": "2020-12-08T15:47:04+00:00",
      "sunsetDate": "2022-01-08T15:47:04+00:00",
      "successors": [
        "sap.xref:dataProduct:Customer:v2"
      ],
      "type": "base",
      "category": "business-object",
      "outputPorts": [
        {
          "ordId": "sap.xref:apiResource:Customer:v1"
        }
      ],
      "responsible": "sap:ach:CIC-DP-CO"
    },
    {
      "ordId": "ns:dataProduct:DATA_PRODUCT_ID_2:v2",
      "localId": "Customer",
      "correlationIds": [
        "sap.xref:foo:bar"
      ],
      "title": "DATA PRODUCT TITLE PRIVATE",
      "shortDescription": "Short description of Data Product",
      "description": "Long description for a private Data Product resource",
      "partOfPackage": "ns:package:PACKAGE_ID:v1",
      "visibility": "private",
      "releaseStatus": "active",
      "version": "2.1.0",
      "lastUpdate": "2022-12-19T15:47:04+00:00",
      "type": "base",
      "category": "business-object",
      "outputPorts": [
        {
          "ordId": "sap.xref:apiResource:Customer:v1"
        }
      ],
      "responsible": "sap:ach:CIC-DP-CO"
    },
    {
      "ordId": "ns:dataProduct:DATA_PRODUCT_ID_3:v3",
      "localId": "Customer",
      "correlationIds": [
        "sap.xref:foo:bar"
      ],
      "title": "DATA PRODUCT TITLE INTERNAL",
      "shortDescription": "Short description of Data Product",
      "description": "Long description for an internal Data Product resource",
      "partOfPackage": "ns:package:PACKAGE_ID:v1",
      "visibility": "internal",
      "releaseStatus": "active",
      "version": "2.1.0",
      "lastUpdate": "2022-12-19T15:47:04+00:00",
      "type": "base",
      "category": "business-object",
      "outputPorts": [
        {
          "ordId": "sap.xref:apiResource:Customer:v1"
        }
      ],
      "responsible": "sap:ach:CIC-DP-CO"
    }
  ],
   "tombstones":[
      {
         "ordId":"ns:apiResource:API_ID2:v1",
         "removalDate":"2020-12-02T14:12:59Z"
		 {{ .additionalProperties }}
      }
   ],
   "vendors":[
      {
         "ordId":"partner:vendor:SAP:",
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
		 {{ .additionalProperties }}
      }
   ]
   {{ .additionalEntities }}
}`
