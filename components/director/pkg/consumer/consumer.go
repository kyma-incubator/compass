package consumer

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
)

// Type missing godoc
type Type string

const (
	// Runtime represents a runtime consumer type
	Runtime Type = "Runtime"
	// ExternalCertificate missing godoc
	ExternalCertificate Type = "External Certificate"
	// Application represents an application consumer type
	Application Type = "Application"
	// IntegrationSystem represents an integration system consumer type
	IntegrationSystem Type = "Integration System"
	// User missing godoc
	User Type = "Static User"
	// SuperAdmin is a consumer type that is used only in our tests
	SuperAdmin Type = "Super Admin"
	// TechnicalClient is a consumer type that is used by Atom
	TechnicalClient Type = "Technical Client"
	// BusinessIntegration is a consumer type that is used by Business Integration operator
	BusinessIntegration Type = "Business Integration"
	// ManagedApplicationProviderOperator is a consumer type that is used by Managed Application Provider operator
	ManagedApplicationProviderOperator Type = "Managed Application Provider Operator"
	// ManagedApplicationConsumer is a consumer type that is used by Managed Application Provider operator
	// when creating Certificate Subject Mappings
	ManagedApplicationConsumer Type = "Managed Application Consumer"
	// ApplicationProvider is type that provides application templates and consumes ORD data
	ApplicationProvider Type = "Application Provider"
	// LandscapeResourceOperator is a consumer type that is used by Landscape Resource operator
	LandscapeResourceOperator Type = "Landscape Resource Operator"
	// TenantDiscoveryOperator is a consumer type that is used by Tenant Discovery operator
	TenantDiscoveryOperator Type = "Tenant Discovery Operator"
	// InstanceCreator is a consumer type that is used by Instance Creator operator
	InstanceCreator Type = "Instance Creator"
	// FormationViewer is a consumer type that is used by Instance Creator operator
	FormationViewer Type = "Formation Viewer"
)

// Consumer missing godoc
type Consumer struct {
	ConsumerID    string `json:"ConsumerID"`
	Type          `json:"ConsumerType"`
	Flow          oathkeeper.AuthFlow `json:"Flow"`
	OnBehalfOf    string              `json:"onBehalfOf"`
	Region        string              `json:"region"`
	Subject       string              `json:"subject"`
	TokenClientID string              `json:"tokenClientID"`
}

// MapSystemAuthToConsumerType missing godoc
func MapSystemAuthToConsumerType(refObj model.SystemAuthReferenceObjectType) (Type, error) {
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
	case model.ApplicationProviderReference:
		return ApplicationProvider, nil
	case model.ManagedApplicationConsumerReference:
		return ManagedApplicationConsumer, nil
	case model.LandscapeResourceOperatorConsumerReference:
		return LandscapeResourceOperator, nil
	case model.TenantDiscoveryOperatorConsumerReference:
		return TenantDiscoveryOperator, nil
	case model.InstanceCreatorConsumerReference:
		return InstanceCreator, nil
	case model.FormationViewerConsumerReference:
		return FormationViewer, nil
	case model.SuperAdminReference:
		return SuperAdmin, nil
	}
	return "", apperrors.NewInternalError("unknown reference object type")
}
