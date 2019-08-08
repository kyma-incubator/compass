package model

const (
	ScenariosKey = "scenarios"
)

var (
	ScenariosDefaultValue = []interface{}{"DEFAULT"}
	ScenariosSchema       = map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type": "string",
			"enum": []string{"DEFAULT"},
		},
	}
	// This schema is used to validate ScenariosSchema (allows only modifications to enum field)
	SchemaForScenariosSchema = map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"type", "minItems", "uniqueItems", "items"},
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"const": "array",
			},
			"minItems": map[string]interface{}{
				"const": 1,
			},
			"uniqueItems": map[string]interface{}{
				"const": true,
			},
			"items": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"type", "enum"},
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"const": "string",
					},
					"enum": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"contains": map[string]interface{}{
							"const": "DEFAULT",
						},
					},
				},
			},
		},
	}
)
