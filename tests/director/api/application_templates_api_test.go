package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestCreateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template-name"
	appTemplateInput := fixApplicationTemplate(name)
	appTemplate, err := tc.graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	createApplicationTemplateRequest := fixCreateApplicationTemplateRequest(appTemplate)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Create application template")
	err = tc.RunOperation(ctx, createApplicationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer deleteApplicationTemplate(t, ctx, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, createApplicationTemplateRequest.Query(), "create application template")

	t.Log("Check if application template was created")

	getApplicationTemplateRequest := fixApplicationTemplateRequest(output.ID)
	appTemplateOutput := graphql.ApplicationTemplate{}

	err = tc.RunOperation(ctx, getApplicationTemplateRequest, &appTemplateOutput)

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

	t.Log("Create application template")
	appTemplate := createApplicationTemplate(t, ctx, name)
	defer deleteApplicationTemplate(t, ctx, appTemplate.ID)

	appTemplateInput := graphql.ApplicationTemplateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateGQL, err := tc.graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
	updateAppTemplateRequest := fixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = tc.RunOperation(ctx, updateAppTemplateRequest, &updateOutput)
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

	t.Log("Create application template")
	appTemplate := createApplicationTemplate(t, ctx, name)

	deleteApplicationTemplateRequest := fixDeleteApplicationTemplate(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Delete application template")
	err := tc.RunOperation(ctx, deleteApplicationTemplateRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application template was deleted")

	out := getApplicationTemplate(t, ctx, appTemplate.ID)

	require.Empty(t, out)
	saveExample(t, deleteApplicationTemplateRequest.Query(), "delete application template")
}

func TestQueryApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"

	t.Log("Create application template")
	appTemplate := createApplicationTemplate(t, ctx, name)
	defer deleteApplicationTemplate(t, ctx, appTemplate.ID)

	getApplicationTemplateRequest := fixApplicationTemplateRequest(appTemplate.ID)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Get application template")
	err := tc.RunOperation(ctx, getApplicationTemplateRequest, &output)
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

	t.Log("Create application templates")
	appTemplate1 := createApplicationTemplate(t, ctx, name1)
	defer deleteApplicationTemplate(t, ctx, appTemplate1.ID)

	appTemplate2 := createApplicationTemplate(t, ctx, name2)
	defer deleteApplicationTemplate(t, ctx, appTemplate2.ID)

	first := 2
	after := ""

	getApplicationTemplatesRequest := fixApplicationTemplates(first, after)
	output := graphql.ApplicationTemplatePage{}

	// WHEN
	t.Log("List application templates")
	err := tc.RunOperation(ctx, getApplicationTemplatesRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application templates were received")
	assert.Equal(t, 2, output.TotalCount)
	saveExample(t, getApplicationTemplatesRequest.Query(), "query application templates")
}

func TestRegisterApplicationFromTemplate(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	tmplName := "template"
	placeholderKey := "new-placeholder"
	appTmplInput := fixApplicationTemplate(tmplName)
	appTmplInput.ApplicationInput.Description = ptr.String("test {{new-placeholder}}")
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        placeholderKey,
			Description: ptr.String("description"),
		}}

	appTmpl := createApplicationTemplateFromInput(t, ctx, appTmplInput)
	defer deleteApplicationTemplate(t, ctx, appTmpl.ID)

	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: tmplName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: placeholderKey,
			Value:       "new-value",
		}}}
	appFromTmplGQL, err := tc.graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}
	//WHEN
	err = tc.RunOperation(ctx, createAppFromTmplRequest, &outputApp)

	//THEN
	require.NoError(t, err)
	unregisterApplication(t, outputApp.ID)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-value", *outputApp.Application.Description)
	saveExample(t, createAppFromTmplRequest.Query(), "register application from template")
}
