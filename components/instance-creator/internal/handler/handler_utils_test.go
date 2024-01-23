package handler_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func Test_SubstituteGJSON(t *testing.T) {
	testCases := []struct {
		name             string
		inputJSON        string
		rootMap          interface{}
		expectedOutput   string
		expectedErrorMsg string
	}{
		{
			name:      "Success - Simple Substitution",
			inputJSON: `{"key": "$.value"}`,
			rootMap: gjson.Parse(`{
				"value": "substituted",
			}`).Value(),
			expectedOutput: `{"key":"substituted"}`,
		},
		{
			name:      "Success - Complex Substitution",
			inputJSON: `{"key": "{$.value}/extra"}`,
			rootMap: gjson.Parse(`{
				"value": "substituted"
			}`).Value(),
			expectedOutput: `{"key":"substituted/extra"}`,
		},
		{
			name:      "Success - Array Substitution",
			inputJSON: `{"array": ["$.value", "{$.other}/extra"]}`,
			rootMap: gjson.Parse(`{
				"value": "substituted",
				"other": "another"
			}`).Value(),
			expectedOutput: `{"array":["substituted","another/extra"]}`,
		},
		{
			name:      "Success - Nested Substitution",
			inputJSON: `{"nested": {"inner": "$.value"}}`,
			rootMap: gjson.Parse(`{
				"value": "substituted",
			}`).Value(),
			expectedOutput: `{"nested":{"inner":"substituted"}}`,
		},
		{
			name:      "Success - Multiple Substitutions",
			inputJSON: `{"key1": "$.value1", "key2": "{$.value2}/extra"}`,
			rootMap: gjson.Parse(`{
				"value1": "substituted1",
				"value2": "substituted2",
			}`).Value(),
			expectedOutput: `{"key1":"substituted1","key2":"substituted2/extra"}`,
		},
		{
			name: "Success - Multiple Complex Substitutions",
			inputJSON: `{
				"name": "{$.user.firstName} {$.user.lastName}",
				"age": "$.user.age",
    		    "address": {
    		        "street": "{$.user.address.street}",
    		        "city": "$.user.address.city"
    		    }
    		}`,
			rootMap: gjson.Parse(`{
				"user": {
					"firstName": "John",
					"lastName":  "Doe",
					"age":       30,
					"address": {
						"street": "123 Main St",
						"city":   "Example City"
					}
				}
			}`).Value(),
			expectedOutput: `{
   			    "name": "John Doe",
   			    "age": 30,
   			    "address": {
   			        "street": "123 Main St",
   			        "city": "Example City"
   			    }
   			}`,
		},
		{
			name: "Success - Multiple Complex Substitutions",
			inputJSON: `{
				"address": "{$.address.url}/{$.address.endpoint}",
				"certificate": "$.address.certificate"
    		}`,
			rootMap: gjson.Parse(`{
				"address": {
					"url":         "www.url.com",
					"endpoint":    "testing",
					"certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
				}
			}`).Value(),
			expectedOutput: `{
   			    "address": "www.url.com/testing",
   			    "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
   			}`,
		},
		{
			name: "Success - More Complex and Nested Substitutions",
			inputJSON: `{
    		    "userInfo": {
    		        "fullName": "{$.user.firstName} {$.user.lastName}",
    		        "contact": {
    		            "email": "{$.user.contact.email}",
    		            "phoneNumbers": ["{$.user.contact.phone[0]}", "{$.user.contact.phone[1]}"]
    		        }
    		    },
    		    "orders": [
    		        {
    		            "id": "$.user.orders[0].id",
    		            "products": ["{$.user.orders[0].products[0]}", "{$.user.orders[0].products[1]}"]
    		        },
    		        {
    		            "id": "$.user.orders[1].id",
    		            "products": ["{$.user.orders[1].products[0]}", "{$.user.orders[1].products[1]}"]
    		        }
    		    ]
    		}`,
			rootMap: gjson.Parse(`{"user": {
					"firstName": "John",
					"lastName":  "Doe",
					"contact": {
						"email": "john.doe@example.com",
						"phone": ["123-456-7890", "987-654-3210"]
					},
					"orders": [
						{
							"id":       101,
							"products": ["Product A", "Product B"]
						},
						{
							"id":       102,
							"products": ["Product C", "Product D"]
						}
					]
				}
			}`).Value(),
			expectedOutput: `{
    		    "userInfo": {
    		        "fullName": "John Doe",
    		        "contact": {
    		            "email": "john.doe@example.com",
    		            "phoneNumbers": ["123-456-7890", "987-654-3210"]
    		        }
    		    },
    		    "orders": [
    		        {
    		            "id": 101,
    		            "products": ["Product A", "Product B"]
    		        },
    		        {
    		            "id": 102,
    		            "products": ["Product C", "Product D"]
    		        }
    		    ]
    		}`,
		},
		{
			name:             "Error in JSONPath",
			inputJSON:        `{"key": "$.nonexistent"}`,
			rootMap:          gjson.Parse(`{"value": "substituted"}`).Value(),
			expectedErrorMsg: "unknown key nonexistent",
		},
		{
			name:             "Error in JSONPath Array",
			inputJSON:        `{"array": ["$.value", "$.nonexistent"]}`,
			rootMap:          gjson.Parse(`{"value": "substituted"}`).Value(),
			expectedErrorMsg: "unknown key nonexistent",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			inputJSON := gjson.Parse(testCase.inputJSON)

			result, err := handler.SubstituteGJSON(context.TODO(), inputJSON, testCase.rootMap)

			if testCase.expectedErrorMsg != "" {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErrorMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, testCase.expectedOutput, result.Raw)
			}
		})
	}
}

func Test_DeepMergeJSON(t *testing.T) {
	testCases := []struct {
		name         string
		srcJSON      string
		destJSON     string
		expectedJSON string
	}{
		{
			name:         "Success - Simple Merge",
			srcJSON:      `{"key1": "value1"}`,
			destJSON:     `{"key2": "value2"}`,
			expectedJSON: `{"key1":"value1","key2":"value2"}`,
		},
		{
			name:         "Success - Nested Merge",
			srcJSON:      `{"user": {"name": "John"}}`,
			destJSON:     `{"user": {"age": 30}}`,
			expectedJSON: `{"user":{"age":30,"name":"John"}}`,
		},
		{
			name:         "Success - Array Merge",
			srcJSON:      `{"items": ["item1", "item2"]}`,
			destJSON:     `{"items": ["item3"]}`,
			expectedJSON: `{"items":["item3","item1","item2"]}`,
		},
		{
			name:         "Success - Overwrite Existing Value",
			srcJSON:      `{"key": "newValue"}`,
			destJSON:     `{"key": "oldValue"}`,
			expectedJSON: `{"key":"newValue"}`,
		},
		{
			name: "Success - Complex Merge",
			srcJSON: `{
				"user": {
					"name": "Alice",
					"address": {
						"city": "Wonderland"
					},
					"lastname": "Doe"
				},
				"items": ["apple", "orange"]
			}`,
			destJSON: `{
				"user": {
					"age": 25,
					"address": {
						"country": "Fantasy"
					},
					"lastname": "Something"
				},
				"items": ["banana"]
			}`,
			expectedJSON: `{
				"user": {
					"age": 25,
					"name": "Alice",
					"address": {
						"city": "Wonderland",
						"country": "Fantasy"
					},
					"lastname": "Doe"
				},
				"items": ["banana", "apple", "orange"]
			}`,
		},
		{
			name: "Success - Complex and Nested Merge",
			srcJSON: `{
				"config": {
					"api": {
						"url": "https://api.example.com",
						"endpoints": {
							"users": "/users",
							"orders": "/orders"
						},
						"security": {
							"apiKey": "api-key-123",
							"certificate": "-----BEGIN CERTIFICATE-----\nYour Certificate\n-----END CERTIFICATE-----"
						}
					},
					"ssl": {
						"enabled": true,
						"certPath": "/path/to/certificate.pem",
						"keyPath": "/path/to/privatekey.pem"
					}
				},
				"modules": {
					"moduleA": {
						"enabled": true,
						"settings": {
							"param1": "valueA1",
							"param2": ["itemA1", "itemA2"]
						}
					},
					"moduleB": {
						"enabled": false,
						"settings": {
							"param1": "valueB1",
							"param3": 42
						}
					}
				}
			}`,
			destJSON: `{
				"config": {
					"api": {
						"url": "https://api.example.com",
						"endpoints": {
							"users": "/users"
						},
						"security": {
							"apiKey": "old-api-key",
							"certificate": "-----BEGIN CERTIFICATE-----\nOld Certificate\n-----END CERTIFICATE-----"
						}
					},
					"ssl": {
						"enabled": false,
						"certPath": "/old/cert.pem"
					}
				},
				"modules": {
					"moduleA": {
						"enabled": false,
						"settings": {
							"param1": "valueOldA1",
							"param2": ["itemOldA1"]
						}
					},
					"moduleC": {
						"enabled": true,
						"settings": {
							"param4": "valueC1"
						}
					}
				}
			}`,
			expectedJSON: `{
				"config": {
					"api": {
						"url": "https://api.example.com",
						"endpoints": {
							"users": "/users",
							"orders": "/orders"
						},
						"security": {
							"apiKey": "api-key-123",
							"certificate": "-----BEGIN CERTIFICATE-----\nYour Certificate\n-----END CERTIFICATE-----"
						}
					},
					"ssl": {
						"enabled": true,
						"certPath": "/path/to/certificate.pem",
						"keyPath": "/path/to/privatekey.pem"
					}
				},
				"modules": {
					"moduleA": {
						"enabled": true,
						"settings": {
							"param1": "valueA1",
							"param2": ["itemOldA1", "itemA1", "itemA2"]
						}
					},
					"moduleB": {
						"enabled": false,
						"settings": {
							"param1": "valueB1",
							"param3": 42
						}
					},
					"moduleC": {
						"enabled": true,
						"settings": {
							"param4": "valueC1"
						}
					}
				}
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srcJSON := gjson.Parse(tc.srcJSON)
			destJSON := gjson.Parse(tc.destJSON)

			result := handler.DeepMergeJSON(srcJSON, destJSON)

			assert.JSONEq(t, tc.expectedJSON, result.Raw)
		})
	}
}
