package director_test

import (
	"github.com/pkg/errors"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/director/automock"
	gcliautomock "github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testErr = errors.New("Test error")

func TestDirectorClient_CreateAPIDefinition(t *testing.T) {
	packageID := "foo"
	in := graphql.APIDefinitionInput{
		Name: "bar",
	}

	tests := []struct {
		Name                 string
		GQLClientFn          func() *gcliautomock.GraphQLClient
		GraphqlizerFn        func() *automock.GraphQLizer
		ExpectedResult       *string
		ExpectedErr          error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\t\tresult: addAPIDefinitionToPackage(packageID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.SuccessAPIDefinition"),
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.SuccessAPIDefinition)
					if !ok {
						return
					}

					res.Result = graphql.APIDefinition{ID: "pkgID"}
				}).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("APIDefinitionInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedResult: str.Ptr("pkgID"),
		},
		{
			Name: "Error - GraphQL input",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("APIDefinitionInputToGQL", in).Return("", testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error - GraphQL client",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\t\tresult: addAPIDefinitionToPackage(packageID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.SuccessAPIDefinition"),
				).Return(testErr).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("APIDefinitionInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			gqlCli := tt.GQLClientFn()
			gqlizer := tt.GraphqlizerFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForAPIDefinition").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, gqlizer, gqlFieldsProvider)

			result, err := dirCli.CreateAPIDefinition(packageID, in)

			if tt.ExpectedResult != nil {
				assert.Equal(t, *tt.ExpectedResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlizer, gqlFieldsProvider)
		})
	}
}

