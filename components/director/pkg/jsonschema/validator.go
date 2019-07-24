package jsonschema

import (
	"github.com/xeipuuv/gojsonschema"
)

type validator struct {
	schema *gojsonschema.Schema
}

func NewValidator(jsonSchema string) (*validator, error) {
	var schema *gojsonschema.Schema
	var err error

	if jsonSchema != "" {
		sl := gojsonschema.NewStringLoader(jsonSchema)
		schema, err = gojsonschema.NewSchema(sl)
		if err != nil {
			return nil, err
		}
	}

	return &validator{
		schema: schema,
	}, nil
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
