package jsonschema

import (
	"github.com/xeipuuv/gojsonschema"
)

type validator struct {
}

func NewValidator() *validator {
	return &validator{}
}

func (v *validator) Validate(jsonschema string, jsons ...string) (bool, error) {

	stringLoader := gojsonschema.NewStringLoader(jsonschema)
	schema, err := gojsonschema.NewSchema(stringLoader)
	if err != nil {
		return false, err
	}

	for _, json := range jsons {
		jsonLoader := gojsonschema.NewStringLoader(json)
		result, err := schema.Validate(jsonLoader)
		if err != nil {
			return false, err
		}

		if result.Valid() == false {
			return false, nil
		}
	}

	return true, nil
}
