package scope

import "github.com/kyma-incubator/compass/components/director/pkg/apperrors"

// HasScopesErrorProvider missing godoc
type HasScopesErrorProvider struct{}

// Error missing godoc
func (s *HasScopesErrorProvider) Error(requiredScopes, actualScopes []string) error {
	return apperrors.NewInsufficientScopesError(requiredScopes, actualScopes)
}

// SanitizeErrorProvider missing godoc
type SanitizeErrorProvider struct{}

// Error missing godoc
func (s *SanitizeErrorProvider) Error(_, _ []string) error {
	return nil
}
