package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestDocumentInput_Validate_Title(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "ExpectedValid",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 128 chars",
			Value:         inputvalidationtest.String129Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidDocument()
			doc.Title = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_DisplayName(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "ExpectedValid",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 128 chars",
			Value:         inputvalidationtest.String129Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidDocument()
			doc.DisplayName = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "this is a valid description",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 2000 chars",
			Value:         inputvalidationtest.String2001Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidDocument()
			doc.Description = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_Format(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         graphql.DocumentFormat
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "MARKDOWN",
			ExpectedValid: true,
		},
		{
			Name:          "Invalid",
			Value:         "INVALID",
			ExpectedValid: false,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidDocument()
			doc.Format = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

		})
	}
}

func TestDocumentInput_Validate_Kind(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("ExpectedValid"),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: true,
		},
		{
			Name:          "String longer than 256 chars",
			Value:         str.Ptr(inputvalidationtest.String257Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidDocument()
			doc.Kind = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_Data(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *graphql.CLOB
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         fixCLOB("ExpectedValid"),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "String longer than 256 chars",
			Value:         fixCLOB(inputvalidationtest.String257Long),
			ExpectedValid: true,
		},
		{

			Name:          "Empty string",
			Value:         fixCLOB(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidDocument()
			doc.Data = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_FetchRequest(t *testing.T) {
	validObj := fixValidFetchRequestInput()
	testCases := []struct {
		Name          string
		Value         *graphql.FetchRequestInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &validObj,
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid, nil value",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid object",
			Value:         &graphql.FetchRequestInput{URL: "test-string"},
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			doc := fixValidDocument()
			doc.FetchRequest = testCase.Value
			//WHEN
			err := doc.Validate()
			//THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidDocument() graphql.DocumentInput {
	return graphql.DocumentInput{
		Title:       "title",
		DisplayName: "name",
		Description: "desc",
		Format:      "MARKDOWN",
		Data:        fixCLOB("data"),
		Kind:        str.Ptr("kind"),
	}
}

func fixCLOB(data string) *graphql.CLOB {
	tmp := graphql.CLOB(data)
	return &tmp
}
