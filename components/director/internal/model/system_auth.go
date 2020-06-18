package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type SystemAuth struct {
	ID                  string
	TenantID            *string
	AppID               *string
	RuntimeID           *string
	IntegrationSystemID *string
	Value               *Auth
}

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

type SystemAuthReferenceObjectType string

const (
	RuntimeReference           SystemAuthReferenceObjectType = "Runtime"
	ApplicationReference       SystemAuthReferenceObjectType = "Application"
	IntegrationSystemReference SystemAuthReferenceObjectType = "Integration System"
)

func IsIntegrationSystemNoTenantFlow(err error, objectType SystemAuthReferenceObjectType) bool {
	return apperrors.IsTenantRequired(err) && objectType == IntegrationSystemReference
}
