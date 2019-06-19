package auth

import (
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToGraphQL(in *model.Auth) *graphql.Auth {
	if in == nil {
		return nil
	}

	var headers graphql.HttpHeaders
	headers = in.AdditionalHeaders

	var params graphql.QueryParams
	params = in.AdditionalQueryParams

	return &graphql.Auth{
		Credential:            c.credentialToGraphQL(in.Credential),
		AdditionalHeaders:     &headers,
		AdditionalQueryParams: &params,
		RequestAuth:           c.requestAuthToGraphQL(in.RequestAuth),
	}
}

func (c *converter) InputFromGraphQL(in *graphql.AuthInput) *model.AuthInput {
	if in == nil {
		return nil
	}

	credential := c.credentialInputFromGraphQL(in.Credential)
	return &model.AuthInput{
		Credential:            credential,
		AdditionalHeaders:     c.headersFromGraphQL(in.AdditionalHeaders),
		AdditionalQueryParams: c.queryParamsFromGraphQL(in.AdditionalQueryParams),
		RequestAuth:           c.requestAuthInputFromGraphQL(in.RequestAuth),
	}
}

func (c *converter) requestAuthToGraphQL(in *model.CredentialRequestAuth) *graphql.CredentialRequestAuth {
	if in == nil {
		return nil
	}

	var csrf *graphql.CSRFTokenCredentialRequestAuth
	if in.Csrf != nil {
		var headers graphql.HttpHeaders
		headers = in.Csrf.AdditionalHeaders

		var params graphql.QueryParams
		params = in.Csrf.AdditionalQueryParams

		csrf = &graphql.CSRFTokenCredentialRequestAuth{
			TokenEndpointURL:      in.Csrf.TokenEndpointURL,
			AdditionalQueryParams: &params,
			AdditionalHeaders:     &headers,
			Credential:            c.credentialToGraphQL(in.Csrf.Credential),
		}
	}

	return &graphql.CredentialRequestAuth{
		Csrf: csrf,
	}
}

func (c *converter) requestAuthInputFromGraphQL(in *graphql.CredentialRequestAuthInput) *model.CredentialRequestAuthInput {
	if in == nil {
		return nil
	}

	var csrf *model.CSRFTokenCredentialRequestAuthInput
	if in.Csrf != nil {
		csrf = &model.CSRFTokenCredentialRequestAuthInput{
			TokenEndpointURL:      in.Csrf.TokenEndpointURL,
			AdditionalQueryParams: c.queryParamsFromGraphQL(in.Csrf.AdditionalQueryParams),
			AdditionalHeaders:     c.headersFromGraphQL(in.Csrf.AdditionalHeaders),
			Credential:            c.credentialInputFromGraphQL(in.Csrf.Credential),
		}
	}

	return &model.CredentialRequestAuthInput{
		Csrf: csrf,
	}
}

func (c *converter) headersFromGraphQL(headers *graphql.HttpHeaders) map[string][]string {
	var h map[string][]string
	if headers != nil {
		h = *headers
	}

	return h
}

func (c *converter) queryParamsFromGraphQL(params *graphql.QueryParams) map[string][]string {
	var h map[string][]string
	if params != nil {
		h = *params
	}

	return h
}

func (c *converter) credentialInputFromGraphQL(in *graphql.CredentialDataInput) *model.CredentialDataInput {
	if in == nil {
		return nil
	}

	var basic model.BasicCredentialDataInput
	var oauth model.OAuthCredentialDataInput

	if in.Basic != nil {
		basic = model.BasicCredentialDataInput{
			Username: in.Basic.Username,
			Password: in.Basic.Password,
		}
	} else if in.Oauth != nil {
		oauth = model.OAuthCredentialDataInput{
			URL:          in.Oauth.URL,
			ClientID:     in.Oauth.ClientID,
			ClientSecret: in.Oauth.ClientSecret,
		}
	}

	return &model.CredentialDataInput{
		Basic: &basic,
		Oauth: &oauth,
	}
}

func (c *converter) credentialToGraphQL(in model.CredentialData) graphql.CredentialData {
	var credential graphql.CredentialData
	if in.Basic != nil {
		credential = graphql.BasicCredentialData{
			Username: in.Basic.Username,
			Password: in.Basic.Password,
		}
	} else if in.Oauth != nil {
		credential = graphql.OAuthCredentialData{
			URL:          in.Oauth.URL,
			ClientID:     in.Oauth.ClientID,
			ClientSecret: in.Oauth.ClientSecret,
		}
	}

	return credential
}
