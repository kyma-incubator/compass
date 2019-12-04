package graphql_test

import (
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestFetchRequestInput_Validate_URL(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: inputvalidationtest.ValidURL,
			Valid: true,
		},
		{
			Name:  "URL longer than 256",
			Value: "https://kyma-project.io/" + strings.Repeat("a", 233),
			Valid: false,
		},
		{
			Name:  "Invalid",
			Value: "kyma-project",
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidFetchRequestInput()
			fr.URL = testCase.Value
			//WHEN
			err := fr.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFetchRequestInput_Validate_Auth(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.AuthInput
		Valid bool
	}{
		//TODO: uncommend and fix those tests after implementing AuthInput.Validate()
		//{
		//	Name:  "Valid",
		//	Value: fixValidInputAuth(),
		//	Valid: true,
		//},
		{
			Name:  "Valid nil value",
			Value: nil,
			Valid: true,
		},
		//{
		//	Name:  "Invalid object",
		//	Value: &model.AuthInput{},
		//	Valid: false,
		//},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidFetchRequestInput()
			fr.Auth = testCase.Value
			//WHEN
			err := fr.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFetchRequestInput_Validate_Mode(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.FetchMode
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: (*graphql.FetchMode)(str.Ptr("SINGLE")),
			Valid: true,
		},
		{
			Name:  "Valid nil value",
			Value: nil,
			Valid: true,
		},
		{
			Name:  "Invalid object",
			Value: (*graphql.FetchMode)(str.Ptr("INVALID")),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidFetchRequestInput()
			fr.Mode = testCase.Value
			//WHEN
			err := fr.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestFetchRequestInput_Validate_Filter(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr("this is a valid string"),
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
			Valid: false,
		},
		{
			Name:  "String bigger than 256 chars",
			Value: str.Ptr(inputvalidationtest.String257Long),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			fr := fixValidFetchRequestInput()
			fr.Filter = testCase.Value
			//WHEN
			err := fr.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidFetchRequestInput() graphql.FetchRequestInput {
	return graphql.FetchRequestInput{
		URL: "https://kyma-project.io",
	}
}
