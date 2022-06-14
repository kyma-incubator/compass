package formationtemplate_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

const (
	testID                   = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	formationTemplateName    = "formation-template-name"
	runtimeType              = "some-runtime-type"
	runtimeTypeDisplayName   = "display-name-for-runtime"
	artifactKindAsString     = "SUBSCRIPTION"
	applicationTypesAsString = "[\"some-application-type\"]"
)

var (
	applicationTypes = []string{"some-application-type"}
)

var (
	nilModelEntity              *model.FormationTemplate
	formationTemplateModelInput = model.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeType:            runtimeType,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
	}
	formationTemplateGraphQLInput = graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeType:            runtimeType,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
	}
	formationTemplateModel = model.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeType:            runtimeType,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
	}
	formationTemplateEntity = formationtemplate.Entity{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypesAsString,
		RuntimeType:            runtimeType,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
	}
	graphQLFormationTemplate = graphql.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeType:            runtimeType,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
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
