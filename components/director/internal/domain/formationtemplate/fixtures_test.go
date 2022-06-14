package formationtemplate_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

const (
	testID = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
)

var (
	nilModelEntity              *model.FormationTemplate
	inputFormationTemplateModel = model.FormationTemplateInput{
		Name:                   "formation-template-name",
		ApplicationTypes:       []string{"some-application-type"},
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    "SUBSCRIPTION",
	}
	inputFormationTemplateGraphQLModel = graphql.FormationTemplateInput{
		Name:                   inputFormationTemplateModel.Name,
		ApplicationTypes:       inputFormationTemplateModel.ApplicationTypes,
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    "SUBSCRIPTION",
	}
	formationTemplateModel = model.FormationTemplate{
		ID:                     testID,
		Name:                   inputFormationTemplateModel.Name,
		ApplicationTypes:       inputFormationTemplateModel.ApplicationTypes,
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    "SUBSCRIPTION",
	}
	formationTemplateEntity = formationtemplate.Entity{
		ID:                     testID,
		Name:                   inputFormationTemplateModel.Name,
		ApplicationTypes:       "[\"some-application-type\"]",
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    "SUBSCRIPTION",
	}
	formationTemplateGraphQLModel = graphql.FormationTemplate{
		ID:                     testID,
		Name:                   inputFormationTemplateModel.Name,
		ApplicationTypes:       inputFormationTemplateModel.ApplicationTypes,
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    graphql.ArtifactTypeSubscription,
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

func fixColumns() []string {
	return []string{"id", "name", "application_types", "runtime_type", "runtime_type_display_name", "runtime_artifact_kind"}
}

func UnusedFormationTemplateService() *automock.FormationTemplateService {
	return &automock.FormationTemplateService{}
}

func UnusedFormationTemplateRepository() *automock.FormationTemplateRepository {
	return &automock.FormationTemplateRepository{}
}

func UnusedFormationTemplateConverter() *automock.FormationTemplateConverter {
	return &automock.FormationTemplateConverter{}
}
