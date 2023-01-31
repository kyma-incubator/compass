package assertions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	json2 "github.com/kyma-incubator/compass/tests/pkg/json"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func AssertApplication(t *testing.T, in graphql.ApplicationRegisterInput, actualApp graphql.ApplicationExt) {
	require.NotEmpty(t, actualApp.ID)

	assert.Equal(t, in.Name, actualApp.Name)
	assert.Equal(t, in.Description, actualApp.Description)
	AssertLabels(t, in.Labels, actualApp.Labels, actualApp)
	assert.Equal(t, in.HealthCheckURL, actualApp.HealthCheckURL)
	assert.Equal(t, in.ProviderName, actualApp.ProviderName)
	AssertWebhooks(t, in.Webhooks, actualApp.Webhooks)
	AssertBundles(t, in.Bundles, actualApp.Bundles.Data)
}

// TODO: After fixing the 'Labels' scalar turn this back into regular assertion
func AssertLabels(t *testing.T, in graphql.Labels, actual graphql.Labels, app graphql.ApplicationExt) {
	appNameNormalizier := normalizer.DefaultNormalizator{}

	if _, ok := in["managed"]; !ok {
		in["managed"] = "false"
	}

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

func AssertWebhooks(t *testing.T, in []*graphql.WebhookInput, actual []graphql.Webhook) {
	assert.Equal(t, len(in), len(actual))
	for _, inWh := range in {
		found := false
		for _, actWh := range actual {
			if urlsAreIdentical(inWh.URL, actWh.URL) {
				found = true
				assert.NotNil(t, actWh.ID)
				assert.Equal(t, inWh.Type, actWh.Type)
				AssertAuth(t, inWh.Auth, actWh.Auth)
			}
		}
		assert.True(t, found)
	}
}

func AssertAuth(t *testing.T, in *graphql.AuthInput, actual *graphql.Auth) {
	if in == nil {
		assert.Nil(t, actual)
		return
	}
	require.NotNil(t, actual)
	AssertHttpHeaders(t, in.AdditionalHeadersSerialized, &actual.AdditionalHeaders)
	AssertQueryParams(t, in.AdditionalQueryParamsSerialized, &actual.AdditionalQueryParams)

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
		} else if in.Credential.CertificateOAuth != nil {
			o, ok := actual.Credential.(*graphql.CertificateOAuthCredentialData)
			require.True(t, ok)
			assert.Equal(t, in.Credential.CertificateOAuth.URL, o.URL)
			assert.Equal(t, in.Credential.CertificateOAuth.Certificate, o.Certificate)
			assert.Equal(t, in.Credential.CertificateOAuth.ClientID, o.ClientID)
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
			} else if in.RequestAuth.Csrf.Credential.CertificateOAuth != nil {
				o, ok := actual.RequestAuth.Csrf.Credential.(*graphql.CertificateOAuthCredentialData)
				require.True(t, ok)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.CertificateOAuth.URL, o.URL)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.CertificateOAuth.Certificate, o.Certificate)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.CertificateOAuth.ClientID, o.ClientID)
			}
		}
		assert.Equal(t, in.RequestAuth.Csrf.AdditionalQueryParams, actual.RequestAuth.Csrf.AdditionalQueryParams)
		assert.Equal(t, in.RequestAuth.Csrf.AdditionalHeaders, actual.RequestAuth.Csrf.AdditionalHeaders)
		assert.Equal(t, in.RequestAuth.Csrf.TokenEndpointURL, actual.RequestAuth.Csrf.TokenEndpointURL)
	}
}

func AssertDocuments(t *testing.T, in []*graphql.DocumentInput, actual []*graphql.DocumentExt) {
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
			AssertFetchRequest(t, inDocu.FetchRequest, actDocu.FetchRequest)
		}
		assert.True(t, found)
	}
}

func AssertBundles(t *testing.T, in []*graphql.BundleCreateInput, actual []*graphql.BundleExt) {
	assert.Equal(t, len(in), len(actual))
	for _, inBndl := range in {
		found := false
		for _, actBndl := range actual {
			if inBndl.Name != actBndl.Name {
				continue
			}
			found = true

			AssertBundle(t, inBndl, actBndl)
		}
		assert.True(t, found)
	}
}

func AssertFetchRequest(t *testing.T, in *graphql.FetchRequestInput, actual *graphql.FetchRequest) {
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
	AssertAuth(t, in.Auth, actual.Auth)
}

func AssertAPI(t *testing.T, in []*graphql.APIDefinitionInput, actual []*graphql.APIDefinitionExt) {
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
			AssertVersion(t, inApi.Version, actApi.Version)
			if inApi.Spec != nil {
				require.NotNil(t, actApi.Spec)
				assert.Equal(t, inApi.Spec.Data, actApi.Spec.Data)
				assert.Equal(t, inApi.Spec.Format, actApi.Spec.Format)
				assert.Equal(t, inApi.Spec.Type, actApi.Spec.Type)
				AssertFetchRequest(t, inApi.Spec.FetchRequest, actApi.Spec.FetchRequest)
			} else {
				assert.Nil(t, actApi.Spec)
			}
		}
		assert.True(t, found)
	}
}

func AssertVersion(t *testing.T, in *graphql.VersionInput, actual *graphql.Version) {
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

func AssertEventsAPI(t *testing.T, in []*graphql.EventDefinitionInput, actual []*graphql.EventAPIDefinitionExt) {
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
			AssertVersion(t, inEv.Version, actEv.Version)
			if inEv.Spec != nil {
				require.NotNil(t, actEv.Spec)
				assert.Equal(t, inEv.Spec.Data, actEv.Spec.Data)
				assert.Equal(t, inEv.Spec.Format, actEv.Spec.Format)
				assert.Equal(t, inEv.Spec.Type, actEv.Spec.Type)
				AssertFetchRequest(t, inEv.Spec.FetchRequest, actEv.Spec.FetchRequest)
			} else {
				assert.Nil(t, actEv.Spec)
			}
		}
		assert.True(t, found)
	}
}

func AssertRuntime(t *testing.T, in graphql.RuntimeRegisterInput, actualRuntime graphql.RuntimeExt) {
	assert.Equal(t, in.Name, actualRuntime.Name)
	assert.Equal(t, in.Description, actualRuntime.Description)
	AssertWebhooks(t, in.Webhooks, actualRuntime.Webhooks)
	AssertRuntimeLabels(t, &in.Labels, actualRuntime.Labels)
}

func AssertUpdatedRuntime(t *testing.T, in graphql.RuntimeUpdateInput, actualRuntime graphql.RuntimeExt) {
	assert.Equal(t, in.Name, actualRuntime.Name)
	assert.Equal(t, in.Description, actualRuntime.Description)
	AssertRuntimeLabels(t, &in.Labels, actualRuntime.Labels)
}

func AssertRuntimePageContainOnlyIDs(t *testing.T, page graphql.RuntimePageExt, ids ...string) {
	require.Equal(t, len(ids), len(page.Data))

	for _, runtime := range page.Data {
		require.Contains(t, ids, runtime.ID)
	}
}

func AssertRuntimeLabels(t *testing.T, inLabels *graphql.Labels, actualLabels graphql.Labels) {
	const (
		isNormalizedKey = "isNormalized"
	)

	if inLabels == nil {
		AssertLabel(t, actualLabels, isNormalizedKey, "true")
		assert.Equal(t, 2, len(actualLabels))
		return
	}

	_, inHasShouldNomalizeKey := (*inLabels)[isNormalizedKey]
	if !inHasShouldNomalizeKey {
		AssertLabel(t, actualLabels, isNormalizedKey, "true")
	}

	for labelKey, labelValues := range *inLabels {
		AssertLabel(t, actualLabels, labelKey, labelValues)
	}
}

func AssertRuntimeContext(t *testing.T, in *graphql.RuntimeContextInput, actual *graphql.RuntimeContextExt) {
	assert.Equal(t, in.Key, actual.Key)
	assert.Equal(t, in.Value, actual.Value)
}

func AssertLabel(t *testing.T, actualLabels graphql.Labels, key string, values interface{}) {
	labelValues, ok := actualLabels[key]
	assert.True(t, ok)
	assert.Equal(t, values, labelValues)
}

func AssertIntegrationSystem(t *testing.T, in graphql.IntegrationSystemInput, actualIntegrationSystem graphql.IntegrationSystemExt) {
	assert.Equal(t, in.Name, actualIntegrationSystem.Name)
	assert.Equal(t, in.Description, actualIntegrationSystem.Description)
}

func AssertApplicationTemplate(t *testing.T, in graphql.ApplicationTemplateInput, actualApplicationTemplate graphql.ApplicationTemplate) {
	assert.Equal(t, in.Name, actualApplicationTemplate.Name)
	assert.Equal(t, in.Description, actualApplicationTemplate.Description)

	gqlAppInput, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(*in.ApplicationInput)
	require.NoError(t, err)

	gqlAppInput = strings.Replace(gqlAppInput, "\t", "", -1)
	gqlAppInput = strings.Replace(gqlAppInput, "\n", "", -1)

	assert.Equal(t, gqlAppInput, actualApplicationTemplate.ApplicationInput)
	AssertApplicationTemplatePlaceholder(t, in.Placeholders, actualApplicationTemplate.Placeholders)
	assert.Equal(t, in.AccessLevel, actualApplicationTemplate.AccessLevel)
	assert.Equal(t, in.Labels, actualApplicationTemplate.Labels)
	AssertWebhooks(t, in.Webhooks, actualApplicationTemplate.Webhooks)
}

func AssertUpdateApplicationTemplate(t *testing.T, in graphql.ApplicationTemplateUpdateInput, actualApplicationTemplate graphql.ApplicationTemplate) {
	assert.Equal(t, in.Name, actualApplicationTemplate.Name)
	assert.Equal(t, in.Description, actualApplicationTemplate.Description)

	gqlAppInput, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(*in.ApplicationInput)
	require.NoError(t, err)

	gqlAppInput = strings.Replace(gqlAppInput, "\t", "", -1)
	gqlAppInput = strings.Replace(gqlAppInput, "\n", "", -1)

	assert.Equal(t, gqlAppInput, actualApplicationTemplate.ApplicationInput)
	AssertApplicationTemplatePlaceholder(t, in.Placeholders, actualApplicationTemplate.Placeholders)
	assert.Equal(t, in.AccessLevel, actualApplicationTemplate.AccessLevel)
}

func AssertFormationTemplate(t *testing.T, in *graphql.FormationTemplateInput, actual *graphql.FormationTemplate) {
	assert.Equal(t, in.Name, actual.Name)
	assert.ElementsMatch(t, in.ApplicationTypes, actual.ApplicationTypes)
	assert.ElementsMatch(t, in.RuntimeTypes, actual.RuntimeTypes)
	assert.Equal(t, in.RuntimeTypeDisplayName, actual.RuntimeTypeDisplayName)
	assert.Equal(t, in.RuntimeArtifactKind, actual.RuntimeArtifactKind)
}

func AssertCertificateSubjectMapping(t *testing.T, in *graphql.CertificateSubjectMappingInput, actual *graphql.CertificateSubjectMapping) {
	require.Equal(t, in.Subject, actual.Subject)
	require.Equal(t, in.ConsumerType, actual.ConsumerType)
	require.Equal(t, in.InternalConsumerID, actual.InternalConsumerID)
	require.Equal(t, in.TenantAccessLevels, actual.TenantAccessLevels)
}

func AssertApplicationTemplatePlaceholder(t *testing.T, in []*graphql.PlaceholderDefinitionInput, actualPlaceholders []*graphql.PlaceholderDefinition) {
	for i := range in {
		assert.Equal(t, in[i].Name, actualPlaceholders[i].Name)
		assert.Equal(t, in[i].Description, actualPlaceholders[i].Description)
		assert.Equal(t, in[i].JSONPath, actualPlaceholders[i].JSONPath)
	}
}

func AssertBundle(t *testing.T, in *graphql.BundleCreateInput, actual *graphql.BundleExt) {
	assert.Equal(t, in.Name, actual.Name)
	assert.Equal(t, in.Description, actual.Description)
	assert.Equal(t, in.InstanceAuthRequestInputSchema, actual.InstanceAuthRequestInputSchema)

	AssertAuth(t, in.DefaultInstanceAuth, actual.DefaultInstanceAuth)
	AssertDocuments(t, in.Documents, actual.Documents.Data)
	AssertAPI(t, in.APIDefinitions, actual.APIDefinitions.Data)
	AssertEventsAPI(t, in.EventDefinitions, actual.EventDefinitions.Data)

	AssertAuth(t, in.DefaultInstanceAuth, actual.DefaultInstanceAuth)
}

func AssertBundleInstanceAuthInput(t *testing.T, expectedAuth graphql.BundleInstanceAuthRequestInput, actualAuth graphql.BundleInstanceAuth) {
	AssertGraphQLJSON(t, expectedAuth.Context, actualAuth.Context)
	AssertGraphQLJSON(t, expectedAuth.InputParams, actualAuth.InputParams)
}

func AssertBundleInstanceAuth(t *testing.T, expectedAuth graphql.BundleInstanceAuth, actualAuth graphql.BundleInstanceAuth) {
	assert.Equal(t, expectedAuth.ID, actualAuth.ID)
	assert.Equal(t, expectedAuth.Context, actualAuth.Context)
	assert.Equal(t, expectedAuth.InputParams, actualAuth.InputParams)
}

func AssertGraphQLJSON(t *testing.T, inExpected *graphql.JSON, inActual *graphql.JSON) {
	inExpectedStr, ok := json2.UnmarshalJSON(t, inExpected).(string)
	assert.True(t, ok)

	var expected map[string]interface{}
	err := json.Unmarshal([]byte(inExpectedStr), &expected)
	require.NoError(t, err)

	var actual map[string]interface{}
	err = json.Unmarshal([]byte(*inActual), &actual)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func AssertGraphQLJSONSchema(t *testing.T, inExpected *graphql.JSONSchema, inActual *graphql.JSONSchema) {
	inExpectedStr, ok := json2.UnmarshalJSONSchema(t, inExpected).(string)
	assert.True(t, ok)

	var expected map[string]interface{}
	err := json.Unmarshal([]byte(inExpectedStr), &expected)
	require.NoError(t, err)

	var actual map[string]interface{}
	err = json.Unmarshal([]byte(*inActual), &actual)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func AssertAutomaticScenarioAssignment(t *testing.T, expected graphql.AutomaticScenarioAssignmentSetInput, actual graphql.AutomaticScenarioAssignment) {
	assert.Equal(t, expected.ScenarioName, actual.ScenarioName)
	require.NotNil(t, actual.Selector)
	require.NotNil(t, expected.Selector)
	assert.Equal(t, expected.Selector.Value, actual.Selector.Value)
	assert.Equal(t, expected.Selector.Key, actual.Selector.Key)
}

func AssertAutomaticScenarioAssignments(t *testing.T, expected []graphql.AutomaticScenarioAssignmentSetInput, actual []*graphql.AutomaticScenarioAssignment) {
	assert.Equal(t, len(expected), len(actual))
	for _, expectedAssignment := range expected {
		found := false
		for _, actualAssignment := range actual {
			require.NotNil(t, actualAssignment)
			if expectedAssignment.ScenarioName == actualAssignment.ScenarioName {
				found = true
				AssertAutomaticScenarioAssignment(t, expectedAssignment, *actualAssignment)
				break
			}
		}
		assert.True(t, found, "Assignment for scenario: '%s' not found", expectedAssignment.ScenarioName)
	}
}

func AssertIntegrationSystemNames(t *testing.T, expectedNames []string, actual graphql.IntegrationSystemPageExt) {
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

func AssertTenants(t *testing.T, in []*graphql.Tenant, actual []*graphql.Tenant) {
	for _, inTnt := range in {
		found := false
		for _, actTnt := range actual {
			if inTnt.ID != actTnt.ID {
				continue
			}
			found = true

			assert.Equal(t, inTnt.Name, actTnt.Name)
		}
		assert.True(t, found)
	}
}

func AssertHttpHeaders(t *testing.T, in *graphql.HTTPHeadersSerialized, actual *graphql.HTTPHeaders) {
	if in == nil && actual == nil {
		return
	}

	require.NotNil(t, in)
	require.NotNil(t, actual)

	unquoted, err := strconv.Unquote(string(*in))
	require.NoError(t, err)

	var headersIn graphql.HTTPHeaders
	err = json.Unmarshal([]byte(unquoted), &headersIn)
	require.NoError(t, err)

	require.Equal(t, &headersIn, actual)
}

func AssertQueryParams(t *testing.T, in *graphql.QueryParamsSerialized, actual *graphql.QueryParams) {
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

func AssertRuntimeScenarios(t *testing.T, runtimes graphql.RuntimePageExt, expectedScenarios map[string][]interface{}) {
	for _, rtm := range runtimes.Data {
		expectedScenarios, found := expectedScenarios[rtm.ID]
		require.True(t, found)
		AssertScenarios(t, rtm.Labels, expectedScenarios)
	}
}

func AssertScenarios(t *testing.T, actual graphql.Labels, expected []interface{}) {
	val := actual["scenarios"]
	scenarios, ok := val.([]interface{})
	if !ok {
		scenarios = []interface{}{}
	}
	assert.ElementsMatch(t, scenarios, expected)
}

func AssertSpecInBundleNotNil(t *testing.T, bndl graphql.BundleExt) {
	assert.True(t, len(bndl.APIDefinitions.Data) > 0)
	assert.NotNil(t, bndl.APIDefinitions.Data[0])
	assert.NotNil(t, bndl.APIDefinitions.Data[0].Spec)
	assert.NotNil(t, bndl.APIDefinitions.Data[0].Spec.Data)
}

func AssertSpecsFromORDService(t *testing.T, respBody string, expectedNumberOfAPIs int, apiSpecsMap map[string]int) []gjson.Result {
	var specs []gjson.Result

	for i := 0; i < expectedNumberOfAPIs; i++ {
		crrSpecs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", i)).Array()
		require.Equal(t, apiSpecsMap[gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()], len(crrSpecs))
		specs = append(specs, crrSpecs...)
	}
	return specs
}

func AssertSingleEntityFromORDService(t *testing.T, respBody string, expectedNumber int, expectedName, expectedDescription, descriptionField string) {
	numberOfEntities := len(gjson.Get(respBody, "value").Array())
	require.Equal(t, expectedNumber, numberOfEntities)

	for i := 0; i < numberOfEntities; i++ {
		require.Equal(t, expectedName, gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String())
		require.Equal(t, expectedDescription, gjson.Get(respBody, fmt.Sprintf("value.%d.%s", i, descriptionField)).String())
	}
}

func AssertMultipleEntitiesFromORDService(t *testing.T, respBody string, entitiesMap map[string]string, expectedNumber int, descriptionField string) {
	numberOfEntities := len(gjson.Get(respBody, "value").Array())
	require.Equal(t, expectedNumber, numberOfEntities)

	for i := 0; i < numberOfEntities; i++ {
		entityTitle := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
		require.NotEmpty(t, entityTitle)

		entityDescription, exists := entitiesMap[entityTitle]
		require.True(t, exists)

		require.Equal(t, entityDescription, gjson.Get(respBody, fmt.Sprintf("value.%d.%s", i, descriptionField)).String())
	}
}

func AssertProducts(t *testing.T, respBody string, expectedProductsMap map[string]string, expectedNumber int, descriptionField string) {
	numberOfProducts := len(gjson.Get(respBody, "value").Array())
	require.Equal(t, expectedNumber, numberOfProducts)

	actualProductsMap := make(map[string]string, expectedNumber)
	for i := 0; i < expectedNumber; i++ {
		entityTitle := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
		require.NotEmpty(t, entityTitle)

		actualProductsMap[entityTitle] = gjson.Get(respBody, fmt.Sprintf("value.%d.%s", i, descriptionField)).String()
	}

	for expectedTitle, expectedDescription := range expectedProductsMap {
		entityDescription, exists := actualProductsMap[expectedTitle]
		require.True(t, exists)
		require.Equal(t, expectedDescription, entityDescription)
	}
}

func AssertDocumentationLabels(t *testing.T, respBody string, expectedLabelKey string, possibleValues []string, expectedNumber int) {
	numberOfEntities := len(gjson.Get(respBody, "value").Array())
	require.Equal(t, expectedNumber, numberOfEntities)

	for i := 0; i < numberOfEntities; i++ {
		documentationLabels := gjson.Get(respBody, fmt.Sprintf("value.%d.documentationLabels", i)).Array()
		for _, label := range documentationLabels {
			key := gjson.Get(label.String(), "key").String()
			value := gjson.Get(label.String(), "value").String()
			require.Equal(t, expectedLabelKey, key)
			assert.Contains(t, possibleValues, value)
		}
	}
}

func AssertBundleCorrelationIds(t *testing.T, respBody string, entitiesMap map[string][]string, expectedNumber int) {
	numberOfEntities := len(gjson.Get(respBody, "value").Array())
	require.Equal(t, expectedNumber, numberOfEntities)

	for i := 0; i < numberOfEntities; i++ {
		entityTitle := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
		require.NotEmpty(t, entityTitle)

		correlationIds := gjson.Get(respBody, fmt.Sprintf("value.%d.correlationIds", i)).Array()
		bundleCorrelationIdsAsString := make([]string, 0)
		for _, id := range correlationIds {
			idAsString := gjson.Get(id.String(), "value").String()
			bundleCorrelationIdsAsString = append(bundleCorrelationIdsAsString, idAsString)
		}
		expectedCorrelationIds := entitiesMap[entityTitle]
		assert.ElementsMatch(t, expectedCorrelationIds, bundleCorrelationIdsAsString)
	}
}

func AssertDefaultBundleID(t *testing.T, respBody string, numberOfEntities int, entityDefaultBundleMap, ordAndInternalIDsMappingForBundles map[string]string) {
	for i := 0; i < numberOfEntities; i++ {
		entityTitle := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
		partOfConsumptionBundles := gjson.Get(respBody, fmt.Sprintf("value.%d.partOfConsumptionBundles", i)).Array()

		for j := 0; j < len(partOfConsumptionBundles); j++ {
			internalBundleID := gjson.Get(respBody, fmt.Sprintf("value.%d.partOfConsumptionBundles.%d.bundleID", i, j)).String()
			ordID := ordAndInternalIDsMappingForBundles[internalBundleID]

			expectedDefaultBundleOrdIDRegex, ok := entityDefaultBundleMap[entityTitle]
			if ok && regexp.MustCompile(expectedDefaultBundleOrdIDRegex).MatchString(ordID) {
				require.True(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfConsumptionBundles.%d.isDefaultBundle", i, j)).Bool())
			} else {
				require.False(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfConsumptionBundles.%d.isDefaultBundle", i, j)).Bool())
			}
		}
	}
}

func AssertRelationBetweenBundleAndEntityFromORDService(t *testing.T, respBody string, entityType string, numberOfEntitiesForBundle map[string]int, entitiesDataForBundle map[string][]string) {
	numberOfBundles := len(gjson.Get(respBody, "value").Array())

	for i := 0; i < numberOfBundles; i++ {
		bundleTitle := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
		numberOfEntities := len(gjson.Get(respBody, fmt.Sprintf("value.%d.%s", i, entityType)).Array())
		require.NotEmpty(t, bundleTitle)
		require.Equal(t, numberOfEntitiesForBundle[bundleTitle], numberOfEntities)

		for j := 0; j < numberOfEntities; j++ {
			entityTitle := gjson.Get(respBody, fmt.Sprintf("value.%d.%s.%d.title", i, entityType, j)).String()
			require.NotEmpty(t, entityTitle)
			require.Contains(t, entitiesDataForBundle[bundleTitle], entityTitle)
		}
	}
}

func AssertTombstoneFromORDService(t *testing.T, respBody string, expectedNumber int, expectedIDRegex string) {
	numberOfEntities := len(gjson.Get(respBody, "value").Array())
	require.Equal(t, expectedNumber, numberOfEntities)

	for i := 0; i < numberOfEntities; i++ {
		require.Regexp(t, expectedIDRegex, gjson.Get(respBody, fmt.Sprintf("value.%d.ordId", i)).String())
	}
}

func AssertVendorFromORDService(t *testing.T, respBody string, expectedNumber int, expectedNumberCreatedByTest int, expectedTitle string) {
	numberOfEntities := len(gjson.Get(respBody, "value").Array())
	require.Equal(t, expectedNumber, numberOfEntities)

	vendorsFromTestsFound := 0
	for i := 0; i < numberOfEntities; i++ {
		if expectedTitle == gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String() {
			vendorsFromTestsFound++
		}
	}
	// LessOrEqual is needed as there may be vendors in the DB which are not created by the test, and their titles may be equal to "expectedTitle" or they can defer from it.
	assert.LessOrEqual(t, expectedNumberCreatedByTest, vendorsFromTestsFound)
}

func AssertApplicationPageContainOnlyIDs(t *testing.T, page graphql.ApplicationPage, ids ...string) {
	require.Equal(t, len(ids), len(page.Data))

	for _, app := range page.Data {
		require.Contains(t, ids, app.ID)
	}
}

func AssertNoErrorForOtherThanNotFound(t require.TestingT, err error) {
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		require.NoError(t, err)
	}
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
