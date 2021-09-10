package claims

import (
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

type scopeBasedClaimsValidator struct {
	requiredScopes []string
}

// NewScopesValidator missing godoc
func NewScopesValidator(requiredScopes []string) *scopeBasedClaimsValidator {
	return &scopeBasedClaimsValidator{
		requiredScopes: requiredScopes,
	}
}

// Validate missing godoc
func (v *scopeBasedClaimsValidator) Validate(claims Claims) error {
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
