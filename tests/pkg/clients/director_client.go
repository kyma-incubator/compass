package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqlTools "github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	consumerID   = "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"
	consumerType = "USER"
)

type Client interface {
	CreateRuntime(in schema.RuntimeInput) (string, error)
	DeleteRuntime(runtimeID string) error
	SetRuntimeLabel(runtimeID, key, value string) error
	CreateApplication(in schema.ApplicationRegisterInput) (string, error)
	DeleteApplication(appID string) error
	SetApplicationLabel(applicationID, key, value string) error
	DeleteApplicationLabel(applicationID, key string) error
	SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error
	GetOneTimeTokenUrl(appID string) (string, string, error)
}

type DirectorClient struct {
	scopes       []string
	tenant       string
	graphqulizer graphqlizer.Graphqlizer
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

func NewDirectorClient(directorURL, directorHealthzURL, tenant string, scopes []string) (Client, error) {

	err := waitUntilDirectorIsReady(directorHealthzURL)
	if err != nil {
		return nil, errors.Wrap(err, "Director is not ready")
	}

	internalTenantID, err := getInternalTenantID(directorURL, tenant)
	if err != nil {
		return nil, err
	}

	gqlClient, err := getClient(directorURL, internalTenantID, scopes)
	if err != nil {
		return nil, err
	}

	return &DirectorClient{
		scopes:       scopes,
		tenant:       tenant,
		graphqulizer: graphqlizer.Graphqlizer{},
		client:       gqlClient,
	}, nil
}

func waitUntilDirectorIsReady(directorHealthzURL string) error {
	httpClient := http.Client{}

	return retry.Do(func() error {
		req, err := http.NewRequest(http.MethodGet, directorHealthzURL, nil)
		if err != nil {
			logrus.Warningf("Failed to create request while waiting for Director: %s", err)
			return err
		}

		res, err := httpClient.Do(req)
		if err != nil {
			logrus.Warningf("Failed to execute request while waiting for Director: %s", err)
			return err
		}

		err = res.Body.Close()
		if err != nil {
			logrus.Warningf("Failed to close request body while waiting for Director: %s", err)
		}

		if res.StatusCode != http.StatusOK {
			return errors.New("Unexpected status code received when waiting for Director: " + res.Status)
		}

		return nil
	}, defaultRetryOptions()...)
}

func (c *DirectorClient) CreateApplication(in schema.ApplicationRegisterInput) (string, error) {

	appGraphql, err := c.graphqulizer.ApplicationRegisterInputToGQL(in)
	if err != nil {
		return "", err
	}

	var result ApplicationResponse
	query := fixtures.FixRegisterApplicationRequest(appGraphql)

	err = c.executeWithRetries(query.Query(), &result)
	if err != nil {
		return "", err
	}

	return result.Result.ID, nil
}

func (c *DirectorClient) DeleteApplication(appID string) error {

	var result ApplicationResponse
	query := fixtures.FixUnregisterApplicationRequest(appID)

	err := c.executeWithRetries(query.Query(), &result)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) SetApplicationLabel(applicationID, key, value string) error {
	query := fixtures.FixSetApplicationLabelRequest(applicationID, key, value)
	var response SetLabelResponse

	err := c.executeWithRetries(query.Query(), &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) DeleteApplicationLabel(applicationID, key string) error {
	query := fixtures.FixDeleteApplicationLabelRequest(applicationID, key)
	var response SetLabelResponse

	err := c.executeWithRetries(query.Query(), &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) CreateRuntime(in schema.RuntimeInput) (string, error) {

	runtimeGraphQL, err := c.graphqulizer.RuntimeInputToGQL(in)

	var result RuntimeResponse
	query := fixtures.FixRegisterRuntimeRequest(runtimeGraphQL)

	err = c.executeWithRetries(query.Query(), &result)
	if err != nil {
		return "", err
	}

	return result.Result.ID, nil
}

func (c *DirectorClient) DeleteRuntime(runtimeID string) error {

	var result RuntimeResponse
	query := fixtures.FixUnregisterRuntimeRequest(runtimeID)

	err := c.executeWithRetries(query.Query(), &result)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) SetRuntimeLabel(runtimeID, key, value string) error {
	query := fixtures.FixSetRuntimeLabelRequest(runtimeID, key, value)
	var response SetLabelResponse

	err := c.executeWithRetries(query.Query(), &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error {

	{
		query := fixtures.FixSetRuntimeLabelRequest(runtimeID, "runtime_eventServiceUrl", eventsBaseURL)
		var response SetLabelResponse

		err := c.executeWithRetries(query.Query(), &response)
		if err != nil {
			return err
		}
	}

	{
		query := fixtures.FixSetDefaultEventingForApplication(appID, runtimeID)
		var response SetDefaultAppEventingResponse

		err := c.executeWithRetries(query.Query(), &response)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DirectorClient) GetOneTimeTokenUrl(appID string) (string, string, error) {

	query := fixtures.FixRequestOneTimeTokenForApplication(appID)
	var response OneTimeTokenResponse

	err := c.executeWithRetries(query.Query(), &response)
	if err != nil {
		return "", "", err
	}

	return response.Result.LegacyConnectorURL, response.Result.Token, nil
}

func getInternalTenantID(directorURL string, externalTenantID string) (string, error) {

	query := fixtures.FixTenantsRequest()

	req := gcli.NewRequest(query.Query())

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

func (c *DirectorClient) executeWithRetries(query string, res interface{}) error {
	return retry.Do(func() error {
		req := gcli.NewRequest(query)
		req.Header.Set("Tenant", c.tenant)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := c.client.Run(ctx, req, res)

		return err
	}, defaultRetryOptions()...)

}

func getClient(url string, tenant string, scopes []string) (*gcli.Client, error) {

	token, err := getToken(tenant, scopes)
	if err != nil {
		return nil, err
	}

	return gqlTools.NewAuthorizedGraphQLClientWithCustomURL(token, url), nil
}

func getToken(tenant string, scopes []string) (string, error) {
	token, err := jwtbuilder.Build(tenant, scopes, &jwtbuilder.Consumer{ID: consumerID, Type: consumerType})
	if err != nil {
		return "", err
	}

	return token, nil
}

func defaultRetryOptions() []retry.Option {
	return []retry.Option{retry.Attempts(20), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second)}
}
