/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package director

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Client
type Client interface {
	Do(ctx context.Context, req *gcli.Request, res interface{}) error
}

//go:generate mockery -name=GqlFieldsProvider
type GqlFieldsProvider interface {
	ForApplication(ctx ...graphqlizer.FieldCtx) string
	ForAPIDefinition(ctx ...graphqlizer.FieldCtx) string
	ForDocument() string
	ForEventDefinition() string
	ForLabel() string
	ForPackage(ctx ...graphqlizer.FieldCtx) string
	ForPackageInstanceAuth() string
	Page(item string) string
}

//go:generate mockery -name=GraphQLizer
type GraphQLizer interface {
	PackageInstanceAuthRequestInputToGQL(in schema.PackageInstanceAuthRequestInput) (string, error)
}

func NewGraphQLClient(gqlClient Client, gqlizer GraphQLizer, gqlFieldsProvider GqlFieldsProvider) *GraphQLClient {
	return &GraphQLClient{
		gcli: gqlClient,
		//queryProvider:     queryProvider{}, - gqlizers are better
		inputGraphqlizer:  &graphqlizer.Graphqlizer{},
		outputGraphqlizer: &graphqlizer.GqlFieldsProvider{},
	}
}

type GraphQLClient struct {
	gcli              Client
	inputGraphqlizer  GraphQLizer
	outputGraphqlizer GqlFieldsProvider
	//queryProvider     queryProvider
}

func (c *GraphQLClient) FetchApplications(ctx context.Context) (*ApplicationsOutput, error) {
	response := ApplicationsOutput{}

	query := fmt.Sprintf(`query {
			result: applications {
					%s
			}
	}`, c.outputGraphqlizer.Page(c.outputGraphqlizer.ForApplication()))

	//applicationsQuery := c.queryProvider.applicationsForRuntimeQuery(runtimeID)
	//TODO make gclirequestprovider so that we can set correlationid
	req := gcli.NewRequest(query)

	err := c.gcli.Do(ctx, req, &response)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching applications in gqlclient")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return nil, errors.New("Failed fetch Applications for Runtime from Director: received nil response.")
	}

	//TODO paging, and dont return page details outside of the client method

	return &response, nil
}

func (c *GraphQLClient) RequestPackageInstanceCredentialsCreation(ctx context.Context, in *RequestPackageInstanceCredentialsInput) (*RequestPackageInstanceCredentialsOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	input, err := in.InputSchema.MarshalToQGLJSON()
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling input schema to GQL JSON")
	}

	inContext, err := in.Context.MarshalToQGLJSON()
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling context to GQL JSON")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`mutation {
			  result: requestPackageInstanceAuthCreation(
				packageID: "%s"
				in: {
				  context: %s
    			  inputParams: %s
				}
			  ) {
					id
					context
					auth {
					  additionalHeaders
					  additionalQueryParams
					  requestAuth {
						csrf {
						  tokenEndpointURL
						}
					  }
					  credential {
						... on OAuthCredentialData {
						  clientId
						  clientSecret
						  url
						}
						... on BasicCredentialData {
						  username
						  password
						}
					  }
					}
					status {
					  condition
					  timestamp
					  message
					  reason
					}
			  	 }
				}`, in.PackageID, inContext, input))

	var resp RequestPackageInstanceCredentialsOutput
	if err = c.gcli.Do(ctx, gqlRequest, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to create package instance auth")
	}

	return &resp, nil
}

func (c *GraphQLClient) FindPackageInstanceCredentialsForContext(ctx context.Context, in *FindPackageInstanceCredentialsByContextInput) (*FindPackageInstanceCredentialsOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s") {
						  instanceAuths {
							id
							context
							auth {
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								}
							  }
							  credential {
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
								... on BasicCredentialData {
								  username
								  password
								}
							  }
							}
							status {
							  condition
							  timestamp
							  message
							  reason
							}
						  }
						}
					  }
					}`, in.ApplicationID, in.PackageID))

	var resp struct {
		Result schema.ApplicationExt `json:"result"`
	}
	err := c.gcli.Do(ctx, gqlRequest, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get package instance auths")
	}

	var authsResp []*schema.PackageInstanceAuth
	for _, auth := range resp.Result.Package.InstanceAuths {
		if auth == nil {
			continue
		}

		var authContext map[string]string
		if err := json.Unmarshal([]byte(*auth.Context), &authContext); err != nil {
			return nil, errors.Wrap(err, "while unmarshaling auth context")
		}

		shouldReturn := true
		for key, value := range in.Context {
			authContextValue, found := authContext[key]
			if !found || authContextValue != value {
				shouldReturn = false
			}
		}

		if shouldReturn {
			authsResp = append(authsResp, auth)
		}
	}

	if len(authsResp) == 0 {
		return nil, &NotFoundError{}
	}

	return &FindPackageInstanceCredentialsOutput{
		InstanceAuths: authsResp,
	}, nil
}

func (c *GraphQLClient) FindPackageInstanceCredentials(ctx context.Context, in *FindPackageInstanceCredentialInput) (*FindPackageInstanceCredentialOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	//TODO replace with provider with correlation id
	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			  result: application(id: %q) {
						package(id: %q) {
						  instanceAuth(id: %q) {
							id
							auth {
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								}
							  }
							  credential {
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
								... on BasicCredentialData {
								  username
								  password
								}
							  }
							}
							status {
							  condition
							  timestamp
							  message
							  reason
							}
						  }
						}
					  }
					}`, in.ApplicationID, in.PackageID, in.InstanceAuthID))

	var response struct {
		Result schema.ApplicationExt `json:"result"`
	}
	if err := c.gcli.Do(ctx, gqlRequest, &response); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get package instance auth")
	}

	if response.Result.Package.InstanceAuth == nil {
		return nil, &NotFoundError{}
	}

	return &FindPackageInstanceCredentialOutput{
		InstanceAuth: response.Result.Package.InstanceAuth,
	}, nil
}

func (c *GraphQLClient) RequestPackageInstanceCredentialsDeletion(ctx context.Context, in *RequestPackageInstanceAuthDeletionInput) (*RequestPackageInstanceAuthDeletionOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`mutation {
			  result: requestPackageInstanceAuthDeletion(authID: %q) {
						id
						status {
						  condition
						  timestamp
						  message
						  reason
						}
					  }
					}`, in.InstanceAuthID))

	var resp struct {
		Result RequestPackageInstanceAuthDeletionOutput `json:"result"`
	}

	if err := c.gcli.Do(ctx, gqlRequest, &resp); err != nil {
		if IsGQLNotFoundError(err) {
			return nil, &NotFoundError{}
		}

		return nil, errors.Wrap(err, "while executing GraphQL call to delete the package instance auth")
	}

	return &resp.Result, nil
}
