package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/stretchr/testify/require"
)

func TestLabelInput_Validate_Key(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "valid",
			Valid: true,
		},
		{
			Name:  "Invalid - Empty",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
		{
			Name:  "Invalid - Too long",
			Value: inputvalidationtest.String257Long,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidLabelInput()
			sut.Key = testCase.Value
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

func TestLabelInput_Validate_Value(t *testing.T) {
	testCases := []struct {
		Name  string
		Value interface{}
		Valid bool
	}{
		{
			Name:  "Valid",
			Value: "valid",
			Valid: true,
		},
		{
			Name:  "Valid - Map with strings",
			Value: map[string]string{"a": "b"},
			Valid: true,
		},
		{
			Name:  "Valid - Slice of ints",
			Value: []int{1, 2, 3},
			Valid: true,
		},
		{
			Name:  "Invalid - Nil",
			Value: nil,
			Valid: false,
		},
		{
			Name:  "Invalid - Empty string",
			Value: inputvalidationtest.EmptyString,
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sut := fixValidLabelInput()
			sut.Value = testCase.Value
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

func fixValidLabelInput() graphql.LabelInput {
	return graphql.LabelInput{
		Key:   "valid",
		Value: "valid",
	}
}
