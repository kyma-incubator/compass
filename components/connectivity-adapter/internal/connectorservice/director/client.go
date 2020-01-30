package director

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/pkg/errors"

	externalSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/machinebox/graphql"
)

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	GetApplication(systemAuthID string) (externalSchema.ApplicationExt, apperrors.AppError)
}

func NewClient(gqlClient *graphql.Client) Client {
	return client{
		gqlClient: gqlClient,
		timeout:   30 * time.Second,
	}
}

type client struct {
	gqlClient *graphql.Client
	timeout   time.Duration
}

func (c client) GetApplication(systemAuthID string) (externalSchema.ApplicationExt, apperrors.AppError) {

	logrus.Info("System auth ID: " + systemAuthID)
	appID, err := c.getApplicationID(systemAuthID)
	if err != nil {
		return externalSchema.ApplicationExt{}, apperrors.Internal(err.Error())
	}

	query := applicationQuery(appID)
	var response ApplicationResponse

	err = c.execute(c.gqlClient, query, &response)
	if err != nil {
		return externalSchema.ApplicationExt{}, apperrors.Internal("Failed to get application: %s", err)
	}

	return response.Result, nil
}

func (c client) getApplicationID(systemAuthID string) (string, error) {
	query := viewerQuery()

	var response ViewerResponse

	err := c.execute(c.gqlClient, query, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get certificate info")
	}

	return response.Result.ID, nil
}

type ViewerResponse struct {
	Result externalSchema.Viewer `json:"result"`
}

type ApplicationResponse struct {
	Result externalSchema.ApplicationExt `json:"result"`
}

func (c *client) execute(client *graphql.Client, query string, res interface{}) error {

	req := graphql.NewRequest(query)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	err := client.Run(ctx, req, res)

	return err
}
