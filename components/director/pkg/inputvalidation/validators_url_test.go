package inputvalidation_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestValidateURL(t *testing.T) {
	testCases := []struct {
		Name  string
		Input interface{}
		Valid bool
	}{
		{
			Name:  "Valid https",
			Input: "https://kyma-project.io",
			Valid: true,
		},
		{
			Name:  "Valid http",
			Input: "http://kyma-project.io",
			Valid: true,
		},
		{
			Name:  "URL without protocol",
			Input: "kyma-project.io",
			Valid: false,
		},
		{
			Name:  "Valid",
			Input: str.Ptr("https://kyma-project.io"),
			Valid: true,
		},
		{
			Name:  "Not string",
			Input: 123,
			Valid: false,
		},
		{
			Name:  "nil",
			Input: nil,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			urlValidator := inputvalidation.IsURL

			//WHEN
			err := validation.Validate(testCase.Input, urlValidator)

			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
