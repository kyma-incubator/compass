package consumer

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
)

// ConsumerType missing godoc
type ConsumerType string

const (
	// Runtime represents a runtime consumer type
	Runtime ConsumerType = "Runtime"
	// ExternalCertificate missing godoc
	ExternalCertificate ConsumerType = "External Certificate"
	// Application represents an application consumer type
	Application ConsumerType = "Application"
	// IntegrationSystem represents an integration system consumer type
	IntegrationSystem ConsumerType = "Integration System"
	// User missing godoc
	User ConsumerType = "Static User"
	// SuperAdmin is a consumer type that is used only in our tests
	SuperAdmin ConsumerType = "Super Admin"
	// TechnicalClient is a consumer type that is used by Atom
	TechnicalClient ConsumerType = "Technical Client"
	// BusinessIntegration is a consumer type that is used by Business Integration operator
	BusinessIntegration ConsumerType = "Business Integration"
	// ManagedApplicationProviderOperator is a consumer type that is used by Managed Application Provider operator
	ManagedApplicationProviderOperator ConsumerType = "Managed Application Provider Operator"
	// ManagedApplicationConsumer is a consumer type that is used by Managed Application Provider operator
	// when creating Certificate Subject Mappings
	ManagedApplicationConsumer ConsumerType = "Managed Application Consumer"
	// LandscapeResourceOperator is a consumer type that is used by Landscape Resource operator
	LandscapeResourceOperator ConsumerType = "Landscape Resource Operator"
	// TenantDiscoveryOperator is a consumer type that is used by Tenant Discovery operator
	TenantDiscoveryOperator ConsumerType = "Tenant Discovery Operator"
	// InstanceCreator is a consumer type that is used by Instance Creator operator
	InstanceCreator ConsumerType = "Instance Creator"
)

// Consumer missing godoc
type Consumer struct {
	ConsumerID    string `json:"ConsumerID"`
	ConsumerType  `json:"ConsumerType"`
	Flow          oathkeeper.AuthFlow `json:"Flow"`
	OnBehalfOf    string              `json:"onBehalfOf"`
	Region        string              `json:"region"`
	TokenClientID string              `json:"tokenClientID"`
}

// MapSystemAuthToConsumerType missing godoc
func MapSystemAuthToConsumerType(refObj model.SystemAuthReferenceObjectType) (ConsumerType, error) {
	switch refObj {
	case model.ApplicationReference:
		return Application, nil
	case model.ExternalCertificateReference:
		return ExternalCertificate, nil
	case model.RuntimeReference:
		return Runtime, nil
	case model.IntegrationSystemReference:
		return IntegrationSystem, nil
	case model.TechnicalClientReference:
		return TechnicalClient, nil
	case model.BusinessIntegrationReference:
		return BusinessIntegration, nil
	case model.ManagedApplicationProviderOperatorReference:
		return ManagedApplicationProviderOperator, nil
	case model.ManagedApplicationConsumerReference:
		return ManagedApplicationConsumer, nil
	case model.LandscapeResourceOperatorConsumerReference:
		return LandscapeResourceOperator, nil
	case model.TenantDiscoveryOperatorConsumerReference:
		return TenantDiscoveryOperator, nil
	case model.InstanceCreatorConsumerReference:
		return InstanceCreator, nil
	case model.SuperAdminReference:
		return SuperAdmin, nil
	}
	return "", apperrors.NewInternalError("unknown reference object type")
}
