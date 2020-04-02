package director

import (
	"context"
	"fmt"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	defaults "github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/retry"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=GraphQLizer -output=automock -outpkg=automock -case=underscore
type GraphQLizer interface {
	APIDefinitionInputToGQL(in graphql.APIDefinitionInput) (string, error)
	EventDefinitionInputToGQL(in graphql.EventDefinitionInput) (string, error)
	DocumentInputToGQL(in *graphql.DocumentInput) (string, error)
	PackageCreateInputToGQL(in graphql.PackageCreateInput) (string, error)
	PackageUpdateInputToGQL(in graphql.PackageUpdateInput) (string, error)
}

//go:generate mockery -name=GqlFieldsProvider -output=automock -outpkg=automock -case=underscore
type GqlFieldsProvider interface {
	ForApplication(ctx ...graphqlizer.FieldCtx) string
	ForAPIDefinition(ctx ...graphqlizer.FieldCtx) string
	ForDocument() string
	ForEventDefinition() string
	ForLabel() string
	ForPackage(ctx ...graphqlizer.FieldCtx) string
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

type CreatePackageResult struct {
	Result graphql.PackageExt `json:"result"`
}

func (c *directorClient) CreatePackage(appID string, in graphql.PackageCreateInput) (string, error) {
	inStr, err := c.graphqlizer.PackageCreateInputToGQL(in)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addPackage(applicationID: "%s", in: %s) {
				id
			}}`, appID, inStr))

	var resp CreatePackageResult

	err = c.runWithRetry(context.Background(), gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (c *directorClient) UpdatePackage(packageID string, in graphql.PackageUpdateInput) error {
	inStr, err := c.graphqlizer.PackageUpdateInputToGQL(in)
	if err != nil {
		return errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updatePackage(id: "%s", in: %s) {
				id
			}
		}`, packageID, inStr))

	err = c.runWithRetry(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}

	return nil
}

type GetPackageResult struct {
	Result graphql.ApplicationExt `json:"result"`
}

func (c *directorClient) GetPackage(appID string, packageID string) (graphql.PackageExt, error) {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, appID, c.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.package": fmt.Sprintf(`package(id: "%s") {%s}`, packageID, c.gqlFieldsProvider.ForPackage()),
		})))

	var resp GetPackageResult

	err := c.runWithRetry(context.Background(), gqlRequest, &resp)
	if err != nil {
		return graphql.PackageExt{}, errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.Package, nil
}

type ListPackagesResult struct {
	Result graphql.ApplicationExt `json:"result"`
}

func (c *directorClient) ListPackages(appID string) ([]*graphql.PackageExt, error) {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, appID, c.gqlFieldsProvider.ForApplication(),
		))

	var resp ListPackagesResult

	err := c.runWithRetry(context.Background(), gqlRequest, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "while doing GraphQL request")
	}

	// No pagination for now. Return first 100 packages;
	// TODO: Implement pagination
	return resp.Result.Packages.Data, nil
}

func (c *directorClient) DeletePackage(packageID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deletePackage(id: "%s") {
			id
		}	
	}`, packageID))

	err := c.runWithRetry(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

type CreateAPIDefinitionResult struct {
	Result graphql.APIDefinition `json:"result"`
}

func (c *directorClient) CreateAPIDefinition(packageID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error) {
	inStr, err := c.graphqlizer.APIDefinitionInputToGQL(apiDefinitionInput)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addAPIDefinitionToPackage(packageID: "%s", in: %s) {
					%s
				}
			}`, packageID, inStr, c.gqlFieldsProvider.ForAPIDefinition()))

	var resp CreateAPIDefinitionResult

	err = c.runWithRetry(context.Background(), gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (c *directorClient) DeleteAPIDefinition(apiID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteAPIDefinition(id: "%s") {
			id
		}	
	}`, apiID))

	err := c.runWithRetry(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

type CreateEventDefinitionResult struct {
	Result graphql.EventDefinition `json:"result"`
}

func (c *directorClient) CreateEventDefinition(packageID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error) {
	inStr, err := c.graphqlizer.EventDefinitionInputToGQL(eventDefinitionInput)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addEventDefinitionToPackage(packageID: "%s", in: %s) {
					%s
				}
			}`, packageID, inStr, c.gqlFieldsProvider.ForEventDefinition()))

	var resp CreateEventDefinitionResult

	err = c.runWithRetry(context.Background(), gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (c *directorClient) DeleteEventDefinition(eventID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteEventDefinition(id: "%s") {
			id
		}	
	}`, eventID))

	err := c.runWithRetry(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

type CreateDocumentResult struct {
	Result graphql.Document `json:"result"`
}

func (c *directorClient) CreateDocument(packageID string, documentInput graphql.DocumentInput) (string, error) {
	inStr, err := c.graphqlizer.DocumentInputToGQL(&documentInput)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addDocumentToPackage(packageID: "%s", in: %s) {
					%s
				}
			}`, packageID, inStr, c.gqlFieldsProvider.ForDocument()))

	var resp CreateDocumentResult

	err = c.runWithRetry(context.Background(), gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (c *directorClient) DeleteDocument(documentID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteDocument(id: "%s") {
			id
		}	
	}`, documentID))

	err := c.runWithRetry(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

func (c *directorClient) runWithRetry(ctx context.Context, req *gcli.Request, resp interface{}) error {
	return retry.Do(func() error {
		return c.cli.Run(ctx, req, resp)
	}, defaults.DefaultRetryOptions()...)
}
