package gqlcli_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var value = gcli.NewClient("foo")

func TestLoadFromContext(t *testing.T) {
	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult     gqlcli.GraphQLClient
		ExpectedErrMessage string
	}{
		{
			Name:               "Success",
			Context:            context.WithValue(context.TODO(), gqlcli.GraphQLClientContextKey{}, value),
			ExpectedResult:     value,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Error",
			Context:            context.TODO(),
			ExpectedResult:     nil,
			ExpectedErrMessage: "cannot read GraphQL client from context",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result, err := gqlcli.LoadFromContext(testCase.Context)

			// then
			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				require.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			if testCase.ExpectedResult != nil {
				assert.Equal(t, testCase.ExpectedResult, result)
			}
		})
	}
}

func TestSaveToLoadFromContext(t *testing.T) {
	// given
	ctx := context.TODO()

	// when
	result := gqlcli.SaveToContext(ctx, value)

	// then
	assert.Equal(t, value, result.Value(gqlcli.GraphQLClientContextKey{}))
}
