package auth

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

// ToGraphQLInput converts model.Auth to graphql.AuthInput
func ToGraphQLInput(in *model.Auth) (*graphql.AuthInput, error) {
	if in == nil {
		return nil, errors.New("Missing system auth")
	}
	credentialDataInput := credentialDataToInput(in.Credential)

	requestAuthInput, err := requestAuthToInput(in.RequestAuth)
	if err != nil {
		return nil, err
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

	return &graphql.AuthInput{
		Credential:                      credentialDataInput,
		AccessStrategy:                  in.AccessStrategy,
		AdditionalHeaders:               headers,
		AdditionalHeadersSerialized:     headersSerialized,
		AdditionalQueryParamsSerialized: paramsSerialized,
		AdditionalQueryParams:           params,
		RequestAuth:                     requestAuthInput,
	}, nil
}

func requestAuthToInput(in *model.CredentialRequestAuth) (*graphql.CredentialRequestAuthInput, error) {
	if in == nil || in.Csrf == nil {
		return nil, nil
	}

	var headers graphql.HTTPHeaders
	var headersSerialized *graphql.HTTPHeadersSerialized
	if len(in.Csrf.AdditionalHeaders) != 0 {
		headers = in.Csrf.AdditionalHeaders

		serialized, err := graphql.NewHTTPHeadersSerialized(in.Csrf.AdditionalHeaders)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling AdditionalHeaders")
		}
		headersSerialized = &serialized
	}

	var params graphql.QueryParams
	var paramsSerialized *graphql.QueryParamsSerialized
	if len(in.Csrf.AdditionalQueryParams) != 0 {
		params = in.Csrf.AdditionalQueryParams

		serialized, err := graphql.NewQueryParamsSerialized(in.Csrf.AdditionalQueryParams)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling AdditionalQueryParams")
		}
		paramsSerialized = &serialized
	}

	credentialDataInput := credentialDataToInput(in.Csrf.Credential)

	requestAuth := &graphql.CredentialRequestAuthInput{
		Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
			TokenEndpointURL:                in.Csrf.TokenEndpointURL,
			AdditionalHeaders:               headers,
			AdditionalHeadersSerialized:     headersSerialized,
			AdditionalQueryParams:           params,
			AdditionalQueryParamsSerialized: paramsSerialized,
			Credential:                      credentialDataInput,
		},
	}

	return requestAuth, nil
}

func credentialDataToInput(in model.CredentialData) *graphql.CredentialDataInput {
	var basicCredentials *model.BasicCredentialData
	var oauthCredentials *model.OAuthCredentialData

	if in.Basic != nil {
		basicCredentials = in.Basic
	}

	if in.Oauth != nil {
		oauthCredentials = in.Oauth
	}

	return &graphql.CredentialDataInput{
		Basic: basicCredentialToInput(basicCredentials),
		Oauth: oauthCredentialToInput(oauthCredentials),
	}
}

func basicCredentialToInput(in *model.BasicCredentialData) *graphql.BasicCredentialDataInput {
	if in == nil {
		return nil
	}

	return &graphql.BasicCredentialDataInput{
		Username: in.Username,
		Password: in.Password,
	}
}

func oauthCredentialToInput(in *model.OAuthCredentialData) *graphql.OAuthCredentialDataInput {
	if in == nil {
		return nil
	}

	return &graphql.OAuthCredentialDataInput{
		ClientID:     in.ClientID,
		ClientSecret: in.ClientSecret,
		URL:          in.URL,
	}
}

// ToModel converts graphql.Auth to model.Auth
func ToModel(in *graphql.Auth) (*model.Auth, error) {
	if in == nil {
		return nil, nil
	}

	var headers map[string][]string
	if len(in.AdditionalHeaders) != 0 {
		if err := in.AdditionalHeaders.UnmarshalGQL(headers); err != nil {
			return nil, err
		}
	}

	var params map[string][]string
	if len(in.AdditionalQueryParams) != 0 {
		if err := in.AdditionalQueryParams.UnmarshalGQL(params); err != nil {
			return nil, err
		}
	}

	reqAuthModel, err := requestAuthToModel(in.RequestAuth)
	if err != nil {
		return nil, err
	}

	return &model.Auth{
		Credential:            credentialToModel(in.Credential),
		AccessStrategy:        in.AccessStrategy,
		AdditionalHeaders:     headers,
		AdditionalQueryParams: params,
		RequestAuth:           reqAuthModel,
		CertCommonName:        str.PtrStrToStr(in.CertCommonName),
	}, nil
}

func requestAuthToModel(in *graphql.CredentialRequestAuth) (*model.CredentialRequestAuth, error) {
	if in == nil {
		return nil, nil
	}

	var csrf *model.CSRFTokenCredentialRequestAuth
	if in.Csrf != nil {
		var headers map[string][]string
		if len(in.Csrf.AdditionalHeaders) != 0 {
			if err := in.Csrf.AdditionalHeaders.UnmarshalGQL(headers); err != nil {
				return nil, err
			}
		}

		var params map[string][]string
		if len(in.Csrf.AdditionalQueryParams) != 0 {
			if err := in.Csrf.AdditionalQueryParams.UnmarshalGQL(params); err != nil {
				return nil, err
			}
		}

		csrf = &model.CSRFTokenCredentialRequestAuth{
			TokenEndpointURL:      in.Csrf.TokenEndpointURL,
			AdditionalQueryParams: params,
			AdditionalHeaders:     headers,
			Credential:            credentialToModel(in.Csrf.Credential),
		}
	}

	return &model.CredentialRequestAuth{
		Csrf: csrf,
	}, nil
}

func credentialToModel(in graphql.CredentialData) model.CredentialData {
	var basic *model.BasicCredentialData
	var oauth *model.OAuthCredentialData

	switch cred := in.(type) {
	case graphql.BasicCredentialData:
		basic = &model.BasicCredentialData{
			Username: cred.Username,
			Password: cred.Password,
		}
	case graphql.OAuthCredentialData:
		oauth = &model.OAuthCredentialData{
			ClientID:     cred.ClientID,
			ClientSecret: cred.ClientSecret,
			URL:          cred.URL,
		}
	}

	return model.CredentialData{Basic: basic, Oauth: oauth}
}
