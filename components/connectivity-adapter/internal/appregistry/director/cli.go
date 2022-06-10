package director

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/retry"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery --name=GraphQLizer --output=automock --outpkg=automock --case=underscore --disable-version-string
type GraphQLizer interface {
	APIDefinitionInputToGQL(in graphql.APIDefinitionInput) (string, error)
	EventDefinitionInputToGQL(in graphql.EventDefinitionInput) (string, error)
	DocumentInputToGQL(in *graphql.DocumentInput) (string, error)
	BundleCreateInputToGQL(in graphql.BundleCreateInput) (string, error)
	BundleUpdateInputToGQL(in graphql.BundleUpdateInput) (string, error)
}

//go:generate mockery --name=GqlFieldsProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type GqlFieldsProvider interface {
	ForApplication(ctx ...graphqlizer.FieldCtx) string
	ForAPIDefinition(ctx ...graphqlizer.FieldCtx) string
	ForDocument() string
	ForEventDefinition() string
	ForLabel() string
	ForBundle(ctx ...graphqlizer.FieldCtx) string
	Page(item string) string
}

const nameKey = "name"

type directorClient struct {
	cli               gqlcli.GraphQLClient
	graphqlizer       GraphQLizer
	gqlFieldsProvider GqlFieldsProvider
}

func NewClient(cli gqlcli.GraphQLClient, graphqlizer GraphQLizer, gqlFieldsProvider GqlFieldsProvider) *directorClient {
	return &directorClient{cli: cli, graphqlizer: graphqlizer, gqlFieldsProvider: gqlFieldsProvider}
}

// TODO: Replace with method which uses GraphQL client
func (c *directorClient) GetApplicationsByNameRequest(appName string) *gcli.Request {
	return gcli.NewRequest(fmt.Sprintf(`query {
			result: applications(filter: {key:"%s", query: "\"%s\""}) {
					%s
			}
	}`, nameKey, appName, c.gqlFieldsProvider.Page(c.gqlFieldsProvider.ForApplication())))
}

type CreateBundleResult struct {
	Result graphql.BundleExt `json:"result"`
}

func (c *directorClient) CreateBundle(ctx context.Context, appID string, in graphql.BundleCreateInput) (string, error) {
	inStr, err := c.graphqlizer.BundleCreateInputToGQL(in)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addBundle(applicationID: "%s", in: %s) {
				id
			}}`, appID, inStr))

	var resp CreateBundleResult

	err = retry.GQLRun(c.cli.Run, ctx, gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (c *directorClient) UpdateBundle(ctx context.Context, bundleID string, in graphql.BundleUpdateInput) error {
	inStr, err := c.graphqlizer.BundleUpdateInputToGQL(in)
	if err != nil {
		return errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateBundle(id: "%s", in: %s) {
				id
			}
		}`, bundleID, inStr))

	err = retry.GQLRun(c.cli.Run, ctx, gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}

	return nil
}

type GetBundleResult struct {
	Result graphql.ApplicationExt `json:"result"`
}

func (c *directorClient) GetBundle(ctx context.Context, appID string, bundleID string) (graphql.BundleExt, error) {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, appID, c.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.bundle": fmt.Sprintf(`bundle(id: "%s") {%s}`, bundleID, c.gqlFieldsProvider.ForBundle()),
		})))

	var resp GetBundleResult

	err := retry.GQLRun(c.cli.Run, ctx, gqlRequest, &resp)
	if err != nil {
		return graphql.BundleExt{}, errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.Bundle, nil
}

type ListBundlesResult struct {
	Result graphql.ApplicationExt `json:"result"`
}

func (c *directorClient) ListBundles(ctx context.Context, appID string) ([]*graphql.BundleExt, error) {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, appID, c.gqlFieldsProvider.ForApplication(),
		))

	var resp ListBundlesResult

	err := retry.GQLRun(c.cli.Run, ctx, gqlRequest, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "while doing GraphQL request")
	}

	// No pagination for now. Return first 100 bundles;
	// TODO: Implement pagination
	return resp.Result.Bundles.Data, nil
}

func (c *directorClient) DeleteBundle(ctx context.Context, bundleID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteBundle(id: "%s") {
			id
		}	
	}`, bundleID))

	err := retry.GQLRun(c.cli.Run, ctx, gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

type CreateAPIDefinitionResult struct {
	Result graphql.APIDefinition `json:"result"`
}

func (c *directorClient) CreateAPIDefinition(ctx context.Context, bundleID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error) {
	inStr, err := c.graphqlizer.APIDefinitionInputToGQL(apiDefinitionInput)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addAPIDefinitionToBundle(bundleID: "%s", in: %s) {
					%s
				}
			}`, bundleID, inStr, c.gqlFieldsProvider.ForAPIDefinition()))

	var resp CreateAPIDefinitionResult

	err = retry.GQLRun(c.cli.Run, ctx, gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (c *directorClient) DeleteAPIDefinition(ctx context.Context, apiID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteAPIDefinition(id: "%s") {
			id
		}	
	}`, apiID))

	err := retry.GQLRun(c.cli.Run, ctx, gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

type CreateEventDefinitionResult struct {
	Result graphql.EventDefinition `json:"result"`
}

func (c *directorClient) CreateEventDefinition(ctx context.Context, bundleID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error) {
	inStr, err := c.graphqlizer.EventDefinitionInputToGQL(eventDefinitionInput)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addEventDefinitionToBundle(bundleID: "%s", in: %s) {
					%s
				}
			}`, bundleID, inStr, c.gqlFieldsProvider.ForEventDefinition()))

	var resp CreateEventDefinitionResult

	err = retry.GQLRun(c.cli.Run, ctx, gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (c *directorClient) DeleteEventDefinition(ctx context.Context, eventID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteEventDefinition(id: "%s") {
			id
		}	
	}`, eventID))

	err := retry.GQLRun(c.cli.Run, ctx, gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

type CreateDocumentResult struct {
	Result graphql.Document `json:"result"`
}

func (c *directorClient) CreateDocument(ctx context.Context, bundleID string, documentInput graphql.DocumentInput) (string, error) {
	inStr, err := c.graphqlizer.DocumentInputToGQL(&documentInput)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addDocumentToBundle(bundleID: "%s", in: %s) {
					%s
				}
			}`, bundleID, inStr, c.gqlFieldsProvider.ForDocument()))

	var resp CreateDocumentResult

	err = retry.GQLRun(c.cli.Run, ctx, gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (c *directorClient) DeleteDocument(ctx context.Context, documentID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteDocument(id: "%s") {
			id
		}	
	}`, documentID))

	err := retry.GQLRun(c.cli.Run, ctx, gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

func (c *directorClient) SetApplicationLabel(ctx context.Context, appID string, label graphql.LabelInput) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setApplicationLabel(applicationID: "%s", key: "%s", value: %s) {
					%s
				}
			}`,
			appID, label.Key, label.Value, c.gqlFieldsProvider.ForLabel()))

	err := retry.GQLRun(c.cli.Run, ctx, gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}

	return nil
}
