package inputvalidation_test

import (
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/stretchr/testify/require"
)

func TestValidateName(t *testing.T) {
	// GIVEN
	testError := errors.New(`[a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]`)

	testCases := []struct {
		Name          string
		Input         interface{}
		ExpectedError error
	}{
		{
			Name:          "Valid input",
			Input:         "thi5-1npu7.15-valid",
			ExpectedError: nil,
		},
		{
			Name:          "Valid pointer input",
			Input:         str.Ptr("thi5-1npu7.15-valid"),
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
			Input:         strings.Repeat("a", 37),
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
			err := inputvalidation.ValidateName(testCase.Input)
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

func TestValidatePrintable(t *testing.T) {
	// GIVEN
	notPrintableError := errors.New("cannot contain not printable characters")

	testCases := []struct {
		Name          string
		Input         interface{}
		ExpectedError error
	}{
		{
			Name:          "Valid input",
			Input:         "汉ʋǟʟɨɖ ɨռքʊȶ!لْحُرُوف ٱلْعَرَبِيَّر",
			ExpectedError: nil,
		},
		{
			Name:          "Valid pointer input",
			Input:         str.Ptr("汉ʋǟʟɨɖ ɨռքʊȶ!لْحُرُوف ٱلْعَرَبِيَّر"),
			ExpectedError: nil,
		},
		{
			Name:          "Valid enum input",
			Input:         model.DocumentFormatMarkdown,
			ExpectedError: nil,
		},
		{
			Name:          "No error when nil string",
			Input:         (*string)(nil),
			ExpectedError: nil,
		},
		{
			Name:          "Error when invalid input",
			Input:         "\u0000",
			ExpectedError: notPrintableError,
		},
		{
			Name: "Error when invalid input #2",
			Input: "	",
			ExpectedError: notPrintableError,
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
			err := inputvalidation.ValidatePrintable(testCase.Input)
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

func TestValidatePrintableWithWhitespace(t *testing.T) {
	// GIVEN
	notPrintableError := errors.New("cannot contain not printable or whitespace characters")

	testCases := []struct {
		Name          string
		Input         interface{}
		ExpectedError error
	}{
		{
			Name:          "Valid printable input",
			Input:         "汉ʋǟʟɨɖ ɨռքʊȶ!لْحُرُوف ٱلْعَرَبِيَّر",
			ExpectedError: nil,
		},
		{
			Name: "Valid whitespace input",
			Input: "	",
			ExpectedError: nil,
		},
		{
			Name:          "Valid pointer input",
			Input:         str.Ptr("汉ʋǟʟɨɖ ɨռքʊȶ!لْحُرُوف ٱلْعَرَبِيَّر"),
			ExpectedError: nil,
		},
		{
			Name:          "No error when nil string",
			Input:         (*string)(nil),
			ExpectedError: nil,
		},
		{
			Name:          "Valid enum input",
			Input:         model.DocumentFormatMarkdown,
			ExpectedError: nil,
		},
		{
			Name:          "Error when invalid input",
			Input:         "\u0000",
			ExpectedError: notPrintableError,
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
			err := inputvalidation.ValidatePrintableWithWhitespace(testCase.Input)
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
