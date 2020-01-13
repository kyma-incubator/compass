package broker

import (
	"github.com/pivotal-cf/brokerapi/v7/domain"
)

const (
	gcpPlanID   = "ca6e5357-707f-4565-bbbd-b3ab732597c6"
	azurePlanID = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
	awsPlanID   = "badf964a-8908-46e3-bbfd-db2eefe35836"
)

var instanceCreateSchema = []byte(`{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "title": "The cluster name"
    },
    "NodeCount": {
      "type": "integer",
      "default": 3,
      "description": "Number of nodes in the cluster"
    }
  }
}`)

// plans is designed to hold plan defaulting logic
var plans = map[string]struct {
	planDefinition        domain.ServicePlan
	provisioningRawSchema []byte
}{
	gcpPlanID: {
		planDefinition: domain.ServicePlan{
			ID:          gcpPlanID,
			Name:        "gcp",
			Description: "GCP",
			Metadata: &domain.ServicePlanMetadata{
				DisplayName: "GCP",
			},
			Schemas: &domain.ServiceSchemas{
				Instance: domain.ServiceInstanceSchema{
					Create: domain.Schema{
						Parameters: make(map[string]interface{}),
					},
				},
			},
		},
		provisioningRawSchema: []byte(`{
		  "$schema": "http://json-schema.org/draft-04/schema#",
		  "type": "object",
		  "properties": {
            "name": {
              "type": "string"
            },
			"diskType": {
			  "type": "string"
			},
			"nodeCount": {
			  "type": "integer"
			},
			"volumeSizeGb": {
			  "type": "integer"
			},
			"machineType": {
			  "type": "string",
			  "enum": ["n1-standard-2", "n1-standard-4", "n1-standard-8", "n1-standard-16", "n1-standard-32", "n1-standard-64"]
			},
			"region": {
			  "type": "string",
               "enum": ["asia-south1", "asia-southeast1", "asia-east2", "asia-east1", "asia-northeast1", "asia-northeast2", "australia-southeast1", "europe-west2", "europe-west1", "europe-west4", "europe-west6", "europe-west3", "europe-north1", "us-west1", "us-west2", "us-central1", "us-east1", "us-east4", "northamerica-northeast1", "southamerica-east1"]
			},
			"zone": {
			  "type": "string",
             "enum": ["asia-east1-a", "asia-east1-b", "asia-east1-c", "asia-east2-a", "asia-east2-b", "asia-east2-c", "asia-northeast1-a", "asia-northeast1-b", "asia-northeast1-c",
                               "asia-northeast2-a", "asia-northeast2-b","asia-northeast2-c", "asia-south1-a", "asia-south1-b", "asia-south1-c",
                               "asia-southeast1-a", "asia-southeast1-b", "asia-southeast1-c", "australia-southeast1-a", "australia-southeast1-b", "australia-southeast1-c",
                               "europe-north1-a", "europe-north1-c", "europe-north1-b", "europe-west1-b", "europe-west1-c", "europe-west1-d",
                               "europe-west2-a", "europe-west2-b", "europe-west2-c", "europe-west3-a", "europe-west3-b", "europe-west3-c",
                               "europe-west4-a", "europe-west4-b", "europe-west4-c", "europe-west6-a", "europe-west6-b", "europe-west6-c",
                               "northamerica-northeast1-a", "northamerica-northeast1-b", "northamerica-northeast1-c", 
                               "southamerica-east1-a", "southamerica-east1-b", "southamerica-east1-c", "us-central1-a", "us-central1-b", "us-central1-c", "us-central1-f",
                               "us-east1-b", "us-east1-c", "us-east1-d", "us-east4-a", "us-east4-b", "us-east4-c", "us-west1-a", "us-west1-b", "us-west1-c",
                               "us-west2-a", "us-west2-b", "us-west2-c", "us-west2-c"]
			},
			"autoScalerMin": {
			  "type": "integer"
			},
			"autoScalerMax": {
			  "type": "integer"
			},
			"maxSurge": {
			  "type": "integer"
			},
			"maxUnavailable": {
			  "type": "integer"
			}
		  },
		  "required": [
			"name"
		  ]
		}`)},
	azurePlanID: {
		planDefinition: domain.ServicePlan{
			ID:          azurePlanID,
			Name:        "azure",
			Description: "Azure",
			Metadata: &domain.ServicePlanMetadata{
				DisplayName: "Azure",
			},
			Schemas: &domain.ServiceSchemas{
				Instance: domain.ServiceInstanceSchema{
					Create: domain.Schema{
						Parameters: make(map[string]interface{}),
					},
				},
			},
		},
		provisioningRawSchema: []byte(`{
		  "$schema": "http://json-schema.org/draft-04/schema#",
		  "type": "object",
		  "properties": {
            "name": {
			  "type": "string"
			},
			"diskType": {
			  "type": "string"
			},
			"nodeCount": {
			  "type": "integer"
			},
			"volumeSizeGb": {
			  "type": "integer",
              "minimum": 50
			},
			"machineType": {
			  "type": "string",
			  "enum": ["n1-standard-2", "n1-standard-4", "n1-standard-8", "n1-standard-16", "n1-standard-32", "n1-standard-64"]
			},
			"region": {
			  "type": "string",
               "enum": ["westeurope", "eastus", "eastus2", "centralus", "northeurope", "southeastasia", "japaneast", "westus2", "uksouth"]
			},
			"zone": {
			  "type": "string"
            },
			"autoScalerMin": {
			  "type": "integer"
			},
			"autoScalerMax": {
			  "type": "integer"
			},
			"maxSurge": {
			  "type": "integer"
			},
			"maxUnavailable": {
			  "type": "integer"
			}
		  },
		  "required": [
			"name"
		  ]
		}`)},
	awsPlanID: {
		planDefinition: domain.ServicePlan{
			ID:          awsPlanID,
			Name:        "aws",
			Description: "AWS",
			Metadata: &domain.ServicePlanMetadata{
				DisplayName: "AWS",
			},
			Schemas: &domain.ServiceSchemas{
				Instance: domain.ServiceInstanceSchema{
					Create: domain.Schema{
						Parameters: make(map[string]interface{}),
					},
				},
			},
		},
		provisioningRawSchema: []byte(`{
		  "$schema": "http://json-schema.org/draft-04/schema#",
		  "type": "object",
		  "properties": {
            "name": {
			  "type": "string"
			},
			"diskType": {
			  "type": "string",
              "enum": ["gp2", "io1"]
			},
			"nodeCount": {
			  "type": "integer"
			},
			"volumeSizeGb": {
			  "type": "string"
			},
			"machineType": {
			  "type": "string",
			  "enum": ["m5.large", "m5.xlarge", "m5.2xlarge", "m5.4xlarge", "m5.12xlarge", "m5.24xlarge",
                       "m4.large", "m4.xlarge", "m4.2xlarge", "m5.4xlarge", "m4.10xlarge", "m4.16xlarge",
                       "c5.large", "c5.xlarge", "c5.2xlarge", "c5.4xlarge",
                       "c5n.large", "c5n.2xlarge", "c5n.4xlarge", "c5n.9xlarge",
                       "p3.2xlarge", "p3.8xlarge", "p3.16xlarge",
                       "p2.xlarge", "p2.8xlarge", "p2.16xlarge",
                       "r4.large", "r4.xlarge", "r4.2xlarge", "r4.4xlarge", "r4.8xlarge", "r4.16xlarge",
                       "r5.large", "r5.xlarge", "r5.2xlarge", "r5.4xlarge", "r5.8xlarge", "r5.16xlarge",
                       "r5a.large", "r5a.xlarge", "r5a.2xlarge", "r5a.4xlarge", "r5a.16xlarge",
                       "r5d.large", "r5d.xlarge", "r5d.2xlarge", "r5d.4xlarge", "r5d.metal",
                       "x1.16xlarge", "x1.32xlarge", "x1e.16xlarge", "x1e.32xlarge",
                       "t3.small", "t3.medium", "t3.large", "t3.xlarge", "t3.2xlarge", "z1d.metal",
                       "g4dn.xlarge", "g4dn.2xlarge", "g4dn.4xlarge", "g4dn.8xlarge", "g4dn.16xlarge", "g4dn.12xlarge"]
			},
			"region": {
			  "type": "string",
               "enum": ["eu-central-1", "ap-northeast-1", "ap-northeast-2", "ap-south-1", "ap-southeast-1", "ap-southeast-1", 
                        "ap-southeast-2", "ca-central-1", "eu-north-1", "eu-west-1", "eu-west-2", "eu-west-3", "sa-east-1",
                        "us-east-1", "us-east-2", "us-west-1", "us-west-2"]
			},
			"zone": {
			  "type": "string",
              "enum": ["eu-central-1a", "eu-central-1b", "eu-central-1c",
                       "ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d",
                       "ap-northeast-2c", "ap-northeast-2a",
                       "ap-south-1a", "ap-south-1b",
                       "ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c",
					   "ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c",
                       "ca-central-1a", "ca-central-1b",
                       "eu-north-1a", "eu-north-1b", "eu-north-1c",
                       "eu-west-1a", "eu-west-1b", "eu-west-1c",
                       "eu-west-2a", "eu-west-2b", "eu-west-2c",
                       "eu-west-3a", "eu-west-3b", "eu-west-3c",
                       "sa-east-1a", "sa-east-1c",
                       "us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f",
                       "us-east-2a", "us-east-2b", "us-east-2c", "us-east-2d", "us-east-2e", "us-east-2f",
                       "us-west-1b", "us-west-1c", 
                       "us-west-2a", "us-west-2b", "us-west-2c", "us-west-2d"]
            },
			"autoScalerMin": {
			  "type": "integer"
			},
			"autoScalerMax": {
			  "type": "integer"
			},
			"maxSurge": {
			  "type": "integer"
			},
			"maxUnavailable": {
			  "type": "integer"
			}
		  },
		  "required": [
			"name"
		  ]
		}`)},
}
