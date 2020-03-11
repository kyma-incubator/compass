package director_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/pkg/errors"

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

func TestDirectorClient_CreatePackage(t *testing.T) {
	appID := "foo"
	in := graphql.PackageCreateInput{
		Name: "bar",
	}

	tests := []struct {
		Name           string
		GQLClientFn    func() *gcliautomock.GraphQLClient
		GraphqlizerFn  func() *automock.GraphQLizer
		ExpectedResult *string
		ExpectedErr    error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\t\tresult: addPackage(applicationID: \"foo\", in: input) {\n\t\t\t\tid\n\t\t\t}}"),
					mock.AnythingOfType("*director.CreatePackageResult"),
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreatePackageResult)
					if !ok {
						return
					}

					res.Result = graphql.PackageExt{Package: graphql.Package{ID: "resID"}}
				}).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("PackageCreateInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedResult: str.Ptr("resID"),
		},
		{
			Name: "Error - GraphQL input",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("PackageCreateInputToGQL", in).Return("", testErr).Once()
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
					gcli.NewRequest("mutation {\n\t\t\tresult: addPackage(applicationID: \"foo\", in: input) {\n\t\t\t\tid\n\t\t\t}}"),
					mock.AnythingOfType("*director.CreatePackageResult"),
				).Return(testErr).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("PackageCreateInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			dirCli := director.NewClient(gqlCli, gqlizer, nil)

			result, err := dirCli.CreatePackage(appID, in)

			if tC.ExpectedResult != nil {
				assert.Equal(t, *tC.ExpectedResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlizer)
		})
	}
}

func TestDirectorClient_UpdatePackage(t *testing.T) {
	packageID := "foo"
	in := graphql.PackageUpdateInput{
		Name: "bar",
	}

	tests := []struct {
		Name          string
		GQLClientFn   func() *gcliautomock.GraphQLClient
		GraphqlizerFn func() *automock.GraphQLizer
		ExpectedErr   error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\t\tresult: updatePackage(id: \"foo\", in: input) {\n\t\t\t\tid\n\t\t\t}\n\t\t}"),
					nil,
				).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("PackageUpdateInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: nil,
		},
		{
			Name: "Error - GraphQL input",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("PackageUpdateInputToGQL", in).Return("", testErr).Once()
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
					gcli.NewRequest("mutation {\n\t\t\tresult: updatePackage(id: \"foo\", in: input) {\n\t\t\t\tid\n\t\t\t}\n\t\t}"),
					nil,
				).Return(testErr).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("PackageUpdateInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			dirCli := director.NewClient(gqlCli, gqlizer, nil)

			err := dirCli.UpdatePackage(packageID, in)

			if tC.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlizer)
		})
	}
}

func TestDirectorClient_GetPackage(t *testing.T) {
	appID := "foo"
	packageID := "foo"
	successResult := graphql.PackageExt{Package: graphql.Package{ID: "1"}}

	tests := []struct {
		Name           string
		GQLClientFn    func() *gcliautomock.GraphQLClient
		ExpectedResult *graphql.PackageExt
		ExpectedErr    error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("query {\n\t\t\tresult: application(id: \"foo\") {\n\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.GetPackageResult"),
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.GetPackageResult)
					if !ok {
						return
					}

					res.Result = graphql.ApplicationExt{Package: successResult}
				}).Return(nil).Once()
				return am
			},
			ExpectedResult: &successResult,
		},
		{
			Name: "Error - GraphQL client",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("query {\n\t\t\tresult: application(id: \"foo\") {\n\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.GetPackageResult"),
				).Return(testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForApplication", graphqlizer.FieldCtx{"Application.package": "package(id: \"foo\") {pkg-fields}"}).Return("fields").Maybe()
			gqlFieldsProvider.On("ForPackage").Return("pkg-fields").Maybe()

			dirCli := director.NewClient(gqlCli, nil, gqlFieldsProvider)

			result, err := dirCli.GetPackage(appID, packageID)

			if tC.ExpectedResult != nil {
				assert.Equal(t, *tC.ExpectedResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlFieldsProvider)
		})
	}
}

func TestDirectorClient_ListPackages(t *testing.T) {
	appID := "foo"
	successResult := []*graphql.PackageExt{{Package: graphql.Package{ID: "1"}}, {Package: graphql.Package{ID: "2"}}}

	tests := []struct {
		Name           string
		GQLClientFn    func() *gcliautomock.GraphQLClient
		ExpectedResult []*graphql.PackageExt
		ExpectedErr    error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("query {\n\t\t\tresult: application(id: \"foo\") {\n\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.ListPackagesResult"),
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.ListPackagesResult)
					if !ok {
						return
					}

					res.Result = graphql.ApplicationExt{Packages: graphql.PackagePageExt{Data: successResult}}
				}).Return(nil).Once()
				return am
			},
			ExpectedResult: successResult,
		},
		{
			Name: "Error - GraphQL client",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("query {\n\t\t\tresult: application(id: \"foo\") {\n\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.ListPackagesResult"),
				).Return(testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForApplication").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, nil, gqlFieldsProvider)

			result, err := dirCli.ListPackages(appID)

			if tC.ExpectedResult != nil {
				assert.Equal(t, tC.ExpectedResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlFieldsProvider)
		})
	}
}

func TestDirectorClient_DeletePackage(t *testing.T) {
	id := "foo"

	tests := []struct {
		Name        string
		GQLClientFn func() *gcliautomock.GraphQLClient
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\tdeletePackage(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}"),
					nil,
				).Return(nil).Once()
				return am
			},
			ExpectedErr: nil,
		},
		{
			Name: "Error - GraphQL client",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\tdeletePackage(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}"),
					nil,
				).Return(testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			dirCli := director.NewClient(gqlCli, nil, nil)

			err := dirCli.DeletePackage(id)

			if tC.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, gqlCli)
		})
	}
}

func TestDirectorClient_CreateAPIDefinition(t *testing.T) {
	packageID := "foo"
	in := graphql.APIDefinitionInput{
		Name: "bar",
	}

	tests := []struct {
		Name           string
		GQLClientFn    func() *gcliautomock.GraphQLClient
		GraphqlizerFn  func() *automock.GraphQLizer
		ExpectedResult *string
		ExpectedErr    error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\t\tresult: addAPIDefinitionToPackage(packageID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.CreateAPIDefinitionResult"),
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateAPIDefinitionResult)
					if !ok {
						return
					}

					res.Result = graphql.APIDefinition{ID: "resID"}
				}).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("APIDefinitionInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedResult: str.Ptr("resID"),
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
					mock.AnythingOfType("*director.CreateAPIDefinitionResult"),
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
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForAPIDefinition").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, gqlizer, gqlFieldsProvider)

			result, err := dirCli.CreateAPIDefinition(packageID, in)

			if tC.ExpectedResult != nil {
				assert.Equal(t, *tC.ExpectedResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlizer, gqlFieldsProvider)
		})
	}
}

func TestDirectorClient_DeleteAPIDefinition(t *testing.T) {
	id := "foo"

	tests := []struct {
		Name        string
		GQLClientFn func() *gcliautomock.GraphQLClient
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\tdeleteAPIDefinition(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}"),
					nil,
				).Return(nil).Once()
				return am
			},
			ExpectedErr: nil,
		},
		{
			Name: "Error - GraphQL client",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\tdeleteAPIDefinition(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}"),
					nil,
				).Return(testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			dirCli := director.NewClient(gqlCli, nil, nil)

			err := dirCli.DeleteAPIDefinition(id)

			if tC.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, gqlCli)
		})
	}
}

func TestDirectorClient_CreateEventDefinition(t *testing.T) {
	packageID := "foo"
	in := graphql.EventDefinitionInput{
		Name: "bar",
	}

	tests := []struct {
		Name           string
		GQLClientFn    func() *gcliautomock.GraphQLClient
		GraphqlizerFn  func() *automock.GraphQLizer
		ExpectedResult *string
		ExpectedErr    error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\t\tresult: addEventDefinitionToPackage(packageID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.CreateEventDefinitionResult"),
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateEventDefinitionResult)
					if !ok {
						return
					}

					res.Result = graphql.EventDefinition{ID: "resID"}
				}).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("EventDefinitionInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedResult: str.Ptr("resID"),
		},
		{
			Name: "Error - GraphQL input",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("EventDefinitionInputToGQL", in).Return("", testErr).Once()
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
					gcli.NewRequest("mutation {\n\t\t\tresult: addEventDefinitionToPackage(packageID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.CreateEventDefinitionResult"),
				).Return(testErr).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("EventDefinitionInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForEventDefinition").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, gqlizer, gqlFieldsProvider)

			result, err := dirCli.CreateEventDefinition(packageID, in)

			if tC.ExpectedResult != nil {
				assert.Equal(t, *tC.ExpectedResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlizer, gqlFieldsProvider)
		})
	}
}

func TestDirectorClient_DeleteEventDefinition(t *testing.T) {
	id := "foo"

	tests := []struct {
		Name        string
		GQLClientFn func() *gcliautomock.GraphQLClient
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\tdeleteEventDefinition(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}"),
					nil,
				).Return(nil).Once()
				return am
			},
			ExpectedErr: nil,
		},
		{
			Name: "Error - GraphQL client",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\tdeleteEventDefinition(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}"),
					nil,
				).Return(testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			dirCli := director.NewClient(gqlCli, nil, nil)

			err := dirCli.DeleteEventDefinition(id)

			if tC.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, gqlCli)
		})
	}
}

func TestDirectorClient_CreateDocument(t *testing.T) {
	packageID := "foo"
	in := graphql.DocumentInput{
		Title: "bar",
	}

	tests := []struct {
		Name           string
		GQLClientFn    func() *gcliautomock.GraphQLClient
		GraphqlizerFn  func() *automock.GraphQLizer
		ExpectedResult *string
		ExpectedErr    error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\t\tresult: addDocumentToPackage(packageID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.CreateDocumentResult"),
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateDocumentResult)
					if !ok {
						return
					}

					res.Result = graphql.Document{ID: "resID"}
				}).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("DocumentInputToGQL", &in).Return("input", nil).Once()
				return am
			},
			ExpectedResult: str.Ptr("resID"),
		},
		{
			Name: "Error - GraphQL input",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("DocumentInputToGQL", &in).Return("", testErr).Once()
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
					gcli.NewRequest("mutation {\n\t\t\tresult: addDocumentToPackage(packageID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}"),
					mock.AnythingOfType("*director.CreateDocumentResult"),
				).Return(testErr).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("DocumentInputToGQL", &in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForDocument").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, gqlizer, gqlFieldsProvider)

			result, err := dirCli.CreateDocument(packageID, in)

			if tC.ExpectedResult != nil {
				assert.Equal(t, *tC.ExpectedResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlizer, gqlFieldsProvider)
		})
	}
}

func TestDirectorClient_DeleteDocument(t *testing.T) {
	apiID := "foo"

	tests := []struct {
		Name        string
		GQLClientFn func() *gcliautomock.GraphQLClient
		ExpectedErr error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\tdeleteDocument(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}"),
					nil,
				).Return(nil).Once()
				return am
			},
			ExpectedErr: nil,
		},
		{
			Name: "Error - GraphQL client",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gcli.NewRequest("mutation {\n\t\tdeleteDocument(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}"),
					nil,
				).Return(testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			dirCli := director.NewClient(gqlCli, nil, nil)

			err := dirCli.DeleteDocument(apiID)

			if tC.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, gqlCli)
		})
	}
}
