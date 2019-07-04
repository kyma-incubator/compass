package director

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertApplication(t *testing.T, in graphql.ApplicationInput, actualApp ApplicationExt) {
	require.NotEmpty(t, actualApp.ID)

	assert.Equal(t, in.Name, actualApp.Name)
	assert.Equal(t, in.Description, actualApp.Description)
	if in.Annotations != nil {
		assert.Equal(t, *in.Annotations, actualApp.Annotations)
	} else {
		assert.Empty(t, actualApp.Annotations)
	}
	if in.Labels != nil {
		assert.Equal(t, *in.Labels, actualApp.Labels)
	} else {
		assert.Empty(t, actualApp.Labels)
	}
	assert.Equal(t, in.HealthCheckURL, actualApp.HealthCheckURL)
	assertWebhooks(t, in.Webhooks, actualApp.Webhooks)
	assertDocuments(t, in.Documents, actualApp.Documents.Data)
	assertAPI(t, in.Apis, actualApp.Apis.Data)
	assertEvents(t, in.EventAPIs, actualApp.EventAPIs.Data)

}

func assertWebhooks(t *testing.T, in []*graphql.ApplicationWebhookInput, actual []graphql.ApplicationWebhook) {
	assert.Equal(t, len(in), len(actual))
	for _, inWh := range in {
		found := false
		for _, actWh := range actual {
			if inWh.URL == actWh.URL {
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
	assert.Equal(t, in.AdditionalHeaders, actual.AdditionalHeaders)
	assert.Equal(t, in.AdditionalQueryParams, actual.AdditionalQueryParams)
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
				basic, ok := actual.Credential.(*graphql.BasicCredentialData)
				require.True(t, ok)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.Basic.Username, basic.Username)
				assert.Equal(t, in.RequestAuth.Csrf.Credential.Basic.Password, basic.Password)

			} else if in.RequestAuth.Csrf.Credential.Oauth != nil {
				o, ok := actual.Credential.(*graphql.OAuthCredentialData)
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

func assertDocuments(t *testing.T, in []*graphql.DocumentInput, actual []*graphql.Document) {
	assert.Equal(t, len(in), len(actual))
	for _, inDocu := range in {
		found := false
		for _, actDocu := range actual {
			if inDocu.Title != actDocu.Title {
				continue
			}
			found = true
			assert.Equal(t, inDocu.Data, actDocu.Data)
			assert.Equal(t, inDocu.Kind, actDocu.Kind)
			assert.Equal(t, inDocu.Format, actDocu.Format)
			assertFetchRequest(t, inDocu.FetchRequest, actDocu.FetchRequest)
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

func assertAPI(t *testing.T, in []*graphql.APIDefinitionInput, actual []*graphql.APIDefinition) {
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
			assertAuth(t, inApi.DefaultAuth, actApi.DefaultAuth)
			assertVersion(t, inApi.Version, actApi.Version)
			if inApi.Spec != nil {
				assert.Equal(t, inApi.Spec.Data, actApi.Spec.Data)
				require.NotNil(t, actApi.Spec.Format)
				assert.Equal(t, inApi.Spec.Format, *actApi.Spec.Format)
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

func assertEvents(t *testing.T, in []*graphql.EventAPIDefinitionInput, actual []*graphql.EventAPIDefinition) {
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
				assert.NotNil(t, actEv.Spec)
				assert.Equal(t, inEv.Spec.Data, actEv.Spec.Data)
				assert.Equal(t, inEv.Spec.EventSpecType, actEv.Spec.Type)
				assertFetchRequest(t, inEv.Spec.FetchRequest, actEv.Spec.FetchRequest)

			} else {
				assert.Nil(t, actEv.Spec)
			}

		}
		assert.True(t, found)
	}
}

func assertRuntime(t *testing.T, in graphql.RuntimeInput, actual graphql.Runtime) {
	assert.Equal(t, in.Name, actual.Name)
	assert.Equal(t, *in.Description, *actual.Description)
	assert.Equal(t, *in.Labels, actual.Labels)
	assert.Equal(t, *in.Annotations, actual.Annotations)
}
