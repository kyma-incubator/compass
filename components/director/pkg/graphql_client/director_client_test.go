package graphqlclient

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql_client/automock"

	gcli "github.com/machinebox/graphql"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDirector_WriteTenants(t *testing.T) {
	tenantsInput := []model.BusinessTenantMappingInput{
		{
			Name:           "0283bc56-406b-11ec-9356-0242ac130003",
			ExternalTenant: "123",
			Parent:         "",
			Subdomain:      "subdomain1",
			Region:         "region1",
			Type:           "account",
			Provider:       "provider1",
		},
		{
			Name:           "109534be-406b-11ec-9356-0242ac130003",
			ExternalTenant: "456",
			Parent:         "",
			Subdomain:      "subdomain2",
			Region:         "region2",
			Type:           "type2",
			Provider:       "account",
		},
	}
	expectedQuery := "mutation { writeTenants(in:[{name: \"0283bc56-406b-11ec-9356-0242ac130003\",externalTenant: \"123\",parent: \"\",subdomain: \"subdomain1\",region:\"region1\",type:\"account\",provider: \"provider1\"},{name: \"109534be-406b-11ec-9356-0242ac130003\",externalTenant: \"456\",parent: \"\",subdomain: \"subdomain2\",region:\"region2\",type:\"type2\",provider: \"account\"}])}"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		GQLClient   func() *automock.GraphQLClient
		Input       []model.BusinessTenantMappingInput
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
					return strings.EqualFold(req.Query(), expectedQuery)
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
			director := NewDirector(gqlClient)

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
	tenantsInput := []model.BusinessTenantMappingInput{
		{
			Name:           "0283bc56-406b-11ec-9356-0242ac130003",
			ExternalTenant: "123",
			Parent:         "",
			Subdomain:      "subdomain1",
			Region:         "region1",
			Type:           "account",
			Provider:       "provider1",
		},
		{
			Name:           "109534be-406b-11ec-9356-0242ac130003",
			ExternalTenant: "456",
			Parent:         "",
			Subdomain:      "subdomain2",
			Region:         "region2",
			Type:           "type2",
			Provider:       "account",
		},
	}
	expectedQuery := "mutation { deleteTenants(in:[{name: \"0283bc56-406b-11ec-9356-0242ac130003\",externalTenant: \"123\",parent: \"\",subdomain: \"subdomain1\",region:\"region1\",type:\"account\",provider: \"provider1\"},{name: \"109534be-406b-11ec-9356-0242ac130003\",externalTenant: \"456\",parent: \"\",subdomain: \"subdomain2\",region:\"region2\",type:\"type2\",provider: \"account\"}])}"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		GQLClient   func() *automock.GraphQLClient
		Input       []model.BusinessTenantMappingInput
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
					return strings.EqualFold(req.Query(), expectedQuery)
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
			director := NewDirector(gqlClient)

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

func TestDirector_CreateLabelDefinition(t *testing.T) {
	tenant := "test-tenant"
	lblDefInput := graphql.LabelDefinitionInput{
		Key:    "test-key",
		Schema: nil,
	}
	expectedQuery := `mutation { result: createLabelDefinition(in: {
		key: "test-key",
	} ) {
		key
		schema}}`
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		GQLClient   func() *automock.GraphQLClient
		Input       graphql.LabelDefinitionInput
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
					isQueryExpected := strings.EqualFold(req.Query(), expectedQuery)
					isTenantExpected := req.Header.Get("Tenant") == tenant
					if isQueryExpected && isTenantExpected {
						return true
					}
					return false
				}), mock.Anything).Return(nil)
				return gqlClient
			},
			Input:       lblDefInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when Run fails",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(testErr)
				return gqlClient
			},
			Input:       lblDefInput,
			ExpectedErr: errors.New("while executing gql query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			ctx := context.TODO()
			gqlClient := testCase.GQLClient()
			director := NewDirector(gqlClient)

			//WHEN
			err := director.CreateLabelDefinition(ctx, testCase.Input, tenant)

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

func TestDirector_UpdateLabelDefinition(t *testing.T) {
	tenant := "test-tenant"
	lblDefInput := graphql.LabelDefinitionInput{
		Key:    "test-key",
		Schema: nil,
	}
	expectedQuery := `mutation { result: updateLabelDefinition(in: {
		key: "test-key",
	} ) {
		key
		schema}}`
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		GQLClient   func() *automock.GraphQLClient
		Input       graphql.LabelDefinitionInput
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
					isQueryExpected := strings.EqualFold(req.Query(), expectedQuery)
					isTenantExpected := req.Header.Get("Tenant") == tenant
					if isQueryExpected && isTenantExpected {
						return true
					}
					return false
				}), mock.Anything).Return(nil)
				return gqlClient
			},
			Input:       lblDefInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when Run fails",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(testErr)
				return gqlClient
			},
			Input:       lblDefInput,
			ExpectedErr: errors.New("while executing gql query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			ctx := context.TODO()
			gqlClient := testCase.GQLClient()
			director := NewDirector(gqlClient)

			//WHEN
			err := director.UpdateLabelDefinition(ctx, testCase.Input, tenant)

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

func TestDirector_SetRuntimeTenant(t *testing.T) {
	runtimeID := "test-runtime-id"
	tenantID := "test-tenant-id"
	expectedQuery := `mutation { result: setRuntimeTenant(runtimeID:"test-runtime-id", tenantID:"test-tenant-id") {
		id
		name
		description
		labels 
		status {condition timestamp}
		metadata { creationTimestamp }
		auths {
		id
		auth {credential {
				... on BasicCredentialData {
					username
					password
				}
				...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				}
			}
			oneTimeToken {
				__typename
				token
				used
				expiresAt
			}
			certCommonName
			additionalHeaders
			additionalQueryParams
			requestAuth { 
			  csrf {
				tokenEndpointURL
				credential {
				  ... on BasicCredentialData {
				  	username
					password
				  }
				  ...  on OAuthCredentialData {
					clientId
					clientSecret
					url
					
				  }
			    }
				additionalHeaders
				additionalQueryParams
			}
			}
		}}
		eventingConfiguration { defaultURL }}}`
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		GQLClient   func() *automock.GraphQLClient
		RuntimeID   string
		TenantID    string
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.MatchedBy(func(req *gcli.Request) bool {
					return strings.EqualFold(req.Query(), expectedQuery)
				}), mock.Anything).Return(nil)
				return gqlClient
			},
			RuntimeID:   runtimeID,
			TenantID:    tenantID,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when Run fails",
			GQLClient: func() *automock.GraphQLClient {
				gqlClient := &automock.GraphQLClient{}
				gqlClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(testErr)
				return gqlClient
			},
			RuntimeID:   runtimeID,
			TenantID:    tenantID,
			ExpectedErr: errors.New("while executing gql query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			ctx := context.TODO()
			gqlClient := testCase.GQLClient()
			director := NewDirector(gqlClient)

			//WHEN
			err := director.SetRuntimeTenant(ctx, testCase.RuntimeID, testCase.TenantID)

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
