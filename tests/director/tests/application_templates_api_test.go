package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestCreateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template-name"
	appTemplateInput := pkg.FixApplicationTemplate(name)
	appTemplate, err := pkg.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	createApplicationTemplateRequest := pkg.FixCreateApplicationTemplateRequest(appTemplate)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Create application template")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createApplicationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer pkg.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, createApplicationTemplateRequest.Query(), "create application template")

	t.Log("Check if application template was created")

	getApplicationTemplateRequest := pkg.FixApplicationTemplateRequest(output.ID)
	appTemplateOutput := graphql.ApplicationTemplate{}

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getApplicationTemplateRequest, &appTemplateOutput)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOutput)
	assertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
	saveExample(t, getApplicationTemplateRequest.Query(), "query application template")
}

func TestUpdateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"
	newName := "new-app-template"
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationRegisterInput{
		Name:           "new-app-create-input",
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTemplate := pkg.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, pkg.FixApplicationTemplate(name))
	defer pkg.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTemplate.ID)

	appTemplateInput := graphql.ApplicationTemplateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateGQL, err := pkg.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
	updateAppTemplateRequest := pkg.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateAppTemplateRequest, &updateOutput)
	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	//THEN
	t.Log("Check if application template was updated")
	assertApplicationTemplate(t, appTemplateInput, updateOutput)
	saveExample(t, updateAppTemplateRequest.Query(), "update application template")
}

func TestDeleteApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTemplate := pkg.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, pkg.FixApplicationTemplate(name))

	deleteApplicationTemplateRequest := pkg.FixDeleteApplicationTemplateRequest(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Delete application template")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteApplicationTemplateRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application template was deleted")

	out := pkg.GetApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTemplate.ID)

	require.Empty(t, out)
	saveExample(t, deleteApplicationTemplateRequest.Query(), "delete application template")
}

func TestQueryApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTemplate := pkg.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, pkg.FixApplicationTemplate(name))
	defer pkg.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTemplate.ID)

	getApplicationTemplateRequest := pkg.FixApplicationTemplateRequest(appTemplate.ID)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Get application template")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getApplicationTemplateRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	//THEN
	t.Log("Check if application template was received")
	assert.Equal(t, name, output.Name)
}

func TestQueryApplicationTemplates(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name1 := "app-template-1"
	name2 := "app-template-2"

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	t.Log("Create application templates")
	appTemplate1 := pkg.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, pkg.FixApplicationTemplate(name1))
	defer pkg.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTemplate1.ID)

	appTemplate2 := pkg.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, pkg.FixApplicationTemplate(name2))
	defer pkg.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTemplate2.ID)

	first := 100
	after := ""

	getApplicationTemplatesRequest := pkg.FixGetApplicationTemplatesWithPagination(first, after)
	output := graphql.ApplicationTemplatePage{}

	// WHEN
	t.Log("List application templates")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, getApplicationTemplatesRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application templates were received")
	assert.Subset(t, output.Data, []*graphql.ApplicationTemplate{&appTemplate1, &appTemplate2})
	saveExample(t, getApplicationTemplatesRequest.Query(), "query application templates")
}

func TestRegisterApplicationFromTemplate(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	tmplName := "template"
	placeholderKey := "new-placeholder"
	appTmplInput := pkg.FixApplicationTemplate(tmplName)
	appTmplInput.ApplicationInput.Description = ptr.String("test {{new-placeholder}}")
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        placeholderKey,
			Description: ptr.String("description"),
		}}

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appTmpl := pkg.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTmplInput)
	defer pkg.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant, appTmpl.ID)

	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: tmplName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: placeholderKey,
			Value:       "new-value",
		}}}
	appFromTmplGQL, err := pkg.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := pkg.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}
	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createAppFromTmplRequest, &outputApp)

	//THEN
	require.NoError(t, err)
	pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, outputApp.ID)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-value", *outputApp.Application.Description)
	saveExample(t, createAppFromTmplRequest.Query(), "register application from template")
}
