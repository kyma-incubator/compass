package broker

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSchemaGenerator(t *testing.T) {
	tests := []struct {
		name      string
		generator func() []byte
		want      string
	}{
		{
			name:      "Azureschema is correct",
			generator: AzureSchema,
			want: `{
			"$schema": "http://json-schema.org/draft-04/schema#",
			"type": "object",
			"properties": {
			"components": {
			"type": "array",
			"items": [
		{
			"type": "string",
			"enum": ["Kiali", "Jaeger"]
		}
		],
			"additionalItems": false,
			"uniqueItems": true
		},
			"name": {
			"type": "string",
			"minLength": 6
		},
			"diskType": {
			"type": "string"
		},
			"volumeSizeGb": {
			"type": "integer",
			"minimum": 50
		},
			"machineType": {
			"type": "string",
			"enum": ["Standard_D8_v3"]
		},
			"region": {
			"type": "string",
			"enum": [ "westus2", "westeurope" ]
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
		}`},
		{
			name:      "GCPschema is correct",
			generator: GCPSchema,
			want: `{
			"$schema": "http://json-schema.org/draft-04/schema#",
			"type": "object",
			"properties": {
			"components": {
			"type": "array",
			"items": [
		{
			"type": "string",
			"enum": ["Kiali", "Jaeger"]
		}
		],
			"additionalItems": false,
			"uniqueItems": true
		},
			"name": {
			"type": "string",
			"minLength": 6
		},
			"diskType": {
			"type": "string"
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
			"us-west2-a", "us-west2-b", "us-west2-c"]
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
		}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prettyWant bytes.Buffer
			err := json.Indent(&prettyWant, []byte(tt.want), "", "  ")
			if err != nil {
				t.Error(err)
				t.Fail()
			}

			got := tt.generator()
			var prettyGot bytes.Buffer
			err = json.Indent(&prettyGot, got, "", "  ")
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			if !reflect.DeepEqual(string(prettyGot.String()), prettyWant.String()) {
				t.Errorf("GCPSchema() = \n######### GOT ###########%v\n######### ENDGOT ########, want \n##### WANT #####%v\n##### ENDWANT #####", prettyGot.String(), prettyWant.String())
			}
		})
	}
}
