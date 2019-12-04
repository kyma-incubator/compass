package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSystemInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "name-123.com",
			Valid: true,
		},
		{
			Name:  "Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid Upper Case Letters",
			Value: "Invalid",
			Valid: false,
		},
		{
			Name:  "String longer than 37 chars",
			Value: inputvalidationtest.String37Long,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			is := fixValidIntegrationSystem()
			is.Name = testCase.Value
			//WHEN
			err := is.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestIntegrationSystemInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: str.Ptr("Valid Value"),
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
			Name:  "String longer than 128 chars",
			Value: str.Ptr(inputvalidationtest.String129Long),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			is := fixValidIntegrationSystem()
			is.Description = testCase.Value
			//WHEN
			err := is.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidIntegrationSystem() graphql.IntegrationSystemInput {
	return graphql.IntegrationSystemInput{
		Name: "valid.name",
	}
}
