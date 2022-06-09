package model

import "github.com/kyma-incubator/compass/components/director/pkg/pagination"

// FormationTemplate missing godoc
type FormationTemplate struct {
	ID                            string   `json:"id"`
	Name                          string   `json:"name"`
	ApplicationTypes              []string `json:"applicationTypes"`
	RuntimeTypes                  []string `json:"runtimeTypes"`
	MissingArtifactInfoMessage    string   `json:"missingArtifactInfoMessage"`
	MissingArtifactWarningMessage string   `json:"missingArtifactWarningMessage"`
}

// FormationTemplateInput missing godoc
type FormationTemplateInput struct {
	Name                          string   `json:"name"`
	ApplicationTypes              []string `json:"applicationTypes"`
	RuntimeTypes                  []string `json:"runtimeTypes"`
	MissingArtifactInfoMessage    string   `json:"missingArtifactInfoMessage"`
	MissingArtifactWarningMessage string   `json:"missingArtifactWarningMessage"`
}

// FormationTemplatePage missing godoc
type FormationTemplatePage struct {
	Data       []*FormationTemplate
	PageInfo   *pagination.Page
	TotalCount int
}
