package systemfetcher

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Authenticator interface {
	GetAuthorization(ctx context.Context) (string, error)
}

type DirectorGraphClient struct {
	*gcli.Client
	Authenticator Authenticator
}

func (d *DirectorGraphClient) DeleteSystemAsync(ctx context.Context, id, tenantID string) error {
	gqlRequest := gcli.NewRequest(fmt.Sprintf(`mutation {
			  unregisterApplication(
				id: %q,
				mode: ASYNC
			  ) {
				  id
			    }
			}`, id))
	ctx = tenant.SaveToContext(ctx, tenantID)
	token, err := d.Authenticator.GetAuthorization(ctx)
	if err != nil {
		return err
	}

	gqlRequest.Header.Set("Authorization", token)
	gqlRequest.Header.Set("tenant", tenantID)

	if err := d.Run(ctx, gqlRequest, nil); err != nil {
		return errors.Wrapf(err, "while executing GraphQL call to delete a system with id %s", id)
	}
	return nil
}
