package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

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
	ID                     string              `json:"id"`
	Name                   string              `json:"name"`
	ApplicationTypes       []string            `json:"applicationTypes"`
	RuntimeType            string              `json:"runtimeType"`
	RuntimeTypeDisplayName string              `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    RuntimeArtifactKind `json:"runtimeArtifactKind"`
}

// FormationTemplateInput missing godoc
type FormationTemplateInput struct {
	Name                   string              `json:"name"`
	ApplicationTypes       []string            `json:"applicationTypes"`
	RuntimeType            string              `json:"runtimeType"`
	RuntimeTypeDisplayName string              `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    RuntimeArtifactKind `json:"runtimeArtifactKind"`
}

// FormationTemplatePage missing godoc
type FormationTemplatePage struct {
	Data       []*FormationTemplate
	PageInfo   *pagination.Page
	TotalCount int
}
