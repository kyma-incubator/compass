package tenantmapping_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types/tenantmapping"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func Test_FindKeyPath(t *testing.T) {
	testCases := []struct {
		name         string
		json         interface{}
		targetKey    string
		expectedPath string
	}{
		{
			name:         "Success - Simple Object Key",
			json:         gjson.Parse(`{"key": "value"}`).Value(),
			targetKey:    "key",
			expectedPath: "key",
		},
		{
			name: "Success - Nested Object Key",
			json: gjson.Parse(`{
				"parent": {
					"child": "value"
				}
			}`).Value(),
			targetKey:    "child",
			expectedPath: "parent.child",
		},
		{
			name:         "Success - Array Element",
			json:         gjson.Parse(`["item1", "item2", "item3"]`).Value(),
			targetKey:    "item2",
			expectedPath: "1",
		},
		{
			name: "Success - Nested Array Element",
			json: gjson.Parse(`{
				"array": [
					{"key": "value1"},
					{"key": "value2"}
				]
			}`).Value(),
			targetKey:    "value2",
			expectedPath: "array.1.key",
		},
		{
			name: "Success - Nested and Complex Structure",
			json: gjson.Parse(`{
				"config": {
					"api": {
						"url": "https://api.example.com",
						"endpoints": {
							"users": "/users",
							"orders": "/orders"
						}
					},
					"ssl": {
						"enabled": true,
						"certPath": "/path/to/certificate.pem",
						"keyPath": "/path/to/privatekey.pem"
					}
				},
				"modules": [
					{
						"name": "moduleA",
						"settings": {
							"param1": "valueA1",
							"param2": ["itemA1", "itemA2"]
						}
					},
					{
						"name": "moduleB",
						"settings": {
							"param1": "valueB1"
						}
					}
				]
			}`).Value(),
			targetKey:    "param2",
			expectedPath: "modules.0.settings.param2",
		},
		{
			name:         "Key Not Found",
			json:         gjson.Parse(`{"key1": "value1", "key3": "value3"}`).Value(),
			targetKey:    "nonexistent",
			expectedPath: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := tenantmapping.FindKeyPath(testCase.json, testCase.targetKey)

			assert.Equal(t, testCase.expectedPath, result)
		})
	}
}

func Test_GetTenantCommunication(t *testing.T) {
	inboundCommunicationKey := "inboundCommunication"
	outboundCommunicationKey := "outboundCommunication"

	testCases := []struct {
		name              string
		body              *tenantmapping.Body
		tenantType        tenantmapping.TenantType
		communicationType string
		expectedResult    string
	}{
		{
			name: "Assigned tenant with inbound communication",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "123",
					Operation:   "assign",
				},
				AssignedTenant: tenantmapping.AssignedTenant{
					Configuration: json.RawMessage(`{"credentials": {"inboundCommunication": {"key": "value"}}}`),
				},
			},
			tenantType:        tenantmapping.AssignedTenantType,
			communicationType: inboundCommunicationKey,
			expectedResult:    `{"key": "value"}`,
		},
		{
			name: "Assigned tenant with inbound communication - more complex",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "123",
					Operation:   "assign",
				},
				AssignedTenant: tenantmapping.AssignedTenant{
					Configuration: json.RawMessage(`{"credentials": {"inboundCommunication": {"key": {"key2": {"key3": "value"}}}}}`),
				},
			},
			tenantType:        tenantmapping.AssignedTenantType,
			communicationType: inboundCommunicationKey,
			expectedResult:    `{"key": {"key2": {"key3": "value"}}}`,
		},
		{
			name: "Assigned tenant with outbound communication",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "123",
					Operation:   "assign",
				},
				AssignedTenant: tenantmapping.AssignedTenant{
					Configuration: json.RawMessage(`{"credentials": {"outboundCommunication": {"key": "value"}}}`),
				},
			},
			tenantType:        tenantmapping.AssignedTenantType,
			communicationType: outboundCommunicationKey,
			expectedResult:    `{"key": "value"}`,
		},
		{
			name: "Receiver tenant with inbound communication",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "123",
					Operation:   "assign",
				},
				ReceiverTenant: tenantmapping.ReceiverTenant{
					Configuration: json.RawMessage(`{"credentials": {"inboundCommunication": {"key": "value"}}}`),
				},
			},
			tenantType:        tenantmapping.ReceiverTenantType,
			communicationType: inboundCommunicationKey,
			expectedResult:    `{"key": "value"}`,
		},
		{
			name: "Receiver tenant with outbound communication",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "123",
					Operation:   "assign",
				},
				ReceiverTenant: tenantmapping.ReceiverTenant{
					Configuration: json.RawMessage(`{"credentials": {"outboundCommunication": {"key": "value"}}}`),
				},
			},
			tenantType:        tenantmapping.ReceiverTenantType,
			communicationType: outboundCommunicationKey,
			expectedResult:    `{"key": "value"}`,
		},
		{
			name: "Unknown tenant",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "123",
					Operation:   "assign",
				},
				ReceiverTenant: tenantmapping.ReceiverTenant{
					Configuration: json.RawMessage(`{"credentials": {"outboundCommunication": {"key": "value"}}}`),
				},
			},
			tenantType:        tenantmapping.TenantType(3),
			communicationType: outboundCommunicationKey,
			expectedResult:    "",
		},
		{
			name: "Unknown communication type",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "123",
					Operation:   "assign",
				},
				ReceiverTenant: tenantmapping.ReceiverTenant{
					Configuration: json.RawMessage(`{"credentials": {"outboundCommunication": {"key": "value"}}}`),
				},
			},
			tenantType:        tenantmapping.ReceiverTenantType,
			communicationType: "unknown",
			expectedResult:    "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := testCase.body.GetTenantCommunication(testCase.tenantType, testCase.communicationType)
			assert.Equal(t, testCase.expectedResult, actual.Raw)
		})
	}
}

func TestAddReceiverTenantOutboundCommunicationIfMissing(t *testing.T) {
	testCases := []struct {
		name           string
		body           *tenantmapping.Body
		expectedResult string
		expectedError  string
	}{
		{
			name: "Success - outboundCommunication exists",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "formation-id",
					Operation:   "assign",
				},
				ReceiverTenant: tenantmapping.ReceiverTenant{
					DeploymentRegion: "region",
					SubaccountID:     "subaccount",
					Configuration:    json.RawMessage(`{"outboundCommunication": {"auth_method": {}}}`),
				},
			},
			expectedResult: `{"outboundCommunication":{"auth_method":{}}}`,
		},
		{
			name: "Success - outboundCommunication does not exist",
			body: &tenantmapping.Body{
				Context: tenantmapping.Context{
					FormationID: "formation-id",
					Operation:   "assign",
				},
				AssignedTenant: tenantmapping.AssignedTenant{
					Configuration: json.RawMessage(`{"inboundCommunication": {"auth_method": {}}}`),
				},
				ReceiverTenant: tenantmapping.ReceiverTenant{
					DeploymentRegion: "region",
					SubaccountID:     "subaccount",
					Configuration:    json.RawMessage(`{}`),
				},
			},
			expectedResult: `{"outboundCommunication":{}}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.body.AddReceiverTenantOutboundCommunicationIfMissing()

			if testCase.expectedError == "" {
				assert.NoError(t, err)
				assert.JSONEq(t, testCase.expectedResult, string(testCase.body.ReceiverTenant.Configuration))
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.expectedError)
			}
		})
	}
}
