package service

import (
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	prefixUnmappedFields          = "compatibility_unmapped_fields_" // TODO
	unmappedFieldIdentifier       = prefixUnmappedFields + "identifier"
	unmappedFieldShortDescription = prefixUnmappedFields + "shortDescription"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

/*
type Converter interface {
	DetailsToGraphQLInput(in model.ServiceDetails) (graphql.ApplicationRegisterInput, error)
	GraphQLToDetailsModel(in graphql.ApplicationExt) (model.ServiceDetails, error)
	GraphQLToModel(in graphql.ApplicationExt) (model.Service, error)
}
*/

func (c *converter) DetailsToGraphQLInput(in model.ServiceDetails) (graphql.ApplicationRegisterInput, error) {

	outLabels := graphql.Labels{}

	out := graphql.ApplicationRegisterInput{
		Name:         in.Name,
		Description:  ptrStringOrNilForEmpty(in.Description),
		ProviderName: ptrStringOrNilForEmpty(in.Provider),
	}

	if in.ShortDescription != "" {
		outLabels[unmappedFieldShortDescription] = in.ShortDescription
	}
	if in.Identifier != "" {
		outLabels[unmappedFieldIdentifier] = in.Identifier
	}
	if in.Api != nil {

		outApi := &graphql.APIDefinitionInput{
			TargetURL: in.Api.TargetUrl,
		}

		if in.Api.ApiType != "" {
			if outApi.Spec == nil {
				outApi.Spec = &graphql.APISpecInput{}
			}
			if strings.ToLower(in.Api.ApiType) == "odata" {
				outApi.Spec.Type = graphql.APISpecTypeOdata
				outApi.Spec.Format = graphql.SpecFormatJSON // this field is required
			} else {
				outApi.Spec.Type = graphql.APISpecTypeOpenAPI // quite brave assumption that it will be OpenAPI
				outApi.Spec.Format = graphql.SpecFormatJSON
			}
		}

		if in.Api.Credentials != nil {
			outApi.DefaultAuth = &graphql.AuthInput{
				Credential: &graphql.CredentialDataInput{},
			}
			if in.Api.Credentials.BasicWithCSRF != nil {
				// TODO not mapped: in.Api.Credentials.BasicWithCSRF.CSRFInfo
				outApi.DefaultAuth.Credential.Basic = &graphql.BasicCredentialDataInput{
					Username: in.Api.Credentials.BasicWithCSRF.Username,
					Password: in.Api.Credentials.BasicWithCSRF.Password,
				}
			}

			if in.Api.Credentials.OauthWithCSRF != nil {
				// TODO not mapped: in.Api.Credentials.OauthWithCSRF.CSRFInfo
				outApi.DefaultAuth.Credential.Oauth = &graphql.OAuthCredentialDataInput{
					ClientID:     in.Api.Credentials.OauthWithCSRF.ClientID,
					ClientSecret: in.Api.Credentials.OauthWithCSRF.ClientSecret,
					URL:          in.Api.Credentials.OauthWithCSRF.URL,
				}
			}

			if in.Api.Credentials.CertificateGenWithCSRF != nil {
				// TODO not supported
			}
		}

		// old way of providing request headers
		if in.Api.Headers != nil {
			if outApi.DefaultAuth == nil {
				outApi.DefaultAuth = &graphql.AuthInput{}
			}
			h := (graphql.HttpHeaders)(*in.Api.Headers)
			outApi.DefaultAuth.AdditionalHeaders = &h
		}

		// old way of providing request headers
		if in.Api.QueryParameters != nil {
			if outApi.DefaultAuth == nil {
				outApi.DefaultAuth = &graphql.AuthInput{}
			}
			q := (graphql.QueryParams)(*in.Api.QueryParameters)
			outApi.DefaultAuth.AdditionalQueryParams = &q
		}

		// new way
		if in.Api.RequestParameters != nil {
			if outApi.DefaultAuth == nil {
				outApi.DefaultAuth = &graphql.AuthInput{}
			}
			if in.Api.RequestParameters.Headers != nil {
				h := (graphql.HttpHeaders)(*in.Api.RequestParameters.Headers)
				outApi.DefaultAuth.AdditionalHeaders = &h
			}
			if in.Api.RequestParameters.QueryParameters != nil {
				q := (graphql.QueryParams)(*in.Api.RequestParameters.QueryParameters)
				outApi.DefaultAuth.AdditionalQueryParams = &q
			}
		}

		if in.Api.Spec != nil {
			if outApi.Spec == nil {
				outApi.Spec = &graphql.APISpecInput{}
			}
			c := graphql.CLOB(string(in.Api.Spec))
			outApi.Spec.Data = &c
			if outApi.Spec.Type == "" {
				outApi.Spec.Type = graphql.APISpecTypeOpenAPI
			}
			if outApi.Spec.Format == "" {
				outApi.Spec.Format = graphql.SpecFormatJSON
			}
		}

		if in.Api.SpecificationUrl != "" {
			if outApi.Spec == nil {
				outApi.Spec = &graphql.APISpecInput{}
			}
			outApi.Spec.FetchRequest = &graphql.FetchRequestInput{
				URL: in.Api.SpecificationUrl,
			}

			if in.Api.SpecificationCredentials != nil || in.Api.SpecificationRequestParameters != nil {
				outApi.Spec.FetchRequest.Auth = &graphql.AuthInput{}
			}

			if in.Api.SpecificationCredentials != nil {
				if in.Api.SpecificationCredentials.Oauth != nil {
					inOauth := in.Api.SpecificationCredentials.Oauth
					outApi.Spec.FetchRequest.Auth.Credential = &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							URL:          inOauth.URL,
							ClientID:     inOauth.ClientID,
							ClientSecret: inOauth.ClientSecret,
						},
					}
				}
				if in.Api.SpecificationCredentials.Basic != nil {
					inBasic := in.Api.SpecificationCredentials.Basic
					outApi.Spec.FetchRequest.Auth.Credential = &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: inBasic.Username,
							Password: inBasic.Password,
						},
					}
				}
			}
		}

		if in.Api.SpecificationRequestParameters != nil {
			if in.Api.SpecificationRequestParameters.Headers != nil {
				h := (graphql.HttpHeaders)(*in.Api.SpecificationRequestParameters.Headers)
				outApi.Spec.FetchRequest.Auth.AdditionalHeaders = &h
			}
			if in.Api.SpecificationRequestParameters.QueryParameters != nil {
				q := (graphql.QueryParams)(*in.Api.SpecificationRequestParameters.QueryParameters)
				outApi.Spec.FetchRequest.Auth.AdditionalQueryParams = &q
			}
		}
		out.APIDefinitions = []*graphql.APIDefinitionInput{outApi}

	}

	if in.Events != nil && in.Events.Spec != nil {
		out.EventDefinitions = []*graphql.EventDefinitionInput{
			{
				Spec: &graphql.EventSpecInput{
					Data: ptrClob(graphql.CLOB(in.Events.Spec)),
				},
			},
		}
	}

	if in.Documentation != nil {
		// TODO later
	}

	if in.Labels != nil && *in.Labels != nil {
		for k, v := range *in.Labels {
			outLabels[k] = v
		}
	}

	out.Labels = getLabelsOrNil(outLabels)
	return out, nil
}

func (c *converter) GraphQLToDetailsModel(in graphql.ApplicationExt) (model.ServiceDetails, error) {
	// TODO
	out := model.ServiceDetails{
		Name: in.Name,
	}
	if in.ProviderName != nil {
		out.Provider = *in.ProviderName
	}

	if in.Description != nil {
		out.Description = *in.Description
	}

	outLabels := make(map[string]string)
	for k, v := range in.Labels {
		asString, ok := v.(string)
		if ok && !strings.HasPrefix(k, prefixUnmappedFields) {
			outLabels[k] = asString
		}
	}
	if len(outLabels) > 0 {
		out.Labels = &outLabels
	}

	out.Identifier = c.getUnmappedFromLabel(in.Labels, "identifier")

	out.ShortDescription = c.getUnmappedFromLabel(in.Labels, "shortDescription")

	// TODO out.Events
	// TODO out.Api
	// TODO out.Documentation

	if in.EventDefinitions.TotalCount != len(in.EventDefinitions.Data) {
		return model.ServiceDetails{}, fmt.Errorf("expected all event definitions [%d], got [%d]", in.EventDefinitions.TotalCount, len(in.EventDefinitions.Data))
	}

	if len(in.EventDefinitions.Data) > 1 {
		return model.ServiceDetails{}, fmt.Errorf("only one event definition is supported, but got [%d]", in.EventDefinitions.TotalCount)
	}

	// Event Definition
	if len(in.EventDefinitions.Data) == 1 {
		inDef := in.EventDefinitions.Data[0]
		// TODO fetch requests
		if inDef != nil && inDef.Spec != nil && inDef.Spec.Data != nil {

			out.Events = &model.Events{
				Spec: []byte(string(*inDef.Spec.Data)),
			}
		}

	}

	if in.APIDefinitions.TotalCount != len(in.APIDefinitions.Data) {
		return model.ServiceDetails{}, fmt.Errorf("expected all api definitinons [%d], got [%d]", in.APIDefinitions.TotalCount, len(in.APIDefinitions.Data))
	}

	if len(in.APIDefinitions.Data) > 1 {
		return model.ServiceDetails{}, fmt.Errorf("only one api definition is supported, but got [%d]", len(in.APIDefinitions.Data))
	}

	// API Definitions
	if len(in.APIDefinitions.Data) == 1 {
		inDef := in.APIDefinitions.Data[0]
		out.Api = &model.API{
			TargetUrl: inDef.TargetURL,
		}

		if inDef.Spec != nil {
			out.Api.ApiType = string(inDef.Spec.Type) // TODO we have enums, they have strings, OMG
		}

		if inDef.DefaultAuth != nil && inDef.DefaultAuth.Credential != nil {
			if out.Api.Credentials == nil {
				out.Api.Credentials = &model.CredentialsWithCSRF{}
			}
			switch actual := inDef.DefaultAuth.Credential.(type) {
			case graphql.BasicCredentialData:
				out.Api.Credentials.BasicWithCSRF = &model.BasicAuthWithCSRF{
					BasicAuth: model.BasicAuth{
						Username: actual.Username,
						Password: actual.Password,
					},
				}
			case graphql.OAuthCredentialData:
				out.Api.Credentials.OauthWithCSRF = &model.OauthWithCSRF{
					Oauth: model.Oauth{
						URL:          actual.URL,
						ClientID:     actual.ClientID,
						ClientSecret: actual.ClientSecret,
					},
				}
			}
		}

		if inDef.DefaultAuth != nil && inDef.DefaultAuth.AdditionalHeaders != nil {
			in := *inDef.DefaultAuth.AdditionalHeaders
			out.Api.Headers = &map[string][]string{}
			if out.Api.RequestParameters == nil {
				out.Api.RequestParameters = &model.RequestParameters{}
			}
			if out.Api.RequestParameters.Headers == nil {
				out.Api.RequestParameters.Headers = &map[string][]string{}
			}

			for k, v := range in {
				(*out.Api.Headers)[k] = v
				(*out.Api.RequestParameters.Headers)[k] = v
			}
		}

		if inDef.DefaultAuth != nil && inDef.DefaultAuth.AdditionalQueryParams != nil {
			in := *inDef.DefaultAuth.AdditionalQueryParams
			out.Api.QueryParameters = &map[string][]string{}
			if out.Api.RequestParameters == nil {
				out.Api.RequestParameters = &model.RequestParameters{}
			}

			if out.Api.RequestParameters.QueryParameters == nil {
				out.Api.RequestParameters.QueryParameters = &map[string][]string{}
			}

			for k, v := range in {
				(*out.Api.QueryParameters)[k] = v
				(*out.Api.RequestParameters.QueryParameters)[k] = v
			}
		}

		if inDef.Spec != nil && inDef.Spec.FetchRequest != nil {
			out.Api.SpecificationUrl = inDef.Spec.FetchRequest.URL
			if inDef.Spec.FetchRequest.Auth != nil {
				if inDef.Spec.FetchRequest.Auth.AdditionalQueryParams != nil || inDef.Spec.FetchRequest.Auth.AdditionalHeaders != nil {
					out.Api.SpecificationRequestParameters = &model.RequestParameters{}
				}

				if inDef.Spec.FetchRequest.Auth.AdditionalQueryParams != nil {
					asMap := ((map[string][]string)(*inDef.Spec.FetchRequest.Auth.AdditionalQueryParams))
					out.Api.SpecificationRequestParameters.QueryParameters = &asMap
				}

				if inDef.Spec.FetchRequest.Auth.AdditionalHeaders != nil {
					asMap := ((map[string][]string)(*inDef.Spec.FetchRequest.Auth.AdditionalHeaders))
					out.Api.SpecificationRequestParameters.Headers = &asMap
				}

				basic, isBasic := (inDef.Spec.FetchRequest.Auth.Credential).(graphql.BasicCredentialData)
				oauth, isOauth := (inDef.Spec.FetchRequest.Auth.Credential).(graphql.OAuthCredentialData)
				outCred := &model.Credentials{}
				if isOauth || isBasic {
					if isBasic {
						outCred.Basic = &model.BasicAuth{
							Username: basic.Username,
							Password: basic.Password,
						}
					}
					if isOauth {
						outCred.Oauth = &model.Oauth{
							URL:               oauth.URL,
							ClientID:          oauth.ClientID,
							ClientSecret:      oauth.ClientSecret,
							RequestParameters: nil, // TODO not supported
						}
					}
					out.Api.SpecificationCredentials = outCred
				}

			}

		}

	}

	if in.Documents.TotalCount != len(in.Documents.Data) {
		return model.ServiceDetails{}, fmt.Errorf("expected all documents [%d], got [%d]", in.Documents.TotalCount, len(in.Documents.Data))
	}

	if len(in.Documents.Data) > 1 {
		return model.ServiceDetails{}, fmt.Errorf("only one Documentation for Application is supported, but got [%d]", len(in.Documents.Data))
	}

	// TODO docs
	return out, nil

}

// GET /v1/metadata/services
func (c *converter) GraphQLToModel(in graphql.ApplicationExt) (model.Service, error) {
	out := model.Service{
		Name: in.Name,
		ID:   in.ID,
	}

	if in.ProviderName != nil {
		out.Provider = *in.ProviderName
	}

	if in.Description != nil {
		out.Description = *in.Description
	}

	outLabels := make(map[string]string)
	for k, v := range in.Labels {
		asString, ok := v.(string)
		if ok && !strings.HasPrefix(k, prefixUnmappedFields) {
			outLabels[k] = asString
		}
	}
	if len(outLabels) > 0 {
		out.Labels = &outLabels
	}

	out.Identifier = c.getUnmappedFromLabel(in.Labels, "identifier")

	return out, nil
}

func (c *converter) getUnmappedFromLabel(labels graphql.Labels, field string) string {
	val := labels[prefixUnmappedFields+field]
	asString, ok := val.(string)
	if ok {
		return asString
	}
	return ""
}

func getLabelsOrNil(in graphql.Labels) *graphql.Labels {
	if len(in) == 0 {
		return nil
	}
	return &in
}

func ptrStringOrNilForEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrClob(in graphql.CLOB) *graphql.CLOB {
	return &in
}
