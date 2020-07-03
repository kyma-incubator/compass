package auth_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
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
	authHeaders           = graphql.HttpHeaders(authMap)
	authHeadersSerialized = graphql.HttpHeadersSerialized(authMapSerialized)
	authParams            = graphql.QueryParams(authMap)
	authParamsSerialized  = graphql.QueryParamsSerialized(authMapSerialized)
)

func fixDetailedAuth() *model.Auth {
	return &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AdditionalHeaders:     authMap,
		AdditionalQueryParams: authMap,
		RequestAuth: &model.CredentialRequestAuth{
			Csrf: &model.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: authEndpoint,
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
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
	return &graphql.Auth{
		Credential: graphql.BasicCredentialData{
			Username: authUsername,
			Password: authPassword,
		},
		AdditionalHeaders:               &authHeaders,
		AdditionalHeadersSerialized:     &authHeadersSerialized,
		AdditionalQueryParams:           &authParams,
		AdditionalQueryParamsSerialized: &authParamsSerialized,
		RequestAuth: &graphql.CredentialRequestAuth{
			Csrf: &graphql.CSRFTokenCredentialRequestAuth{
				TokenEndpointURL: authEndpoint,
				Credential: graphql.BasicCredentialData{
					Username: authUsername,
					Password: authPassword,
				},
				AdditionalHeaders:     &authHeaders,
				AdditionalQueryParams: &authParams,
			},
		},
	}
}

func fixDetailedAuthInput() *model.AuthInput {
	return &model.AuthInput{
		Credential: &model.CredentialDataInput{
			Basic: &model.BasicCredentialDataInput{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AdditionalHeaders:     authMap,
		AdditionalQueryParams: authMap,
		RequestAuth: &model.CredentialRequestAuthInput{
			Csrf: &model.CSRFTokenCredentialRequestAuthInput{
				TokenEndpointURL: authEndpoint,
				Credential: &model.CredentialDataInput{
					Basic: &model.BasicCredentialDataInput{
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
		AdditionalHeaders:     &authHeaders,
		AdditionalQueryParams: &authParams,
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
				AdditionalHeaders:     &authHeaders,
				AdditionalQueryParams: &authParams,
			},
		},
	}
}
