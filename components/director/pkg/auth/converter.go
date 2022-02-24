package auth

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type Credential struct {
	Basic *graphql.BasicCredentialData
	OAuth *graphql.OAuthCredentialData
}

func (Credential) IsCredentialData() {}

func ToGraphQLInput(in graphql.Auth) (*graphql.AuthInput, error) {
	credential, ok := in.Credential.(Credential)
	if !ok {
		return nil, errors.New("Could not cast credentials")
	}

	var csrfCredential Credential
	var requestAuth *graphql.CredentialRequestAuthInput

	if in.RequestAuth != nil && in.RequestAuth.Csrf != nil && in.RequestAuth.Csrf.Credential != nil {
		csrfCredential, ok = in.RequestAuth.Csrf.Credential.(Credential)
		if !ok {
			return nil, errors.New("Could not cast csrf credentials")
		}
	}

	basicCredentials := basicCredentialToInput(credential.Basic)
	oauthCredentials := oauthCredentialToInput(credential.OAuth)
	basicCsrfCredentials := basicCredentialToInput(csrfCredential.Basic)
	oauthCsrfCredentials := oauthCredentialToInput(csrfCredential.OAuth)

	if in.RequestAuth != nil && in.RequestAuth.Csrf != nil {
		requestAuth = &graphql.CredentialRequestAuthInput{
			Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
				TokenEndpointURL: in.RequestAuth.Csrf.TokenEndpointURL,
				Credential: &graphql.CredentialDataInput{
					Basic: basicCsrfCredentials,
					Oauth: oauthCsrfCredentials,
				},
				AdditionalHeaders:               in.RequestAuth.Csrf.AdditionalHeaders,
				AdditionalHeadersSerialized:     in.RequestAuth.Csrf.AdditionalHeadersSerialized,
				AdditionalQueryParams:           in.RequestAuth.Csrf.AdditionalQueryParams,
				AdditionalQueryParamsSerialized: in.RequestAuth.Csrf.AdditionalQueryParamsSerialized,
			},
		}
	}

	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Oauth: oauthCredentials,
			Basic: basicCredentials,
		},
		AccessStrategy:                  in.AccessStrategy,
		AdditionalHeaders:               in.AdditionalHeaders,
		AdditionalQueryParamsSerialized: in.AdditionalQueryParamsSerialized,
		AdditionalQueryParams:           in.AdditionalQueryParams,
		RequestAuth:                     requestAuth,
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
