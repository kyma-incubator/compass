package auth

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

func ToGraphQLInput(in graphql.Auth) (*graphql.AuthInput, error) {
	credentialDataInput, err := credentialDataToInput(in.Credential)
	if err != nil {
		return nil, err
	}

	requestAuthInput, err := requestAuthToInput(in.RequestAuth)
	if err != nil {
		return nil, err
	}

	return &graphql.AuthInput{
		Credential:                      credentialDataInput,
		AccessStrategy:                  in.AccessStrategy,
		AdditionalHeaders:               in.AdditionalHeaders,
		AdditionalQueryParamsSerialized: in.AdditionalQueryParamsSerialized,
		AdditionalQueryParams:           in.AdditionalQueryParams,
		RequestAuth:                     requestAuthInput,
	}, nil
}

func basicCredentialToInput(in *graphql.BasicCredentialData) *graphql.BasicCredentialDataInput {
	if in == nil {
		return nil
	}

	return &graphql.BasicCredentialDataInput{
		Username: in.Username,
		Password: in.Password,
	}
}

func oauthCredentialToInput(in *graphql.OAuthCredentialData) *graphql.OAuthCredentialDataInput {
	if in == nil {
		return nil
	}

	return &graphql.OAuthCredentialDataInput{
		ClientID:     in.ClientID,
		ClientSecret: in.ClientSecret,
		URL:          in.URL,
	}
}

func credentialDataToInput(in graphql.CredentialData) (*graphql.CredentialDataInput, error) {
	if in == nil {
		return nil, nil
	}

	var basicCredentials *graphql.BasicCredentialData
	var oauthCredentials *graphql.OAuthCredentialData

	switch actual := in.(type) {
	case graphql.BasicCredentialData:
		basicCredentials = &actual
	case graphql.OAuthCredentialData:
		oauthCredentials = &actual
	default:
		return nil, errors.New("Could not cast credentials")
	}

	return &graphql.CredentialDataInput{
		Basic: basicCredentialToInput(basicCredentials),
		Oauth: oauthCredentialToInput(oauthCredentials),
	}, nil
}

func requestAuthToInput(in *graphql.CredentialRequestAuth) (*graphql.CredentialRequestAuthInput, error) {
	if in == nil || in.Csrf == nil {
		return nil, nil
	}

	requestAuth := &graphql.CredentialRequestAuthInput{
		Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
			TokenEndpointURL:                in.Csrf.TokenEndpointURL,
			AdditionalHeaders:               in.Csrf.AdditionalHeaders,
			AdditionalHeadersSerialized:     in.Csrf.AdditionalHeadersSerialized,
			AdditionalQueryParams:           in.Csrf.AdditionalQueryParams,
			AdditionalQueryParamsSerialized: in.Csrf.AdditionalQueryParamsSerialized,
		},
	}

	if in.Csrf.Credential != nil {
		csrfCredentialDataInput, err := credentialDataToInput(in.Csrf.Credential)
		if err != nil {
			return nil, err
		}

		requestAuth.Csrf.Credential = csrfCredentialDataInput
	}

	return requestAuth, nil
}
