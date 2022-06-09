package formationtemplate_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

const (
	testID = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
)

var (
	inputFormationTemplateModel = model.FormationTemplateInput{
		Name:                          "formation-template-name",
		ApplicationTypes:              []string{"some-application-type"},
		RuntimeTypes:                  []string{"some-runtime-type"},
		MissingArtifactInfoMessage:    "some missing info message",
		MissingArtifactWarningMessage: "some missing warning message",
	}
	inputFormationTemplateGraphQLModel = graphql.FormationTemplateInput{
		Name:                          inputFormationTemplateModel.Name,
		ApplicationTypes:              inputFormationTemplateModel.ApplicationTypes,
		RuntimeTypes:                  inputFormationTemplateModel.RuntimeTypes,
		MissingArtifactInfoMessage:    inputFormationTemplateModel.MissingArtifactInfoMessage,
		MissingArtifactWarningMessage: inputFormationTemplateModel.MissingArtifactWarningMessage,
	}
	formationTemplateModel = model.FormationTemplate{
		ID:                            testID,
		Name:                          inputFormationTemplateModel.Name,
		ApplicationTypes:              inputFormationTemplateModel.ApplicationTypes,
		RuntimeTypes:                  inputFormationTemplateModel.RuntimeTypes,
		MissingArtifactInfoMessage:    inputFormationTemplateModel.MissingArtifactInfoMessage,
		MissingArtifactWarningMessage: inputFormationTemplateModel.MissingArtifactWarningMessage,
	}
	formationTemplateGraphQLModel = graphql.FormationTemplate{
		ID:                            testID,
		Name:                          inputFormationTemplateModel.Name,
		ApplicationTypes:              inputFormationTemplateModel.ApplicationTypes,
		RuntimeTypes:                  inputFormationTemplateModel.RuntimeTypes,
		MissingArtifactInfoMessage:    inputFormationTemplateModel.MissingArtifactInfoMessage,
		MissingArtifactWarningMessage: inputFormationTemplateModel.MissingArtifactWarningMessage,
	}
	formationTemplateModelPage = model.FormationTemplatePage{
		Data: []*model.FormationTemplate{&formationTemplateModel},
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 1,
	}
	formationTemplateGraphQLModelPage = graphql.FormationTemplatePage{
		Data: []*graphql.FormationTemplate{&formationTemplateGraphQLModel},
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 1,
	}
)

func UnusedFormationTemplateService() *automock.FormationTemplateService {
	return &automock.FormationTemplateService{}
}

func UnusedFormationTemplateRepository() *automock.FormationTemplateRepository {
	return &automock.FormationTemplateRepository{}
}

func UnusedFormationTemplateConverter() *automock.FormationTemplateConverter {
	return &automock.FormationTemplateConverter{}
}
