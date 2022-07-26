package formation_test

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

var (
	tenantID          = uuid.New()
	externalTenantID  = uuid.New()
	nilFormationModel *model.Formation

	modelFormation = model.Formation{
		ID:                  FormationID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
	graphqlFormation = graphql.Formation{
		ID:                  FormationID,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
	defaultFormation = model.Formation{
		Name: model.DefaultScenario,
	}
	formationTemplate = model.FormationTemplate{
		ID:          FormationTemplateID,
		Name:        "formation-template",
		RuntimeType: runtimeType,
	}
	runtimeLblFilters = []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("runtimeType", fmt.Sprintf("\"%s\"", runtimeType))}
)

const (
	TargetTenantID          = "targetTenantID"
	ScenarioName            = "scenario-A"
	ScenarioName2           = "scenario-B"
	ErrMsg                  = "some error"
	Tnt                     = "953ac686-5773-4ad0-8eb1-2349e931f852"
	TargetTenant            = "targetTenant"
	ExternalTnt             = "external-tnt"
	TenantID2               = "18271026-3998-4391-be58-b783a09fcca8"
	TargetTenantID2         = "targetTenantID2"
	RuntimeID               = "rt-id"
	RuntimeContextID        = "rt-ctx-id"
	FormationTemplateID     = "bda5378d-caa1-4ee4-b8bf-f733e180fbf9"
	FormationID             = "cf7e396b-ee70-4a47-9aff-9fa9bfa466c1"
	testFormationName       = "test-formation"
	secondTestFormationName = "second-formation"
	runtimeType             = "runtimeType"
)

func unusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func unusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func unusedASAService() *automock.AutomaticFormationAssignmentService {
	return &automock.AutomaticFormationAssignmentService{}
}

func unusedLabelDefServiceFn() *automock.LabelDefService {
	return &automock.LabelDefService{}
}

func unusedASARepo() *automock.AutomaticFormationAssignmentRepository {
	return &automock.AutomaticFormationAssignmentRepository{}
}

func unusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func unusedRuntimeContextRepo() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func unusedLabelDefService() *automock.LabelDefService {
	return &automock.LabelDefService{}
}

// UnusedUUIDService returns a mock uid service that does not expect to get called
func unusedUUIDService() *automock.UuidService {
	return &automock.UuidService{}
}

func unusedConverter() *automock.Converter {
	return &automock.Converter{}
}

func unusedService() *automock.Service {
	return &automock.Service{}
}

func unusedFormationRepo() *automock.FormationRepository {
	return &automock.FormationRepository{}
}

func unusedFormationTemplateRepo() *automock.FormationTemplateRepository {
	return &automock.FormationTemplateRepository{}
}

func fixCtxWithTenant() context.Context {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID.String(), externalTenantID.String())

	return ctx
}

func fixModel(scenarioName string) model.AutomaticScenarioAssignment {
	return fixModelWithScenarioName(scenarioName)
}

func fixModelWithScenarioName(scenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   scenario,
		Tenant:         tenantID.String(),
		TargetTenantID: TargetTenantID,
	}
}

func fixError() error {
	return errors.New(ErrMsg)
}

func mockScenarioDefServiceThatReturns(scenarios []string) *automock.LabelDefService {
	mockScenarioDefSvc := &automock.LabelDefService{}
	mockScenarioDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, tenantID.String()).Return(nil)
	mockScenarioDefSvc.On("GetAvailableScenarios", mock.Anything, tenantID.String()).Return(scenarios, nil)
	return mockScenarioDefSvc
}

func fixUUID() string {
	return FormationID
}

func fixColumns() []string {
	return []string{"id", "tenant_id", "formation_template_id", "name"}
}

func fixDefaultScenariosLabelDefinition(tenantID string, schema interface{}) model.LabelDefinition {
	return model.LabelDefinition{
		Key:     model.ScenariosKey,
		Tenant:  tenantID,
		Schema:  &schema,
		Version: 1,
	}
}

func fixAutomaticScenarioAssigment(selectorScenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   selectorScenario,
		Tenant:         tenantID.String(),
		TargetTenantID: TargetTenantID,
	}
}

func fixFormationTemplateModel() *model.FormationTemplate {
	return &model.FormationTemplate{
		ID:                     FormationTemplateID,
		Name:                   "formation-tmpl-name",
		ApplicationTypes:       []string{"appType1", "appType2"},
		RuntimeType:            "runtimeTypes",
		RuntimeTypeDisplayName: "runtimeDisplayName",
		RuntimeArtifactKind:    model.RuntimeArtifactKindEnvironmentInstance,
	}
}

func fixFormationModel() *model.Formation {
	return &model.Formation{
		ID:                  FormationID,
		TenantID:            Tnt,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
}

func fixFormationEntity() *formation.Entity {
	return &formation.Entity{
		ID:                  FormationID,
		TenantID:            Tnt,
		FormationTemplateID: FormationTemplateID,
		Name:                testFormationName,
	}
}
