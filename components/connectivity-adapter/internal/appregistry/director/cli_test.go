package director_test

import (
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

func TestDirectorClient_CreateAPIDefinition(t *testing.T) {
	packageID := "foo"
	in := graphql.APIDefinitionInput{
		Name: "bar",
	}

	tests := []struct {
		Name                 string
		GQLClientFn          func() *gcliautomock.GraphQLClient
		GraphqlizerFn        func() *automock.GraphQLizer
		GqQLFieldsProviderFn func() *automock.GqlFieldsProvider
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
			GqQLFieldsProviderFn: func() *automock.GqlFieldsProvider {
				am := &automock.GqlFieldsProvider{}
				am.On("ForAPIDefinition").Return("fields").Once()
				return am
			},
			ExpectedResult: str.Ptr("pkgID"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			gqlCli := tt.GQLClientFn()
			gqlizer := tt.GraphqlizerFn()
			gqlFieldsProvider := tt.GqQLFieldsProviderFn()

			dirCli := director.NewClient(gqlCli, gqlizer, gqlFieldsProvider)

			result, err := dirCli.CreateAPIDefinition(packageID, in)

			if tt.ExpectedResult != nil {
				assert.Equal(t, *tt.ExpectedResult, result)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.ExpectedErr.Error(), err.Error())
			}

			mock.AssertExpectationsForObjects(t, gqlCli, gqlizer, gqlFieldsProvider)
		})
	}
}

//
//func Test_directorClient_CreateDocument(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		packageID     string
//		documentInput graphql.DocumentInput
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    string
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			got, err := c.CreateDocument(tt.args.packageID, tt.args.documentInput)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("CreateDocument() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("CreateDocument() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_directorClient_CreateEventDefinition(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		packageID            string
//		eventDefinitionInput graphql.EventDefinitionInput
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    string
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			got, err := c.CreateEventDefinition(tt.args.packageID, tt.args.eventDefinitionInput)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("CreateEventDefinition() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("CreateEventDefinition() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_directorClient_CreatePackage(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		appID string
//		in    graphql.PackageCreateInput
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    string
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			got, err := c.CreatePackage(tt.args.appID, tt.args.in)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("CreatePackage() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("CreatePackage() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_directorClient_DeleteAPIDefinition(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		apiID string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			if err := c.DeleteAPIDefinition(tt.args.apiID); (err != nil) != tt.wantErr {
//				t.Errorf("DeleteAPIDefinition() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func Test_directorClient_DeleteDocument(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		documentID string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			if err := c.DeleteDocument(tt.args.documentID); (err != nil) != tt.wantErr {
//				t.Errorf("DeleteDocument() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func Test_directorClient_DeleteEventDefinition(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		eventID string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			if err := c.DeleteEventDefinition(tt.args.eventID); (err != nil) != tt.wantErr {
//				t.Errorf("DeleteEventDefinition() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func Test_directorClient_DeletePackage(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		packageID string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			if err := c.DeletePackage(tt.args.packageID); (err != nil) != tt.wantErr {
//				t.Errorf("DeletePackage() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func Test_directorClient_GetApplicationsByNameRequest(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		appName string
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//		want   *gcli.Request
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			if got := c.GetApplicationsByNameRequest(tt.args.appName); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetApplicationsByNameRequest() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_directorClient_GetPackage(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		appID     string
//		packageID string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    graphql.PackageExt
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			got, err := c.GetPackage(tt.args.appID, tt.args.packageID)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetPackage() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetPackage() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_directorClient_ListPackages(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		appID string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    []*graphql.PackageExt
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			got, err := c.ListPackages(tt.args.appID)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("ListPackages() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("ListPackages() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_directorClient_UpdatePackage(t *testing.T) {
//	type fields struct {
//		cli               gqlcli.GraphQLClient
//		graphqlizer       GraphQLizer
//		gqlFieldsProvider GqlFieldsProvider
//	}
//	type args struct {
//		packageID string
//		in        graphql.PackageUpdateInput
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &directorClient{
//				cli:               tt.fields.cli,
//				graphqlizer:       tt.fields.graphqlizer,
//				gqlFieldsProvider: tt.fields.gqlFieldsProvider,
//			}
//			if err := c.UpdatePackage(tt.args.packageID, tt.args.in); (err != nil) != tt.wantErr {
//				t.Errorf("UpdatePackage() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
