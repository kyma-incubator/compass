package auth

import (
	"reflect"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

var (
	authUsername   = "user"
	authPassword   = "password"
	authEndpoint   = "url"
	accessStrategy = "testAccessStrategy"
	authMap        = map[string][]string{
		"foo": {"bar", "baz"},
		"too": {"tar", "taz"},
	}
	authMapSerialized     = "{\"foo\":[\"bar\",\"baz\"],\"too\":[\"tar\",\"taz\"]}"
	authHeaders           = graphql.HTTPHeaders(authMap)
	authHeadersSerialized = graphql.HTTPHeadersSerialized(authMapSerialized)
	authParams            = graphql.QueryParams(authMap)
	authParamsSerialized  = graphql.QueryParamsSerialized(authMapSerialized)
)

func TestConverter_ToGraphQLInput(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name             string
		Input            *model.Auth
		Expected         *graphql.AuthInput
		ExpectedErrorMsg string
	}{
		{
			Name:     "All properties given",
			Input:    fixModelAuth(),
			Expected: fixGQLAuthInput(),
		},
		{
			Name:     "Empty",
			Input:    &model.Auth{},
			Expected: &graphql.AuthInput{Credential: &graphql.CredentialDataInput{}},
		},
		{
			Name:             "Nil",
			Input:            nil,
			Expected:         nil,
			ExpectedErrorMsg: "Missing system auth",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			res, err := ToGraphQLInput(testCase.Input)

			// then
			if testCase.ExpectedErrorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
				reflect.DeepEqual(testCase.Expected, res)
			}
		})
	}
}

func TestConverter_ToModel(t *testing.T) {
	// GIVEN
	expectedAuthInputWithNilReqAuth := fixGQLAuthInput()
	expectedAuthInputWithNilReqAuth.RequestAuth = nil

	testCases := []struct {
		Name             string
		Input            *graphql.Auth
		Expected         *model.Auth
		ExpectedErrorMsg string
	}{
		{
			Name:     "All properties given",
			Input:    fixGqlAuth(),
			Expected: fixModelAuth(),
		},
		{
			Name:     "Empty",
			Input:    &graphql.Auth{},
			Expected: &model.Auth{},
		},
		{
			Name:             "Nil",
			Input:            nil,
			Expected:         nil,
			ExpectedErrorMsg: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			res, err := ToModel(testCase.Input)

			// then
			if testCase.ExpectedErrorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
				reflect.DeepEqual(testCase.Expected, res)
			}
		})
	}
}

func fixModelAuth() *model.Auth {
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

func fixGqlAuth() *graphql.Auth {
	return &graphql.Auth{
		Credential: graphql.BasicCredentialData{
			Username: authUsername,
			Password: authPassword,
		},
		AccessStrategy:        &accessStrategy,
		AdditionalHeaders:     authHeaders,
		AdditionalQueryParams: authParams,
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
	}
}

func fixGQLAuthInput() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: authUsername,
				Password: authPassword,
			},
			Oauth: nil,
		},
		AccessStrategy:                  &accessStrategy,
		AdditionalHeaders:               authHeaders,
		AdditionalHeadersSerialized:     &authHeadersSerialized,
		AdditionalQueryParams:           authParams,
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
				AdditionalHeaders:               authHeaders,
				AdditionalHeadersSerialized:     &authHeadersSerialized,
				AdditionalQueryParams:           authParams,
				AdditionalQueryParamsSerialized: &authParamsSerialized,
			},
		},
	}
}
