package inputvalidation_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/stretchr/testify/require"
)

const (
	dns1123Error = `a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character`
)

func TestDNSNameValidator_Validate(t *testing.T) {
	// GIVEN
	testError := errors.New(dns1123Error)

	rule := inputvalidation.DNSName

	testCases := []struct {
		Name          string
		Input         interface{}
		ExpectedError error
	}{
		{
			Name:          "Valid input",
			Input:         inputvalidationtest.ValidName,
			ExpectedError: nil,
		},
		{
			Name:          "Valid pointer input",
			Input:         str.Ptr(inputvalidationtest.ValidName),
			ExpectedError: nil,
		},
		{
			Name:          "No error when nil string",
			Input:         (*string)(nil),
			ExpectedError: nil,
		},
		{
			Name:          "Error when starts with digit",
			Input:         "0invalid",
			ExpectedError: errors.New("cannot start with digit"),
		},
		{
			Name:          "Error when too long input",
			Input:         inputvalidationtest.String37Long,
			ExpectedError: errors.New("must be no more than 36 characters"),
		},
		{
			Name:          "Error when upper case letter",
			Input:         "Test",
			ExpectedError: testError,
		},
		{
			Name:          "Error when not allowed character",
			Input:         "imię",
			ExpectedError: testError,
		},
		{
			Name:          "Error when not allowed character #2",
			Input:         "name;",
			ExpectedError: testError,
		},
		{
			Name:          "Error when invalid type",
			Input:         10,
			ExpectedError: errors.New("type has to be a string"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			err := rule.Validate(testCase.Input)
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
