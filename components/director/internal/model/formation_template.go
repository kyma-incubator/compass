package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

// FormationTemplate missing godoc
type FormationTemplate struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	ApplicationTypes       []string `json:"applicationTypes"`
	RuntimeType            string   `json:"runtimeType"`
	RuntimeTypeDisplayName string   `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    string   `json:"runtimeArtifactKind"`
}

// FormationTemplateInput missing godoc
type FormationTemplateInput struct {
	Name                   string   `json:"name"`
	ApplicationTypes       []string `json:"applicationTypes"`
	RuntimeType            string   `json:"runtimeType"`
	RuntimeTypeDisplayName string   `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    string   `json:"runtimeArtifactKind"`
}

// FormationTemplatePage missing godoc
type FormationTemplatePage struct {
	Data       []*FormationTemplate
	PageInfo   *pagination.Page
	TotalCount int
}
