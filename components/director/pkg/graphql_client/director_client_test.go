package graphqlclient_test

import (
	"context"
	"strings"
	"testing"

	graphqlclient "github.com/kyma-incubator/compass/components/director/pkg/graphql_client"
	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql_client/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDirector_WriteTenants(t *testing.T) {
	tenantsInput := []graphql.BusinessTenantMappingInput{
		{
			Name:           "0283bc56-406b-11ec-9356-0242ac130003",
			ExternalTenant: "123",
			Parent:         str.Ptr(""),
			Subdomain:      str.Ptr("subdomain1"),
			Region:         str.Ptr("region1"),
			Type:           "account",
			Provider:       "provider1",
		},
		{
			Name:           "109534be-406b-11ec-9356-0242ac130003",
			ExternalTenant: "456",
			Parent:         str.Ptr(""),
			Subdomain:      str.Ptr("subdomain2"),
			Region:         str.Ptr("region2"),
			Type:           "type2",
			Provider:       "account",
		},
	}
	expectedQuery := "mutation { writeTenants(in:[{name: \"0283bc56-406b-11ec-9356-0242ac130003\",externalTenant: \"123\",parent: \"\", region:\"region1\", subdomain: \"subdomain1\",type:\"account\",provider: \"provider1\"},{name: \"109534be-406b-11ec-9356-0242ac130003\",externalTenant: \"456\",parent: \"\", region:\"region2\", subdomain: \"subdomain2\",type:\"type2\",provider: \"account\"}])}"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		GQLClient   func() *automock.GraphQLClient
		Input       []graphql.BusinessTenantMappingInput
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
					return strings.EqualFold(removeWhitespaces(req.Query()), removeWhitespaces(expectedQuery))
				}), mock.Anything).Return(nil)
				return gqlClient
			},
			Input:       tenantsInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when Run fails",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(testErr)
				return gqlClient
			},
			Input:       tenantsInput,
			ExpectedErr: errors.New("while executing gql query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			ctx := context.TODO()
			gqlClient := testCase.GQLClient()
			director := graphqlclient.NewDirector(gqlClient)

			//WHEN
			err := director.WriteTenants(ctx, testCase.Input)

			//THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}
		})
	}
}

func TestDirector_DeleteTenants(t *testing.T) {
	tenantsInput := []graphql.BusinessTenantMappingInput{
		{
			Name:           "0283bc56-406b-11ec-9356-0242ac130003",
			ExternalTenant: "123",
			Parent:         str.Ptr(""),
			Subdomain:      str.Ptr("subdomain1"),
			Region:         str.Ptr("region1"),
			Type:           "account",
			Provider:       "provider1",
		},
		{
			Name:           "109534be-406b-11ec-9356-0242ac130003",
			ExternalTenant: "456",
			Parent:         str.Ptr(""),
			Subdomain:      str.Ptr("subdomain2"),
			Region:         str.Ptr("region2"),
			Type:           "type2",
			Provider:       "account",
		},
	}
	expectedQuery := `mutation { deleteTenants(in:["123", "456"])}`
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		GQLClient   func() *automock.GraphQLClient
		Input       []graphql.BusinessTenantMappingInput
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
					return strings.EqualFold(removeWhitespaces(req.Query()), removeWhitespaces(expectedQuery))
				}), mock.Anything).Return(nil)
				return gqlClient
			},
			Input:       tenantsInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when Run fails",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(testErr)
				return gqlClient
			},
			Input:       tenantsInput,
			ExpectedErr: errors.New("while executing gql query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			ctx := context.TODO()
			gqlClient := testCase.GQLClient()
			director := graphqlclient.NewDirector(gqlClient)

			//WHEN
			err := director.DeleteTenants(ctx, testCase.Input)

			//THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}
		})
	}
}

func TestDirector_UpdateTenant(t *testing.T) {
	tenantID := "id1"
	tenantInput := graphql.BusinessTenantMappingInput{
		Name:           "0283bc56-406b-11ec-9356-0242ac130003",
		ExternalTenant: "123",
		Parent:         str.Ptr(""),
		Subdomain:      str.Ptr("subdomain1"),
		Region:         str.Ptr("region1"),
		Type:           "account",
		Provider:       "provider1",
	}
	expectedQuery := `mutation { updateTenant(id: "id1", in:{
		name: "0283bc56-406b-11ec-9356-0242ac130003",
		externalTenant: "123",
		parent: "",
		region: "region1",
		subdomain: "subdomain1",
		type: "account",
		provider: "provider1"
	}) {
		id
		internalID
		name
		initialized
		labels
		parentID
		type
	}}`
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		GQLClient   func() *automock.GraphQLClient
		TenantID    string
		Input       graphql.BusinessTenantMappingInput
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
					return strings.EqualFold(removeWhitespaces(req.Query()), removeWhitespaces(expectedQuery))
				}), mock.Anything).Return(nil)
				return gqlClient
			},
			TenantID:    tenantID,
			Input:       tenantInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when Run fails",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(testErr)
				return gqlClient
			},
			TenantID:    tenantID,
			Input:       tenantInput,
			ExpectedErr: errors.New("while executing gql query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			ctx := context.TODO()
			gqlClient := testCase.GQLClient()
			director := graphqlclient.NewDirector(gqlClient)

			//WHEN
			err := director.UpdateTenant(ctx, testCase.TenantID, testCase.Input)

			//THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}
		})
	}
}

func removeWhitespaces(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}
