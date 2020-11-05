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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Client
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
	ForPackageInstanceAuthStatus() string
	Page(item string) string
}

//go:generate mockery -name=GraphQLizer
type GraphQLizer interface {
	PackageInstanceAuthRequestInputToGQL(in schema.PackageInstanceAuthRequestInput) (string, error)
}

func NewGraphQLClient(gqlClient Client, gqlizer GraphQLizer, gqlFieldsProvider GqlFieldsProvider) *GraphQLClient {
	return &GraphQLClient{
		gcli:              gqlClient,
		inputGraphqlizer:  gqlizer,
		outputGraphqlizer: gqlFieldsProvider,
	}
}

type GraphQLClient struct {
	gcli              Client
	inputGraphqlizer  GraphQLizer
	outputGraphqlizer GqlFieldsProvider
}

func (c *GraphQLClient) FetchApplications(ctx context.Context) (*ApplicationsOutput, error) {
	query := fmt.Sprintf(`query {
			result: applications {
					%s
			}
	}`, c.outputGraphqlizer.Page(c.outputGraphqlizer.ForApplication()))

	apps := ApplicationsOutput{}

	req := gcli.NewRequest(query)

	err := c.gcli.Do(ctx, req, &apps)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching applications in gqlclient")
	}
	if apps.Result == nil {
		return nil, errors.New("failed to fetch applications")
	}

	return &apps, nil
}

func (c *GraphQLClient) RequestPackageInstanceCredentialsCreation(ctx context.Context, in *PackageInstanceCredentialsInput) (*PackageInstanceAuthOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	inputParams, err := in.InputSchema.MarshalToQGLJSON()
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling input schema to GQL JSON")
	}

	inContext, err := in.Context.MarshalToQGLJSON()
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling context to GQL JSON")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`mutation {
			  result: requestPackageInstanceAuthCreation(
				packageID: %q
				in: {
				  id: %q
				  context: %s
    			  inputParams: %s
				}
			  ) {
					status {
					  condition
					  timestamp
					  message
					  reason
					}
			  	 }
				}`, in.PackageID, in.AuthID, inContext, inputParams))

	var resp PackageInstanceAuthOutput
	if err = c.gcli.Do(ctx, gqlRequest, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to create package instance auth")
	}

	return &resp, nil
}

func (c *GraphQLClient) FetchPackageInstanceCredentials(ctx context.Context, in *PackageInstanceInput) (*PackageInstanceCredentialsOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query{
			  result:packageByInstanceAuth(authID:%q){
				apiDefinitions{
				  data{
					name
					targetURL
				  }
				}
				instanceAuth(id: %q){
				  %s
				}
			  }
	}`, in.InstanceAuthID, in.InstanceAuthID, c.outputGraphqlizer.ForPackageInstanceAuth()))

	var response struct {
		Package *schema.PackageExt `json:"result"`
	}
	if err := c.gcli.Do(ctx, gqlRequest, &response); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get package instance auth")
	}

	if response.Package == nil || response.Package.InstanceAuth == nil || response.Package.InstanceAuth.Context == nil {
		return nil, &NotFoundError{}
	}

	var authContext map[string]string
	if err := json.Unmarshal([]byte(*response.Package.InstanceAuth.Context), &authContext); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling auth context")
	}

	if authContext["instance_id"] != in.Context["instance_id"] || authContext["binding_id"] != in.Context["binding_id"] {
		return nil, &NotFoundError{}
	}

	targetURLs := make(map[string]string, response.Package.APIDefinitions.TotalCount)
	for _, apiDefinition := range response.Package.APIDefinitions.Data {
		targetURLs[apiDefinition.Name] = apiDefinition.TargetURL
	}

	return &PackageInstanceCredentialsOutput{
		InstanceAuth: response.Package.InstanceAuth,
		TargetURLs:   targetURLs,
	}, nil
}

func (c *GraphQLClient) FetchPackageInstanceAuth(ctx context.Context, in *PackageInstanceInput) (*PackageInstanceAuthOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			  result: packageInstanceAuth(id: %q) {
				context
				status {
				  condition
				  timestamp
				  message
				  reason
				}
			  }
	}`, in.InstanceAuthID))

	var response struct {
		PackageInstanceAuth *schema.PackageInstanceAuth `json:"result"`
	}
	if err := c.gcli.Do(ctx, gqlRequest, &response); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get package instance auth")
	}

	if response.PackageInstanceAuth == nil || response.PackageInstanceAuth.Context == nil {
		return nil, &NotFoundError{}
	}

	var authContext map[string]string
	if err := json.Unmarshal([]byte(*response.PackageInstanceAuth.Context), &authContext); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling auth context")
	}

	if authContext["instance_id"] != in.Context["instance_id"] || authContext["binding_id"] != in.Context["binding_id"] {
		return nil, &NotFoundError{}
	}

	return &PackageInstanceAuthOutput{
		InstanceAuth: response.PackageInstanceAuth,
	}, nil
}

func (c *GraphQLClient) RequestPackageInstanceCredentialsDeletion(ctx context.Context, in *PackageInstanceAuthDeletionInput) (*PackageInstanceAuthDeletionOutput, error) {
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
		Result PackageInstanceAuthDeletionOutput `json:"result"`
	}

	if err := c.gcli.Do(ctx, gqlRequest, &resp); err != nil {
		if IsGQLNotFoundError(err) {
			return nil, &NotFoundError{}
		}

		return nil, errors.Wrap(err, "while executing GraphQL call to delete the package instance auth")
	}

	return &resp.Result, nil
}

func (c *GraphQLClient) FindSpecification(ctx context.Context, in *PackageSpecificationInput) (*PackageSpecificationOutput, error) {
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
		return &PackageSpecificationOutput{
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
		return &PackageSpecificationOutput{
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
