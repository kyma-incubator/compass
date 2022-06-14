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
	formationTemplateModelInput = model.FormationTemplateInput{
		Name:                   "formation-template-name",
		ApplicationTypes:       []string{"some-application-type"},
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    "SUBSCRIPTION",
	}
	formationTemplateGraphQLInput = graphql.FormationTemplateInput{
		Name:                   formationTemplateModelInput.Name,
		ApplicationTypes:       formationTemplateModelInput.ApplicationTypes,
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    "SUBSCRIPTION",
	}
	formationTemplateModel = model.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateModelInput.Name,
		ApplicationTypes:       formationTemplateModelInput.ApplicationTypes,
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    "SUBSCRIPTION",
	}
	formationTemplateEntity = formationtemplate.Entity{
		ID:                     testID,
		Name:                   formationTemplateModelInput.Name,
		ApplicationTypes:       "[\"some-application-type\"]",
		RuntimeType:            "some-runtime-type",
		RuntimeTypeDisplayName: "display-name-for-runtime",
		RuntimeArtifactKind:    "SUBSCRIPTION",
	}
	graphQLFormationTemplate = graphql.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateModelInput.Name,
		ApplicationTypes:       formationTemplateModelInput.ApplicationTypes,
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
	graphQLFormationTemplatePage = graphql.FormationTemplatePage{
		Data: []*graphql.FormationTemplate{&graphQLFormationTemplate},
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
