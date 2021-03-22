package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertApplication(t *testing.T, in graphql.ApplicationRegisterInput, actualApp graphql.ApplicationExt) {
	require.NotEmpty(t, actualApp.ID)

	assert.Equal(t, in.Name, actualApp.Name)
	assert.Equal(t, in.Description, actualApp.Description)
	assertLabels(t, *in.Labels, actualApp.Labels, actualApp)
	assert.Equal(t, in.HealthCheckURL, actualApp.HealthCheckURL)
	assert.Equal(t, in.ProviderName, actualApp.ProviderName)
	assertWebhooks(t, in.Webhooks, actualApp.Webhooks)
	assertBundles(t, in.Bundles, actualApp.Bundles.Data)
}

//TODO: After fixing the 'Labels' scalar turn this back into regular assertion
func assertLabels(t *testing.T, in graphql.Labels, actual graphql.Labels, app graphql.ApplicationExt) {
	appNameNormalizier := normalizer.DefaultNormalizator{}
	for key, value := range actual {
		if key == "integrationSystemID" {
			if app.IntegrationSystemID == nil {
				continue
			}
			assert.Equal(t, value, app.IntegrationSystemID)
			continue
		} else if key == "name" {
			assert.Equal(t, value, appNameNormalizier.Normalize(app.Name))
			continue
		}
		assert.Equal(t, value, in[key])
	}
}

func assertWebhooks(t *testing.T, in []*graphql.WebhookInput, actual []graphql.Webhook) {
	assert.Equal(t, len(in), len(actual))
	for _, inWh := range in {
		found := false
		for _, actWh := range actual {
			if urlsAreIdentical(inWh.URL, actWh.URL) {
				found = true
				assert.NotNil(t, actWh.ID)
				assert.Equal(t, inWh.Type, actWh.Type)
				assertAuth(t, inWh.Auth, actWh.Auth)
			}
		}
		assert.True(t, found)
	}
}

func assertAuth(t *testing.T, in *graphql.AuthInput, actual *graphql.Auth) {
	if in == nil {
		assert.Nil(t, actual)
		return
	}
	require.NotNil(t, actual)
	assertHttpHeaders(t, in.AdditionalHeadersSerialized, actual.AdditionalHeaders)
	assertQueryParams(t, in.AdditionalQueryParamsSerialized, actual.AdditionalQueryParams)

	if in.Credential != nil {
		if in.Credential.Basic != nil {
			basic, ok := actual.Credential.(*graphql.BasicCredentialData)
			require.True(t, ok)
			assert.Equal(t, in.Credential.Basic.Username, basic.Username)
			assert.Equal(t, in.Credential.Basic.Password, basic.Password)

		} else if in.Credential.Oauth != nil {
			o, ok := actual.Credential.(*graphql.OAuthCredentialData)
			require.True(t, ok)
			assert.Equal(t, in.Credential.Oauth.URL, o.URL)
			assert.Equal(t, in.Credential.Oauth.ClientSecret, o.ClientSecret)
			assert.Equal(t, in.Credential.Oauth.ClientID, o.ClientID)
		}
	}

	if in.RequestAuth != nil && in.RequestAuth.Csrf != nil {
		require.NotNil(t, actual.RequestAuth)
		require.NotNil(t, actual.RequestAuth.Csrf)
		if in.RequestAuth.Csrf.Credential != nil {
			if in.RequestAuth.Csrf.Credential.Basic != nil {
				basic, ok := actual.RequestAuth.Csrf.Credential.(*graphql.BasicCredentialData)
				require.True(t, ok)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.Basic.Username, basic.Username)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.Basic.Password, basic.Password)

			} else if in.RequestAuth.Csrf.Credential.Oauth != nil {
				o, ok := actual.RequestAuth.Csrf.Credential.(*graphql.OAuthCredentialData)
				require.True(t, ok)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.Oauth.URL, o.URL)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.Oauth.ClientSecret, o.ClientSecret)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.Oauth.ClientID, o.ClientID)
			}
		}
		assert.Equal(t, in.RequestAuth.Csrf.AdditionalQueryParams, actual.RequestAuth.Csrf.AdditionalQueryParams)
		assert.Equal(t, in.RequestAuth.Csrf.AdditionalHeaders, actual.RequestAuth.Csrf.AdditionalHeaders)
		assert.Equal(t, in.RequestAuth.Csrf.TokenEndpointURL, actual.RequestAuth.Csrf.TokenEndpointURL)
	}
}

func assertDocuments(t *testing.T, in []*graphql.DocumentInput, actual []*graphql.DocumentExt) {
	assert.Equal(t, len(in), len(actual))
	for _, inDocu := range in {
		found := false
		for _, actDocu := range actual {
			if inDocu.Title != actDocu.Title {
				continue
			}
			found = true
			assert.Equal(t, inDocu.DisplayName, actDocu.DisplayName)
			assert.Equal(t, inDocu.Description, actDocu.Description)
			assert.Equal(t, inDocu.Data, actDocu.Data)
			assert.Equal(t, inDocu.Kind, actDocu.Kind)
			assert.Equal(t, inDocu.Format, actDocu.Format)
			assertFetchRequest(t, inDocu.FetchRequest, actDocu.FetchRequest)
		}
		assert.True(t, found)
	}
}

func assertBundles(t *testing.T, in []*graphql.BundleCreateInput, actual []*graphql.BundleExt) {
	assert.Equal(t, len(in), len(actual))
	for _, inBndl := range in {
		found := false
		for _, actBndl := range actual {
			if inBndl.Name != actBndl.Name {
				continue
			}
			found = true

			assertBundle(t, inBndl, actBndl)
		}
		assert.True(t, found)
	}
}

func assertFetchRequest(t *testing.T, in *graphql.FetchRequestInput, actual *graphql.FetchRequest) {
	if in == nil {
		assert.Nil(t, actual)
		return
	}
	assert.NotNil(t, actual)
	assert.Equal(t, in.URL, actual.URL)
	assert.Equal(t, in.Filter, actual.Filter)
	if in.Mode != nil {
		assert.Equal(t, *in.Mode, actual.Mode)
	} else {
		assert.Equal(t, graphql.FetchModeSingle, actual.Mode)
	}
	assertAuth(t, in.Auth, actual.Auth)
}

func assertAPI(t *testing.T, in []*graphql.APIDefinitionInput, actual []*graphql.APIDefinitionExt) {
	assert.Equal(t, len(in), len(actual))
	for _, inApi := range in {
		found := false
		for _, actApi := range actual {
			if inApi.Name != actApi.Name {
				continue
			}
			found = true
			assert.Equal(t, inApi.Description, actApi.Description)
			assert.Equal(t, inApi.TargetURL, actApi.TargetURL)
			assert.Equal(t, inApi.Group, actApi.Group)
			assertVersion(t, inApi.Version, actApi.Version)
			if inApi.Spec != nil {
				require.NotNil(t, actApi.Spec)
				assert.Equal(t, inApi.Spec.Data, actApi.Spec.Data)
				assert.Equal(t, inApi.Spec.Format, actApi.Spec.Format)
				assert.Equal(t, inApi.Spec.Type, actApi.Spec.Type)
				assertFetchRequest(t, inApi.Spec.FetchRequest, actApi.Spec.FetchRequest)
			} else {
				assert.Nil(t, actApi.Spec)
			}
		}
		assert.True(t, found)
	}
}

func assertVersion(t *testing.T, in *graphql.VersionInput, actual *graphql.Version) {
	if in != nil {
		assert.NotNil(t, actual)
		assert.Equal(t, in.Value, actual.Value)
		assert.Equal(t, in.DeprecatedSince, actual.DeprecatedSince)
		assert.Equal(t, in.Deprecated, actual.Deprecated)
		assert.Equal(t, in.ForRemoval, actual.ForRemoval)
	} else {
		assert.Nil(t, actual)
	}
}

func assertEventsAPI(t *testing.T, in []*graphql.EventDefinitionInput, actual []*graphql.EventAPIDefinitionExt) {
	assert.Equal(t, len(in), len(actual))
	for _, inEv := range in {
		found := false
		for _, actEv := range actual {
			if actEv.Name != inEv.Name {
				continue
			}
			found = true
			assert.Equal(t, inEv.Group, actEv.Group)
			assert.Equal(t, inEv.Description, actEv.Description)
			assertVersion(t, inEv.Version, actEv.Version)
			if inEv.Spec != nil {
				require.NotNil(t, actEv.Spec)
				assert.Equal(t, inEv.Spec.Data, actEv.Spec.Data)
				assert.Equal(t, inEv.Spec.Format, actEv.Spec.Format)
				assert.Equal(t, inEv.Spec.Type, actEv.Spec.Type)
				assertFetchRequest(t, inEv.Spec.FetchRequest, actEv.Spec.FetchRequest)
			} else {
				assert.Nil(t, actEv.Spec)
			}
		}
		assert.True(t, found)
	}
}

func assertRuntime(t *testing.T, in graphql.RuntimeInput, actualRuntime graphql.RuntimeExt) {
	assert.Equal(t, in.Name, actualRuntime.Name)
	assert.Equal(t, in.Description, actualRuntime.Description)
	assertRuntimeLabels(t, in.Labels, actualRuntime.Labels)
}

func assertRuntimeLabels(t *testing.T, inLabels *graphql.Labels, actualLabels graphql.Labels) {
	const (
		scenariosKey    = "scenarios"
		isNormalizedKey = "isNormalized"
	)

	if inLabels == nil {
		assertLabel(t, actualLabels, scenariosKey, []interface{}{"DEFAULT"})
		assertLabel(t, actualLabels, isNormalizedKey, "true")
		assert.Equal(t, 2, len(actualLabels))
		return
	}

	_, inHasScenarios := (*inLabels)[scenariosKey]
	if !inHasScenarios {
		assertLabel(t, actualLabels, scenariosKey, []interface{}{"DEFAULT"})
	}

	_, inHasShouldNomalizeKey := (*inLabels)[isNormalizedKey]
	if !inHasShouldNomalizeKey {
		assertLabel(t, actualLabels, isNormalizedKey, "true")
	}

	for labelKey, labelValues := range *inLabels {
		assertLabel(t, actualLabels, labelKey, labelValues)
	}
}

func assertLabel(t *testing.T, actualLabels graphql.Labels, key string, values interface{}) {
	labelValues, ok := actualLabels[key]
	assert.True(t, ok)
	assert.Equal(t, values, labelValues)
}

func assertIntegrationSystem(t *testing.T, in graphql.IntegrationSystemInput, actualIntegrationSystem graphql.IntegrationSystemExt) {
	assert.Equal(t, in.Name, actualIntegrationSystem.Name)
	assert.Equal(t, in.Description, actualIntegrationSystem.Description)
}

func assertApplicationTemplate(t *testing.T, in graphql.ApplicationTemplateInput, actualApplicationTemplate graphql.ApplicationTemplate) {
	assert.Equal(t, in.Name, actualApplicationTemplate.Name)
	assert.Equal(t, in.Description, actualApplicationTemplate.Description)

	gqlAppInput, err := tc.graphqlizer.ApplicationRegisterInputToGQL(*in.ApplicationInput)
	require.NoError(t, err)

	gqlAppInput = strings.Replace(gqlAppInput, "\t", "", -1)
	gqlAppInput = strings.Replace(gqlAppInput, "\n", "", -1)

	assert.Equal(t, gqlAppInput, actualApplicationTemplate.ApplicationInput)
	assertApplicationTemplatePlaceholder(t, in.Placeholders, actualApplicationTemplate.Placeholders)
	assert.Equal(t, in.AccessLevel, actualApplicationTemplate.AccessLevel)

	assertWebhooks(t, in.Webhooks, actualApplicationTemplate.Webhooks)
}

func assertUpdateApplicationTemplate(t *testing.T, in graphql.ApplicationTemplateUpdateInput, actualApplicationTemplate graphql.ApplicationTemplate) {
	assert.Equal(t, in.Name, actualApplicationTemplate.Name)
	assert.Equal(t, in.Description, actualApplicationTemplate.Description)

	gqlAppInput, err := tc.graphqlizer.ApplicationRegisterInputToGQL(*in.ApplicationInput)
	require.NoError(t, err)

	gqlAppInput = strings.Replace(gqlAppInput, "\t", "", -1)
	gqlAppInput = strings.Replace(gqlAppInput, "\n", "", -1)

	assert.Equal(t, gqlAppInput, actualApplicationTemplate.ApplicationInput)
	assertApplicationTemplatePlaceholder(t, in.Placeholders, actualApplicationTemplate.Placeholders)
	assert.Equal(t, in.AccessLevel, actualApplicationTemplate.AccessLevel)
}

func assertApplicationTemplatePlaceholder(t *testing.T, in []*graphql.PlaceholderDefinitionInput, actualPlaceholders []*graphql.PlaceholderDefinition) {
	for i, _ := range in {
		assert.Equal(t, in[i].Name, actualPlaceholders[i].Name)
		assert.Equal(t, in[i].Description, actualPlaceholders[i].Description)
	}
}

func assertBundle(t *testing.T, in *graphql.BundleCreateInput, actual *graphql.BundleExt) {
	assert.Equal(t, in.Name, actual.Name)
	assert.Equal(t, in.Description, actual.Description)
	assert.Equal(t, in.InstanceAuthRequestInputSchema, actual.InstanceAuthRequestInputSchema)

	assertAuth(t, in.DefaultInstanceAuth, actual.DefaultInstanceAuth)
	assertDocuments(t, in.Documents, actual.Documents.Data)
	assertAPI(t, in.APIDefinitions, actual.APIDefinitions.Data)
	assertEventsAPI(t, in.EventDefinitions, actual.EventDefinitions.Data)

	assertAuth(t, in.DefaultInstanceAuth, actual.DefaultInstanceAuth)
}

func assertBundleInstanceAuthInput(t *testing.T, expectedAuth graphql.BundleInstanceAuthRequestInput, actualAuth graphql.BundleInstanceAuth) {
	assertGraphQLJSON(t, expectedAuth.Context, actualAuth.Context)
	assertGraphQLJSON(t, expectedAuth.InputParams, actualAuth.InputParams)
}

func assertBundleInstanceAuth(t *testing.T, expectedAuth graphql.BundleInstanceAuth, actualAuth graphql.BundleInstanceAuth) {
	assert.Equal(t, expectedAuth.ID, actualAuth.ID)
	assert.Equal(t, expectedAuth.Context, actualAuth.Context)
	assert.Equal(t, expectedAuth.InputParams, actualAuth.InputParams)
}

func assertGraphQLJSON(t *testing.T, inExpected *graphql.JSON, inActual *graphql.JSON) {
	inExpectedStr, ok := unmarshalJSON(t, inExpected).(string)
	assert.True(t, ok)

	var expected map[string]interface{}
	err := json.Unmarshal([]byte(inExpectedStr), &expected)
	require.NoError(t, err)

	var actual map[string]interface{}
	err = json.Unmarshal([]byte(*inActual), &actual)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func assertGraphQLJSONSchema(t *testing.T, inExpected *graphql.JSONSchema, inActual *graphql.JSONSchema) {
	inExpectedStr, ok := unmarshalJSONSchema(t, inExpected).(string)
	assert.True(t, ok)

	var expected map[string]interface{}
	err := json.Unmarshal([]byte(inExpectedStr), &expected)
	require.NoError(t, err)

	var actual map[string]interface{}
	err = json.Unmarshal([]byte(*inActual), &actual)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func marshalJSON(t *testing.T, data interface{}) *graphql.JSON {
	out, err := json.Marshal(data)
	require.NoError(t, err)
	output := strconv.Quote(string(out))
	j := graphql.JSON(output)
	return &j
}

func unmarshalJSON(t *testing.T, j *graphql.JSON) interface{} {
	require.NotNil(t, j)
	var output interface{}
	err := json.Unmarshal([]byte(*j), &output)
	require.NoError(t, err)

	return output
}

func marshalJSONSchema(t *testing.T, schema interface{}) *graphql.JSONSchema {
	out, err := json.Marshal(schema)
	require.NoError(t, err)
	output := strconv.Quote(string(out))
	jsonSchema := graphql.JSONSchema(output)
	return &jsonSchema
}

func unmarshalJSONSchema(t *testing.T, schema *graphql.JSONSchema) interface{} {
	require.NotNil(t, schema)
	var output interface{}
	err := json.Unmarshal([]byte(*schema), &output)
	require.NoError(t, err)

	return output
}

func assertAutomaticScenarioAssignment(t *testing.T, expected graphql.AutomaticScenarioAssignmentSetInput, actual graphql.AutomaticScenarioAssignment) {
	assert.Equal(t, expected.ScenarioName, actual.ScenarioName)
	require.NotNil(t, actual.Selector)
	require.NotNil(t, expected.Selector)
	assert.Equal(t, expected.Selector.Value, actual.Selector.Value)
	assert.Equal(t, expected.Selector.Key, actual.Selector.Key)
}

func assertAutomaticScenarioAssignments(t *testing.T, expected []graphql.AutomaticScenarioAssignmentSetInput, actual []*graphql.AutomaticScenarioAssignment) {
	assert.Equal(t, len(expected), len(actual))
	for _, expectedAssignment := range expected {
		found := false
		for _, actualAssignment := range actual {
			require.NotNil(t, actualAssignment)
			if expectedAssignment.ScenarioName == actualAssignment.ScenarioName {
				found = true
				assertAutomaticScenarioAssignment(t, expectedAssignment, *actualAssignment)
				break
			}
		}
		assert.True(t, found, "Assignment for scenario: '%s' not found", expectedAssignment.ScenarioName)
	}
}

func assertIntegrationSystemNames(t *testing.T, expectedNames []string, actual graphql.IntegrationSystemPageExt) {
	for _, intSysName := range expectedNames {
		found := false
		require.NotEmpty(t, actual.Data)
		for _, intSys := range actual.Data {
			require.NotNil(t, intSys)
			if intSysName == intSys.Name {
				found = true
				break
			}
		}
		assert.True(t, found, "Integration system: '%s' not found", intSysName)
	}
}

func assertTenants(t *testing.T, in []*graphql.Tenant, actual []*graphql.Tenant) {
	for _, inTnt := range in {
		found := false
		for _, actTnt := range actual {
			if inTnt.ID != actTnt.ID {
				continue
			}
			found = true

			assert.Equal(t, inTnt.Name, actTnt.Name)

			if inTnt.Initialized != nil {
				assert.Equal(t, inTnt.Initialized, actTnt.Initialized)
			}
		}
		assert.True(t, found)
	}
}

func assertHttpHeaders(t *testing.T, in *graphql.HttpHeadersSerialized, actual *graphql.HttpHeaders) {
	if in == nil && actual == nil {
		return
	}

	require.NotNil(t, in)
	require.NotNil(t, actual)

	unquoted, err := strconv.Unquote(string(*in))
	require.NoError(t, err)

	var headersIn graphql.HttpHeaders
	err = json.Unmarshal([]byte(unquoted), &headersIn)
	require.NoError(t, err)

	require.Equal(t, &headersIn, actual)
}

func assertQueryParams(t *testing.T, in *graphql.QueryParamsSerialized, actual *graphql.QueryParams) {
	if in == nil && actual == nil {
		return
	}

	require.NotNil(t, in)
	require.NotNil(t, actual)

	unquoted, err := strconv.Unquote(string(*in))
	require.NoError(t, err)

	var queryParamsIn graphql.QueryParams
	err = json.Unmarshal([]byte(unquoted), &queryParamsIn)
	require.NoError(t, err)

	require.Equal(t, &queryParamsIn, actual)
}

func urlsAreIdentical(url1, url2 *string) bool {
	identical := url1 == url2
	if !identical {
		if url1 != nil && url2 != nil {
			identical = *url1 == *url2
		}
	}
	return identical
}
