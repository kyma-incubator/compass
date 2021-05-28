package scenarioassignment

import (
	"context"
	"fmt"

	labelpkg "github.com/kyma-incubator/compass/components/director/internal/domain/label"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore
type LabelRepository interface {
	GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error)
	GetRuntimesIDsByStringLabel(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]string, error)
	GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
}

//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore
type LabelUpsertService interface {
	UpsertScenarios(ctx context.Context, tenantID string, labels []model.Label, newScenarios []string, mergeFn func(scenarios []string, diffScenario string) []string) error
}

type BundleInstanceAuthService interface {
	AssociateBundleInstanceAuthForNewRuntimeScenarios(ctx context.Context, existingScenarios, inputScenarios []string, runtimeId string) error
	IsAnyExistForRuntimeAndScenario(ctx context.Context, scenarios []string, runtimeId string) (bool, error)
}

type engine struct {
	labelRepo                 LabelRepository
	scenarioAssignmentRepo    Repository
	labelService              LabelUpsertService
	bundleInstanceAuthService BundleInstanceAuthService
}

func NewEngine(labelService LabelUpsertService, labelRepo LabelRepository, scenarioAssignmentRepo Repository, bundleInstanceAuthService BundleInstanceAuthService) *engine {
	return &engine{
		labelRepo:                 labelRepo,
		scenarioAssignmentRepo:    scenarioAssignmentRepo,
		labelService:              labelService,
		bundleInstanceAuthService: bundleInstanceAuthService,
	}
}

func (e *engine) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	runtimesIDs, err := e.labelRepo.GetRuntimesIDsByStringLabel(ctx, in.Tenant, in.Selector.Key, in.Selector.Value)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtimes id which match given selector:%+v", in)
	}

	if len(runtimesIDs) == 0 {
		return nil
	}

	runtimeLabels, err := e.labelRepo.GetScenarioLabelsForRuntimes(ctx, in.Tenant, runtimesIDs)
	if err != nil {
		return errors.Wrap(err, "while fetching scenarios labels for matched runtimes")
	}

	err = e.addNewScenarioToExistingBundleInstanceAuthFromMatchedApplication(ctx, runtimeLabels, in.ScenarioName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("while adding new scenario to already existing bundle instance auth between app and runtime matched by scenario: "))
	}

	runtimeLabels = e.appendMissingScenarioLabelsForRuntimes(in.Tenant, runtimesIDs, runtimeLabels)
	return e.labelService.UpsertScenarios(ctx, in.Tenant, runtimeLabels, []string{in.ScenarioName}, labelpkg.UniqueScenarios)
}

func (e *engine) addNewScenarioToExistingBundleInstanceAuthFromMatchedApplication(ctx context.Context, runtimeLabels []model.Label, scenario string) error {
	for _, runtimeLabel := range runtimeLabels {
		runtimeScenarios, err := labelpkg.GetScenariosAsStringSlice(runtimeLabel)
		if err != nil {
			return errors.Wrap(err, "while parsing runtime label value")
		}

		err = e.bundleInstanceAuthService.AssociateBundleInstanceAuthForNewRuntimeScenarios(ctx, runtimeScenarios, []string{scenario}, runtimeLabel.ObjectID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *engine) RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	labels, err := e.labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, in.Tenant, in.Selector.Key, in.Selector.Value)
	if err != nil {
		return errors.Wrap(err, "while getting runtimes scenarios which match given selector")
	}
	for _, label := range labels {
		runtimeID := label.ObjectID
		scenarios, err := labelpkg.GetScenariosAsStringSlice(label)
		if err != nil {
			return errors.Wrap(err, "while parsing runtime label value")
		}

		anyExist, err := e.bundleInstanceAuthService.IsAnyExistForRuntimeAndScenario(ctx, scenarios, runtimeID)
		if err != nil {
			return err
		}

		if anyExist {
			return errors.New("Unable to delete label .....Bundle Instance Auths should be deleted first")
		}
	}
	return e.labelService.UpsertScenarios(ctx, in.Tenant, labels, []string{in.ScenarioName}, e.removeScenario)
}

func (e *engine) RemoveAssignedScenarios(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	for _, asa := range in {
		err := e.RemoveAssignedScenario(ctx, *asa)
		if err != nil {
			return errors.Wrapf(err, "while deleting automatic scenario assigment: %s", asa.ScenarioName)
		}
	}
	return nil
}

func (e engine) GetScenariosForSelectorLabels(ctx context.Context, inputLabels map[string]string) ([]string, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	scenariosSet := make(map[string]struct{})

	for k, v := range inputLabels {
		selector := model.LabelSelector{
			Key:   k,
			Value: v,
		}

		scenarioAssignments, err := e.scenarioAssignmentRepo.ListForSelector(ctx, selector, tenantID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting Automatic Scenario Assignments for selector [key: %s, val: %s]", k, v)
		}

		for _, sa := range scenarioAssignments {
			scenariosSet[sa.ScenarioName] = struct{}{}
		}

	}

	scenarios := make([]string, 0)
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios, nil
}

func (e engine) MergeScenariosFromInputLabelsAndAssignments(ctx context.Context, inputLabels map[string]interface{}) ([]interface{}, error) {
	scenariosSet := make(map[string]struct{})

	possibleSelectorLabels := e.convertMapStringInterfaceToMapStringString(inputLabels)

	scenariosFromAssignments, err := e.GetScenariosForSelectorLabels(ctx, possibleSelectorLabels)
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

func (e engine) MergeScenarios(baseScenarios, scenariosToDelete, scenariosToAdd []interface{}) []interface{} {
	scenariosSet := map[interface{}]struct{}{}
	for _, scenario := range baseScenarios {
		scenariosSet[scenario] = struct{}{}
	}

	for _, scenario := range scenariosToDelete {
		delete(scenariosSet, scenario)
	}

	for _, scenario := range scenariosToAdd {
		scenariosSet[scenario] = struct{}{}
	}

	scenarios := make([]interface{}, 0)
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios
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
	return model.Label{Tenant: tenantID,
		Key:        model.ScenariosKey,
		Value:      []string{},
		ObjectID:   rtmID,
		ObjectType: model.RuntimeLabelableObject,
	}
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

func (e engine) convertMapStringInterfaceToMapStringString(inputLabels map[string]interface{}) map[string]string {
	convertedLabels := make(map[string]string)

	for k, v := range inputLabels {
		val, ok := v.(string)
		if ok {
			convertedLabels[k] = val
		}
	}

	return convertedLabels
}
