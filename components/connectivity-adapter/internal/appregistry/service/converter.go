package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	// fields that exist in Application Registry but don't exist in Director GraphQL API, are stored in labels with special prefix
	prefixUnmappedFields          = "compatibility_unmapped_fields/"
	unmappedFieldIdentifier       = prefixUnmappedFields + "identifier"
	unmappedFieldShortDescription = prefixUnmappedFields + "shortDescription"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) DetailsToGraphQLInput(deprecated model.ServiceDetails) (graphql.ApplicationRegisterInput, error) {

	outLabels := graphql.Labels{}

	out := graphql.ApplicationRegisterInput{
		Name:         deprecated.Name,
		Description:  ptrStringOrNilForEmpty(deprecated.Description),
		ProviderName: ptrStringOrNilForEmpty(deprecated.Provider),
	}

	if deprecated.ShortDescription != "" {
		outLabels[unmappedFieldShortDescription] = deprecated.ShortDescription
	}
	if deprecated.Identifier != "" {
		outLabels[unmappedFieldIdentifier] = deprecated.Identifier
	}
	if deprecated.Api != nil {

		outApi := &graphql.APIDefinitionInput{
			TargetURL: deprecated.Api.TargetUrl,
		}

		if deprecated.Api.ApiType != "" {
			if outApi.Spec == nil {
				outApi.Spec = &graphql.APISpecInput{}
			}

			if strings.ToLower(deprecated.Api.ApiType) == "odata" {
				outApi.Spec.Type = graphql.APISpecTypeOdata
			} else {
				outApi.Spec.Type = graphql.APISpecTypeOpenAPI // quite brave assumption that it will be OpenAPI
			}
		}

		if deprecated.Api.Credentials != nil {
			outApi.DefaultAuth = &graphql.AuthInput{
				Credential: &graphql.CredentialDataInput{},
			}
			if deprecated.Api.Credentials.BasicWithCSRF != nil {
				// TODO later: not mapped: deprecated.Api.Credentials.BasicWithCSRF.CSRFInfo
				outApi.DefaultAuth.Credential.Basic = &graphql.BasicCredentialDataInput{
					Username: deprecated.Api.Credentials.BasicWithCSRF.Username,
					Password: deprecated.Api.Credentials.BasicWithCSRF.Password,
				}
			}

			if deprecated.Api.Credentials.OauthWithCSRF != nil {
				// TODO later: not mapped: deprecated.Api.Credentials.OauthWithCSRF.CSRFInfo
				outApi.DefaultAuth.Credential.Oauth = &graphql.OAuthCredentialDataInput{
					ClientID:     deprecated.Api.Credentials.OauthWithCSRF.ClientID,
					ClientSecret: deprecated.Api.Credentials.OauthWithCSRF.ClientSecret,
					URL:          deprecated.Api.Credentials.OauthWithCSRF.URL,
				}
			}

			if deprecated.Api.Credentials.CertificateGenWithCSRF != nil {
				// TODO not supported
			}
		}

		// old way of providing request headers
		if deprecated.Api.Headers != nil {
			if outApi.DefaultAuth == nil {
				outApi.DefaultAuth = &graphql.AuthInput{}
			}
			h := (graphql.HttpHeaders)(*deprecated.Api.Headers)
			outApi.DefaultAuth.AdditionalHeaders = &h
		}

		// old way of providing request headers
		if deprecated.Api.QueryParameters != nil {
			if outApi.DefaultAuth == nil {
				outApi.DefaultAuth = &graphql.AuthInput{}
			}
			q := (graphql.QueryParams)(*deprecated.Api.QueryParameters)
			outApi.DefaultAuth.AdditionalQueryParams = &q
		}

		// new way
		if deprecated.Api.RequestParameters != nil {
			if outApi.DefaultAuth == nil {
				outApi.DefaultAuth = &graphql.AuthInput{}
			}
			if deprecated.Api.RequestParameters.Headers != nil {
				h := (graphql.HttpHeaders)(*deprecated.Api.RequestParameters.Headers)
				outApi.DefaultAuth.AdditionalHeaders = &h
			}
			if deprecated.Api.RequestParameters.QueryParameters != nil {
				q := (graphql.QueryParams)(*deprecated.Api.RequestParameters.QueryParameters)
				outApi.DefaultAuth.AdditionalQueryParams = &q
			}
		}

		if deprecated.Api.Spec != nil {
			if outApi.Spec == nil {
				outApi.Spec = &graphql.APISpecInput{}
			}
			asClob := graphql.CLOB(string(deprecated.Api.Spec))
			outApi.Spec.Data = &asClob
			if outApi.Spec.Type == "" {
				outApi.Spec.Type = graphql.APISpecTypeOpenAPI
			}
			if outApi.Spec.Format == "" {
				if c.isXML([]byte(deprecated.Api.Spec)) {
					outApi.Spec.Format = graphql.SpecFormatXML
				} else if c.isJSON([]byte(deprecated.Api.Spec)) {
					outApi.Spec.Format = graphql.SpecFormatJSON
				} else {
					outApi.Spec.Format = graphql.SpecFormatYaml
				}

			}
		}

		if deprecated.Api.SpecificationUrl != "" {
			if outApi.Spec == nil {
				outApi.Spec = &graphql.APISpecInput{}
			}
			outApi.Spec.FetchRequest = &graphql.FetchRequestInput{
				URL: deprecated.Api.SpecificationUrl,
			}

			if deprecated.Api.SpecificationCredentials != nil || deprecated.Api.SpecificationRequestParameters != nil {
				outApi.Spec.FetchRequest.Auth = &graphql.AuthInput{}
			}

			if deprecated.Api.SpecificationCredentials != nil {
				if deprecated.Api.SpecificationCredentials.Oauth != nil {
					inOauth := deprecated.Api.SpecificationCredentials.Oauth
					outApi.Spec.FetchRequest.Auth.Credential = &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							URL:          inOauth.URL,
							ClientID:     inOauth.ClientID,
							ClientSecret: inOauth.ClientSecret,
						},
					}
				}
				if deprecated.Api.SpecificationCredentials.Basic != nil {
					inBasic := deprecated.Api.SpecificationCredentials.Basic
					outApi.Spec.FetchRequest.Auth.Credential = &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: inBasic.Username,
							Password: inBasic.Password,
						},
					}
				}
			}
		}

		if deprecated.Api.SpecificationRequestParameters != nil {
			if deprecated.Api.SpecificationRequestParameters.Headers != nil {
				h := (graphql.HttpHeaders)(*deprecated.Api.SpecificationRequestParameters.Headers)
				outApi.Spec.FetchRequest.Auth.AdditionalHeaders = &h
			}
			if deprecated.Api.SpecificationRequestParameters.QueryParameters != nil {
				q := (graphql.QueryParams)(*deprecated.Api.SpecificationRequestParameters.QueryParameters)
				outApi.Spec.FetchRequest.Auth.AdditionalQueryParams = &q
			}
		}
		out.APIDefinitions = []*graphql.APIDefinitionInput{outApi}

	}

	if deprecated.Events != nil && deprecated.Events.Spec != nil {
		out.EventDefinitions = []*graphql.EventDefinitionInput{
			{
				Spec: &graphql.EventSpecInput{
					Data: ptrClob(graphql.CLOB(deprecated.Events.Spec)),
				},
			},
		}
	}

	if deprecated.Documentation != nil {
		// TODO later
	}

	if deprecated.Labels != nil && *deprecated.Labels != nil {
		for k, v := range *deprecated.Labels {
			outLabels[k] = v
		}
	}

	out.Labels = getLabelsOrNilIfEmpty(outLabels)
	return out, nil
}

func (c *converter) GraphQLToDetailsModel(in graphql.ApplicationExt) (model.ServiceDetails, error) {
	outDeprecated := model.ServiceDetails{
		Name: in.Name,
	}
	if in.ProviderName != nil {
		outDeprecated.Provider = *in.ProviderName
	}

	if in.Description != nil {
		outDeprecated.Description = *in.Description
	}

	outLabels := make(map[string]string)
	for k, v := range in.Labels {
		asString, ok := v.(string)
		if ok && !strings.HasPrefix(k, prefixUnmappedFields) {
			outLabels[k] = asString
		}
	}
	if len(outLabels) > 0 {
		outDeprecated.Labels = &outLabels
	}

	outDeprecated.Identifier = c.getUnmappedFromLabel(in.Labels, "identifier")

	outDeprecated.ShortDescription = c.getUnmappedFromLabel(in.Labels, "shortDescription")

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
			outDeprecated.Events = &model.Events{
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
		outDeprecated.Api = &model.API{
			TargetUrl: inDef.TargetURL,
		}

		if inDef.Spec != nil {
			outDeprecated.Api.ApiType = string(inDef.Spec.Type)
		}

		if inDef.DefaultAuth != nil && inDef.DefaultAuth.Credential != nil {
			if outDeprecated.Api.Credentials == nil {
				outDeprecated.Api.Credentials = &model.CredentialsWithCSRF{}
			}
			switch actual := inDef.DefaultAuth.Credential.(type) {
			case graphql.BasicCredentialData:
				outDeprecated.Api.Credentials.BasicWithCSRF = &model.BasicAuthWithCSRF{
					BasicAuth: model.BasicAuth{
						Username: actual.Username,
						Password: actual.Password,
					},
				}
			case graphql.OAuthCredentialData:
				outDeprecated.Api.Credentials.OauthWithCSRF = &model.OauthWithCSRF{
					Oauth: model.Oauth{
						URL:          actual.URL,
						ClientID:     actual.ClientID,
						ClientSecret: actual.ClientSecret,
					},
				}
			}
		}

		if inDef.DefaultAuth != nil && inDef.DefaultAuth.AdditionalHeaders != nil {
			inHeaders := *inDef.DefaultAuth.AdditionalHeaders
			outDeprecated.Api.Headers = &map[string][]string{}
			if outDeprecated.Api.RequestParameters == nil {
				outDeprecated.Api.RequestParameters = &model.RequestParameters{}
			}
			if outDeprecated.Api.RequestParameters.Headers == nil {
				outDeprecated.Api.RequestParameters.Headers = &map[string][]string{}
			}

			for k, v := range inHeaders {
				(*outDeprecated.Api.Headers)[k] = v
				(*outDeprecated.Api.RequestParameters.Headers)[k] = v
			}
		}

		if inDef.DefaultAuth != nil && inDef.DefaultAuth.AdditionalQueryParams != nil {
			in := *inDef.DefaultAuth.AdditionalQueryParams
			outDeprecated.Api.QueryParameters = &map[string][]string{}
			if outDeprecated.Api.RequestParameters == nil {
				outDeprecated.Api.RequestParameters = &model.RequestParameters{}
			}

			if outDeprecated.Api.RequestParameters.QueryParameters == nil {
				outDeprecated.Api.RequestParameters.QueryParameters = &map[string][]string{}
			}

			for k, v := range in {
				(*outDeprecated.Api.QueryParameters)[k] = v
				(*outDeprecated.Api.RequestParameters.QueryParameters)[k] = v
			}
		}

		if inDef.Spec != nil && inDef.Spec.FetchRequest != nil {
			outDeprecated.Api.SpecificationUrl = inDef.Spec.FetchRequest.URL
			if inDef.Spec.FetchRequest.Auth != nil {
				if inDef.Spec.FetchRequest.Auth.AdditionalQueryParams != nil || inDef.Spec.FetchRequest.Auth.AdditionalHeaders != nil {
					outDeprecated.Api.SpecificationRequestParameters = &model.RequestParameters{}
				}

				if inDef.Spec.FetchRequest.Auth.AdditionalQueryParams != nil {
					asMap := (map[string][]string)(*inDef.Spec.FetchRequest.Auth.AdditionalQueryParams)
					outDeprecated.Api.SpecificationRequestParameters.QueryParameters = &asMap
				}

				if inDef.Spec.FetchRequest.Auth.AdditionalHeaders != nil {
					asMap := (map[string][]string)(*inDef.Spec.FetchRequest.Auth.AdditionalHeaders)
					outDeprecated.Api.SpecificationRequestParameters.Headers = &asMap
				}

				basic, isBasic := (inDef.Spec.FetchRequest.Auth.Credential).(graphql.BasicCredentialData)
				oauth, isOauth := (inDef.Spec.FetchRequest.Auth.Credential).(graphql.OAuthCredentialData)

				if isOauth || isBasic {
					outCred := &model.Credentials{}
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
					outDeprecated.Api.SpecificationCredentials = outCred
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

	// TODO docs later
	return outDeprecated, nil

}

func (c *converter) GraphQLToModel(in graphql.ApplicationExt) (model.Service, error) {
	outDeprecated := model.Service{
		Name: in.Name,
		ID:   in.ID,
	}

	if in.ProviderName != nil {
		outDeprecated.Provider = *in.ProviderName
	}

	if in.Description != nil {
		outDeprecated.Description = *in.Description
	}

	outLabels := make(map[string]string)
	for k, v := range in.Labels {
		asString, ok := v.(string)
		if ok && !strings.HasPrefix(k, prefixUnmappedFields) {
			outLabels[k] = asString
		}
	}
	if len(outLabels) > 0 {
		outDeprecated.Labels = &outLabels
	}

	outDeprecated.Identifier = c.getUnmappedFromLabel(in.Labels, "identifier")

	return outDeprecated, nil
}

func (c *converter) getUnmappedFromLabel(labels graphql.Labels, field string) string {
	val := labels[prefixUnmappedFields+field]
	asString, ok := val.(string)
	if ok {
		return asString
	}
	return ""
}

func getLabelsOrNilIfEmpty(in graphql.Labels) *graphql.Labels {
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

func (c *converter) isXML(content []byte) bool {
	openingIndex := strings.Index(string(content), "<")
	closingIndex := strings.Index(string(content), ">")

	return openingIndex != -1 && openingIndex < closingIndex
}

func (c *converter) isJSON(content []byte) bool {
	out := map[string]interface{}{}
	err := json.Unmarshal(content, &out)
	return err == nil
}
