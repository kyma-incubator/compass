package formationtemplate_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

const (
	testID                   = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	formationTemplateName    = "formation-template-name"
	runtimeTypeDisplayName   = "display-name-for-runtime"
	artifactKindAsString     = "SUBSCRIPTION"
	applicationTypesAsString = "[\"some-application-type\"]"
	runtimeTypesAsString     = "[\"some-runtime-type\"]"
	testTenantID             = "d9fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	parentTenantID           = "d8fddec6-5456-4a1e-9ae0-74447f5d6ae9"
)

var (
	nilModelEntity              *model.FormationTemplate
	applicationTypes            = []string{"some-application-type"}
	runtimeTypes                = []string{"some-runtime-type"}
	formationTemplateModelInput = model.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
	}
	formationTemplateGraphQLInput = graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
	}
	formationTemplateModel = model.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
		TenantID:               str.Ptr(testTenantID),
	}
	formationTemplateModelNullTenant = model.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
		TenantID:               nil,
	}
	formationTemplateEntity = formationtemplate.Entity{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypesAsString,
		RuntimeTypes:           runtimeTypesAsString,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
		TenantID:               repo.NewValidNullableString(testTenantID),
	}
	formationTemplateEntityNullTenant = formationtemplate.Entity{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypesAsString,
		RuntimeTypes:           runtimeTypesAsString,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeArtifactKind:    artifactKindAsString,
		TenantID:               repo.NewValidNullableString(""),
	}
	graphQLFormationTemplate = graphql.FormationTemplate{
		ID:                     testID,
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
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

func newModelBusinessTenantMappingWithType(tenantType tenant.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             testTenantID,
		Name:           "name",
		ExternalTenant: "external",
		Parent:         parentTenantID,
		Type:           tenantType,
		Provider:       "test-provider",
		Status:         tenant.Active,
	}
}

func fixColumns() []string {
	return []string{"id", "name", "application_types", "runtime_types", "runtime_type_display_name", "runtime_artifact_kind", "tenant_id"}
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

func UnusedTenantService() *automock.TenantService {
	return &automock.TenantService{}
}
