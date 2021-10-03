package claims

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

type validator struct{}

// NewValidator missing godoc
func NewValidator() *validator {
	return &validator{}
}

// Validate missing godoc
func (*validator) Validate(claims Claims) error {
	if err := claims.Valid(); err != nil {
		return errors.Wrapf(err, "while validating claims")
	}

	if claims.Tenant["consumerTenant"] == "" && claims.Tenant["externalTenant"] != "" {
		return apperrors.NewTenantNotFoundError(claims.Tenant["externalTenant"])
	}

	if len(claims.Consumers) > 1{
		return apperrors.NewMultipleConsumersNotSupportedError()
	}

	return nil
}
