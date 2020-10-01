package graphql_test

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql/graphqlfakes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient_DoDoesNotReturnErrorWhenGQLSucceeds(t *testing.T) {
	fakeGQLClient := &graphqlfakes.FakeGraphQLClient{}
	fakeGQLClient.RunReturns(nil)

	gqlClient, err := graphql.NewClient(graphql.DefaultConfig(), fakeGQLClient)
	require.NoError(t, err)

	require.NotNil(t, gqlClient)
	err = gqlClient.Do(context.TODO(), nil, nil)
	require.NoError(t, err)
}

func TestClient_DoDoesReturnErrorWhenGQLFails(t *testing.T) {
	fakeGQLClient := &graphqlfakes.FakeGraphQLClient{}
	expectedError := errors.New("GraphQL client error")
	fakeGQLClient.RunReturns(expectedError)

	gqlClient, err := graphql.NewClient(graphql.DefaultConfig(), fakeGQLClient)
	require.NoError(t, err)

	require.NotNil(t, gqlClient)
	err = gqlClient.Do(context.TODO(), nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedError.Error())
}
