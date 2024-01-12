package tenantmapping_test

import (
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types/tenantmapping"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"testing"
)

func TestFindKeyPath(t *testing.T) {
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
