package director

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/retry"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/machinebox/graphql"
)

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	GetApplication(systemAuthID string) (schema.ApplicationExt, apperrors.AppError)
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

func (c client) GetApplication(systemAuthID string) (schema.ApplicationExt, apperrors.AppError) {
	appID, err := c.getApplicationID(systemAuthID)
	if err != nil {
		return schema.ApplicationExt{}, apperrors.Internal(err.Error())
	}

	query := applicationQuery(appID)
	var response ApplicationResponse

	err = c.execute(c.gqlClient, query, &response)
	if err != nil {
		return schema.ApplicationExt{}, apperrors.Internal(err.Error())
	}

	return response.Result, nil
}

func (c client) getApplicationID(systemAuthID string) (string, error) {
	query := viewerQuery()

	var response ViewerResponse

	err := c.execute(c.gqlClient, query, &response)
	if err != nil {
		return "", err
	}

	return response.Result.ID, nil
}

type ViewerResponse struct {
	Result schema.Viewer `json:"result"`
}

type ApplicationResponse struct {
	Result schema.ApplicationExt `json:"result"`
}

func (c *client) execute(client *graphql.Client, query string, res interface{}) error {
	req := graphql.NewRequest(query)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return retry.GQLRun(client.Run, ctx, req, res)
}
