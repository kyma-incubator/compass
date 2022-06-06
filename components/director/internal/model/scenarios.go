package model

const (
	// ScenariosKey is scenarios label key
	ScenariosKey = "scenarios"
	// DefaultScenario is default value for scenarios label
	DefaultScenario = "DEFAULT"
)

var (
	// ScenariosDefaultValue missing godoc
	ScenariosDefaultValue = []interface{}{DefaultScenario}
	// ScenariosSchema missing godoc
	ScenariosSchema = map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type":      "string",
			"pattern":   "^[A-Za-z0-9]([-_A-Za-z0-9\\s]*[A-Za-z0-9])$",
			"enum":      []string{"DEFAULT"},
			"maxLength": 128,
		},
	}
	// SchemaForScenariosSchema is used to validate ScenariosSchema (allows only modifications to enum field)
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
					"pattern": map[string]interface{}{
						"const": "^[A-Za-z0-9]([-_A-Za-z0-9\\s]*[A-Za-z0-9])$",
					},
					"maxLength": map[string]interface{}{
						"const": 128,
					},
					"enum": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":      "string",
							"pattern":   "^[A-Za-z0-9]([-_A-Za-z0-9\\s]*[A-Za-z0-9])$",
							"maxLength": 128,
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

// NewScenariosSchema returns new default scenario schema
func NewScenariosSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type":      "string",
			"pattern":   "^[A-Za-z0-9]([-_A-Za-z0-9\\s]*[A-Za-z0-9])$",
			"enum":      []string{"DEFAULT"},
			"maxLength": 128,
		},
	}
}
