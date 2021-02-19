/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"

	"github.com/stretchr/testify/require"
)


//Application
func GetApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.ApplicationExt {
	appRequest := FixGetApplicationRequest(id)
	app := graphql.ApplicationExt{}

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, appRequest, &app)
	require.NoError(t, err)
	return app
}

func UpdateApplicationWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string, in graphql.ApplicationUpdateInput) (graphql.ApplicationExt, error) {
	appInputGQL, err := Tc.Graphqlizer.ApplicationUpdateInputToGQL(in)
	require.NoError(t, err)

	createRequest := FixUpdateApplicationRequest(id, appInputGQL)
	app := graphql.ApplicationExt{}
	err = Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, createRequest, &app)
	return app, err
}

func RegisterApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, name, tenant string) graphql.ApplicationExt {
	in := FixSampleApplicationRegisterInputWithName("first", name)
	return RegisterApplicationFromInputWithinTenant(t, ctx, gqlClient, tenant, in)
}

func RegisterApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID string, in graphql.ApplicationRegisterInput) graphql.ApplicationExt {
	appInputGQL, err := Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := FixRegisterApplicationRequest(appInputGQL)

	app := graphql.ApplicationExt{}

	require.NoError(t, Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createRequest, &app))
	require.NotEmpty(t, app.ID)
	return app
}

func RequestClientCredentialsForApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.SystemAuth {
	req := FixRequestClientCredentialsForApplication(id)
	systemAuth := graphql.SystemAuth{}

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	return systemAuth
}

func UnregisterApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.ApplicationExt {
	deleteRequest := FixUnregisterApplicationRequest(id)
	app := graphql.ApplicationExt{}

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteRequest, &app)
	require.NoError(t, err)
	return app
}

func DeleteApplicationLabel(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id, labelKey string) {
	deleteRequest := FixDeleteApplicationLabelRequest(id, labelKey)

	require.NoError(t, Tc.RunOperation(ctx, gqlClient, deleteRequest, nil))
}

func SetApplicationLabel(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string, labelKey string, labelValue interface{}) graphql.Label {
	setLabelRequest := FixSetApplicationLabelRequest(id, labelKey, labelValue)
	label := graphql.Label{}

	err := Tc.RunOperation(ctx, gqlClient, setLabelRequest, &label)
	require.NoError(t, err)

	return label
}

func GenerateClientCredentialsForApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client,id string) graphql.SystemAuth {
	req := FixRequestClientCredentialsForApplication(id)

	out := graphql.SystemAuth{}
	err := Tc.RunOperation(ctx, gqlClient, req, &out)
	require.NoError(t, err)

	return out
}

func DeleteSystemAuthForApplication(t *testing.T, ctx context.Context,gqlClient *gcli.Client, id string) {
	req := FixDeleteSystemAuthForApplicationRequest(id)
	err := Tc.RunOperation(ctx,gqlClient, req, nil)
	require.NoError(t, err)
}

func SetDefaultEventingForApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, appID string, runtimeID string) {
	req := FixSetDefaultEventingForApplication(appID, runtimeID)
	err := Tc.RunOperation(ctx,gqlClient, req, nil)
	require.NoError(t, err)
}

func RegisterSimpleApp(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID string) func() {
	placeholder := "foo"
	in := FixSampleApplicationRegisterInputWithWebhooks(placeholder)
	appInputGQL, err := Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	var res graphql.Application
	req := FixRegisterApplicationRequest(appInputGQL)
	err = Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, req, &res)
	require.NoError(t, err)

	return func() {
		UnregisterApplication(t,ctx, gqlClient, tenantID, res.ID)
	}
}

// Runtime
func RegisterRuntimeFromInputWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, input *graphql.RuntimeInput) graphql.RuntimeExt {
	inputGQL, err := Tc.Graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	registerRuntimeRequest := FixRegisterRuntimeRequest(inputGQL)
	var runtime graphql.RuntimeExt

	err = Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, registerRuntimeRequest, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)
	return runtime
}

func RequestClientCredentialsForRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.SystemAuth {
	req := FixRequestClientCredentialsForRuntime(id)
	systemAuth := graphql.SystemAuth{}

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	return systemAuth
}

func UnregisterRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	delReq := FixUnregisterRuntimeRequest(id)

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, delReq, nil)
	require.NoError(t, err)
}

func GetRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.RuntimeExt {
	req := FixGetRuntimeRequest(id)
	runtime := graphql.RuntimeExt{}

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &runtime)
	require.NoError(t, err)
	return runtime
}

func ListRuntimes(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string) graphql.RuntimePageExt {
	runtimesPage := graphql.RuntimePageExt{}
	queryReq := FixGetRuntimesRequestWithPagination()
	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, queryReq, &runtimesPage)
	require.NoError(t, err)
	return runtimesPage
}

func SetRuntimeLabel(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, runtimeID string, labelKey string, labelValue interface{}) *graphql.Label {
	setLabelRequest := FixSetRuntimeLabelRequest(runtimeID, labelKey, labelValue)
	label := graphql.Label{}

	err := Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenant, setLabelRequest, &label)
	require.NoError(t, err)

	return &label
}

func DeleteSystemAuthForRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client,id string) {
	req := FixDeleteSystemAuthForRuntimeRequest(id)
	err := Tc.RunOperation(ctx,gqlClient, req, nil)
	require.NoError(t, err)
}

//Bundle
func CreateBundleWithInput(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, appID string, input graphql.BundleCreateInput) graphql.BundleExt {
	in, err := Tc.Graphqlizer.BundleCreateInputToGQL(input)
	require.NoError(t, err)

	req := FixAddBundleRequest(appID, in)
	var resp graphql.BundleExt

	err = Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &resp)
	require.NoError(t, err)

	return resp
}

func CreateBundle(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, appID, bndlName string) graphql.BundleExt {
	in, err := Tc.Graphqlizer.BundleCreateInputToGQL(FixBundleCreateInput(bndlName))
	require.NoError(t, err)

	req := FixAddBundleRequest(appID, in)
	var resp graphql.BundleExt

	err = Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &resp)
	require.NoError(t, err)

	return resp
}

func DeleteBundle(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := FixDeleteBundleRequest(id)

	require.NoError(t, Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil))
}

func GetBundle(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, appID, bndlID string) graphql.BundleExt {
	req := FixBundleRequest(appID, bndlID)
	bndl := graphql.ApplicationExt{}
	require.NoError(t, Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &bndl))
	return bndl.Bundle
}

func AddAPIToBundleWithInput(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, bndlID string, input graphql.APIDefinitionInput) graphql.APIDefinitionExt {
	inStr, err := Tc.Graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := FixAddAPIToBundleRequest(bndlID, inStr)
	err = Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &actualApi)
	require.NoError(t, err)
	return actualApi
}

func AddAPIToBundle(t *testing.T, ctx context.Context, gqlClient *gcli.Client, bndlID string) graphql.APIDefinitionExt {
	return AddAPIToBundleWithInput(t, ctx,gqlClient, TestTenants.GetDefaultTenantID(), bndlID, FixAPIDefinitionInput())
}

func AddEventToBundleWithInput(t *testing.T, ctx context.Context, gqlClient *gcli.Client, bndlID string, input graphql.EventDefinitionInput) graphql.EventDefinition {
	inStr, err := Tc.Graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	event := graphql.EventDefinition{}
	req := FixAddEventAPIToBundleRequest(bndlID, inStr)
	err = Tc.RunOperation(ctx,gqlClient, req, &event)
	require.NoError(t, err)
	return event
}

func AddEventToBundle(t *testing.T, ctx context.Context, gqlClient *gcli.Client, bndlID string) graphql.EventDefinition {
	return AddEventToBundleWithInput(t, ctx,gqlClient, bndlID, FixEventAPIDefinitionInput())
}

func AddDocumentToBundleWithInput(t *testing.T, ctx context.Context, gqlClient *gcli.Client, bndlID string, input graphql.DocumentInput) graphql.DocumentExt {
	inStr, err := Tc.Graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDoc := graphql.DocumentExt{}
	req := FixAddDocumentToBundleRequest(bndlID, inStr)
	err = Tc.RunOperation(ctx,gqlClient, req, &actualDoc)
	require.NoError(t, err)
	return actualDoc
}

func AddDocumentToBundle(t *testing.T, ctx context.Context, gqlClient *gcli.Client, bndlID string) graphql.DocumentExt {
	return AddDocumentToBundleWithInput(t, ctx,gqlClient, bndlID, FixDocumentInput(t))
}

func CreateBundleInstanceAuth(t *testing.T, ctx context.Context, gqlClient *gcli.Client, bndlID string) graphql.BundleInstanceAuth {
	authCtx, inputParams := FixBundleInstanceAuthContextAndInputParams(t)
	in, err := Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(FixBundleInstanceAuthRequestInput(authCtx, inputParams))
	require.NoError(t, err)

	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestBundleInstanceAuthCreation(bundleID: "%s", in: %s) {
				id
			}}`, bndlID, in))

	var resp graphql.BundleInstanceAuth

	err = Tc.RunOperation(ctx,gqlClient, req, &resp)
	require.NoError(t, err)

	return resp
}

// Integration System
func GetIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.IntegrationSystemExt {
	intSysRequest := FixGetIntegrationSystemRequest(id)
	intSys := graphql.IntegrationSystemExt{}
	require.NoError(t, Tc.RunOperation(ctx,gqlClient, intSysRequest, &intSys))
	return &intSys
}

func RegisterIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant , name string) *graphql.IntegrationSystemExt {
	input := graphql.IntegrationSystemInput{Name: name}
	in, err := Tc.Graphqlizer.IntegrationSystemInputToGQL(input)
	if err != nil {
		return nil
	}

	req := FixRegisterIntegrationSystemRequest(in)
	intSys := &graphql.IntegrationSystemExt{}

	err = Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys)
	return intSys
}

func UnregisterIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := FixUnregisterIntegrationSystem(id)
	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.NoError(t, err)
}

func UnregisterIntegrationSystemWithErr(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := FixUnregisterIntegrationSystem(id)
	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "The record cannot be deleted because another record refers to it")
}

func GetSystemAuthsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) []*graphql.SystemAuth {
	req := FixGetIntegrationSystemRequest(id)
	intSys := graphql.IntegrationSystemExt{}
	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &intSys)
	require.NoError(t, err)
	return intSys.Auths
}

func RequestClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) *graphql.SystemAuth {
	req := FixRequestClientCredentialsForIntegrationSystem(id)
	systemAuth := graphql.SystemAuth{}

	// WHEN
	t.Log("Generate client credentials for integration system")
	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	require.NotEmpty(t, systemAuth.Auth)

	t.Log("Check if client credentials were generated")
	assert.NotEmpty(t, systemAuth.Auth.Credential)
	intSysOauthCredentialData, ok := systemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
	require.NotEmpty(t, intSysOauthCredentialData.ClientID)
	assert.Equal(t, systemAuth.ID, intSysOauthCredentialData.ClientID)
	return &systemAuth
}

func GenerateOneTimeTokenForApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.OneTimeTokenForApplicationExt {
	req := FixRequestOneTimeTokenForApplication(id)
	oneTimeToken := graphql.OneTimeTokenForApplicationExt{}

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &oneTimeToken)
	require.NoError(t, err)

	require.NotEmpty(t, oneTimeToken.ConnectorURL)
	require.NotEmpty(t, oneTimeToken.Token)
	require.NotEmpty(t, oneTimeToken.Raw)
	require.NotEmpty(t, oneTimeToken.RawEncoded)
	require.NotEmpty(t, oneTimeToken.LegacyConnectorURL)
	return oneTimeToken
}

func DeleteSystemAuthForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) {
	req := FixDeleteSystemAuthForIntegrationSystemRequest(id)
	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, "", req, nil)
	require.NoError(t, err)
}

//Application Template
func CreateApplicationTemplate(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, input graphql.ApplicationTemplateInput) graphql.ApplicationTemplate {
	appTemplate, err := Tc.Graphqlizer.ApplicationTemplateInputToGQL(input)
	require.NoError(t, err)

	req := FixCreateApplicationTemplateRequest(appTemplate)
	appTpl := graphql.ApplicationTemplate{}
	err = Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &appTpl)
	require.NoError(t, err)
	return appTpl
}

func GetApplicationTemplate(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.ApplicationTemplate {
	req := FixApplicationTemplateRequest(id)
	appTpl := graphql.ApplicationTemplate{}

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &appTpl)
	require.NoError(t, err)
	return appTpl
}

func DeleteApplicationTemplate(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := FixDeleteApplicationTemplateRequest(id)

	err := Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.NoError(t, err)
}

// Label Definitions
func CreateLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, key string, schema interface{}, tenantID string) *graphql.LabelDefinition {
	input := graphql.LabelDefinitionInput{
		Key:    key,
		Schema: MarshalJSONSchema(t, schema),
	}

	in, err := Tc.Graphqlizer.LabelDefinitionInputToGQL(input)
	if err != nil {
		return nil
	}

	createRequest := FixCreateLabelDefinitionRequest(in)

	output := graphql.LabelDefinition{}
	err = Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, createRequest, &output)
	require.NoError(t, err)

	return &output
}

func CreateScenariosLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID string, scenarios []string) *graphql.LabelDefinition {
	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": scenarios,
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	return CreateLabelDefinitionWithinTenant(t, ctx,gqlClient, "scenarios", jsonSchema, tenantID)
}

func UpdateLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, key string, schema interface{}, tenantID string) *graphql.LabelDefinition {
	input := graphql.LabelDefinitionInput{
		Key:    key,
		Schema: MarshalJSONSchema(t, schema),
	}

	in, err := Tc.Graphqlizer.LabelDefinitionInputToGQL(input)
	if err != nil {
		return nil
	}

	updateRequest := FixUpdateLabelDefinitionRequest(in)

	output := graphql.LabelDefinition{}
	err = Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, updateRequest, &output)
	require.NoError(t, err)

	return &output
}

func UpdateScenariosLabelDefinitionWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID string, scenarios []string) *graphql.LabelDefinition {
	jsonSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"enum": scenarios,
			"type": "string",
		},
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
	}

	return UpdateLabelDefinitionWithinTenant(t, ctx,gqlClient, "scenarios", jsonSchema, tenantID)
}

func DeleteLabelDefinition(t *testing.T, ctx context.Context, gqlClient *gcli.Client, labelDefinitionKey string, deleteRelatedResources bool, tenantID string) {
	deleteRequest := FixDeleteLabelDefinitionRequest(labelDefinitionKey, deleteRelatedResources)

	require.NoError(t, Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, deleteRequest, nil))
}

func ListLabelDefinitionsWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID string) ([]*graphql.LabelDefinition, error) {
	labelDefinitionsRequest := FixLabelDefinitionsRequest()

	var labelDefinitions []*graphql.LabelDefinition

	err := Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, labelDefinitionsRequest, &labelDefinitions)
	return labelDefinitions, err
}


//OneTimeToken

func RequestOneTimeTokenForApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) graphql.OneTimeTokenForApplicationExt {
	tokenRequest := FixRequestOneTimeTokenForApplication(id)
	token := graphql.OneTimeTokenForApplicationExt{}
	err := Tc.RunOperation(ctx,gqlClient, tokenRequest, &token)
	require.NoError(t, err)
	return token
}

func RequestOneTimeTokenForRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) graphql.OneTimeTokenForRuntimeExt {
	tokenRequest := FixRequestOneTimeTokenForRuntime(id)
	token := graphql.OneTimeTokenForRuntimeExt{}
	err := Tc.RunOperation(ctx,gqlClient, tokenRequest, &token)
	require.NoError(t, err)
	return token
}

func CreateAutomaticScenarioAssignmentInTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.AutomaticScenarioAssignmentSetInput, tenantID string) *graphql.AutomaticScenarioAssignment {
	assignmentInput, err := Tc.Graphqlizer.AutomaticScenarioAssignmentSetInputToGQL(in)
	require.NoError(t, err)

	createRequest := FixCreateAutomaticScenarioAssignmentRequest(assignmentInput)

	assignment := graphql.AutomaticScenarioAssignment{}

	require.NoError(t, Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, createRequest, &assignment))
	require.NotEmpty(t, assignment.ScenarioName)
	return &assignment
}

func ListAutomaticScenarioAssignmentsWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID string) graphql.AutomaticScenarioAssignmentPage {
	assignmentsPage := graphql.AutomaticScenarioAssignmentPage{}
	req := FixAutomaticScenarioAssignmentsRequest()
	err := Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, req, &assignmentsPage)
	require.NoError(t, err)
	return assignmentsPage
}

func DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID, scenarioName string) graphql.AutomaticScenarioAssignment {
	assignment := graphql.AutomaticScenarioAssignment{}
	req := FixDeleteAutomaticScenarioAssignmentForScenarioRequest(scenarioName)
	err := Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, req, &assignment)
	require.NoError(t, err)
	return assignment
}

func DeleteAutomaticScenarioAssigmentForSelector(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID string, selector graphql.LabelSelectorInput) []graphql.AutomaticScenarioAssignment {
	paylaod, err := Tc.Graphqlizer.LabelSelectorInputToGQL(selector)
	require.NoError(t, err)
	req := FixDeleteAutomaticScenarioAssignmentsForSelectorRequest(paylaod)

	assignment := []graphql.AutomaticScenarioAssignment{}
	err = Tc.RunOperationWithCustomTenant(ctx,gqlClient, tenantID, req, &assignment)
	require.NoError(t, err)
	return assignment
}
