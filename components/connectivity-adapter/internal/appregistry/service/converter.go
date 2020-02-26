package service

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) DetailsToGraphQLInput(id string, deprecated model.ServiceDetails) (model.GraphQLServiceDetailsInput, error) {
	out := model.GraphQLServiceDetailsInput{
		ID: id,
	}

	if deprecated.Api != nil {

		out.API = &graphql.APIDefinitionInput{
			Name:      deprecated.Name,
			TargetURL: deprecated.Api.TargetUrl,
		}

		if deprecated.Description != "" {
			out.API.Description = &deprecated.Description
		}

		if deprecated.Api.ApiType != "" {
			if out.API.Spec == nil {
				out.API.Spec = &graphql.APISpecInput{}
			}

			if strings.ToLower(deprecated.Api.ApiType) == "odata" {
				out.API.Spec.Type = graphql.APISpecTypeOdata
			} else {
				out.API.Spec.Type = graphql.APISpecTypeOpenAPI // quite brave assumption that it will be OpenAPI
			}
			out.API.Spec.Format = graphql.SpecFormatYaml
		}

		if deprecated.Api.Credentials != nil {
			out.API.DefaultAuth = &graphql.AuthInput{
				Credential: &graphql.CredentialDataInput{},
			}
			if deprecated.Api.Credentials.BasicWithCSRF != nil {
				// TODO later: not mapped: deprecated.Api.Credentials.BasicWithCSRF.CSRFInfo
				out.API.DefaultAuth.Credential.Basic = &graphql.BasicCredentialDataInput{
					Username: deprecated.Api.Credentials.BasicWithCSRF.Username,
					Password: deprecated.Api.Credentials.BasicWithCSRF.Password,
				}
			}

			if deprecated.Api.Credentials.OauthWithCSRF != nil {
				// TODO later: not mapped: deprecated.Api.Credentials.OauthWithCSRF.CSRFInfo
				out.API.DefaultAuth.Credential.Oauth = &graphql.OAuthCredentialDataInput{
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
			if out.API.DefaultAuth == nil {
				out.API.DefaultAuth = &graphql.AuthInput{}
			}
			h := (graphql.HttpHeaders)(*deprecated.Api.Headers)
			out.API.DefaultAuth.AdditionalHeaders = &h
		}

		// old way of providing request headers
		if deprecated.Api.QueryParameters != nil {
			if out.API.DefaultAuth == nil {
				out.API.DefaultAuth = &graphql.AuthInput{}
			}
			q := (graphql.QueryParams)(*deprecated.Api.QueryParameters)
			out.API.DefaultAuth.AdditionalQueryParams = &q
		}

		// new way
		if deprecated.Api.RequestParameters != nil {
			if out.API.DefaultAuth == nil {
				out.API.DefaultAuth = &graphql.AuthInput{}
			}
			if deprecated.Api.RequestParameters.Headers != nil {
				h := (graphql.HttpHeaders)(*deprecated.Api.RequestParameters.Headers)
				out.API.DefaultAuth.AdditionalHeaders = &h
			}
			if deprecated.Api.RequestParameters.QueryParameters != nil {
				q := (graphql.QueryParams)(*deprecated.Api.RequestParameters.QueryParameters)
				out.API.DefaultAuth.AdditionalQueryParams = &q
			}
		}

		if deprecated.Api.Spec != nil {
			if out.API.Spec == nil {
				out.API.Spec = &graphql.APISpecInput{}
			}
			asClob := graphql.CLOB(string(deprecated.Api.Spec))
			out.API.Spec.Data = &asClob
			if out.API.Spec.Type == "" {
				out.API.Spec.Type = graphql.APISpecTypeOpenAPI
			}

			if c.isXML(string(deprecated.Api.Spec)) {
				out.API.Spec.Format = graphql.SpecFormatXML
			} else if c.isJSON([]byte(deprecated.Api.Spec)) {
				out.API.Spec.Format = graphql.SpecFormatJSON
			} else {
				out.API.Spec.Format = graphql.SpecFormatYaml
			}
		}

		if deprecated.Api.Spec == nil { // TODO provide test for that
			if deprecated.Api.SpecificationUrl != "" || deprecated.Api.SpecificationCredentials != nil || deprecated.Api.SpecificationRequestParameters != nil {
				if out.API.Spec == nil {
					out.API.Spec = &graphql.APISpecInput{}
				}
				out.API.Spec.FetchRequest = &graphql.FetchRequestInput{
					URL: deprecated.Api.SpecificationUrl,
				}

				out.API.Spec.Type = toNewSpecType(deprecated.Api.ApiType)
				out.API.Spec.Format = graphql.SpecFormatJSON
			}

			if deprecated.Api.SpecificationCredentials != nil || deprecated.Api.SpecificationRequestParameters != nil {
				out.API.Spec.FetchRequest.Auth = &graphql.AuthInput{}
			}

			if deprecated.Api.SpecificationCredentials != nil {
				if deprecated.Api.SpecificationCredentials.Oauth != nil {
					inOauth := deprecated.Api.SpecificationCredentials.Oauth
					out.API.Spec.FetchRequest.Auth.Credential = &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							URL:          inOauth.URL,
							ClientID:     inOauth.ClientID,
							ClientSecret: inOauth.ClientSecret,
						},
					}
				}
				if deprecated.Api.SpecificationCredentials.Basic != nil {
					inBasic := deprecated.Api.SpecificationCredentials.Basic
					out.API.Spec.FetchRequest.Auth.Credential = &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: inBasic.Username,
							Password: inBasic.Password,
						},
					}
				}
			}

			if deprecated.Api.SpecificationRequestParameters != nil && out.API.Spec.FetchRequest != nil {
				if deprecated.Api.SpecificationRequestParameters.Headers != nil {
					h := (graphql.HttpHeaders)(*deprecated.Api.SpecificationRequestParameters.Headers)
					out.API.Spec.FetchRequest.Auth.AdditionalHeaders = &h
				}
				if deprecated.Api.SpecificationRequestParameters.QueryParameters != nil {
					q := (graphql.QueryParams)(*deprecated.Api.SpecificationRequestParameters.QueryParameters)
					out.API.Spec.FetchRequest.Auth.AdditionalQueryParams = &q
				}
			}
		}
	}

	if deprecated.Events != nil && deprecated.Events.Spec != nil {
		// TODO add tests
		var format graphql.SpecFormat
		if c.isXML(string(deprecated.Events.Spec)) {
			format = graphql.SpecFormatXML
		} else if c.isJSON(deprecated.Events.Spec) {
			format = graphql.SpecFormatJSON
		} else {
			format = graphql.SpecFormatYaml
		}

		out.Event =
			&graphql.EventDefinitionInput{
				Name: deprecated.Name,
				Spec: &graphql.EventSpecInput{
					Data:   ptrClob(graphql.CLOB(deprecated.Events.Spec)),
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: format,
				},
			}

		if deprecated.Description != "" {
			out.Event.Description = &deprecated.Description
		}

	}

	return out, nil
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

func (c *converter) GraphQLToServiceDetails(in model.GraphQLServiceDetails) (model.ServiceDetails, error) {
	outDeprecated := model.ServiceDetails{Labels: &map[string]string{}}
	if in.API != nil {
		outDeprecated.Name = in.API.Name
		outDeprecated.Api = &model.API{
			TargetUrl: in.API.TargetURL,
		}

		if in.API.Description != nil {
			outDeprecated.Description = *in.API.Description
		}

		if in.API.Spec != nil {
			outDeprecated.Api.ApiType = string(in.API.Spec.Type)
			if in.API.Spec.Data != nil {
				outDeprecated.Api.Spec = json.RawMessage(*in.API.Spec.Data)
			}
		}

		if in.API.DefaultAuth != nil && in.API.DefaultAuth.Credential != nil {
			if outDeprecated.Api.Credentials == nil {
				outDeprecated.Api.Credentials = &model.CredentialsWithCSRF{}
			}
			switch actual := in.API.DefaultAuth.Credential.(type) {
			case *graphql.BasicCredentialData:
				outDeprecated.Api.Credentials.BasicWithCSRF = &model.BasicAuthWithCSRF{
					BasicAuth: model.BasicAuth{
						Username: actual.Username,
						Password: actual.Password,
					},
				}
			case *graphql.OAuthCredentialData:
				outDeprecated.Api.Credentials.OauthWithCSRF = &model.OauthWithCSRF{
					Oauth: model.Oauth{
						URL:          actual.URL,
						ClientID:     actual.ClientID,
						ClientSecret: actual.ClientSecret,
					},
				}
			}
		}

		if in.API.DefaultAuth != nil && in.API.DefaultAuth.AdditionalHeaders != nil {
			inHeaders := *in.API.DefaultAuth.AdditionalHeaders
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

		if in.API.DefaultAuth != nil && in.API.DefaultAuth.AdditionalQueryParams != nil {
			in := *in.API.DefaultAuth.AdditionalQueryParams
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

		if in.API.Spec != nil && in.API.Spec.FetchRequest != nil {
			outDeprecated.Api.SpecificationUrl = in.API.Spec.FetchRequest.URL
			if in.API.Spec.FetchRequest.Auth != nil {
				if in.API.Spec.FetchRequest.Auth.AdditionalQueryParams != nil || in.API.Spec.FetchRequest.Auth.AdditionalHeaders != nil {
					outDeprecated.Api.SpecificationRequestParameters = &model.RequestParameters{}
				}

				if in.API.Spec.FetchRequest.Auth.AdditionalQueryParams != nil {
					asMap := (map[string][]string)(*in.API.Spec.FetchRequest.Auth.AdditionalQueryParams)
					outDeprecated.Api.SpecificationRequestParameters.QueryParameters = &asMap
				}

				if in.API.Spec.FetchRequest.Auth.AdditionalHeaders != nil {
					asMap := (map[string][]string)(*in.API.Spec.FetchRequest.Auth.AdditionalHeaders)
					outDeprecated.Api.SpecificationRequestParameters.Headers = &asMap
				}

				basic, isBasic := (in.API.Spec.FetchRequest.Auth.Credential).(*graphql.BasicCredentialData)
				oauth, isOauth := (in.API.Spec.FetchRequest.Auth.Credential).(*graphql.OAuthCredentialData)

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
	if in.Event != nil {
		outDeprecated.Name = in.Event.Name

		if in.Event.Description != nil {
			outDeprecated.Description = *in.Event.Description
		}

		if in.Event.Spec != nil && in.Event.Spec.Data != nil {
			outDeprecated.Events = &model.Events{
				Spec: []byte(string(*in.Event.Spec.Data)),
			}
		}
		//TODO: convert also fetchRequest
	}
	return outDeprecated, nil
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

func ptrClob(in graphql.CLOB) *graphql.CLOB {
	return &in
}

func (c *converter) isXML(content string) bool {
	const snippetLength = 512

	if unquoted, err := strconv.Unquote(content); err == nil {
		content = unquoted
	}

	var snippet string
	length := len(content)
	if length < snippetLength {
		snippet = content
	} else {
		snippet = content[:snippetLength]
	}

	openingIndex := strings.Index(snippet, "<")
	closingIndex := strings.Index(snippet, ">")

	return openingIndex == 0 && openingIndex < closingIndex
}

func (c *converter) isJSON(content []byte) bool {
	out := map[string]interface{}{}
	err := json.Unmarshal(content, &out)
	return err == nil
}
