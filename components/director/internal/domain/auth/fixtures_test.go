package auth_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var (
	authUsername = "user"
	authPassword = "password"
	authEndpoint = "url"
	authMap      = map[string][]string{
		"foo": {"bar", "baz"},
		"too": {"tar", "taz"},
	}
	authMapSerialized     = "{\"foo\":[\"bar\",\"baz\"],\"too\":[\"tar\",\"taz\"]}"
	authHeaders           = graphql.HTTPHeaders(authMap)
	authHeadersSerialized = graphql.HTTPHeadersSerialized(authMapSerialized)
	authParams            = graphql.QueryParams(authMap)
	authParamsSerialized  = graphql.QueryParamsSerialized(authMapSerialized)
	accessStrategy        = "testAccessStrategy"
)

func fixDetailedAuth() *auth.Auth {
	return &auth.Auth{
		Credential: auth.CredentialData{
			Basic: &auth.BasicCredentialData{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AccessStrategy:        &accessStrategy,
		AdditionalHeaders:     authMap,
		AdditionalQueryParams: authMap,
		RequestAuth: &auth.CredentialRequestAuth{
			Csrf: &auth.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: authEndpoint,
				Credential: auth.CredentialData{
					Basic: &auth.BasicCredentialData{
						Username: authUsername,
						Password: authPassword,
					},
					Oauth: nil,
				},
				AdditionalHeaders:     authMap,
				AdditionalQueryParams: authMap,
			},
		},
	}
}

func fixDetailedGQLAuth() *graphql.Auth {
	emptyCertCommonName := ""
	return &graphql.Auth{
		Credential: graphql.BasicCredentialData{
			Username: authUsername,
			Password: authPassword,
		},
		AccessStrategy:                  &accessStrategy,
		AdditionalHeaders:               authHeaders,
		AdditionalHeadersSerialized:     &authHeadersSerialized,
		AdditionalQueryParams:           authParams,
		AdditionalQueryParamsSerialized: &authParamsSerialized,
		RequestAuth: &graphql.CredentialRequestAuth{
			Csrf: &graphql.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: authEndpoint,
				Credential: graphql.BasicCredentialData{
					Username: authUsername,
					Password: authPassword,
				},
				AdditionalHeaders:     authHeaders,
				AdditionalQueryParams: authParams,
			},
		},
		CertCommonName: &emptyCertCommonName,
	}
}

func fixDetailedAuthInput() *auth.AuthInput {
	return &auth.AuthInput{
		Credential: &auth.CredentialDataInput{
			Basic: &auth.BasicCredentialDataInput{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AccessStrategy:        &accessStrategy,
		AdditionalHeaders:     authMap,
		AdditionalQueryParams: authMap,
		RequestAuth: &auth.CredentialRequestAuthInput{
			Csrf: &auth.CSRFTokenCredentialRequestAuthInput{
				TokenEndpointURL: authEndpoint,
				Credential: &auth.CredentialDataInput{
					Basic: &auth.BasicCredentialDataInput{
						Username: authUsername,
						Password: authPassword,
					},
					Oauth: nil,
				},
				AdditionalHeaders:     authMap,
				AdditionalQueryParams: authMap,
			},
		},
	}
}

func fixDetailedGQLAuthInput() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AccessStrategy:                  &accessStrategy,
		AdditionalHeadersSerialized:     &authHeadersSerialized,
		AdditionalQueryParamsSerialized: &authParamsSerialized,
		RequestAuth: &graphql.CredentialRequestAuthInput{
			Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
				TokenEndpointURL: authEndpoint,
				Credential: &graphql.CredentialDataInput{
					Basic: &graphql.BasicCredentialDataInput{
						Username: authUsername,
						Password: authPassword,
					},
					Oauth: nil,
				},
				AdditionalHeadersSerialized:     &authHeadersSerialized,
				AdditionalQueryParamsSerialized: &authParamsSerialized,
			},
		},
	}
}

func fixDetailedGQLAuthInputDeprecated() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AccessStrategy:        &accessStrategy,
		AdditionalHeaders:     authHeaders,
		AdditionalQueryParams: authParams,
		RequestAuth: &graphql.CredentialRequestAuthInput{
			Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
				TokenEndpointURL: authEndpoint,
				Credential: &graphql.CredentialDataInput{
					Basic: &graphql.BasicCredentialDataInput{
						Username: authUsername,
						Password: authPassword,
					},
					Oauth: nil,
				},
				AdditionalHeaders:     authHeaders,
				AdditionalQueryParams: authParams,
			},
		},
	}
}
