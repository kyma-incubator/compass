package inputvalidation_test

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

func TestValidateExactlyOneNotNil(t *testing.T) {
	// GIVEN
	testErrorMessage := "test error message"
	testError := errors.New(testErrorMessage)

	testCases := []struct {
		Name          string
		Pointers      []interface{}
		ExpectedError error
	}{
		{
			Name: "Success",
			Pointers: []interface{}{
				nil, nil, str.Ptr("ok"),
			},
			ExpectedError: nil,
		},
		{
			Name: "Success when pointer to type",
			Pointers: []interface{}{
				str.Ptr("ok"), nil, (*string)(nil),
			},
			ExpectedError: nil,
		},
		{
			Name:          "Success when one element",
			Pointers:      []interface{}{str.Ptr("ok")},
			ExpectedError: nil,
		},
		{
			Name: "Error when all nil",
			Pointers: []interface{}{
				nil, nil, (*string)(nil),
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when more than one not nil",
			Pointers: []interface{}{
				str.Ptr("ok"), nil, str.Ptr("notok"),
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when invalid use",
			Pointers: []interface{}{
				5, str.Ptr("notok"), str.Ptr("notok"),
			},
			ExpectedError: errors.New("internal server error: field is not a pointer"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var ptrs []interface{}
			if len(testCase.Pointers) > 1 {
				ptrs = append(ptrs, testCase.Pointers[1:]...)
			}

			// WHEN
			err := inputvalidation.ValidateExactlyOneNotNil(testErrorMessage, testCase.Pointers[0], ptrs...)

			// THEN
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
