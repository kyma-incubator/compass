package director

import (
	"context"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/avast/retry-go"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqlTools "github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Client interface {
	CreateRuntime(in schema.RuntimeInput) (string, error)
	DeleteRuntime(runtimeID string) error
	CreateApplication(in schema.ApplicationRegisterInput) (string, error)
	DeleteApplication(appID string) error
	SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error
	GetOneTimeTokenUrl(appID string) (string, string, error)
}

type client struct {
	scopes       []string
	tenant       string
	graphqulizer gqlTools.Graphqlizer
	client       *gcli.Client
}

type ApplicationResponse struct {
	Result schema.ApplicationExt `json:"result"`
}

type RuntimeResponse struct {
	Result schema.Runtime `json:"result"`
}

type SetLabelResponse struct {
	Result schema.Label `json:"result"`
}

type SetDefaultAppEventingResponse struct {
	Result schema.ApplicationEventingConfiguration `json:"result"`
}

type OneTimeTokenResponse struct {
	Result schema.OneTimeTokenForApplication `json:"result"`
}

type TenantsResponse struct {
	Result []*schema.Tenant
}

func NewClient(directorURL, tenant string, scopes []string) (Client, error) {
	var directorClient Client

	err := retry.Do(func() error {
		client, err := newClient(directorURL, tenant, scopes)
		if err == nil {
			directorClient = client
		} else {
			logrus.Warningf("Failed to create Director client: %s", err)
		}

		return err
	}, retry.Delay(time.Second*15))

	return directorClient, err
}

func newClient(directorURL, tenant string, scopes []string) (Client, error) {
	internalTenantID, err := getInternalTenantID(directorURL, tenant)
	if err != nil {
		return nil, err
	}

	gqlClient, err := getClient(directorURL, internalTenantID, scopes)
	if err != nil {
		return nil, err
	}

	return client{
		scopes:       scopes,
		tenant:       tenant,
		graphqulizer: gqlTools.Graphqlizer{},
		client:       gqlClient,
	}, nil
}

func (c client) CreateApplication(in schema.ApplicationRegisterInput) (string, error) {

	appGraphql, err := c.graphqulizer.ApplicationRegisterInputToGQL(in)
	if err != nil {
		return "", err
	}

	var result ApplicationResponse
	query := createApplicationQuery(appGraphql)

	err = c.execute(query, &result)
	if err != nil {
		return "", err
	}

	return result.Result.ID, nil
}

func (c client) DeleteApplication(appID string) error {

	var result ApplicationResponse
	query := deleteApplicationQuery(appID)

	err := c.execute(query, &result)
	if err != nil {
		return err
	}

	return nil
}

func (c client) CreateRuntime(in schema.RuntimeInput) (string, error) {

	runtimeGraphQL, err := c.graphqulizer.RuntimeInputToGQL(in)

	var result RuntimeResponse
	query := createRuntimeQuery(runtimeGraphQL)

	err = c.execute(query, &result)
	if err != nil {
		return "", err
	}

	return result.Result.ID, nil
}

func (c client) DeleteRuntime(runtimeID string) error {

	var result RuntimeResponse
	query := deleteRuntimeQuery(runtimeID)

	err := c.execute(query, &result)
	if err != nil {
		return err
	}

	return nil
}

func (c client) SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error {

	{
		query := setEventBaseURLQuery(runtimeID, eventsBaseURL)
		var response SetLabelResponse

		err := c.execute(query, &response)
		if err != nil {
			return err
		}
	}

	{
		query := setDefaultEventingQuery(runtimeID, appID)
		var response SetDefaultAppEventingResponse

		err := c.execute(query, &response)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c client) GetOneTimeTokenUrl(appID string) (string, string, error) {

	query := getOneTimeTokenQuery(appID)
	var response OneTimeTokenResponse

	err := c.execute(query, &response)
	if err != nil {
		return "", "", err
	}

	return response.Result.LegacyConnectorURL, response.Result.Token, nil
}

func getInternalTenantID(directorURL string, externalTenantID string) (string, error) {

	query := getTenantsQuery()

	req := gcli.NewRequest(query)

	token, err := getToken(externalTenantID, []string{"tenant:read"})
	if err != nil {
		return "", err
	}

	client := gqlTools.NewAuthorizedGraphQLClientWithCustomURL(token, directorURL)

	var response TenantsResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Run(ctx, req, &response)
	if err != nil {
		return "", err
	}

	for _, tenant := range response.Result {
		if tenant.ID == externalTenantID {
			return tenant.InternalID, nil
		}
	}

	return "", errors.New("Cannot find test tenant.")
}

func (c client) execute(query string, res interface{}) error {

	req := gcli.NewRequest(query)
	req.Header.Set("Tenant", c.tenant)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.client.Run(ctx, req, res)

	return err
}

func getClient(url string, tenant string, scopes []string) (*gcli.Client, error) {

	token, err := getToken(tenant, scopes)
	if err != nil {
		return nil, err
	}

	return gqlTools.NewAuthorizedGraphQLClientWithCustomURL(token, url), nil
}

func getToken(tenant string, scopes []string) (string, error) {
	token, err := jwtbuilder.Do(tenant, scopes)
	if err != nil {
		return "", err
	}

	return token, nil
}
