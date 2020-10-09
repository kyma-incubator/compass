package graphql_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql/graphqlfakes"
	"github.com/stretchr/testify/require"
)

func TestClient_DoDoesNotReturnErrorWhenGQLSucceeds(t *testing.T) {
	fakeGQLClient := &graphqlfakes.FakeGraphQLClient{}
	fakeGQLClient.RunReturns(nil)

	gqlClient := graphql.NewClient(graphql.DefaultConfig(), fakeGQLClient)

	require.NotNil(t, gqlClient)
	err := gqlClient.Do(context.TODO(), nil, nil)
	require.NoError(t, err)
}

func TestClient_DoDoesReturnErrorWhenGQLFails(t *testing.T) {
	fakeGQLClient := &graphqlfakes.FakeGraphQLClient{}
	expectedError := errors.New("GraphQL client error")
	fakeGQLClient.RunReturns(expectedError)

	gqlClient := graphql.NewClient(graphql.DefaultConfig(), fakeGQLClient)

	require.NotNil(t, gqlClient)
	err := gqlClient.Do(context.TODO(), nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedError.Error())
}
