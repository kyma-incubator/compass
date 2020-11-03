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

	"github.com/kyma-incubator/compass/components/system-broker/pkg/scheduler"

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
	ForApplicationPageable(ctx ...graphqlizer.FieldCtx) string
	ForAPIDefinition(ctx ...graphqlizer.FieldCtx) string
	ForDocument() string
	ForEventDefinition() string
	ForLabel() string
	ForPackage(ctx ...graphqlizer.FieldCtx) string
	ForPackagePageable(ctx ...graphqlizer.FieldCtx) string
	ForPackageInstanceAuth() string
	Page(item string) string
}

//go:generate mockery -name=GraphQLizer
type GraphQLizer interface {
	PackageInstanceAuthRequestInputToGQL(in schema.PackageInstanceAuthRequestInput) (string, error)
}

func NewGraphQLClient(gqlClient Client, gqlizer GraphQLizer, gqlFieldsProvider GqlFieldsProvider, c *Config) *GraphQLClient {
	return &GraphQLClient{
		gcli: gqlClient,
		//queryProvider:     queryProvider{}, - gqlizers are better
		inputGraphqlizer:  gqlizer,
		outputGraphqlizer: gqlFieldsProvider,
		pageSize:          c.PageSize,
		pageConcurrency:   c.PageConcurrency,
	}
}

type GraphQLClient struct {
	gcli              Client
	inputGraphqlizer  GraphQLizer
	outputGraphqlizer GqlFieldsProvider
	pageSize          int
	pageConcurrency   int
}

func (c *GraphQLClient) FetchApplications(ctx context.Context) (ApplicationsOutput, error) {
	query := fmt.Sprintf(`query($first: Int, $after: PageCursor) {
			result: applications(first: $first, after: $after) {
					%s
			}
	}`, c.outputGraphqlizer.Page(c.outputGraphqlizer.ForApplicationPageable()))

	pager := NewPaginator(query, c.pageSize, c.gcli)
	apps := &ApplicationResponse{}

	appsResult, err := apps.ListAll(ctx, pager)
	if err != nil {
		return nil, errors.Wrap(err, "while listing all apps")
	}

	s := scheduler.New(ctx, c.pageConcurrency)
	s.Schedule(func(ctx context.Context) error {
		c.fetchPackages(appsResult, s)
		return nil
	})

	err = s.Wait()
	if err != nil {
		return nil, errors.Wrap(err, "while fetching applicaiotns")
	}
	return appsResult, nil
}

func (c *GraphQLClient) fetchPackages(apps ApplicationsOutput, s *scheduler.Scheduler) {
	for i, app := range apps {
		f := c.fetchPackagesForApp(i, app, apps, s)
		s.Schedule(f)
	}
}

func (c *GraphQLClient) fetchPackagesForApp(i int, app schema.ApplicationExt, apps ApplicationsOutput, s *scheduler.Scheduler) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		query := fmt.Sprintf(`query($first: Int, $after: PageCursor) {
				result: application(id: %q) {
					packages(first: $first, after: $after) {
					%s
					}
				}
			}`, app.ID, c.outputGraphqlizer.Page(c.outputGraphqlizer.ForPackagePageable()))

		pager := NewPaginator(query, c.pageSize, c.gcli)
		packages := &PackagesResponse{}
		packagesResult, err := packages.ListAll(ctx, pager)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return errors.Wrap(err, "while fetching packages")
			}
		}

		apps[i].Packages = schema.PackagePageExt{
			Data: packagesResult,
		}

		s.Schedule(func(ctx context.Context) error {
			c.fetchApiDefinitions(app.ID, packagesResult, s)
			return nil
		})
		s.Schedule(func(ctx context.Context) error {
			c.fetchEventDefinitions(app.ID, packagesResult, s)
			return nil
		})
		s.Schedule(func(ctx context.Context) error {
			c.fetchDocuments(app.ID, packagesResult, s)
			return nil
		})
		return nil
	}
}

func (c *GraphQLClient) fetchApiDefinitions(appID string, packages PackagessOutput, s *scheduler.Scheduler) {
	for i, packaged := range packages {
		f := c.fetchApiDefForPackage(i, packaged, appID, packages)
		s.Schedule(f)
	}
}

func (c *GraphQLClient) fetchEventDefinitions(appID string, packages PackagessOutput, s *scheduler.Scheduler) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		for i, packaged := range packages {
			f := c.fetchEventDefForPackage(i, packaged, appID, packages)
			s.Schedule(f)
		}
		return nil
	}
}

func (c *GraphQLClient) fetchDocuments(appID string, packages PackagessOutput, s *scheduler.Scheduler) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		for i, packaged := range packages {
			f := c.fetchDocumentsForPackage(i, packaged, appID, packages)
			s.Schedule(f)
		}
		return nil
	}
}

func (c *GraphQLClient) fetchApiDefForPackage(i int, packaged *schema.PackageExt, appID string, packages PackagessOutput) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		query := fmt.Sprintf(`query($first: Int, $after: PageCursor) {
	result: application(id: %q) {
		package(id: %q) {
			apiDefinitions(first: $first, after: $after) {
				%s
			}
		  }
	}
}`, appID, packaged.ID, c.outputGraphqlizer.Page(c.outputGraphqlizer.ForAPIDefinition()))

		pager := NewPaginator(query, c.pageSize, c.gcli)
		definitions := &ApiDefinitionsResponse{}
		responseApiDefinitions, err := definitions.ListAll(ctx, pager)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return errors.Wrap(err, "while fetching api definitions")
			}
		}
		packages[i].APIDefinitions = schema.APIDefinitionPageExt{
			Data: responseApiDefinitions,
		}
		return nil
	}
}

func (c *GraphQLClient) fetchEventDefForPackage(i int, packaged *schema.PackageExt, appID string, packages PackagessOutput) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		query := fmt.Sprintf(`query($first: Int, $after: PageCursor) {
result: application(id: %q) {
	package(id: %q) {
		eventDefinitions(first: $first, after: $after) {
			%s
		}
	  }
}
}`, appID, packaged.ID, c.outputGraphqlizer.Page(c.outputGraphqlizer.ForEventDefinition()))

		pager := NewPaginator(query, c.pageSize, c.gcli)
		definitions := &EventDefinitionsResponse{}
		responseEventDefinitions, err := definitions.ListAll(ctx, pager)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return errors.Wrap(err, "while fetching event definitions")
			}
		}

		packages[i].EventDefinitions = schema.EventAPIDefinitionPageExt{
			Data: responseEventDefinitions,
		}
		return nil
	}
}

func (c *GraphQLClient) fetchDocumentsForPackage(i int, packaged *schema.PackageExt, appID string, packages PackagessOutput) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		query := fmt.Sprintf(`query($first: Int, $after: PageCursor) {
result: application(id: %q) {
	package(id: %q) {
		documents(first: $first, after: $after) {
			%s
		}
	  }
}
}`, appID, packaged.ID, c.outputGraphqlizer.Page(c.outputGraphqlizer.ForDocument()))

		pager := NewPaginator(query, c.pageSize, c.gcli)
		definitions := &DocumentsResponse{}
		responseDocuments, err := definitions.ListAll(ctx, pager)

		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return errors.Wrap(err, "while fetching documents")
			}
		}
		packages[i].Documents = schema.DocumentPageExt{
			Data: responseDocuments,
		}
		return nil
	}
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
							%s
					 	}
					}
				}`, in.ApplicationID, in.PackageID, c.outputGraphqlizer.ForPackage()))

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

	targetURLs := make(map[string]string, resp.Result.Package.APIDefinitions.TotalCount)

	for _, apiDefinition := range resp.Result.Package.APIDefinitions.Data {
		targetURLs[apiDefinition.Name] = apiDefinition.TargetURL
	}

	return &FindPackageInstanceCredentialsOutput{
		InstanceAuths: authsResp,
		TargetURLs:    targetURLs,
	}, nil
}

func (c *GraphQLClient) FindPackageInstanceCredentials(ctx context.Context, in *FindPackageInstanceCredentialInput) (*FindPackageInstanceCredentialOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			  result: application(id: %q) {
						package(id: %q) {
						  instanceAuth(id: %q) {
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

func (c *GraphQLClient) FindSpecification(ctx context.Context, in *FindPackageSpecificationInput) (*FindPackageSpecificationOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			  result: application(id: %q) {
						package(id: %q) {
						  apiDefinition(id: %q) {
							  spec {
								data
								type
								format
							  }
						  }
						  eventDefinition(id: %q) {
							  spec {
								data
								type
								format
							  }
						  }
						}
					  }
					}`, in.ApplicationID, in.PackageID, in.DefinitionID, in.DefinitionID))

	var response struct {
		Result schema.ApplicationExt `json:"result"`
	}
	if err := c.gcli.Do(ctx, gqlRequest, &response); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get package instance auth")
	}

	apidef := response.Result.Package.APIDefinition
	if apidef.Spec != nil {
		return &FindPackageSpecificationOutput{
			Name:        apidef.Name,
			Description: apidef.Description,
			Data:        apidef.Spec.Data,
			Format:      apidef.Spec.Format,
			Type:        string(apidef.Spec.Type),
			Version:     apidef.Version,
		}, nil
	}

	eventdef := response.Result.Package.EventDefinition
	if eventdef.Spec != nil {
		return &FindPackageSpecificationOutput{
			Name:        eventdef.Name,
			Description: eventdef.Description,
			Data:        eventdef.Spec.Data,
			Format:      eventdef.Spec.Format,
			Type:        string(eventdef.Spec.Type),
			Version:     eventdef.Version,
		}, nil
	}

	return nil, errors.New("definition missing from director response")
}
