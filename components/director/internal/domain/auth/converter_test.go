package auth_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.Auth
		Expected *graphql.Auth
		Error    error
	}{
		{
			Name:     "All properties given",
			Input:    fixDetailedAuth(),
			Expected: fixDetailedGQLAuth(),
		},
		{
			Name:     "Empty",
			Input:    &model.Auth{},
			Expected: &graphql.Auth{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			converter := auth.NewConverter()
			res, err := converter.ToGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *graphql.AuthInput
		Expected *model.AuthInput
	}{
		{
			Name:     "All properties given",
			Input:    fixDetailedGQLAuthInput(),
			Expected: fixDetailedAuthInput(),
		},
		{
			Name:     "All properties given - deprecated",
			Input:    fixDetailedGQLAuthInputDeprecated(),
			Expected: fixDetailedAuthInput(),
		},
		{
			Name:     "Empty",
			Input:    &graphql.AuthInput{},
			Expected: &model.AuthInput{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			converter := auth.NewConverter()
			res, err := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
