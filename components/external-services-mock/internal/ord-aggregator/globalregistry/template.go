package ord_global_registry

const ordConfigTemplate = `{
    "$schema": "../spec/v1/generated/Configuration.schema.json",
    "baseUrl": "%s",
	"openResourceDiscoveryV1": {
        "documents": [
            {
                "url": "/open-resource-discovery/v1/documents/example1",
                "systemInstanceAware": true,
                "accessStrategies": [
                    {
                        "type": "sap:cmp-mtls:v1",
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
   "description":"Global Registry Test Document",
   "products": [
    {
      "ordId": "sap:product:SAPCustomerExperience:",
      "title": "SAP Customer Experience",
      "shortDescription": "Bring together customer data, machine learning technology, and microservices to power real-time customer engagements across sales, service, marketing, and commerce.",
      "vendor": "sap:vendor:SAP:"
    },
    {
      "ordId": "sap:product:SAPServiceCloudV2:",
      "title": "SAP Service Cloud Version 2",
      "shortDescription": "Enables you to run service processes efficiently with service agents having customer information at their fingertips.",
      "vendor": "sap:vendor:SAP:",
      "parent": "sap:product:SAPCustomerExperience:"
    },
    {
      "ordId": "sap:product:SAPGraph:",
      "title": "SAP Graph",
      "shortDescription": "The easy-to-use API for the data of the Intelligent Enterprise from SAP.",
      "vendor": "sap:vendor:SAP:"
    },
    {
      "ordId": "sap:product:SAPS4HANACloud:",
      "title": "SAP S/4HANA Cloud",
      "shortDescription": "The next generation digital core designed to help you run simple in a digital economy. It provides the industry-specific capabilities and cloud benefits that your business needs.",
      "vendor": "sap:vendor:SAP:",
      "labels": {
        "logo": [
          "https://cloudintegration.hana.ondemand.com/falcon-assets/logos/products/SAPS4HANACloud_MINI.svg"
        ]
      }
    },
    {
      "ordId": "sap:product:SAPS4HANA:",
      "title": "SAP S/4HANA",
      "shortDescription": "A future-ready ERP system with built-in intelligent technologies, including AI, machine learning, and advanced analytics.",
      "vendor": "sap:vendor:SAP:",
      "labels": {
        "logo": [
          "https://cloudintegration.hana.ondemand.com/falcon-assets/logos/products/SAPS4HANA_MINI.svg"
        ]
      }
    },
    {
      "ordId": "sap:product:SAPCloudPlatform:",
      "title": "SAP Business Technology Platform",
      "shortDescription": "Accelerate business outcomes with integration, data to value, and extensibility.",
      "vendor": "sap:vendor:SAP:"
    },
    {
      "ordId": "sap:product:SAPS4HANAUtilities:",
      "title": "SAP S/4HANA Utilities",
      "shortDescription": "Provides an intelligent and integrated ERP system for utilities that runs on our in-memory database, SAP HANA.",
      "vendor": "sap:vendor:SAP:",
      "parent": "sap:product:SAPS4HANA:"
    },
	{
		"ordId": "sap:product:SAPSuccessFactors:",
		"title": "SAP SuccessFactors",
		"shortDescription": "World-leading provider of cloud human experience management (HXM).",
		"vendor": "sap:vendor:SAP:"
	},
    {
      "ordId": "sap:product:SAPTransactionalBankingforSAPS4HANA:",
      "title": "SAP Transactional Banking for SAP S/4HANA",
      "shortDescription": "Open core banking platform, which is based on an architecture that ensures real-time processing and continuous availability.",
      "vendor": "sap:vendor:SAP:",
      "parent": "sap:product:SAPS4HANA:"
    },
    {
      "ordId": "sap:product:SAPLandscapeManagementCloud:",
      "title": "SAP Landscape Management Cloud",
      "shortDescription": "Manage operations and scalability for SAP solutions on cloud infrastructure providers with SAP Landscape Management Cloud.",
      "vendor": "sap:vendor:SAP:"
    },
    {
      "ordId": "sap:product:SAPS4HANAPrivateCloud:",
      "title": "SAP S/4HANA Cloud Private Edition",
      "shortDescription": "The next generation digital core designed to help you run simple in a digital economy. It provides the industry-specific capabilities and cloud benefits that your business needs.",
      "vendor": "sap:vendor:SAP:"
    },
	{
		"ordId": "sap:product:SAPMasterDataIntegration:",
		"title": "SAP Master Data Integration",
		"shortDescription": "SAP Master Data Integration provide and manage generic replication to keep consistent view on master data.",
		"vendor": "sap:vendor:SAP:"
	}
  ],
   "vendors":[
		{
		  "ordId": "sap:vendor:SAP:",
		  "title": "SAP SE",
		  "partners": []
		},
		{
		  "ordId": "customer:vendor:SAPCustomer:",
		  "title": "SAP Customer",
		  "partners": []
		}
   ]
}`
