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

//TODO: Fix

func (c *converter) ToGraphQL(in *model.Auth) *graphql.Auth {
	if in == nil {
		return nil
	}

	var headers graphql.HttpHeaders
	headers = in.AdditionalHeaders

	var params graphql.QueryParams
	params = in.AdditionalQueryParams

	return &graphql.Auth{
		Credential:            in.Credential,
		AdditionalHeaders:     &headers,
		AdditionalQueryParams: &params,
		RequestAuth:           c.requestAuthToGraphQL(in.RequestAuth),
	}
}

func (c *converter) InputFromGraphQL(in graphql.AuthInput) model.AuthInput {
	var credential *model.CredentialDataInput
	if in.Credential != nil {

		credential = &model.CredentialDataInput{
			Basic: in.Credential.Basic,
			Oauth: in.Credential.Oauth,
		}
	}

	return model.AuthInput{
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
			Credential:            in.Csrf.Credential,
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
			Credential:            in.Csrf.Credential,
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
