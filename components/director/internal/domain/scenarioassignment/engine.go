package scenarioassignment

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore
type labelRepository interface {
	GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore
type runtimeRepository interface {
	ListAll(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=labelUpsertService --output=automock --outpkg=automock --case=underscore
type labelUpsertService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

type engine struct {
	labelRepo              labelRepository
	scenarioAssignmentRepo Repository
	labelService           labelUpsertService
	runtimeRepo            runtimeRepository
}

// NewEngine missing godoc
func NewEngine(labelService labelUpsertService, labelRepo labelRepository, scenarioAssignmentRepo Repository, runtimeRepo runtimeRepository) *engine {
	return &engine{
		labelRepo:              labelRepo,
		scenarioAssignmentRepo: scenarioAssignmentRepo,
		labelService:           labelService,
		runtimeRepo:            runtimeRepo,
	}
}

// EnsureScenarioAssigned ensures that the scenario is assigned to all the runtimes that are in the ASAs target_tenant_id
func (e *engine) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	labels, runtimeIDs, err := e.getScenarioLabelsForRuntimes(ctx, in)
	if err != nil {
		return err
	}

	labels = e.appendMissingScenarioLabelsForRuntimes(in.Tenant, runtimeIDs, labels)
	return e.upsertScenarios(ctx, in.Tenant, labels, in.ScenarioName, e.uniqueScenarios)
}

// RemoveAssignedScenario removes all the scenarios that are coming from the provided ASA
func (e *engine) RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	labels, _, err := e.getScenarioLabelsForRuntimes(ctx, in)
	if err != nil {
		return err
	}

	return e.upsertScenarios(ctx, in.Tenant, labels, in.ScenarioName, e.removeScenario)
}

// RemoveAssignedScenarios removes all the scenarios that are coming from any of the provided ASAs
func (e *engine) RemoveAssignedScenarios(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	for _, asa := range in {
		err := e.RemoveAssignedScenario(ctx, *asa)
		if err != nil {
			return errors.Wrapf(err, "while deleting automatic scenario assigment: %s", asa.ScenarioName)
		}
	}
	return nil
}

// MergeScenariosFromInputLabelsAndAssignments merges all the scenarios that are part of the resource labels (already added + to be added with the current operation)
// with all the scenarios that should be assigned based on ASAs.
func (e *engine) MergeScenariosFromInputLabelsAndAssignments(ctx context.Context, inputLabels map[string]interface{}, runtimeID string) ([]interface{}, error) {
	scenariosSet := make(map[string]struct{})

	scenariosFromAssignments, err := e.GetScenariosFromMatchingASAs(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting scenarios for selector labels")
	}

	for _, scenario := range scenariosFromAssignments {
		scenariosSet[scenario] = struct{}{}
	}

	scenariosFromInput, isScenarioLabelInInput := inputLabels[model.ScenariosKey]

	if isScenarioLabelInInput {
		scenariosFromInputInterfaceSlice, ok := scenariosFromInput.([]interface{})
		if !ok {
			return nil, apperrors.NewInternalError("while converting scenarios label to an interface slice")
		}

		for _, scenario := range scenariosFromInputInterfaceSlice {
			scenariosSet[fmt.Sprint(scenario)] = struct{}{}
		}
	}

	scenarios := make([]interface{}, 0)
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios, nil
}

func (e *engine) getScenarioLabelsForRuntimes(ctx context.Context, in model.AutomaticScenarioAssignment) ([]model.Label, []string, error) {
	// Currently. it is not possible to have non-owner access of a runtime in a tenant.
	// It is enough to list all the runtimes in the target tenant.
	runtimes, err := e.runtimeRepo.ListAll(ctx, in.TargetTenantID, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while fetching runtimes in target tenant: %s", in.TargetTenantID)
	}

	if len(runtimes) == 0 {
		return nil, nil, nil
	}

	runtimeIDs := make([]string, 0, len(runtimes))
	for _, runtime := range runtimes {
		runtimeIDs = append(runtimeIDs, runtime.ID)
	}

	labels, err := e.labelRepo.GetScenarioLabelsForRuntimes(ctx, in.Tenant, runtimeIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while fetching scenarios labels for matched runtimes")
	}

	return labels, runtimeIDs, nil
}

// GetScenariosFromMatchingASAs gets all the scenarios that should be added to the runtime based on the matching Automatic Scenario Assignments
// In order to do that, the ASAs should be searched in the caller tenant as this is the tenant that modifies the runtime and this is the tenant that the ASA
// produced labels should be added to.
func (e *engine) GetScenariosFromMatchingASAs(ctx context.Context, runtimeID string) ([]string, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	scenariosSet := make(map[string]struct{})

	scenarioAssignments, err := e.scenarioAssignmentRepo.ListAll(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listinng Automatic Scenario Assignments in tenant: %s", tenantID)
	}

	matchingASAs := make([]*model.AutomaticScenarioAssignment, 0, len(scenarioAssignments))
	for _, scenarioAssignment := range scenarioAssignments {
		matches, err := e.isASAMatchingRuntime(ctx, scenarioAssignment, runtimeID)
		if err != nil {
			return nil, errors.Wrapf(err, "while checkig if asa matches runtime with ID %s", runtimeID)
		}
		if matches {
			matchingASAs = append(matchingASAs, scenarioAssignment)
		}
	}

	for _, sa := range matchingASAs {
		scenariosSet[sa.ScenarioName] = struct{}{}
	}

	scenarios := make([]string, 0)
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios, nil
}

func (e *engine) isASAMatchingRuntime(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeID string) (bool, error) {
	return e.runtimeRepo.Exists(ctx, asa.TargetTenantID, runtimeID)
}

func (e *engine) appendMissingScenarioLabelsForRuntimes(tenantID string, runtimesIDs []string, labels []model.Label) []model.Label {
	rtmWithScenario := make(map[string]struct{})
	for _, label := range labels {
		rtmWithScenario[label.ObjectID] = struct{}{}
	}

	for _, rtmID := range runtimesIDs {
		_, ok := rtmWithScenario[rtmID]
		if !ok {
			labels = append(labels, e.createNewEmptyScenarioLabel(tenantID, rtmID))
		}
	}

	return labels
}

func (e *engine) createNewEmptyScenarioLabel(tenantID string, rtmID string) model.Label {
	return model.Label{
		Tenant:     &tenantID,
		Key:        model.ScenariosKey,
		Value:      []string{},
		ObjectID:   rtmID,
		ObjectType: model.RuntimeLabelableObject,
	}
}

func (e *engine) upsertScenarios(ctx context.Context, tenantID string, labels []model.Label, newScenario string, mergeFn func(scenarios []string, diffScenario string) []string) error {
	for _, label := range labels {
		var scenariosString []string
		switch value := label.Value.(type) {
		case []string:
			{
				scenariosString = value
			}
		case []interface{}:
			{
				convertedScenarios, err := e.convertInterfaceArrayToStringArray(value)
				if err != nil {
					return errors.Wrap(err, "while converting array of interfaces to array of strings")
				}
				scenariosString = convertedScenarios
			}
		default:
			return errors.Errorf("scenarios value is invalid type: %t", label.Value)
		}

		newScenarios := mergeFn(scenariosString, newScenario)
		err := e.updateScenario(ctx, tenantID, label, newScenarios)
		if err != nil {
			return errors.Wrap(err, "while updating scenarios label")
		}
	}
	return nil
}

func (e *engine) updateScenario(ctx context.Context, tenantID string, label model.Label, scenarios []string) error {
	if len(scenarios) == 0 {
		return e.labelRepo.Delete(ctx, tenantID, model.RuntimeLabelableObject, label.ObjectID, model.ScenariosKey)
	}
	labelInput := model.LabelInput{
		Key:        label.Key,
		Value:      scenarios,
		ObjectID:   label.ObjectID,
		ObjectType: label.ObjectType,
	}
	return e.labelService.UpsertLabel(ctx, tenantID, &labelInput)
}

func (e *engine) convertInterfaceArrayToStringArray(scenarios []interface{}) ([]string, error) {
	scenariosString := make([]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		item, ok := scenario.(string)
		if !ok {
			return nil, apperrors.NewInternalError("scenario value is not a string")
		}
		scenariosString = append(scenariosString, item)
	}
	return scenariosString, nil
}

func (e *engine) uniqueScenarios(scenarios []string, newScenario string) []string {
	scenarios = append(scenarios, newScenario)
	return str.Unique(scenarios)
}

func (e *engine) removeScenario(scenarios []string, toRemove string) []string {
	var newScenarios []string
	for _, scenario := range scenarios {
		if scenario != toRemove {
			newScenarios = append(newScenarios, scenario)
		}
	}
	return newScenarios
}
