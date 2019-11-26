package inputvalidation_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestEach(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name          string
		Value         interface{}
		ExpectedError error
	}{
		{
			Name:          "Success when nil map",
			Value:         map[string]string{},
			ExpectedError: nil,
		},
		{
			Name:          "Success when map",
			Value:         map[string]string{"key1": "val1", "key2": "val2"},
			ExpectedError: nil,
		},
		{
			Name:          "Success when pointer to map",
			Value:         &map[string]string{"key1": "val1", "key2": "val2"},
			ExpectedError: nil,
		},
		{
			Name:          "Success when slice",
			Value:         []string{"a", "b", "c"},
			ExpectedError: nil,
		},
		{
			Name:          "Success when nil slice",
			Value:         []string{},
			ExpectedError: nil,
		},
		{
			Name:          "Success when pointer slice",
			Value:         &[]string{"a", "b", "c"},
			ExpectedError: nil,
		},
		{
			Name:          "Returns error when map value is empty",
			Value:         map[string]string{"key1": "", "key2": "val2"},
			ExpectedError: errors.New("key1: cannot be blank."),
		},
		{
			Name:          "Returns error when pointer to map with empty value",
			Value:         &map[string]string{"key1": "", "key2": "val2"},
			ExpectedError: errors.New("key1: cannot be blank."),
		},
		{
			Name:          "Returns error when multiple slice values are empty",
			Value:         []string{"", "test", ""},
			ExpectedError: errors.New("0: cannot be blank; 2: cannot be blank."),
		},
		{
			Name:          "Returns error when pointer to slice with empty value",
			Value:         &[]string{"ok", "", "ok"},
			ExpectedError: errors.New("1: cannot be blank."),
		},
		{
			Name:          "Returns error when when nil",
			Value:         nil,
			ExpectedError: errors.New("must be an iterable (map, slice or array) or a pointer to iterable"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := inputvalidation.Each(validation.Required)
			// WHEN
			err := r.Validate(testCase.Value)
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
