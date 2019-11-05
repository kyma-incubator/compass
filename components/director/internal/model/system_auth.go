package model

import "github.com/pkg/errors"

type SystemAuth struct {
	ID                  string
	TenantID            string
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

	return "", errors.New("unknown reference object type")
}

const IntegrationSystemTenant = "00000000-00000000-00000000-00000000"

type SystemAuthReferenceObjectType string

const (
	RuntimeReference           SystemAuthReferenceObjectType = "Runtime"
	ApplicationReference       SystemAuthReferenceObjectType = "Application"
	IntegrationSystemReference SystemAuthReferenceObjectType = "Integration System"
)
