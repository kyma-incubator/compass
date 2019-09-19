package model

type SystemAuth struct {
	ID                  string
	TenantID            string
	AppID               *string
	RuntimeID           *string
	IntegrationSystemID *string
	Value               *Auth
}

const IntegrationSystemTenant = "00000000-00000000-00000000-00000000"

type SystemAuthReferenceObjectType string

const (
	RuntimeReference           SystemAuthReferenceObjectType = "Runtime"
	ApplicationReference       SystemAuthReferenceObjectType = "Application"
	IntegrationSystemReference SystemAuthReferenceObjectType = "Integration System"
)
