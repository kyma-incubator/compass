package claims

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

type validator struct{}

func NewValidator() *validator {
	return &validator{}
}

func (*validator) Validate(claims Claims) error {
	if err := claims.Valid(); err != nil {
		return errors.Wrapf(err, "while validating claims")
	}

	if claims.Tenant == "" && claims.ExternalTenant != "" {
		return apperrors.NewTenantNotFoundError(claims.ExternalTenant)
	}

	return nil
}
