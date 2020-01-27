package service

import (
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service/automock"
	"github.com/stretchr/testify/require"

	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

func TestGqlRequestBuilder_RegisterApplicationRequest(t *testing.T) {
	// given
	input := fixGQLApplicationRegisterInput("test", "Lorem ipsum")

	t.Run("Success", func(t *testing.T) {
		graphqlizer := &automock.GraphQLizer{}
		graphqlizer.On("ApplicationRegisterInputToGQL", input).Return("{foo}", nil)
		defer graphqlizer.AssertExpectations(t)

		builder := NewGqlRequestBuilder(graphqlizer, nil)
		expectedRq := gcli.NewRequest(`mutation {
			result: registerApplication(in: {foo}) {
				id
			}	
		}`)

		// when
		rq, err := builder.RegisterApplicationRequest(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedRq, rq)
	})

	t.Run("Error", func(t *testing.T) {
		graphqlizer := &automock.GraphQLizer{}
		graphqlizer.On("ApplicationRegisterInputToGQL", input).Return("", testErr)
		defer graphqlizer.AssertExpectations(t)

		builder := NewGqlRequestBuilder(graphqlizer, nil)

		// when
		_, err := builder.RegisterApplicationRequest(input)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
}

func TestGqlRequestBuilder_UnregisterApplicationRequest(t *testing.T) {
	// given
	id := "test"
	builder := NewGqlRequestBuilder(nil, nil)
	expectedRq := gcli.NewRequest(`mutation {
		result: unregisterApplication(id: "test") {
			id
		}	
	}`)

	// when
	rq := builder.UnregisterApplicationRequest(id)

	// then
	assert.Equal(t, expectedRq, rq)
}

func TestGqlRequestBuilder_GetApplicationRequest(t *testing.T) {
	// given
	gqlFieldsProvider := &automock.GqlFieldsProvider{}
	gqlFieldsProvider.On("ForApplication").Return("{foo}")
	defer gqlFieldsProvider.AssertExpectations(t)

	id := "test"
	builder := NewGqlRequestBuilder(nil, gqlFieldsProvider)
	expectedRq := gcli.NewRequest(`query {
			result: application(id: "test") {
					{foo}
			}
		}`)

	// when
	rq := builder.GetApplicationRequest(id)

	// then
	assert.Equal(t, expectedRq, rq)
}
