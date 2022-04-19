package auth_test

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var (
	authUsername   = "user"
	authPassword   = "password"
	authEndpoint   = "url"
	connectorURL   = "connectorURL"
	modelTokenType = tokens.ApplicationToken
	gqlTokenType   = graphql.OneTimeTokenTypeApplication
	authMap        = map[string][]string{
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

func fixDetailedAuth() *model.Auth {
	return &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AccessStrategy:        &accessStrategy,
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

func fixDetailedAuthInput() *model.AuthInput {
	return &model.AuthInput{
		Credential: &model.CredentialDataInput{
			Basic: &model.BasicCredentialDataInput{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AccessStrategy:        &accessStrategy,
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

func fixOneTimeTokenInput() *model.OneTimeToken {
	return &model.OneTimeToken{
		Token:        "token",
		ConnectorURL: connectorURL,
		Used:         false,
		Type:         modelTokenType,
		CreatedAt:    time.Time{},
		ExpiresAt:    time.Time{},
		UsedAt:       time.Time{},
	}
}

func fixOneTimeTokenGQLInput() *graphql.OneTimeTokenInput {
	return &graphql.OneTimeTokenInput{
		Token:        "token",
		ConnectorURL: &connectorURL,
		Used:         false,
		Type:         &gqlTokenType,
		CreatedAt:    graphql.Timestamp{},
		ExpiresAt:    graphql.Timestamp{},
		UsedAt:       graphql.Timestamp{},
	}
}
