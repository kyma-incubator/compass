package claims

import (
	"context"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"
	"github.com/pkg/errors"
)

type claimsValidator struct{}

// NewClaimsValidator implements the ClaimsValidator interface
func NewClaimsValidator() *claimsValidator {
	return &claimsValidator{}
}

// Validate validates the claims and asserts that the consumerTenant and externalTenant are not empty
func (v *claimsValidator) Validate(_ context.Context, claims Claims) error {
	if err := claims.Valid(); err != nil {
		return errors.Wrapf(err, "while validating claims")
	}

	if claims.Tenant[tenantmapping.ConsumerTenantKey] == "" || claims.Tenant[tenantmapping.ExternalTenantKey] == "" {
		return errors.New("missing tenant")
	}

	return nil
}
