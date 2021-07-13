package systemfetcher

import (
	"context"
	"fmt"

	"github.com/form3tech-oss/jwt-go"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Authenticator interface {
	GetAuthorization(ctx context.Context, tenant string) (string, error)
}

type DirectorGraphClient struct {
	*gcli.Client
	authenticator Authenticator
}

func (d *DirectorGraphClient) DeleteSystemAsync(ctx context.Context, id, tenant string) error {
	gqlRequest := gcli.NewRequest(fmt.Sprintf(`mutation {
			  unregisterApplication(
				id: %q,
				mode: ASYNC
			  ) {
				  id
			    }
			}`, id))
	token, err := d.authenticator.GetAuthorization(ctx, tenant)
	if err != nil {
		return err
	}

	gqlRequest.Header.Set("Authorization", "Bearer "+token)

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
