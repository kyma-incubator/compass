package claimsvalidator

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/token_claims"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"
	"github.com/pkg/errors"
)

// ClaimsValidator implements the ClaimsValidator interface
type ClaimsValidator struct{}

// Validate validates given id_token claims
func (v *ClaimsValidator) Validate(ctx context.Context, claims token_claims.Claims) error {
	if err := claims.Valid(); err != nil {
		return errors.Wrapf(err, "while validating claims")
	}

	if claims.Tenant[tenantmapping.ConsumerTenantKey] == "" && claims.Tenant[tenantmapping.ExternalTenantKey] != "" {
		return apperrors.NewTenantNotFoundError(claims.Tenant[tenantmapping.ExternalTenantKey])
	}

	return nil
}
