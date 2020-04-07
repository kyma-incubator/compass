package broker

import (
	"encoding/json"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

const (
	GcpPlanID   = "ca6e5357-707f-4565-bbbd-b3ab732597c6"
	AzurePlanID = "4deee563-e5ec-4731-b9b1-53b42d855f0c"
)

func AzureRegions() []string {
	return []string{
		"westus2",
		"westeurope",
	}
}

type Type struct {
	Type            string        `json:"type"`
	Minimum         int           `json:"minimum,omitempty"`
	Enum            []interface{} `json:"enum,omitempty"`
	Items           []Type        `json:"items,omitempty"`
	AdditionalItems *bool         `json:"additionalItems,omitempty"`
	UniqueItems     *bool         `json:"uniqueItems,omitempty"`
	MinLength       int           `json:"minLength,omitempty"`
}
type RootSchema struct {
	Schema string `json:"$schema"`
	Type
	Properties interface{} `json:"properties"`
	Required   []string    `json:"required"`
}

type ProvisioningProperties struct {
	Components     Type `json:"components"`
	Name           Type `json:"name"`
	DiskType       Type `json:"diskType"`
	VolumeSizeGb   Type `json:"volumeSizeGb"`
	MachineType    Type `json:"machineType"`
	Region         Type `json:"region"`
	Zone           Type `json:"zone"`
	AutoScalerMin  Type `json:"autoScalerMin"`
	AutoScalerMax  Type `json:"autoScalerMax"`
	MaxSurge       Type `json:"maxSurge"`
	MaxUnavailable Type `json:"maxUnavailable"`
}

func GCPSchema() []byte {
	f := new(bool)
	*f = false
	t := new(bool)
	*t = true

	rs := RootSchema{
		Schema: "http://json-schema.org/draft-04/schema#",
		Type: Type{
			Type: "object",
		},
		Properties: ProvisioningProperties{
			Components: Type{
				Type: "array",
				Items: []Type{{
					Type: "string",
					Enum: ToInterfaceSlice([]string{"Kiali", "Jaeger"}),
				}},
				AdditionalItems: f,
				UniqueItems:     t,
			},
			Name: Type{
				Type:      "string",
				MinLength: 6,
			},
			DiskType: Type{Type: "string"},
			VolumeSizeGb: Type{
				Type: "integer",
			},
			MachineType: Type{
				Type: "string",
				Enum: ToInterfaceSlice([]string{"n1-standard-2", "n1-standard-4", "n1-standard-8", "n1-standard-16", "n1-standard-32", "n1-standard-64"}),
			},
			Region: Type{
				Type: "string",
				Enum: ToInterfaceSlice([]string{
					"asia-south1", "asia-southeast1",
					"asia-east2", "asia-east1",
					"asia-northeast1", "asia-northeast2",
					"australia-southeast1",
					"europe-west2", "europe-west1", "europe-west4", "europe-west6", "europe-west3",
					"europe-north1",
					"us-west1", "us-west2",
					"us-central1",
					"us-east1", "us-east4",
					"northamerica-northeast1", "southamerica-east1"}),
			},
			Zone: Type{
				Type: "string",
				Enum: ToInterfaceSlice([]string{
					"asia-east1-a", "asia-east1-b", "asia-east1-c", "asia-east2-a", "asia-east2-b", "asia-east2-c",
					"asia-northeast1-a", "asia-northeast1-b", "asia-northeast1-c", "asia-northeast2-a", "asia-northeast2-b", "asia-northeast2-c",
					"asia-south1-a", "asia-south1-b", "asia-south1-c", "asia-southeast1-a",
					"asia-southeast1-b", "asia-southeast1-c", "australia-southeast1-a",
					"australia-southeast1-b", "australia-southeast1-c", "europe-north1-a",
					"europe-north1-c", "europe-north1-b", "europe-west1-b",
					"europe-west1-c", "europe-west1-d", "europe-west2-a", "europe-west2-b", "europe-west2-c", "europe-west3-a", "europe-west3-b", "europe-west3-c", "europe-west4-a", "europe-west4-b", "europe-west4-c", "europe-west6-a", "europe-west6-b", "europe-west6-c", "northamerica-northeast1-a",
					"northamerica-northeast1-b", "northamerica-northeast1-c",
					"southamerica-east1-a", "southamerica-east1-b", "southamerica-east1-c", "us-central1-a",
					"us-central1-b", "us-central1-c", "us-central1-f", "us-east1-b",
					"us-east1-c", "us-east1-d", "us-east4-a", "us-east4-b", "us-east4-c", "us-west1-a", "us-west1-b", "us-west1-c", "us-west2-a", "us-west2-b", "us-west2-c",
				}),
			},
			AutoScalerMin: Type{
				Type: "integer",
			},
			AutoScalerMax: Type{
				Type: "integer",
			},
			MaxSurge: Type{
				Type: "integer",
			},
			MaxUnavailable: Type{
				Type: "integer",
			},
		},
		Required: []string{"name"},
	}

	bytes, err := json.Marshal(rs)
	if err != nil {
		panic(err)
	}
	return bytes
}
func AzureSchema() []byte {
	f := new(bool)
	*f = false
	t := new(bool)
	*t = true
	rs := RootSchema{
		Schema: "http://json-schema.org/draft-04/schema#",
		Type: Type{
			Type: "object",
		},
		Properties: ProvisioningProperties{
			Components: Type{
				Type: "array",
				Items: []Type{{
					Type: "string",
					Enum: ToInterfaceSlice([]string{"Kiali", "Jaeger"}),
				}},
				AdditionalItems: f,
				UniqueItems:     t,
			},
			Name: Type{
				Type:      "string",
				MinLength: 6,
			},
			DiskType: Type{Type: "string"},
			VolumeSizeGb: Type{
				Type:    "integer",
				Minimum: 50,
			},
			MachineType: Type{
				Type: "string",
				Enum: ToInterfaceSlice([]string{"Standard_D8_v3"}),
			},
			Region: Type{
				Type: "string",
				Enum: ToInterfaceSlice(AzureRegions()),
			},
			Zone: Type{
				Type: "string",
			},
			AutoScalerMin: Type{
				Type: "integer",
			},
			AutoScalerMax: Type{
				Type: "integer",
			},
			MaxSurge: Type{
				Type: "integer",
			},
			MaxUnavailable: Type{
				Type: "integer",
			},
		},
		Required: []string{"name"},
	}

	bytes, err := json.Marshal(rs)
	if err != nil {
		panic(err)
	}
	return bytes
}

func ToInterfaceSlice(input []string) []interface{} {
	interfaces := make([]interface{}, len(input))
	for i, item := range input {
		interfaces[i] = item
	}
	return interfaces
}

// plans is designed to hold plan defaulting logic
// keep internal/hyperscaler/azure/config.go in sync with any changes to available zones
var plans = map[string]struct {
	planDefinition        domain.ServicePlan
	provisioningRawSchema []byte
}{
	GcpPlanID: {
		planDefinition: domain.ServicePlan{
			ID:          GcpPlanID,
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
		provisioningRawSchema: GCPSchema(),
	},
	AzurePlanID: {
		planDefinition: domain.ServicePlan{
			ID:          AzurePlanID,
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
		provisioningRawSchema: AzureSchema(),
	},
}
