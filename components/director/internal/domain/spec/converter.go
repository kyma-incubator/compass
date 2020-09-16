package spec

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c converter) FromEntity(specEnt Entity) model.Spec {
	spec := model.Spec{
		ID:                specEnt.ID,
		Tenant:            specEnt.TenantID,
		APIDefinitionID:   repo.StringPtrFromNullableString(specEnt.APIDefinitionID),
		EventDefinitionID: repo.StringPtrFromNullableString(specEnt.EventDefinitionID),
	}
	if !specEnt.SpecData.Valid && !specEnt.SpecFormat.Valid && !specEnt.SpecType.Valid {
		return spec
	}

	specFormat := repo.StringPtrFromNullableString(specEnt.SpecFormat)
	if specFormat != nil {
		spec.Format = model.SpecFormat(*specFormat)
	}

	specType := repo.StringPtrFromNullableString(specEnt.SpecType)
	if specFormat != nil {
		spec.Type = model.SpecType(*specType)
	}
	spec.CustomType = repo.StringPtrFromNullableString(specEnt.CustomType)
	spec.Data = repo.StringPtrFromNullableString(specEnt.SpecData)
	return spec
}

func (c converter) ToEntity(apiModel model.Spec) Entity {
	return Entity{
		ID:                apiModel.ID,
		TenantID:          apiModel.Tenant,
		APIDefinitionID:   repo.NewNullableString(apiModel.APIDefinitionID),
		EventDefinitionID: repo.NewNullableString(apiModel.EventDefinitionID),
		SpecData:          repo.NewNullableString(apiModel.Data),
		SpecFormat:        repo.NewNullableString(str.Ptr(string(apiModel.Format))),
		SpecType:          repo.NewNullableString(str.Ptr(string(apiModel.Type))),
		CustomType:        repo.NewNullableString(apiModel.CustomType),
	}
}

func (c converter) APISpecInputFromSpec(spec *model.Spec, fr *model.FetchRequest) *model.APISpecInput {
	var auth *model.AuthInput
	if fr.Auth != nil {
		var basicCreds *model.BasicCredentialDataInput
		var oauthCreds *model.OAuthCredentialDataInput
		if fr.Auth.Credential.Basic != nil {
			basicCreds = &model.BasicCredentialDataInput{
				Username: fr.Auth.Credential.Basic.Username,
				Password: fr.Auth.Credential.Basic.Password,
			}
		}
		if fr.Auth.Credential.Oauth != nil {
			oauthCreds = &model.OAuthCredentialDataInput{
				ClientID:     fr.Auth.Credential.Oauth.ClientID,
				ClientSecret: fr.Auth.Credential.Oauth.ClientSecret,
				URL:          fr.Auth.Credential.Oauth.URL,
			}
		}
		var requestAuth *model.CredentialRequestAuthInput
		if fr.Auth.RequestAuth != nil {
			var csrf *model.CSRFTokenCredentialRequestAuthInput
			if fr.Auth.RequestAuth.Csrf != nil {
				var basicCreds *model.BasicCredentialDataInput
				var oauthCreds *model.OAuthCredentialDataInput
				if fr.Auth.RequestAuth.Csrf.Credential.Basic != nil {
					basicCreds = &model.BasicCredentialDataInput{
						Username: fr.Auth.RequestAuth.Csrf.Credential.Basic.Username,
						Password: fr.Auth.RequestAuth.Csrf.Credential.Basic.Password,
					}
				}
				if fr.Auth.RequestAuth.Csrf.Credential.Oauth != nil {
					oauthCreds = &model.OAuthCredentialDataInput{
						ClientID:     fr.Auth.RequestAuth.Csrf.Credential.Oauth.ClientID,
						ClientSecret: fr.Auth.RequestAuth.Csrf.Credential.Oauth.ClientSecret,
						URL:          fr.Auth.RequestAuth.Csrf.Credential.Oauth.URL,
					}
				}
				csrf = &model.CSRFTokenCredentialRequestAuthInput{
					TokenEndpointURL: fr.Auth.RequestAuth.Csrf.TokenEndpointURL,
					Credential: &model.CredentialDataInput{
						Basic: basicCreds,
						Oauth: oauthCreds,
					},
					AdditionalHeaders:     fr.Auth.RequestAuth.Csrf.AdditionalHeaders,
					AdditionalQueryParams: fr.Auth.RequestAuth.Csrf.AdditionalQueryParams,
				}
			}
			requestAuth = &model.CredentialRequestAuthInput{
				Csrf: csrf,
			}
		}

		auth = &model.AuthInput{
			Credential: &model.CredentialDataInput{
				Basic: basicCreds,
				Oauth: oauthCreds,
			},
			AdditionalHeaders:     fr.Auth.AdditionalHeaders,
			AdditionalQueryParams: fr.Auth.AdditionalQueryParams,
			RequestAuth:           requestAuth,
		}
	}

	return &model.APISpecInput{
		Data:       spec.Data,
		Format:     spec.Format,
		Type:       model.APISpecType(spec.Type),
		CustomType: spec.CustomType,
		FetchRequest: &model.FetchRequestInput{
			URL:    fr.URL,
			Auth:   auth,
			Mode:   &fr.Mode,
			Filter: fr.Filter,
		},
	}
}

func (c converter) EventSpecInputFromSpec(spec *model.Spec, fr *model.FetchRequest) *model.EventSpecInput {
	var auth *model.AuthInput
	if fr.Auth != nil {
		var basicCreds *model.BasicCredentialDataInput
		var oauthCreds *model.OAuthCredentialDataInput
		if fr.Auth.Credential.Basic != nil {
			basicCreds = &model.BasicCredentialDataInput{
				Username: fr.Auth.Credential.Basic.Username,
				Password: fr.Auth.Credential.Basic.Password,
			}
		}
		if fr.Auth.Credential.Oauth != nil {
			oauthCreds = &model.OAuthCredentialDataInput{
				ClientID:     fr.Auth.Credential.Oauth.ClientID,
				ClientSecret: fr.Auth.Credential.Oauth.ClientSecret,
				URL:          fr.Auth.Credential.Oauth.URL,
			}
		}
		var requestAuth *model.CredentialRequestAuthInput
		if fr.Auth.RequestAuth != nil {
			var csrf *model.CSRFTokenCredentialRequestAuthInput
			if fr.Auth.RequestAuth.Csrf != nil {
				var basicCreds *model.BasicCredentialDataInput
				var oauthCreds *model.OAuthCredentialDataInput
				if fr.Auth.RequestAuth.Csrf.Credential.Basic != nil {
					basicCreds = &model.BasicCredentialDataInput{
						Username: fr.Auth.RequestAuth.Csrf.Credential.Basic.Username,
						Password: fr.Auth.RequestAuth.Csrf.Credential.Basic.Password,
					}
				}
				if fr.Auth.RequestAuth.Csrf.Credential.Oauth != nil {
					oauthCreds = &model.OAuthCredentialDataInput{
						ClientID:     fr.Auth.RequestAuth.Csrf.Credential.Oauth.ClientID,
						ClientSecret: fr.Auth.RequestAuth.Csrf.Credential.Oauth.ClientSecret,
						URL:          fr.Auth.RequestAuth.Csrf.Credential.Oauth.URL,
					}
				}
				csrf = &model.CSRFTokenCredentialRequestAuthInput{
					TokenEndpointURL: fr.Auth.RequestAuth.Csrf.TokenEndpointURL,
					Credential: &model.CredentialDataInput{
						Basic: basicCreds,
						Oauth: oauthCreds,
					},
					AdditionalHeaders:     fr.Auth.RequestAuth.Csrf.AdditionalHeaders,
					AdditionalQueryParams: fr.Auth.RequestAuth.Csrf.AdditionalQueryParams,
				}
			}
			requestAuth = &model.CredentialRequestAuthInput{
				Csrf: csrf,
			}
		}

		auth = &model.AuthInput{
			Credential: &model.CredentialDataInput{
				Basic: basicCreds,
				Oauth: oauthCreds,
			},
			AdditionalHeaders:     fr.Auth.AdditionalHeaders,
			AdditionalQueryParams: fr.Auth.AdditionalQueryParams,
			RequestAuth:           requestAuth,
		}
	}

	return &model.EventSpecInput{
		Data:          spec.Data,
		Format:        spec.Format,
		EventSpecType: model.EventSpecType(spec.Type),
		CustomType:    spec.CustomType,
		FetchRequest: &model.FetchRequestInput{
			URL:    fr.URL,
			Auth:   auth,
			Mode:   &fr.Mode,
			Filter: fr.Filter,
		},
	}
}
