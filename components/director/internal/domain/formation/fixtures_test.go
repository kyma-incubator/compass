package formation

import (
	"context"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

var TenantID = uuid.New()
var ExternalTenantID = uuid.New()

const (
	TargetTenantID   = "targetTenantID"
	ScenarioName     = "scenario-A"
	ScenarioName2    = "scenario-B"
	ErrMsg           = "some error"
	Tnt              = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	TargetTenant     = "targetTenant"
	ExternalTnt      = "external-tnt"
	TenantID2        = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	TargetTenantID2  = "targetTenantID2"
	RuntimeID        = "rt-id"
	RuntimeContextID = "rt-ctx-id"
)

func UnusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func UnusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func UnusedASAService() *automock.AutomaticFormationAssignmentService {
	return &automock.AutomaticFormationAssignmentService{}
}

func UnusedLabelDefServiceFn() *automock.LabelDefService {
	return &automock.LabelDefService{}
}

func UnusedASARepo() *automock.AutomaticFormationAssignmentRepository {
	return &automock.AutomaticFormationAssignmentRepository{}
}

func UnusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func UnusedRuntimeContextRepo() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func UnusedLabelDefService() *automock.LabelDefService {
	return &automock.LabelDefService{}
}

func FixCtxWithTenant() context.Context {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TenantID.String(), ExternalTenantID.String())

	return ctx
}

func FixModel() model.AutomaticScenarioAssignment {
	return fixModelWithScenarioName(ScenarioName)
}

func fixModelWithScenarioName(scenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   scenario,
		Tenant:         TenantID.String(),
		TargetTenantID: TargetTenantID,
	}
}

func FixError() error {
	return errors.New(ErrMsg)
}

func MockScenarioDefServiceThatReturns(scenarios []string) *automock.LabelDefService {
	mockScenarioDefSvc := &automock.LabelDefService{}
	mockScenarioDefSvc.On("EnsureScenariosLabelDefinitionExists", mock.Anything, TenantID.String()).Return(nil)
	mockScenarioDefSvc.On("GetAvailableScenarios", mock.Anything, TenantID.String()).Return(scenarios, nil)
	return mockScenarioDefSvc
}

func FixUUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}

func FixDefaultScenariosLabelDefinition(tenantID string, schema interface{}) model.LabelDefinition {
	return model.LabelDefinition{
		Key:     model.ScenariosKey,
		Tenant:  tenantID,
		Schema:  &schema,
		Version: 1,
	}
}

func FixAutomaticScenarioAssigment(selectorScenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   selectorScenario,
		Tenant:         TenantID.String(),
		TargetTenantID: TargetTenantID,
	}
}
