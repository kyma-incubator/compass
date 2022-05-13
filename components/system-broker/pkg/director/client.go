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

//go:generate mockery --name=GqlFieldsProvider --disable-version-string
type GqlFieldsProvider interface {
	OmitForApplication(omit []string) string
	ForApplication(ctx ...graphqlizer.FieldCtx) string
	ForAPIDefinition(ctx ...graphqlizer.FieldCtx) string
	ForDocument() string
	ForEventDefinition() string
	ForLabel() string
	ForBundle(ctx ...graphqlizer.FieldCtx) string
	ForBundleInstanceAuth() string
	ForBundleInstanceAuthStatus() string
	Page(item string) string
}

//go:generate mockery --name=GraphQLizer --disable-version-string
type GraphQLizer interface {
	BundleInstanceAuthRequestInputToGQL(in schema.BundleInstanceAuthRequestInput) (string, error)
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

func (c *GraphQLClient) FetchApplication(ctx context.Context, id string) (*ApplicationOutput, error) {
	query := fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
			}
	}`, id, c.outputGraphqlizer.OmitForApplication([]string{
		"providerName",
		"description",
		"integrationSystemID",
		"labels",
		"status",
		"healthCheckURL",
		"bundles",
		"auths",
		"eventingConfiguration",
	}))
	app := ApplicationOutput{}
	req := gcli.NewRequest(query)

	if err := c.gcli.Do(ctx, req, &app); err != nil {
		return nil, errors.Wrap(err, "while fetching application in gqlclient")
	}
	if app.Result == nil {
		return nil, &NotFoundError{}
	}

	return &app, nil
}

func (c *GraphQLClient) FetchApplications(ctx context.Context) (*ApplicationsOutput, error) {
	query := fmt.Sprintf(`query {
			result: applications {
					%s
			}
	}`, c.outputGraphqlizer.Page(c.outputGraphqlizer.OmitForApplication([]string{
		"auths",
		"webhooks",
		"status",
		"bundles.instanceAuths",
		"bundles.documents",
		"bundles.apiDefinitions.spec.fetchRequest",
		"bundles.eventDefinitions.spec.fetchRequest",
	})))
	apps := ApplicationsOutput{}
	req := gcli.NewRequest(query)

	if err := c.gcli.Do(ctx, req, &apps); err != nil {
		return nil, errors.Wrap(err, "while fetching applications in gqlclient")
	}
	if apps.Result == nil {
		return nil, errors.New("failed to fetch applications")
	}

	return &apps, nil
}

func (c *GraphQLClient) RequestBundleInstanceCredentialsCreation(ctx context.Context, in *BundleInstanceCredentialsInput) (*BundleInstanceAuthOutput, error) {
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
			  result: requestBundleInstanceAuthCreation(
				bundleID: %q
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
				}`, in.BundleID, in.AuthID, inContext, inputParams))

	var resp BundleInstanceAuthOutput
	if err = c.gcli.Do(ctx, gqlRequest, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to create bundle instance auth")
	}

	return &resp, nil
}

func (c *GraphQLClient) FetchBundleInstanceCredentials(ctx context.Context, in *BundleInstanceInput) (*BundleInstanceCredentialsOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query{
			  result:bundleByInstanceAuth(authID:%q){
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
	}`, in.InstanceAuthID, in.InstanceAuthID, c.outputGraphqlizer.ForBundleInstanceAuth()))

	var response struct {
		Bundle *schema.BundleExt `json:"result"`
	}
	if err := c.gcli.Do(ctx, gqlRequest, &response); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get bundle instance auth")
	}

	if response.Bundle == nil || response.Bundle.InstanceAuth == nil || response.Bundle.InstanceAuth.Context == nil {
		return nil, &NotFoundError{}
	}

	var authContext map[string]string
	if err := json.Unmarshal([]byte(*response.Bundle.InstanceAuth.Context), &authContext); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling auth context")
	}

	if authContext["instance_id"] != in.Context["instance_id"] || authContext["binding_id"] != in.Context["binding_id"] {
		return nil, errors.New("found binding with mismatched context coordinates")
	}

	targetURLs := make(map[string]string, response.Bundle.APIDefinitions.TotalCount)
	for _, apiDefinition := range response.Bundle.APIDefinitions.Data {
		targetURLs[apiDefinition.Name] = apiDefinition.TargetURL
	}

	return &BundleInstanceCredentialsOutput{
		InstanceAuth: response.Bundle.InstanceAuth,
		TargetURLs:   targetURLs,
	}, nil
}

func (c *GraphQLClient) FetchBundleInstanceAuth(ctx context.Context, in *BundleInstanceInput) (*BundleInstanceAuthOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			  result: bundleInstanceAuth(id: %q) {
				id
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
		BundleInstanceAuth *schema.BundleInstanceAuth `json:"result"`
	}
	if err := c.gcli.Do(ctx, gqlRequest, &response); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get bundle instance auth")
	}

	if response.BundleInstanceAuth == nil || response.BundleInstanceAuth.Context == nil {
		return nil, &NotFoundError{}
	}

	var authContext map[string]string
	if err := json.Unmarshal([]byte(*response.BundleInstanceAuth.Context), &authContext); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling auth context")
	}

	if authContext["instance_id"] != in.Context["instance_id"] || authContext["binding_id"] != in.Context["binding_id"] {
		return nil, errors.New("found binding with mismatched context coordinates")
	}

	return &BundleInstanceAuthOutput{
		InstanceAuth: response.BundleInstanceAuth,
	}, nil
}

func (c *GraphQLClient) RequestBundleInstanceCredentialsDeletion(ctx context.Context, in *BundleInstanceAuthDeletionInput) (*BundleInstanceAuthDeletionOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`mutation {
			  result: requestBundleInstanceAuthDeletion(authID: %q) {
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
		Result BundleInstanceAuthDeletionOutput `json:"result"`
	}

	if err := c.gcli.Do(ctx, gqlRequest, &resp); err != nil {
		if IsGQLNotFoundError(err) {
			return nil, &NotFoundError{}
		}

		return nil, errors.Wrap(err, "while executing GraphQL call to delete the bundle instance auth")
	}

	return &resp.Result, nil
}

func (c *GraphQLClient) FindSpecification(ctx context.Context, in *BundleSpecificationInput) (*BundleSpecificationOutput, error) {
	if _, err := govalidator.ValidateStruct(in); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	gqlRequest := gcli.NewRequest(fmt.Sprintf(`query {
			  result: application(id: %q) {
						bundle(id: %q) {
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
					}`, in.ApplicationID, in.BundleID, in.DefinitionID, in.DefinitionID))

	var response struct {
		Result schema.ApplicationExt `json:"result"`
	}
	if err := c.gcli.Do(ctx, gqlRequest, &response); err != nil {
		return nil, errors.Wrap(err, "while executing GraphQL call to get bundle instance auth")
	}

	apidef := response.Result.Bundle.APIDefinition
	if apidef.Spec != nil {
		return &BundleSpecificationOutput{
			Name:        apidef.Name,
			Description: apidef.Description,
			Data:        apidef.Spec.Data,
			Format:      apidef.Spec.Format,
			Type:        string(apidef.Spec.Type),
			Version:     apidef.Version,
		}, nil
	}

	eventdef := response.Result.Bundle.EventDefinition
	if eventdef.Spec != nil {
		return &BundleSpecificationOutput{
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
