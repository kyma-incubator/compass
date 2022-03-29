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
   "products":[
		{
		  "ordId": "sap:product:SAPCloudPlatform:",
		  "title": "SAP Business Technology Platform",
		  "shortDescription": "Accelerate business outcomes with integration, data to value, and extensibility.",
		  "vendor": "sap:vendor:SAP:"
		}
   ],
   "vendors":[
		{
		  "ordId": "sap:vendor:SAP:",
		  "title": "SAP SE",
		  "partners": []
		}
   ]
}`
