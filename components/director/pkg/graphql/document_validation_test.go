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
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "Valid",
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "String longer than 128 chars",
			Value: inputvalidationtest.String129Long,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_DisplayName(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "Valid",
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "String longer than 128 chars",
			Value: inputvalidationtest.String129Long,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "this is a valid description",
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "String longer than 128 chars",
			Value: inputvalidationtest.String129Long,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_Format(t *testing.T) {
	testCases := []struct {
		Name  string
		Value graphql.DocumentFormat
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "MARKDOWN",
			Valid: true,
		},
		{
			Name:  "Invalid",
			Value: "INVALID",
			Valid: false,
		},
		{
			Name:  "Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

		})
	}
}

func TestDocumentInput_Validate_Kind(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr("Valid"),
			Valid: true,
		},
		{
			Name:  "Valid nil pointer",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: str.Ptr(inputvalidationtest.EmptyString),
			Valid: true,
		},
		{
			Name:  "String longer than 256 chars",
			Value: str.Ptr(inputvalidationtest.String257Long),
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_Data(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.CLOB
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: fixCLOB("Valid"),
			Valid: true,
		},
		{
			Name:  "Valid nil pointer",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "String longer than 256 chars",
			Value: fixCLOB(inputvalidationtest.String257Long),
			Valid: true,
		},
		{

			Name:  "Empty string",
			Value: fixCLOB(inputvalidationtest.EmptyString),
			Valid: false,
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
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocumentInput_Validate_FetchRequest(t *testing.T) {
	valid := fixValidFetchRequestInput()
	testCases := []struct {
		Name  string
		Value *graphql.FetchRequestInput
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: &valid,
			Valid: true,
		},
		{
			Name:  "Valid, nil value",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Invalid object",
			Value: &graphql.FetchRequestInput{URL: "test-string"},
			Valid: false,
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
			if testCase.Valid {
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
