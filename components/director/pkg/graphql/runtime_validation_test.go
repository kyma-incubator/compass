package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

func TestRuntimeInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: inputvalidationtest.ValidName,
			Valid: true,
		},
		{
			Name:  "Invalid - Empty",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Invalid Name",
			Value: inputvalidationtest.InvalidName,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeInput()
			sut.Name = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestRuntimeInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *string
		Valid bool
	}{
		{
			Name: "Valid",
			Value: str.Ptr("valid	valid"),
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: (*string)(nil),
			Valid: true,
		},
		{
			Name:  "Valid - Empty",
			Value: str.Ptr(inputvalidationtest.EmptyString),
			Valid: true,
		},
		{
			Name:  "Invalid - Too long",
			Value: str.Ptr(inputvalidationtest.String129Long),
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeInput()
			sut.Description = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
func TestRuntimeInput_Validate_Labels(t *testing.T) {
	testCases := []struct {
		Name  string
		Value *graphql.Labels
		Valid bool
	}{
		{
			Name: "Valid",
			Value: &graphql.Labels{
				"test": "ok",
			},
			Valid: true,
		},
		{
			Name:  "Valid - Nil",
			Value: nil,
			Valid: true,
		},
		{
			Name: "Valid - Nil map value",
			Value: &graphql.Labels{
				"test": nil,
			},
			Valid: true,
		},
		{
			Name: "Invalid - Empty map key",
			Value: &graphql.Labels{
				inputvalidationtest.EmptyString: "val",
			},
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidRuntimeInput()
			sut.Labels = testCase.Value
			//WHEN
			err := sut.Validate()
			//THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidRuntimeInput() graphql.RuntimeInput {
	return graphql.RuntimeInput{
		Name: inputvalidationtest.ValidName,
	}
}
