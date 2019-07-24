package jsonschema

import (
	"github.com/xeipuuv/gojsonschema"
)

type validator struct {
}

func NewValidator() *validator {
	return &validator{}
}

func (v *validator) Validate(jsonschema, json string) (bool, error) {
	schemaLoader := gojsonschema.NewSchemaLoader()
	stringLoader := gojsonschema.NewStringLoader(jsonschema)

	schema, err := schemaLoader.Compile(stringLoader)
	if err != nil {
		return false, err
	}

	jsonLoader := gojsonschema.NewStringLoader(json)
	result, err := schema.Validate(jsonLoader)
	if err != nil {
		return false, err
	}

	return result.Valid(), nil
}
