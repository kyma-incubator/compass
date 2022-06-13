package model

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

// SystemAuth missing godoc
type SystemAuth struct {
	ID                  string
	TenantID            *string
	AppID               *string
	RuntimeID           *string
	IntegrationSystemID *string
	Value               *model.Auth
}

// GetReferenceObjectType missing godoc
func (sa SystemAuth) GetReferenceObjectType() (SystemAuthReferenceObjectType, error) {
	if sa.AppID != nil {
		return ApplicationReference, nil
	}

	if sa.RuntimeID != nil {
		return RuntimeReference, nil
	}

	if sa.IntegrationSystemID != nil {
		return IntegrationSystemReference, nil
	}

	return "", apperrors.NewInternalError("unknown reference object type")
}

// GetReferenceObjectID missing godoc
func (sa SystemAuth) GetReferenceObjectID() (string, error) {
	if sa.AppID != nil {
		return *sa.AppID, nil
	}

	if sa.RuntimeID != nil {
		return *sa.RuntimeID, nil
	}

	if sa.IntegrationSystemID != nil {
		return *sa.IntegrationSystemID, nil
	}

	return "", apperrors.NewInternalError("unknown reference object ID")
}

// SystemAuthReferenceObjectType missing godoc
type SystemAuthReferenceObjectType string

const (
	// RuntimeReference missing godoc
	RuntimeReference SystemAuthReferenceObjectType = "Runtime"
	// ExternalCertificateReference missing godoc
	ExternalCertificateReference SystemAuthReferenceObjectType = "External Certificate"
	// ApplicationReference missing godoc
	ApplicationReference SystemAuthReferenceObjectType = "Application"
	// IntegrationSystemReference missing godoc
	IntegrationSystemReference SystemAuthReferenceObjectType = "Integration System"
)

// IsIntegrationSystemNoTenantFlow missing godoc
func IsIntegrationSystemNoTenantFlow(err error, objectType SystemAuthReferenceObjectType) bool {
	return apperrors.IsTenantRequired(err) && objectType == IntegrationSystemReference
}
