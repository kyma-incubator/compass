package clients

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	consumerID   = "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"
	consumerType = "USER"

	cleanupRetryCount = 3
	defaultRetryCount = 20
)

type Client interface {
	CreateRuntime(in schema.RuntimeRegisterInput) (string, error)
	CleanupRuntime(runtimeID string) error
	SetRuntimeLabel(runtimeID, key, value string) error
	CreateApplication(in schema.ApplicationRegisterInput) (string, error)
	CleanupApplication(appID string) error
	SetApplicationLabel(applicationID, key, value string) error
	DeleteApplicationLabel(applicationID, key string) error
	SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error
	GetOneTimeTokenUrl(appID string) (string, string, error)
}

type DirectorClient struct {
	tenant       string
	graphqulizer graphqlizer.Graphqlizer
	client       *gcli.Client
}

func (c *DirectorClient) CleanupRuntime(runtimeID string) error {
	if runtimeID == "" {
		return nil
	}

	var result RuntimeResponse
	query := fixtures.FixUnregisterRuntimeRequest(runtimeID)

	err := c.executeWithRetries(query.Query(), &result, retryOptions(cleanupRetryCount)...)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}

	return nil
}

func (c *DirectorClient) CleanupApplication(appID string) error {
	if appID == "" {
		return nil
	}
	var result ApplicationResponse
	query := fixtures.FixUnregisterApplicationRequest(appID)

	err := c.executeWithRetries(query.Query(), &result, retryOptions(cleanupRetryCount)...)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}

	return nil
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

func NewDirectorClient(gqlClient *graphql.Client, tenant, readyzURL string) (Client, error) {
	err := waitUntilDirectorIsReady(readyzURL)
	if err != nil {
		return nil, errors.Wrap(err, "Director is not ready")
	}

	return &DirectorClient{
		tenant:       tenant,
		graphqulizer: graphqlizer.Graphqlizer{},
		client:       gqlClient,
	}, nil
}

func waitUntilDirectorIsReady(directorReadyzURL string) error {
	httpClient := http.Client{}

	return retry.Do(func() error {
		req, err := http.NewRequest(http.MethodGet, directorReadyzURL, nil)
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
	}, retryOptions(defaultRetryCount)...)
}

func (c *DirectorClient) CreateApplication(in schema.ApplicationRegisterInput) (string, error) {
	appGraphql, err := c.graphqulizer.ApplicationRegisterInputToGQL(in)
	if err != nil {
		return "", err
	}

	var result ApplicationResponse
	query := fixtures.FixRegisterApplicationRequest(appGraphql)

	err = c.executeWithDefaultRetries(query.Query(), &result)
	if err != nil {
		return "", err
	}

	return result.Result.ID, nil
}

func (c *DirectorClient) SetApplicationLabel(applicationID, key, value string) error {
	query := fixtures.FixSetApplicationLabelRequest(applicationID, key, value)
	var response SetLabelResponse

	err := c.executeWithDefaultRetries(query.Query(), &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) DeleteApplicationLabel(applicationID, key string) error {
	query := fixtures.FixDeleteApplicationLabelRequest(applicationID, key)
	var response SetLabelResponse

	err := c.executeWithDefaultRetries(query.Query(), &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) CreateRuntime(in schema.RuntimeRegisterInput) (string, error) {
	runtimeGraphQL, err := c.graphqulizer.RuntimeRegisterInputToGQL(in)
	if err != nil {
		return "", err
	}

	var result RuntimeResponse
	query := fixtures.FixRegisterRuntimeRequest(runtimeGraphQL)

	if err = c.executeWithDefaultRetries(query.Query(), &result); err != nil {
		return "", err
	}

	return result.Result.ID, nil
}

func (c *DirectorClient) SetRuntimeLabel(runtimeID, key, value string) error {
	query := fixtures.FixSetRuntimeLabelRequest(runtimeID, key, value)
	var response SetLabelResponse

	err := c.executeWithDefaultRetries(query.Query(), &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error {
	query := fixtures.FixSetRuntimeLabelRequest(runtimeID, "runtime_eventServiceUrl", eventsBaseURL)
	var response SetLabelResponse

	err := c.executeWithDefaultRetries(query.Query(), &response)
	if err != nil {
		return err
	}

	query = fixtures.FixSetDefaultEventingForApplication(appID, runtimeID)
	var resp SetDefaultAppEventingResponse

	err = c.executeWithDefaultRetries(query.Query(), &resp)
	if err != nil {
		return err
	}

	return nil
}

func (c *DirectorClient) GetOneTimeTokenUrl(appID string) (string, string, error) {

	query := fixtures.FixRequestOneTimeTokenForApplication(appID)
	var response OneTimeTokenResponse

	err := c.executeWithDefaultRetries(query.Query(), &response)
	if err != nil {
		return "", "", err
	}

	return response.Result.LegacyConnectorURL, response.Result.Token, nil
}

func (c *DirectorClient) executeWithRetries(query string, res interface{}, opts ...retry.Option) error {
	return retry.Do(func() error {
		req := gcli.NewRequest(query)
		req.Header.Set("Tenant", c.tenant)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := c.client.Run(ctx, req, res)

		return err
	}, opts...)

}

func (c *DirectorClient) executeWithDefaultRetries(query string, res interface{}) error {
	return c.executeWithRetries(query, res, retryOptions(defaultRetryCount)...)
}

func retryOptions(retryCount uint) []retry.Option {
	return []retry.Option{retry.Attempts(retryCount), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second)}
}
