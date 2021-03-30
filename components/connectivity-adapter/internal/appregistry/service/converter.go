package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

const oDataSpecFormat = "%s/$metadata"
const oDataSpecType = "odata"

func (c *converter) DetailsToGraphQLCreateInput(deprecated model.ServiceDetails) (graphql.BundleCreateInput, error) {
	out := graphql.BundleCreateInput{}
	out.Name = deprecated.Name
	if deprecated.Description != "" {
		out.Description = &deprecated.Description
	}

	defaultInstanceAuth := &graphql.AuthInput{}
	if deprecated.Api != nil {
		var apiDef *graphql.APIDefinitionInput
		apiDef = &graphql.APIDefinitionInput{
			Name:      deprecated.Name,
			TargetURL: deprecated.Api.TargetUrl,
		}

		if deprecated.Description != "" {
			apiDef.Description = &deprecated.Description
		}

		if deprecated.Api.ApiType != "" {
			if apiDef.Spec == nil {
				apiDef.Spec = &graphql.APISpecInput{}
			}

			if strings.ToLower(deprecated.Api.ApiType) == "odata" {
				apiDef.Spec.Type = graphql.APISpecTypeOdata
			} else {
				apiDef.Spec.Type = graphql.APISpecTypeOpenAPI // quite brave assumption that it will be OpenAPI
			}
			apiDef.Spec.Format = graphql.SpecFormatYaml
		}

		if deprecated.Api.Credentials != nil && (deprecated.Api.Credentials.OauthWithCSRF != nil || deprecated.Api.Credentials.BasicWithCSRF != nil) {
			defaultInstanceAuth.Credential = &graphql.CredentialDataInput{}

			if deprecated.Api.Credentials.BasicWithCSRF != nil {
				defaultInstanceAuth.Credential.Basic = &graphql.BasicCredentialDataInput{
					Username: deprecated.Api.Credentials.BasicWithCSRF.Username,
					Password: deprecated.Api.Credentials.BasicWithCSRF.Password,
				}

				if deprecated.Api.Credentials.BasicWithCSRF.CSRFInfo != nil {
					defaultInstanceAuth.RequestAuth = &graphql.CredentialRequestAuthInput{
						Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
							TokenEndpointURL: deprecated.Api.Credentials.BasicWithCSRF.CSRFInfo.TokenEndpointURL,
						}}
				}
			}

			if deprecated.Api.Credentials.OauthWithCSRF != nil {
				defaultInstanceAuth.Credential.Oauth = &graphql.OAuthCredentialDataInput{
					ClientID:     deprecated.Api.Credentials.OauthWithCSRF.ClientID,
					ClientSecret: deprecated.Api.Credentials.OauthWithCSRF.ClientSecret,
					URL:          deprecated.Api.Credentials.OauthWithCSRF.URL,
				}

				if deprecated.Api.Credentials.OauthWithCSRF.CSRFInfo != nil {
					defaultInstanceAuth.RequestAuth = &graphql.CredentialRequestAuthInput{
						Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
							TokenEndpointURL: deprecated.Api.Credentials.OauthWithCSRF.CSRFInfo.TokenEndpointURL,
						}}
				}
			}

			// TODO deprecated.Api.Credentials.CertificateGenWithCSRF not supported
		}

		// old way of providing request headers
		if deprecated.Api.Headers != nil {
			h, err := graphql.NewHttpHeadersSerialized(*deprecated.Api.Headers)
			if err != nil {
				return graphql.BundleCreateInput{}, err
			}

			defaultInstanceAuth.AdditionalHeadersSerialized = &h
		}

		// old way of providing request headers
		if deprecated.Api.QueryParameters != nil {
			q, err := graphql.NewQueryParamsSerialized(*deprecated.Api.QueryParameters)
			if err != nil {
				return graphql.BundleCreateInput{}, err
			}

			defaultInstanceAuth.AdditionalQueryParamsSerialized = &q
		}

		// new way
		if deprecated.Api.RequestParameters != nil {
			if deprecated.Api.RequestParameters.Headers != nil {
				h, err := graphql.NewHttpHeadersSerialized(*deprecated.Api.RequestParameters.Headers)
				if err != nil {
					return graphql.BundleCreateInput{}, err
				}

				defaultInstanceAuth.AdditionalHeadersSerialized = &h
			}
			if deprecated.Api.RequestParameters.QueryParameters != nil {
				q, err := graphql.NewQueryParamsSerialized(*deprecated.Api.RequestParameters.QueryParameters)
				if err != nil {
					return graphql.BundleCreateInput{}, err
				}

				defaultInstanceAuth.AdditionalQueryParamsSerialized = &q
			}
		}

		if deprecated.Api.Spec != nil {
			if apiDef.Spec == nil {
				apiDef.Spec = &graphql.APISpecInput{}
			}
			apiDef.Spec.Data = ptrClob(graphql.CLOB(*deprecated.Api.Spec))
			if apiDef.Spec.Type == "" {
				apiDef.Spec.Type = graphql.APISpecTypeOpenAPI
			}

			if model.IsXML(string(*deprecated.Api.Spec)) {
				apiDef.Spec.Format = graphql.SpecFormatXML
			} else if c.isJSON(string(*deprecated.Api.Spec)) {
				apiDef.Spec.Format = graphql.SpecFormatJSON
			} else {
				apiDef.Spec.Format = graphql.SpecFormatYaml
			}
		}

		if deprecated.Api.Spec == nil { // TODO provide test for that

			lowercaseDeprecatedAPIType := strings.ToLower(deprecated.Api.ApiType)
			if deprecated.Api.SpecificationUrl != "" || deprecated.Api.SpecificationCredentials != nil || deprecated.Api.SpecificationRequestParameters != nil || lowercaseDeprecatedAPIType == oDataSpecType {
				if apiDef.Spec == nil {
					apiDef.Spec = &graphql.APISpecInput{}
				}

				apiDef.Spec.Type = toNewSpecType(deprecated.Api.ApiType)
				if lowercaseDeprecatedAPIType == oDataSpecType {
					apiDef.Spec.Format = graphql.SpecFormatXML
				} else {
					apiDef.Spec.Format = graphql.SpecFormatJSON
				}

				url := deprecated.Api.SpecificationUrl
				if lowercaseDeprecatedAPIType == oDataSpecType && url == "" {
					targetURL := strings.TrimSuffix(apiDef.TargetURL, "/")
					url = fmt.Sprintf(oDataSpecFormat, targetURL)
				}

				apiDef.Spec.FetchRequest = &graphql.FetchRequestInput{
					URL: url,
				}
			}

			if deprecated.Api.SpecificationCredentials != nil || deprecated.Api.SpecificationRequestParameters != nil {
				apiDef.Spec.FetchRequest.Auth = &graphql.AuthInput{}
			}

			if deprecated.Api.SpecificationCredentials != nil {
				if deprecated.Api.SpecificationCredentials.Oauth != nil {
					inOauth := deprecated.Api.SpecificationCredentials.Oauth
					apiDef.Spec.FetchRequest.Auth.Credential = &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							URL:          inOauth.URL,
							ClientID:     inOauth.ClientID,
							ClientSecret: inOauth.ClientSecret,
						},
					}
				}
				if deprecated.Api.SpecificationCredentials.Basic != nil {
					inBasic := deprecated.Api.SpecificationCredentials.Basic
					apiDef.Spec.FetchRequest.Auth.Credential = &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: inBasic.Username,
							Password: inBasic.Password,
						},
					}
				}
			}

			if deprecated.Api.SpecificationRequestParameters != nil && apiDef.Spec.FetchRequest != nil {
				if deprecated.Api.SpecificationRequestParameters.Headers != nil {
					h, err := graphql.NewHttpHeadersSerialized(*deprecated.Api.SpecificationRequestParameters.Headers)
					if err != nil {
						return graphql.BundleCreateInput{}, err
					}

					apiDef.Spec.FetchRequest.Auth.AdditionalHeadersSerialized = &h
				}
				if deprecated.Api.SpecificationRequestParameters.QueryParameters != nil {
					q, err := graphql.NewQueryParamsSerialized(*deprecated.Api.SpecificationRequestParameters.QueryParameters)
					if err != nil {
						return graphql.BundleCreateInput{}, err
					}

					apiDef.Spec.FetchRequest.Auth.AdditionalQueryParamsSerialized = &q
				}
			}
		}

		out.APIDefinitions = []*graphql.APIDefinitionInput{apiDef}
	}

	out.DefaultInstanceAuth = defaultInstanceAuth

	if deprecated.Events != nil && deprecated.Events.Spec != nil {
		var eventDef *graphql.EventDefinitionInput

		// TODO add tests
		var format graphql.SpecFormat
		if model.IsXML(string(*deprecated.Events.Spec)) {
			format = graphql.SpecFormatXML
		} else if c.isJSON(string(*deprecated.Events.Spec)) {
			format = graphql.SpecFormatJSON
		} else {
			format = graphql.SpecFormatYaml
		}

		eventDef =
			&graphql.EventDefinitionInput{
				Name: deprecated.Name,
				Spec: &graphql.EventSpecInput{
					Data:   ptrClob(graphql.CLOB(*deprecated.Events.Spec)),
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: format,
				},
			}

		if deprecated.Description != "" {
			eventDef.Description = &deprecated.Description
		}

		out.EventDefinitions = []*graphql.EventDefinitionInput{eventDef}
	}

	out.Documents = c.legacyDocumentationToDocuments(deprecated.Documentation)

	return out, nil
}

func (c *converter) legacyDocumentationToDocuments(legacyDocumentation *model.Documentation) []*graphql.DocumentInput {
	if legacyDocumentation == nil {
		return nil
	}

	var docs []*graphql.DocumentInput
	for _, legacyDoc := range legacyDocumentation.Docs {
		data := graphql.CLOB(legacyDoc.Source)
		doc := &graphql.DocumentInput{
			Title:       legacyDoc.Title,
			DisplayName: legacyDoc.Title,
			Description: " ",                            // to workaround our strict validation
			Format:      graphql.DocumentFormatMarkdown, // we don't have any other format in our API anyway
			Kind:        nil,
			Data:        &data,
		}

		docs = append(docs, doc)
	}

	return docs
}

func (c *converter) documentsToLegacyDocumentation(documents []*graphql.DocumentExt) *model.Documentation {
	if documents == nil || len(documents) == 0 {
		return nil
	}

	var legacyDocs []model.DocsObject
	for _, doc := range documents {
		var source string
		if doc.Data != nil {
			source = string(*doc.Data)
		}

		legacyDoc := model.DocsObject{
			Title:  doc.Title,
			Type:   ".md", // we don't have any other format in our API anyway
			Source: source,
		}

		legacyDocs = append(legacyDocs, legacyDoc)
	}

	return &model.Documentation{
		DisplayName: "ServiceDocumentation",
		Description: "Documents for legacy Service",
		Tags:        nil,
		Docs:        legacyDocs,
	}
}

func toNewSpecType(apiType string) graphql.APISpecType {
	switch strings.ToLower(apiType) {
	case "openapi":
		return graphql.APISpecTypeOpenAPI
	case "odata":
		return graphql.APISpecTypeOdata
	default:
		return graphql.APISpecTypeOpenAPI
	}
}

func (c *converter) GraphQLToServiceDetails(in graphql.BundleExt, legacyServiceReference LegacyServiceReference) (model.ServiceDetails, error) {
	var desc string
	if in.Description != nil {
		desc = *in.Description
	}
	outDeprecated := model.ServiceDetails{
		Name:        in.Name,
		Description: desc,
		Identifier:  legacyServiceReference.Identifier,
		Labels:      &map[string]string{},
	}
	if in.APIDefinitions.Data != nil && len(in.APIDefinitions.Data) > 0 {
		if len(in.APIDefinitions.Data) > 1 {
			return model.ServiceDetails{}, errors.New("found more API Definitions than one supported for legacy Service")
		}
		var apiDef = in.APIDefinitions.Data[0]

		outDeprecated.Api = &model.API{
			TargetUrl: apiDef.TargetURL,
		}

		if apiDef.Description != nil {
			outDeprecated.Description = *apiDef.Description
		}

		if apiDef.Spec != nil {
			outDeprecated.Api.ApiType = string(apiDef.Spec.Type)
			if apiDef.Spec.Data != nil {
				outDeprecated.Api.Spec = ptrSpecResponse(model.SpecResponse(*apiDef.Spec.Data))
			}
		}

		if in.DefaultInstanceAuth != nil && in.DefaultInstanceAuth.Credential != nil {
			basicCreds, isBasic := in.DefaultInstanceAuth.Credential.(*graphql.BasicCredentialData)
			oauthCreds, isOauth := in.DefaultInstanceAuth.Credential.(*graphql.OAuthCredentialData)

			if (isBasic && basicCreds != nil) || (isOauth && oauthCreds != nil) {
				if outDeprecated.Api.Credentials == nil {
					outDeprecated.Api.Credentials = &model.CredentialsWithCSRF{}
				}
				switch actual := in.DefaultInstanceAuth.Credential.(type) {
				case *graphql.BasicCredentialData:
					outDeprecated.Api.Credentials.BasicWithCSRF = &model.BasicAuthWithCSRF{
						BasicAuth: model.BasicAuth{
							Username: actual.Username,
							Password: actual.Password,
						},
					}
					if in.DefaultInstanceAuth.RequestAuth != nil && in.DefaultInstanceAuth.RequestAuth.Csrf != nil {
						outDeprecated.Api.Credentials.BasicWithCSRF.CSRFInfo = &model.CSRFInfo{
							TokenEndpointURL: in.DefaultInstanceAuth.RequestAuth.Csrf.TokenEndpointURL,
						}
					}

				case *graphql.OAuthCredentialData:
					outDeprecated.Api.Credentials.OauthWithCSRF = &model.OauthWithCSRF{
						Oauth: model.Oauth{
							URL:          actual.URL,
							ClientID:     actual.ClientID,
							ClientSecret: actual.ClientSecret,
						},
					}
					if in.DefaultInstanceAuth.RequestAuth != nil && in.DefaultInstanceAuth.RequestAuth.Csrf != nil {
						outDeprecated.Api.Credentials.OauthWithCSRF.CSRFInfo = &model.CSRFInfo{
							TokenEndpointURL: in.DefaultInstanceAuth.RequestAuth.Csrf.TokenEndpointURL,
						}
					}
				}
			}
		}

		if in.DefaultInstanceAuth != nil && in.DefaultInstanceAuth.AdditionalHeaders != nil {
			inHeaders := in.DefaultInstanceAuth.AdditionalHeaders
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

		if in.DefaultInstanceAuth != nil && in.DefaultInstanceAuth.AdditionalQueryParams != nil {
			in := in.DefaultInstanceAuth.AdditionalQueryParams
			outQueryParameters := &map[string][]string{}

			for k, v := range in {
				(*outQueryParameters)[k] = v
			}
			outDeprecated.Api.QueryParameters = outQueryParameters
			if outDeprecated.Api.RequestParameters == nil {
				outDeprecated.Api.RequestParameters = &model.RequestParameters{}
			}
			outDeprecated.Api.RequestParameters.QueryParameters = outQueryParameters
		}

		if apiDef.Spec != nil && apiDef.Spec.FetchRequest != nil {
			outDeprecated.Api.SpecificationUrl = apiDef.Spec.FetchRequest.URL
			if apiDef.Spec.FetchRequest.Auth != nil {
				if apiDef.Spec.FetchRequest.Auth.AdditionalQueryParams != nil || apiDef.Spec.FetchRequest.Auth.AdditionalHeaders != nil {
					outDeprecated.Api.SpecificationRequestParameters = &model.RequestParameters{}
				}

				if apiDef.Spec.FetchRequest.Auth.AdditionalQueryParams != nil {
					asMap := (map[string][]string)(apiDef.Spec.FetchRequest.Auth.AdditionalQueryParams)
					outDeprecated.Api.SpecificationRequestParameters.QueryParameters = &asMap
				}

				if apiDef.Spec.FetchRequest.Auth.AdditionalHeaders != nil {
					asMap := (map[string][]string)(apiDef.Spec.FetchRequest.Auth.AdditionalHeaders)
					outDeprecated.Api.SpecificationRequestParameters.Headers = &asMap
				}

				basic, isBasic := (apiDef.Spec.FetchRequest.Auth.Credential).(*graphql.BasicCredentialData)
				oauth, isOauth := (apiDef.Spec.FetchRequest.Auth.Credential).(*graphql.OAuthCredentialData)

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
	if in.EventDefinitions.Data != nil && len(in.EventDefinitions.Data) > 0 {
		if len(in.EventDefinitions.Data) > 1 {
			return model.ServiceDetails{}, errors.New("found more Event Definitions than one supported for legacy Service")
		}
		var eventDef = in.EventDefinitions.Data[0]

		if eventDef.Description != nil {
			outDeprecated.Description = *eventDef.Description
		}

		if eventDef.Spec != nil && eventDef.Spec.Data != nil {
			outDeprecated.Events = &model.Events{
				Spec: ptrSpecResponse(model.SpecResponse(*eventDef.Spec.Data)),
			}
		}
		//TODO: convert also fetchRequest
	}

	if in.Documents.Data != nil {
		outDeprecated.Documentation = c.documentsToLegacyDocumentation(in.Documents.Data)
	}

	return outDeprecated, nil
}

func (c *converter) GraphQLCreateInputToUpdateInput(in graphql.BundleCreateInput) graphql.BundleUpdateInput {
	return graphql.BundleUpdateInput{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: in.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            in.DefaultInstanceAuth,
	}
}

func (c *converter) ServiceDetailsToService(in model.ServiceDetails, serviceID string) (model.Service, error) {
	return model.Service{
		ID:          serviceID,
		Provider:    in.Provider,
		Name:        in.Name,
		Description: in.Description,
		Identifier:  in.Identifier,
		Labels:      in.Labels,
	}, nil
}

func (c *converter) isJSON(content string) bool {
	out := map[string]interface{}{}
	err := json.Unmarshal([]byte(content), &out)
	return err == nil
}

func ptrClob(in graphql.CLOB) *graphql.CLOB {
	return &in
}

func ptrSpecResponse(in model.SpecResponse) *model.SpecResponse {
	return &in
}
