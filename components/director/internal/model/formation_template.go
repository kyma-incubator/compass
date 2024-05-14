package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// RuntimeArtifactKind missing godoc
type RuntimeArtifactKind string

const (
	// RuntimeArtifactKindSubscription missing godoc
	RuntimeArtifactKindSubscription RuntimeArtifactKind = "SUBSCRIPTION"
	// RuntimeArtifactKindServiceInstance missing godoc
	RuntimeArtifactKindServiceInstance RuntimeArtifactKind = "SERVICE_INSTANCE"
	// RuntimeArtifactKindEnvironmentInstance missing godoc
	RuntimeArtifactKindEnvironmentInstance RuntimeArtifactKind = "ENVIRONMENT_INSTANCE"
)

// FormationTemplate missing godoc
type FormationTemplate struct {
	ID                     string                 `json:"id"`
	Name                   string                 `json:"name"`
	ApplicationTypes       []string               `json:"applicationTypes"`
	RuntimeTypes           []string               `json:"runtimeTypes"`
	RuntimeTypeDisplayName *string                `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    *RuntimeArtifactKind   `json:"runtimeArtifactKind"`
	TenantID               *string                `json:"tenant_id"`
	Webhooks               []*Webhook             `json:"webhooks"`
	LeadingProductIDs      []string               `json:"leadingProductIDs"`
	SupportsReset          bool                   `json:"supportsReset"`
	DiscoveryConsumers     []string               `json:"discoveryConsumers"`
	Labels                 map[string]interface{} `json:"labels"`
	CreatedAt              time.Time              `json:"createdAt"`
	UpdatedAt              *time.Time             `json:"updatedAt"`
}

// FormationTemplateRegisterInput is an input used for formation template registration
type FormationTemplateRegisterInput struct {
	Name                   string                 `json:"name"`
	ApplicationTypes       []string               `json:"applicationTypes"`
	RuntimeTypes           []string               `json:"runtimeTypes"`
	RuntimeTypeDisplayName *string                `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    *RuntimeArtifactKind   `json:"runtimeArtifactKind"`
	Webhooks               []*WebhookInput        `json:"webhooks"`
	LeadingProductIDs      []string               `json:"leadingProductIDs"`
	SupportsReset          bool                   `json:"supportsReset"`
	DiscoveryConsumers     []string               `json:"discoveryConsumers"`
	Labels                 map[string]interface{} `json:"labels"`
}

// FormationTemplateUpdateInput is an input used for formation template update operations
type FormationTemplateUpdateInput struct {
	Name                   string               `json:"name"`
	ApplicationTypes       []string             `json:"applicationTypes"`
	RuntimeTypes           []string             `json:"runtimeTypes"`
	RuntimeTypeDisplayName *string              `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    *RuntimeArtifactKind `json:"runtimeArtifactKind"`
	LeadingProductIDs      []string             `json:"leadingProductIDs"`
	SupportsReset          bool                 `json:"supportsReset"`
	DiscoveryConsumers     []string             `json:"discoveryConsumers"`
}

// FormationTemplatePage missing godoc
type FormationTemplatePage struct {
	Data       []*FormationTemplate
	PageInfo   *pagination.Page
	TotalCount int
}
