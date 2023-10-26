package claims

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/idtokenclaims"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"
	"github.com/pkg/errors"
)

// Validator implements the Validator interface
type Validator struct{}

// Validate validates given id_token claims
func (v *Validator) Validate(ctx context.Context, claims idtokenclaims.Claims) error {
	if err := claims.Valid(); err != nil {
		return errors.Wrapf(err, "while validating claims")
	}

	if claims.Tenant[tenantmapping.ConsumerTenantKey] == "" && claims.Tenant[tenantmapping.ExternalTenantKey] != "" {
		return apperrors.NewTenantNotFoundError(claims.Tenant[tenantmapping.ExternalTenantKey])
	}

	return nil
}
