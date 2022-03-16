package auth

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// ToGraphQLInput converts graphql.Auth to graphql.AuthInput
func ToGraphQLInput(in *Auth) (*graphql.AuthInput, error) {
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

func basicCredentialToInput(in *BasicCredentialData) *graphql.BasicCredentialDataInput {
	if in == nil {
		return nil
	}

	return &graphql.BasicCredentialDataInput{
		Username: in.Username,
		Password: in.Password,
	}
}

func oauthCredentialToInput(in *OAuthCredentialData) *graphql.OAuthCredentialDataInput {
	if in == nil {
		return nil
	}

	return &graphql.OAuthCredentialDataInput{
		ClientID:     in.ClientID,
		ClientSecret: in.ClientSecret,
		URL:          in.URL,
	}
}

func credentialDataToInput(in CredentialData) *graphql.CredentialDataInput {
	var basicCredentials *BasicCredentialData
	var oauthCredentials *OAuthCredentialData

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

func requestAuthToInput(in *CredentialRequestAuth) (*graphql.CredentialRequestAuthInput, error) {
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

	//if in.Csrf.Credential != nil {
	//	csrfCredentialDataInput, err := credentialDataToInput(in.Csrf.Credential)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	requestAuth.Csrf.Credential = csrfCredentialDataInput
	//}

	return requestAuth, nil
}

// ToModel missing godoc
func ToModel(in *graphql.Auth) (*Auth, error) {
	if in == nil {
		return nil, nil
	}

	var headers map[string][]string
	if len(in.AdditionalHeaders) != 0 {
		in.AdditionalHeaders.UnmarshalGQL(headers)
	}

	var params map[string][]string
	if len(in.AdditionalQueryParams) != 0 {
		in.AdditionalQueryParams.UnmarshalGQL(params)
	}

	return &Auth{
		Credential:            credentialToModel(in.Credential),
		AccessStrategy:        in.AccessStrategy,
		AdditionalHeaders:     headers,
		AdditionalQueryParams: params,
		RequestAuth:           requestAuthToModel(in.RequestAuth),
		CertCommonName:        *in.CertCommonName,
	}, nil
}

func requestAuthToModel(in *graphql.CredentialRequestAuth) *CredentialRequestAuth {
	if in == nil {
		return nil
	}

	var csrf *CSRFTokenCredentialRequestAuth
	if in.Csrf != nil {
		var headers map[string][]string
		if len(in.Csrf.AdditionalHeaders) != 0 {
			in.Csrf.AdditionalHeaders.UnmarshalGQL(headers)
		}

		var params map[string][]string
		if len(in.Csrf.AdditionalQueryParams) != 0 {
			in.Csrf.AdditionalQueryParams.UnmarshalGQL(params)
		}

		csrf = &CSRFTokenCredentialRequestAuth{
			TokenEndpointURL:      in.Csrf.TokenEndpointURL,
			AdditionalQueryParams: params,
			AdditionalHeaders:     headers,
			Credential:            credentialToModel(in.Csrf.Credential),
		}
	}

	return &CredentialRequestAuth{
		Csrf: csrf,
	}
}

func credentialToModel(in graphql.CredentialData) CredentialData {
	var basic *BasicCredentialData
	var oauth *OAuthCredentialData

	switch cred := in.(type) {
	case *graphql.BasicCredentialData:
		basic = &BasicCredentialData{
			Username: cred.Username,
			Password: cred.Password,
		}
	case *graphql.OAuthCredentialData:
		oauth = &OAuthCredentialData{
			ClientID:     cred.ClientID,
			ClientSecret: cred.ClientSecret,
			URL:          cred.URL,
		}
	}

	return CredentialData{Basic: basic, Oauth: oauth}
}
