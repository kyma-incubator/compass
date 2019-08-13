package jsonschema

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type validator struct {
	schema *gojsonschema.Schema
}

type ValidationResult struct {
	Valid bool
	Error error
}

func NewValidatorFromStringSchema(jsonSchema string) (*validator, error) {
	if jsonSchema == "" {
		return &validator{schema: nil}, nil
	}

	sl := gojsonschema.NewStringLoader(jsonSchema)
	schema, err := gojsonschema.NewSchema(sl)
	if err != nil {
		return nil, err
	}

	return &validator{
		schema: schema,
	}, nil
}

func NewValidatorFromRawSchema(jsonSchema interface{}) (*validator, error) {
	if jsonSchema == nil {
		return &validator{}, nil
	}

	sl := gojsonschema.NewGoLoader(jsonSchema)
	schema, err := gojsonschema.NewSchema(sl)
	if err != nil {
		return nil, err
	}

	return &validator{
		schema: schema,
	}, nil
}

func (v *validator) ValidateString(json string) (ValidationResult, error) {
	if v.schema == nil {
		return ValidationResult{
			Valid: true,
			Error: nil,
		}, nil
	}

	jsonLoader := gojsonschema.NewStringLoader(json)
	result, err := v.schema.Validate(jsonLoader)
	if err != nil {
		return ValidationResult{}, err
	}

	var validationError *multierror.Error
	for _, e := range result.Errors() {
		validationError = multierror.Append(validationError, errors.New(e.String()))
	}

	if validationError != nil {
		validationError.ErrorFormat = func(i []error) string {
			var s []string
			for _, v := range i {
				s = append(s, v.Error())
			}
			return strings.Join(s, ", ")
		}
	}

	return ValidationResult{
		Valid: result.Valid(),
		Error: validationError,
	}, nil
}

func (v *validator) ValidateRaw(value interface{}) (ValidationResult, error) {
	if v.schema == nil {
		return ValidationResult{
			Valid: true,
			Error: nil,
		}, nil
	}

	valueMarshalled, err := json.Marshal(value)
	if err != nil {
		return ValidationResult{}, err
	}

	return v.ValidateString(string(valueMarshalled))
}
