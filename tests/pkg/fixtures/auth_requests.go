package fixtures

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/stretchr/testify/require"
)

func FixBasicAuth(t *testing.T) *graphql.AuthInput {
	additionalHeaders, err := graphql.NewHttpHeadersSerialized(map[string][]string{
		"header-A": []string{"ha1", "ha2"},
		"header-B": []string{"hb1", "hb2"},
	})
	require.NoError(t, err)

	additionalQueryParams, err := graphql.NewQueryParamsSerialized(map[string][]string{
		"qA": []string{"qa1", "qa2"},
		"qB": []string{"qb1", "qb2"},
	})
	require.NoError(t, err)

	return &graphql.AuthInput{
		Credential:                      FixBasicCredential(),
		AdditionalHeadersSerialized:     &additionalHeaders,
		AdditionalQueryParamsSerialized: &additionalQueryParams,
	}
}

func FixOauthAuth() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: FixOAuthCredential(),
	}
}

func FixBasicCredential() *graphql.CredentialDataInput {
	return &graphql.CredentialDataInput{
		Basic: &graphql.BasicCredentialDataInput{
			Username: "admin",
			Password: "secret",
		}}
}

func FixOAuthCredential() *graphql.CredentialDataInput {
	return &graphql.CredentialDataInput{
		Oauth: &graphql.OAuthCredentialDataInput{
			URL:          "url.net",
			ClientSecret: "grazynasecret",
			ClientID:     "clientid",
		}}
}

func FixDeprecatedVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:           "v1",
		Deprecated:      ptr.Bool(true),
		ForRemoval:      ptr.Bool(false),
		DeprecatedSince: ptr.String("v5"),
	}
}

func FixDecommissionedVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:      "v1",
		Deprecated: ptr.Bool(true),
		ForRemoval: ptr.Bool(true),
	}
}

func FixActiveVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:      "v2",
		Deprecated: ptr.Bool(false),
		ForRemoval: ptr.Bool(false),
	}
}
