package systemfetcher

import (
	"context"
	"fmt"

	"github.com/form3tech-oss/jwt-go"
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

	if err := d.Run(ctx, gqlRequest, nil); err != nil {
		return errors.Wrapf(err, "while executing GraphQL call to delete a system with id %s", id)
	}
	return nil
}

type claims struct {
	Scopes string `json:"scopes"`
	Tenant string `json:"tenant"`
	jwt.StandardClaims
}

type authProvider struct {
}

func (a *authProvider) GetAuthorization(ctx context.Context, tenant string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims{
		Scopes: "application:write",
		Tenant: tenant,
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	return signedToken, err
}
