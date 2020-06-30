package auth

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToGraphQL(in *model.Auth) (*graphql.Auth, error) {
	if in == nil {
		return nil, nil
	}

	var headers *graphql.HttpHeaders
	var headersSerialized *graphql.HttpHeadersSerialized
	if len(in.AdditionalHeaders) != 0 {
		var value graphql.HttpHeaders = in.AdditionalHeaders
		headers = &value

		serialized, err := graphql.NewHttpHeadersSerialized(in.AdditionalHeaders)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling AdditionalHeaders")
		}
		headersSerialized = &serialized
	}

	var params *graphql.QueryParams
	var paramsSerialized *graphql.QueryParamsSerialized
	if len(in.AdditionalQueryParams) != 0 {
		var value graphql.QueryParams = in.AdditionalQueryParams
		params = &value

		serialized, err := graphql.NewQueryParamsSerialized(in.AdditionalQueryParams)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling AdditionalQueryParams")
		}
		paramsSerialized = &serialized
	}

	return &graphql.Auth{
		Credential:                      c.credentialToGraphQL(in.Credential),
		AdditionalHeaders:               headers,
		AdditionalHeadersSerialized:     headersSerialized,
		AdditionalQueryParams:           params,
		AdditionalQueryParamsSerialized: paramsSerialized,
		RequestAuth:                     c.requestAuthToGraphQL(in.RequestAuth),
	}, nil
}

func (c *converter) InputFromGraphQL(in *graphql.AuthInput) (*model.AuthInput, error) {
	if in == nil {
		return nil, nil
	}

	credential := c.credentialInputFromGraphQL(in.Credential)

	additionalHeaders, err := c.headersFromGraphQL(in.AdditionalHeaders, in.AdditionalHeadersSerialized)
	if err != nil {
		return nil, errors.Wrap(err, "while converting AdditionalHeaders from GraphQL input")
	}

	additionalQueryParams, err := c.queryParamsFromGraphQL(in.AdditionalQueryParams, in.AdditionalQueryParamsSerialized)
	if err != nil {
		return nil, errors.Wrap(err, "while converting AdditionalQueryParams from GraphQL input")
	}

	reqAuth, err := c.requestAuthInputFromGraphQL(in.RequestAuth)
	if err != nil {
		return nil, err
	}

	return &model.AuthInput{
		Credential:            credential,
		AdditionalHeaders:     additionalHeaders,
		AdditionalQueryParams: additionalQueryParams,
		RequestAuth:           reqAuth,
	}, nil
}

func (c *converter) requestAuthToGraphQL(in *model.CredentialRequestAuth) *graphql.CredentialRequestAuth {
	if in == nil {
		return nil
	}

	var csrf *graphql.CSRFTokenCredentialRequestAuth
	if in.Csrf != nil {
		var headers *graphql.HttpHeaders
		if len(in.Csrf.AdditionalHeaders) != 0 {
			var value graphql.HttpHeaders = in.Csrf.AdditionalHeaders
			headers = &value
		}

		var params *graphql.QueryParams
		if len(in.Csrf.AdditionalQueryParams) != 0 {
			var value graphql.QueryParams = in.Csrf.AdditionalQueryParams
			params = &value
		}

		csrf = &graphql.CSRFTokenCredentialRequestAuth{
			TokenEndpointURL:      in.Csrf.TokenEndpointURL,
			AdditionalQueryParams: params,
			AdditionalHeaders:     headers,
			Credential:            c.credentialToGraphQL(in.Csrf.Credential),
		}
	}

	return &graphql.CredentialRequestAuth{
		Csrf: csrf,
	}
}

func (c *converter) requestAuthInputFromGraphQL(in *graphql.CredentialRequestAuthInput) (*model.CredentialRequestAuthInput, error) {
	if in == nil {
		return nil, nil
	}

	var csrf *model.CSRFTokenCredentialRequestAuthInput
	if in.Csrf != nil {
		additionalHeaders, err := c.headersFromGraphQL(in.Csrf.AdditionalHeaders, in.Csrf.AdditionalHeadersSerialized)
		if err != nil {
			return nil, errors.Wrap(err, "while converting CSRF AdditionalHeaders from GraphQL input")
		}

		additionalQueryParams, err := c.queryParamsFromGraphQL(in.Csrf.AdditionalQueryParams, in.Csrf.AdditionalQueryParamsSerialized)
		if err != nil {
			return nil, errors.Wrap(err, "while converting CSRF AdditionalQueryParams from GraphQL input")
		}

		csrf = &model.CSRFTokenCredentialRequestAuthInput{
			TokenEndpointURL:      in.Csrf.TokenEndpointURL,
			AdditionalQueryParams: additionalQueryParams,
			AdditionalHeaders:     additionalHeaders,
			Credential:            c.credentialInputFromGraphQL(in.Csrf.Credential),
		}
	}

	return &model.CredentialRequestAuthInput{
		Csrf: csrf,
	}, nil
}

func (c *converter) headersFromGraphQL(headers *graphql.HttpHeaders, headersSerialized *graphql.HttpHeadersSerialized) (map[string][]string, error) {
	var h map[string][]string

	if headersSerialized != nil {
		return headersSerialized.Unmarshal()
	} else if headers != nil {
		h = *headers
	}

	return h, nil
}

func (c *converter) queryParamsFromGraphQL(params *graphql.QueryParams, paramsSerialized *graphql.QueryParamsSerialized) (map[string][]string, error) {
	var p map[string][]string

	if paramsSerialized != nil {
		return paramsSerialized.Unmarshal()
	} else if params != nil {
		p = *params
	}

	return p, nil
}

func (c *converter) credentialInputFromGraphQL(in *graphql.CredentialDataInput) *model.CredentialDataInput {
	if in == nil {
		return nil
	}

	var basic *model.BasicCredentialDataInput
	var oauth *model.OAuthCredentialDataInput

	if in.Basic != nil {
		basic = &model.BasicCredentialDataInput{
			Username: in.Basic.Username,
			Password: in.Basic.Password,
		}
	} else if in.Oauth != nil {
		oauth = &model.OAuthCredentialDataInput{
			URL:          in.Oauth.URL,
			ClientID:     in.Oauth.ClientID,
			ClientSecret: in.Oauth.ClientSecret,
		}
	}

	return &model.CredentialDataInput{
		Basic: basic,
		Oauth: oauth,
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
