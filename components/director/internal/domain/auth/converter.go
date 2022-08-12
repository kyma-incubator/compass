package auth

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

// TokenConverter missing godoc
//go:generate mockery --name=TokenConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type TokenConverter interface {
	ToGraphQLForApplication(model model.OneTimeToken) (graphql.OneTimeTokenForApplication, error)
}

type converter struct {
}

type converterOTT struct {
	*converter
	tokenConverter TokenConverter
}

// NewConverterWithOTT is meant to be used for converting system auth with the one time token also converted
func NewConverterWithOTT(tokenConverter TokenConverter) *converterOTT {
	return &converterOTT{
		converter:      &converter{},
		tokenConverter: tokenConverter,
	}
}

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

// ToGraphQL missing godoc
func (c *converter) ToGraphQL(in *model.Auth) (*graphql.Auth, error) {
	if in == nil {
		return nil, nil
	}

	var headers graphql.HTTPHeaders
	var headersSerialized *graphql.HTTPHeadersSerialized
	if len(in.AdditionalHeaders) != 0 {
		headers = in.AdditionalHeaders

		serialized, err := graphql.NewHTTPHeadersSerialized(in.AdditionalHeaders)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling AdditionalHeaders")
		}
		headersSerialized = &serialized
	}

	var params graphql.QueryParams
	var paramsSerialized *graphql.QueryParamsSerialized
	if len(in.AdditionalQueryParams) != 0 {
		params = in.AdditionalQueryParams

		serialized, err := graphql.NewQueryParamsSerialized(in.AdditionalQueryParams)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling AdditionalQueryParams")
		}
		paramsSerialized = &serialized
	}

	return &graphql.Auth{
		Credential:                      c.credentialToGraphQL(in.Credential),
		AccessStrategy:                  in.AccessStrategy,
		AdditionalHeaders:               headers,
		AdditionalHeadersSerialized:     headersSerialized,
		AdditionalQueryParams:           params,
		AdditionalQueryParamsSerialized: paramsSerialized,
		RequestAuth:                     c.requestAuthToGraphQL(in.RequestAuth),
		CertCommonName:                  &in.CertCommonName,
	}, nil
}

func (c *converterOTT) ToGraphQL(in *model.Auth) (*graphql.Auth, error) {
	auth, err := c.converter.ToGraphQL(in)
	if err != nil {
		return nil, err
	}
	if auth == nil {
		return nil, nil
	}

	if in.OneTimeToken != nil && in.OneTimeToken.Type == tokens.ApplicationToken {
		oneTimeToken, err := c.tokenConverter.ToGraphQLForApplication(*in.OneTimeToken)
		if err != nil {
			return nil, err
		}
		auth.OneTimeToken = &oneTimeToken
	}

	return auth, nil
}

// InputFromGraphQL missing godoc
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
		AccessStrategy:        in.AccessStrategy,
		AdditionalHeaders:     additionalHeaders,
		AdditionalQueryParams: additionalQueryParams,
		RequestAuth:           reqAuth,
	}, nil
}

// ModelFromGraphQLTokenInput converts a graphql.OneTimeTokenInput to model.OneTimeToken
func (c *converter) ModelFromGraphQLTokenInput(in *graphql.OneTimeTokenInput) *model.OneTimeToken {
	if in == nil {
		return nil
	}

	var connectorURL string
	if in.ConnectorURL != nil {
		connectorURL = *in.ConnectorURL
	}

	var tokenType tokens.TokenType
	if in.Type != nil {
		tokenType = tokens.TokenType(*in.Type)
	}

	return &model.OneTimeToken{
		Token:        in.Token,
		ConnectorURL: connectorURL,
		Type:         tokenType,
		CreatedAt:    timestampToTime(in.CreatedAt),
		UsedAt:       timestampToTime(in.UsedAt),
		ExpiresAt:    timestampToTime(in.ExpiresAt),
		Used:         in.Used,
	}
}

// ModelFromGraphQLInput converts a graphql.AuthInput to model.Auth
func (c *converter) ModelFromGraphQLInput(in graphql.AuthInput) (*model.Auth, error) {
	credential := c.credentialModelFromGraphQL(in.Credential)

	additionalHeaders, err := c.headersFromGraphQL(in.AdditionalHeaders, in.AdditionalHeadersSerialized)
	if err != nil {
		return nil, errors.Wrap(err, "while converting AdditionalHeaders from GraphQL input")
	}

	additionalQueryParams, err := c.queryParamsFromGraphQL(in.AdditionalQueryParams, in.AdditionalQueryParamsSerialized)
	if err != nil {
		return nil, errors.Wrap(err, "while converting AdditionalQueryParams from GraphQL input")
	}

	reqAuth, err := c.requestAuthFromGraphQL(in.RequestAuth)
	if err != nil {
		return nil, err
	}

	return &model.Auth{
		OneTimeToken:          c.ModelFromGraphQLTokenInput(in.OneTimeToken),
		Credential:            credential,
		AccessStrategy:        in.AccessStrategy,
		AdditionalHeaders:     additionalHeaders,
		AdditionalQueryParams: additionalQueryParams,
		RequestAuth:           reqAuth,
		CertCommonName:        str.PtrStrToStr(in.CertCommonName),
	}, nil
}

func (c *converter) requestAuthToGraphQL(in *model.CredentialRequestAuth) *graphql.CredentialRequestAuth {
	if in == nil {
		return nil
	}

	var csrf *graphql.CSRFTokenCredentialRequestAuth
	if in.Csrf != nil {
		var headers graphql.HTTPHeaders
		if len(in.Csrf.AdditionalHeaders) != 0 {
			headers = in.Csrf.AdditionalHeaders
		}

		var params graphql.QueryParams
		if len(in.Csrf.AdditionalQueryParams) != 0 {
			params = in.Csrf.AdditionalQueryParams
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

func (c *converter) requestAuthFromGraphQL(in *graphql.CredentialRequestAuthInput) (*model.CredentialRequestAuth, error) {
	if in == nil {
		return nil, nil
	}

	var csrf *model.CSRFTokenCredentialRequestAuth
	if in.Csrf != nil {
		additionalHeaders, err := c.headersFromGraphQL(in.Csrf.AdditionalHeaders, in.Csrf.AdditionalHeadersSerialized)
		if err != nil {
			return nil, errors.Wrap(err, "while converting CSRF AdditionalHeaders from GraphQL input")
		}

		additionalQueryParams, err := c.queryParamsFromGraphQL(in.Csrf.AdditionalQueryParams, in.Csrf.AdditionalQueryParamsSerialized)
		if err != nil {
			return nil, errors.Wrap(err, "while converting CSRF AdditionalQueryParams from GraphQL input")
		}

		csrf = &model.CSRFTokenCredentialRequestAuth{
			TokenEndpointURL:      in.Csrf.TokenEndpointURL,
			AdditionalQueryParams: additionalQueryParams,
			AdditionalHeaders:     additionalHeaders,
			Credential:            c.credentialModelFromGraphQL(in.Csrf.Credential),
		}
	}

	return &model.CredentialRequestAuth{
		Csrf: csrf,
	}, nil
}

func (c *converter) headersFromGraphQL(headers graphql.HTTPHeaders, headersSerialized *graphql.HTTPHeadersSerialized) (map[string][]string, error) {
	var h map[string][]string

	if headersSerialized != nil {
		return headersSerialized.Unmarshal()
	} else if headers != nil {
		h = headers
	}

	return h, nil
}

func (c *converter) queryParamsFromGraphQL(params graphql.QueryParams, paramsSerialized *graphql.QueryParamsSerialized) (map[string][]string, error) {
	var p map[string][]string

	if paramsSerialized != nil {
		return paramsSerialized.Unmarshal()
	} else if params != nil {
		p = params
	}

	return p, nil
}

func (c *converter) credentialInputFromGraphQL(in *graphql.CredentialDataInput) *model.CredentialDataInput {
	if in == nil {
		return nil
	}

	var basic *model.BasicCredentialDataInput
	var oauth *model.OAuthCredentialDataInput
	var certOAuth *model.CertificateOAuthCredentialDataInput

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
	} else if in.CertificateOAuth != nil {
		certOAuth = &model.CertificateOAuthCredentialDataInput{
			ClientID:    in.CertificateOAuth.ClientID,
			Certificate: in.CertificateOAuth.Certificate,
			URL:         in.CertificateOAuth.URL,
		}
	}

	return &model.CredentialDataInput{
		Basic:            basic,
		Oauth:            oauth,
		CertificateOAuth: certOAuth,
	}
}

func (c *converter) credentialModelFromGraphQL(in *graphql.CredentialDataInput) model.CredentialData {
	if in == nil {
		return model.CredentialData{}
	}

	var basic *model.BasicCredentialData
	var oauth *model.OAuthCredentialData
	var certOAuth *model.CertificateOAuthCredentialData

	if in.Basic != nil {
		basic = &model.BasicCredentialData{
			Username: in.Basic.Username,
			Password: in.Basic.Password,
		}
	} else if in.Oauth != nil {
		oauth = &model.OAuthCredentialData{
			URL:          in.Oauth.URL,
			ClientID:     in.Oauth.ClientID,
			ClientSecret: in.Oauth.ClientSecret,
		}
	} else if in.CertificateOAuth != nil {
		certOAuth = &model.CertificateOAuthCredentialData{
			ClientID:    in.CertificateOAuth.ClientID,
			Certificate: in.CertificateOAuth.Certificate,
			URL:         in.CertificateOAuth.URL,
		}
	}

	return model.CredentialData{
		Basic:            basic,
		Oauth:            oauth,
		CertificateOAuth: certOAuth,
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
	} else if in.CertificateOAuth != nil {
		credential = graphql.CertificateOAuthCredentialData{
			ClientID:    in.CertificateOAuth.ClientID,
			Certificate: in.CertificateOAuth.Certificate,
			URL:         in.CertificateOAuth.URL,
		}
	}

	return credential
}

func timestampToTime(timestamp graphql.Timestamp) time.Time {
	return time.Time(timestamp)
}
