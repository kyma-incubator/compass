package claims

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/idtokenclaims"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

type scopeBasedClaimsValidator struct {
	requiredScopes []string
}

// NewScopesValidator creates new scopes validator for given scopes.
func NewScopesValidator(requiredScopes []string) *scopeBasedClaimsValidator {
	return &scopeBasedClaimsValidator{
		requiredScopes: requiredScopes,
	}
}

// Validate validates the scopes in given token claims.
func (v *scopeBasedClaimsValidator) Validate(_ context.Context, claims idtokenclaims.Claims) error {
	if err := claims.Valid(); err != nil {
		return errors.Wrapf(err, "while validating claims")
	}

	if !containsAll(v.requiredScopes, claims.Scopes) {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Not all required scopes %q were found in claim with scopes %q", v.requiredScopes, claims.Scopes))
	}
	return nil
}

func containsAll(stringSlice []string, str string) bool {
	for _, v := range stringSlice {
		if !strings.Contains(str, v) {
			return false
		}
	}
	return true
}
