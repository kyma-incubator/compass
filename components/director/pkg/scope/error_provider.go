package scope

import "github.com/kyma-incubator/compass/components/director/pkg/apperrors"

type HasScopesErrorProvider struct{}

func (s *HasScopesErrorProvider) Error(requiredScopes, actualScopes []string) error {
	return apperrors.NewInsufficientScopesError(requiredScopes, actualScopes)
}

type SanitizeErrorProvider struct{}

func (s *SanitizeErrorProvider) Error(_, _ []string) error {
	return nil
}
