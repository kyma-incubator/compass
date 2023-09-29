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
      "description": "Bring together customer data, machine learning technology, and microservices to power real-time customer engagements across sales, service, marketing, and commerce.",
      "shortDescription": "Optimize customer interactions in many aspects in real time.",
      "vendor": "sap:vendor:SAP:"
    },
    {
      "ordId": "sap:product:SAPServiceCloudV2:",
      "title": "SAP Service Cloud Version 2",
      "description": "Enables you to run service processes efficiently with service agents having customer information at their fingertips.",
	  "shortDescription": "Boosts efficiency with immediate customer data access for agents.",
      "vendor": "sap:vendor:SAP:",
      "parent": "sap:product:SAPCustomerExperience:"
    },
    {
      "ordId": "sap:product:SAPGraph:",
      "title": "SAP Graph",
      "description": "SAP Graph is the easy-to-use API for the data of the Intelligent Enterprise from SAP.\nIt provides an intuitive programming model that you can use to easily build new extensions and applications using SAP data.",
      "shortDescription": "SAP Graph: Streamlined API for Intelligent Enterprise data, enabling intuitive extension and app development with SAP data.",
      "vendor": "sap:vendor:SAP:"
    },
    {
      "ordId": "sap:product:SAPS4HANACloud:",
      "title": "SAP S/4HANA Cloud",
      "description": "The next generation digital core designed to help you run simple\nin a digital economy. It provides the industry-specific capabilities and cloud\nbenefits that your business needs.",
	  "shortDescription": "Cutting-edge digital core for streamlined operations and industry-specific capabilities in the digital era.",
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
      "description": "A future-ready ERP system with built-in intelligent technologies,\nincluding AI, machine learning, and advanced analytics which transforms business\nprocesses with intelligent automation.",
	  "shortDescription": "Modern ERP with integrated AI, ML, and analytics for intelligent business process automation.",
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
      "description": "Accelerate business outcomes with integration, data to value, and extensibility.",
      "shortDescription": "Optimize business impact with integration and data value.",
      "vendor": "sap:vendor:SAP:"
    },
    {
      "ordId": "sap:product:SAPS4HANAUtilities:",
      "title": "SAP S/4HANA Utilities",
      "description": "Provides an intelligent and integrated ERP system for utilities that runs on our in-memory database, SAP HANA.",
      "shortDescription": "Offers smart, integrated ERP for various sectors on SAP HANA in-memory database.",
      "vendor": "sap:vendor:SAP:",
      "parent": "sap:product:SAPS4HANA:"
    },
	{
	  "ordId": "sap:product:SAPSuccessFactors:",
	  "title": "SAP SuccessFactors",
	  "description": "SAP SuccessFactors is a world-leading provider of cloud human experience management (HXM).",
	  "shortDescription": "SAP SuccessFactors: Leading in cloud-based Human Experience Management (HXM).",
	  "vendor": "sap:vendor:SAP:"
	},
    {
      "ordId": "sap:product:SAPTransactionalBankingforSAPS4HANA:",
      "title": "SAP Transactional Banking for SAP S/4HANA",
      "description": "SAP Transactional Banking for SAP S/4HANA is an open core banking platform, which is based on an architecture that ensures real-time processing and continuous availability.",
	  "shortDescription": "SAP Transactional Banking: Open core banking platform on SAP S/4HANA with real-time processing and constant availability.",
      "vendor": "sap:vendor:SAP:",
      "parent": "sap:product:SAPS4HANA:"
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
