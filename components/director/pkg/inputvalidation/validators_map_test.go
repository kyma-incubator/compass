package inputvalidation_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestEachKey(t *testing.T) {
	structToValidate := struct {
		String  string
		Map     map[string]string
		Nil     map[string]string
		Pointer *map[string]string
	}{
		String: "test",
		Map: map[string]string{
			"aaa":  "bbbb",
			"AAAA": "BBBB",
			"Ę":    "Ę",
		},
		Nil: nil,
		Pointer: &map[string]string{
			"aaa": "bbb",
		},
	}

	// GIVEN
	testCases := []struct {
		Name          string
		Rules         []*validation.FieldRules
		ExpectedError error
	}{
		{
			Name: "Success",
			Rules: []*validation.FieldRules{
				validation.Field(&structToValidate.Map, inputvalidation.EachKey(validation.Required)),
			},
			ExpectedError: nil,
		},
		{
			Name: "Success when map is nil",
			Rules: []*validation.FieldRules{
				validation.Field(&structToValidate.Nil, inputvalidation.EachKey(validation.Required)),
			},
			ExpectedError: nil,
		},
		{
			Name: "Success when pointer to map",
			Rules: []*validation.FieldRules{
				validation.Field(&structToValidate.Map, inputvalidation.EachKey(validation.Required)),
			},
			ExpectedError: nil,
		},
		{
			Name: "Works with custom validators",
			Rules: []*validation.FieldRules{
				validation.Field(&structToValidate.Map, inputvalidation.EachKey(inputvalidation.Name)),
			},
			ExpectedError: errors.New(dns1123Error),
		},
		{
			Name: "Returns error when field is not a map",
			Rules: []*validation.FieldRules{
				validation.Field(&structToValidate.String, inputvalidation.EachKey(validation.Required)),
			},
			ExpectedError: errors.New("String: the value must be a map."),
		},
		{
			Name: "Returns error when one map key is invalid",
			Rules: []*validation.FieldRules{
				validation.Field(&structToValidate.Map, inputvalidation.EachKey(validation.Required, validation.Length(1, 3))),
			},
			ExpectedError: errors.New("Map: (AAAA: the length must be between 1 and 3.)."),
		},
		{
			Name: "Returns error when multiple map keys are invalid",
			Rules: []*validation.FieldRules{
				validation.Field(&structToValidate.Map, inputvalidation.EachKey(validation.Required, validation.Length(1, 2))),
			},
			ExpectedError: errors.New("Map: (AAAA: the length must be between 1 and 2; aaa: the length must be between 1 and 2.)."),
		},
		{
			Name: "Returns error when pointer to invalid map",
			Rules: []*validation.FieldRules{
				validation.Field(&structToValidate.Pointer, inputvalidation.EachKey(validation.Required, validation.Length(100, 200))),
			},
			ExpectedError: errors.New("Pointer: (aaa: the length must be between 100 and 200.)."),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			err := validation.ValidateStruct(&structToValidate, testCase.Rules...)
			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
}
