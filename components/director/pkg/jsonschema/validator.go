package jsonschema

import (
	"github.com/xeipuuv/gojsonschema"
)

type validator struct {
	schema *gojsonschema.Schema
}

func NewValidator(schema *gojsonschema.Schema) *validator {
	return &validator{
		schema: schema,
	}
}

func (v *validator) Validate(json string) (bool, error) {

	if v.schema == nil {
		return true, nil
	}

	jsonLoader := gojsonschema.NewStringLoader(json)
	result, err := v.schema.Validate(jsonLoader)
	if err != nil {
		return false, err
	}

	return result.Valid(), nil
}
