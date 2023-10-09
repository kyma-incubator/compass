package director_test

import (
	"context"
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

var (
	testErr       = errors.New("Test error")
	connectionErr = errors.New("connection refused")
)

func TestDirectorClient_CreateBundle(t *testing.T) {
	appID := "foo"
	in := graphql.BundleCreateInput{
		Name: "bar",
	}
	gqlRequest := gcli.NewRequest("mutation {\n\t\t\tresult: addBundle(applicationID: \"foo\", in: input) {\n\t\t\t\tid\n\t\t\t}}")

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
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateBundleResult)
					if !ok {
						return
					}

					res.Result = graphql.BundleExt{Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "resID"}}}
				}).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("BundleCreateInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedResult: str.Ptr("resID"),
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateBundleResult)
					if !ok {
						return
					}

					res.Result = graphql.BundleExt{Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "resID"}}}
				}).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("BundleCreateInputToGQL", in).Return("input", nil).Once()
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
				am.On("BundleCreateInputToGQL", in).Return("", testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Twice()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("BundleCreateInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			dirCli := director.NewClient(gqlCli, gqlizer, nil)

			result, err := dirCli.CreateBundle(context.TODO(), appID, in)

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

func TestDirectorClient_UpdateBundle(t *testing.T) {
	bundleID := "foo"
	in := graphql.BundleUpdateInput{
		Name: "bar",
	}
	gqlRequest := gcli.NewRequest("mutation {\n\t\t\tresult: updateBundle(id: \"foo\", in: input) {\n\t\t\t\tid\n\t\t\t}\n\t\t}")

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
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("BundleUpdateInputToGQL", in).Return("input", nil).Once()
				return am
			},
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("BundleUpdateInputToGQL", in).Return("input", nil).Once()
				return am
			},
		},
		{
			Name: "Error - GraphQL input",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("BundleUpdateInputToGQL", in).Return("", testErr).Once()
				return am
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Twice()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("BundleUpdateInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			dirCli := director.NewClient(gqlCli, gqlizer, nil)

			err := dirCli.UpdateBundle(context.TODO(), bundleID, in)

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

func TestDirectorClient_GetBundle(t *testing.T) {
	appID := "foo"
	bundleID := "foo"
	successResult := graphql.BundleExt{Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "1"}}}
	gqlRequest := gcli.NewRequest("query {\n\t\t\tresult: application(id: \"foo\") {\n\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}")

	tests := []struct {
		Name           string
		GQLClientFn    func() *gcliautomock.GraphQLClient
		ExpectedResult *graphql.BundleExt
		ExpectedErr    error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.GetBundleResult)
					if !ok {
						return
					}

					res.Result = graphql.ApplicationExt{Bundle: successResult}
				}).Return(nil).Once()
				return am
			},
			ExpectedResult: &successResult,
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.GetBundleResult)
					if !ok {
						return
					}

					res.Result = graphql.ApplicationExt{Bundle: successResult}
				}).Return(nil).Once()
				return am
			},
			ExpectedResult: &successResult,
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Twice()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForApplication", graphqlizer.FieldCtx{"Application.bundle": "bundle(id: \"foo\") {bndl-fields}"}).Return("fields").Maybe()
			gqlFieldsProvider.On("ForBundle").Return("bndl-fields").Maybe()

			dirCli := director.NewClient(gqlCli, nil, gqlFieldsProvider)

			result, err := dirCli.GetBundle(context.TODO(), appID, bundleID)

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

func TestDirectorClient_ListBundles(t *testing.T) {
	appID := "foo"
	successResult := []*graphql.BundleExt{{Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "1"}}}, {Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "2"}}}}
	gqlRequest := gcli.NewRequest("query {\n\t\t\tresult: application(id: \"foo\") {\n\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}")

	tests := []struct {
		Name           string
		GQLClientFn    func() *gcliautomock.GraphQLClient
		ExpectedResult []*graphql.BundleExt
		ExpectedErr    error
	}{
		{
			Name: "Success",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.ListBundlesResult)
					if !ok {
						return
					}

					res.Result = graphql.ApplicationExt{Bundles: graphql.BundlePageExt{Data: successResult}}
				}).Return(nil).Once()
				return am
			},
			ExpectedResult: successResult,
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.ListBundlesResult)
					if !ok {
						return
					}

					res.Result = graphql.ApplicationExt{Bundles: graphql.BundlePageExt{Data: successResult}}
				}).Return(nil).Once()
				return am
			},
			ExpectedResult: successResult,
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Twice()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForApplication").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, nil, gqlFieldsProvider)

			result, err := dirCli.ListBundles(context.TODO(), appID)

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

func TestDirectorClient_DeleteBundle(t *testing.T) {
	id := "foo"
	gqlRequest := gcli.NewRequest("mutation {\n\t\tdeleteBundle(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}")

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
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Twice()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			dirCli := director.NewClient(gqlCli, nil, nil)

			err := dirCli.DeleteBundle(context.TODO(), id)

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
	bundleID := "foo"
	in := graphql.APIDefinitionInput{
		Name: "bar",
	}
	gqlRequest := gcli.NewRequest("mutation {\n\t\t\tresult: addAPIDefinitionToBundle(bundleID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}")

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
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateAPIDefinitionResult)
					if !ok {
						return
					}

					res.Result = graphql.APIDefinition{BaseEntity: &graphql.BaseEntity{ID: "resID"}}
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
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateAPIDefinitionResult)
					if !ok {
						return
					}

					res.Result = graphql.APIDefinition{BaseEntity: &graphql.BaseEntity{ID: "resID"}}
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
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Twice()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("APIDefinitionInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForAPIDefinition").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, gqlizer, gqlFieldsProvider)

			result, err := dirCli.CreateAPIDefinition(context.TODO(), bundleID, in)

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
	gqlRequest := gcli.NewRequest("mutation {\n\t\tdeleteAPIDefinition(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}")

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
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Twice()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			dirCli := director.NewClient(gqlCli, nil, nil)

			err := dirCli.DeleteAPIDefinition(context.TODO(), id)

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
	bundleID := "foo"
	in := graphql.EventDefinitionInput{
		Name: "bar",
	}
	gqlRequest := gcli.NewRequest("mutation {\n\t\t\tresult: addEventDefinitionToBundle(bundleID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}")

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
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateEventDefinitionResult)
					if !ok {
						return
					}

					res.Result = graphql.EventDefinition{BaseEntity: &graphql.BaseEntity{ID: "resID"}}
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
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateEventDefinitionResult)
					if !ok {
						return
					}

					res.Result = graphql.EventDefinition{BaseEntity: &graphql.BaseEntity{ID: "resID"}}
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
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Twice()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("EventDefinitionInputToGQL", in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForEventDefinition").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, gqlizer, gqlFieldsProvider)

			result, err := dirCli.CreateEventDefinition(context.TODO(), bundleID, in)

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
	gqlRequest := gcli.NewRequest("mutation {\n\t\tdeleteEventDefinition(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}")

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
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Twice()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			dirCli := director.NewClient(gqlCli, nil, nil)

			err := dirCli.DeleteEventDefinition(context.TODO(), id)

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
	bundleID := "foo"
	in := graphql.DocumentInput{
		Title: "bar",
	}
	gqlRequest := gcli.NewRequest("mutation {\n\t\t\tresult: addDocumentToBundle(bundleID: \"foo\", in: input) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}")

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
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateDocumentResult)
					if !ok {
						return
					}

					res.Result = graphql.Document{BaseEntity: &graphql.BaseEntity{ID: "resID"}}
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
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					arg := args.Get(2)
					res, ok := arg.(*director.CreateDocumentResult)
					if !ok {
						return
					}

					res.Result = graphql.Document{BaseEntity: &graphql.BaseEntity{ID: "resID"}}
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
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Twice()
				return am
			},
			GraphqlizerFn: func() *automock.GraphQLizer {
				am := &automock.GraphQLizer{}
				am.On("DocumentInputToGQL", &in).Return("input", nil).Once()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()
			gqlizer := tC.GraphqlizerFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForDocument").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, gqlizer, gqlFieldsProvider)

			result, err := dirCli.CreateDocument(context.TODO(), bundleID, in)

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
	gqlRequest := gcli.NewRequest("mutation {\n\t\tdeleteDocument(id: \"foo\") {\n\t\t\tid\n\t\t}\t\n\t}")

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
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					nil,
				).Return(connectionErr).Twice()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			dirCli := director.NewClient(gqlCli, nil, nil)

			err := dirCli.DeleteDocument(context.TODO(), apiID)

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

func TestDirectorClient_SetApplicationLabel(t *testing.T) {
	appID := "foo"
	labelInput := graphql.LabelInput{
		Key:   "testKey",
		Value: "testVal",
	}
	gqlRequest := gcli.NewRequest("mutation {\n\t\t\tresult: setApplicationLabel(applicationID: \"foo\", key: \"testKey\", value: testVal) {\n\t\t\t\t\tfields\n\t\t\t\t}\n\t\t\t}")

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
					gqlRequest,
					mock.Anything,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Success - retry",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Once()
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(nil).Once()
				return am
			},
		},
		{
			Name: "Error - GraphQL client has connectivity problems",
			GQLClientFn: func() *gcliautomock.GraphQLClient {
				am := &gcliautomock.GraphQLClient{}
				am.On("Run",
					mock.Anything,
					gqlRequest,
					mock.Anything,
				).Return(connectionErr).Twice()
				return am
			},
			ExpectedErr: connectionErr,
		},
	}
	for _, tC := range tests {
		t.Run(tC.Name, func(t *testing.T) {
			gqlCli := tC.GQLClientFn()

			gqlFieldsProvider := &automock.GqlFieldsProvider{}
			gqlFieldsProvider.On("ForLabel").Return("fields").Maybe()

			dirCli := director.NewClient(gqlCli, nil, gqlFieldsProvider)

			err := dirCli.SetApplicationLabel(context.TODO(), appID, labelInput)

			if tC.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tC.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlFieldsProvider)
		})
	}
}
