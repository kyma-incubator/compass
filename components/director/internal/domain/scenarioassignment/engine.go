package scenarioassignment

import (
	"context"
	"fmt"

	labelpkg "github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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

//go:generate mockery --name=RuntimeRepository --output=automock --outpkg=automock --case=underscore
type RuntimeRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
}

//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore
type LabelUpsertService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

//go:generate mockery --name=BundleInstanceAuthService --output=automock --outpkg=automock --case=underscore
type BundleInstanceAuthService interface {
	AssociateBundleInstanceAuthForNewRuntimeScenarios(ctx context.Context, existingScenarios, inputScenarios []string, runtimeId string) error
	GetForRuntimeAndAnyMatchingScenarios(ctx context.Context, runtimeId string, scenarios []string) ([]*model.BundleInstanceAuth, error)
}

type engine struct {
	labelRepo                 LabelRepository
	runtimeRepo               RuntimeRepository
	scenarioAssignmentRepo    Repository
	labelService              LabelUpsertService
	bundleInstanceAuthService BundleInstanceAuthService
}

func NewEngine(labelService LabelUpsertService, labelRepo LabelRepository, scenarioAssignmentRepo Repository, bundleInstanceAuthService BundleInstanceAuthService, runtimeRepo RuntimeRepository) *engine {
	return &engine{
		labelRepo:                 labelRepo,
		scenarioAssignmentRepo:    scenarioAssignmentRepo,
		labelService:              labelService,
		bundleInstanceAuthService: bundleInstanceAuthService,
		runtimeRepo:               runtimeRepo,
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
	for _, label := range runtimeLabels {
		rmtLabelInput, err := labelpkg.MergeScenarios(label, []string{in.ScenarioName}, labelpkg.UniqueScenarios)
		if err != nil {
			return err
		}

		if rmtLabelInput == nil {
			return errors.New("unable to update scenarios for runtime label")
		}

		err = e.labelService.UpsertLabel(ctx, in.Tenant, rmtLabelInput)
		if err != nil {
			return errors.Wrapf(err, "while upserting scenarios label for runtime with id: %s", label.ObjectID)
		}
	}
	return nil
}

func (e *engine) addNewScenarioToExistingBundleInstanceAuthFromMatchedApplication(ctx context.Context, runtimeLabels []model.Label, scenario string) error {
	for _, runtimeLabel := range runtimeLabels {
		runtimeScenarios, err := labelpkg.GetScenariosAsStringSlice(runtimeLabel)
		if err != nil {
			return errors.Wrap(err, "while parsing runtime label value")
		}

		inputScenarios := []string{scenario}
		inputScenarios = append(inputScenarios, runtimeScenarios...)
		if err = e.bundleInstanceAuthService.AssociateBundleInstanceAuthForNewRuntimeScenarios(ctx, runtimeScenarios, inputScenarios, runtimeLabel.ObjectID); err != nil {
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

		if err = e.validateNoBundleInstanceAuthsExist(ctx, in.Tenant, runtimeID, in.ScenarioName); err != nil {
			return err
		}
	}

	for _, label := range labels {
		rmtLabelInput, err := labelpkg.MergeScenarios(label, []string{in.ScenarioName}, e.removeScenario)
		if err != nil {
			return err
		}

		if rmtLabelInput == nil {
			err := e.labelRepo.Delete(ctx, in.Tenant, label.ObjectType, label.ObjectID, model.ScenariosKey)
			if err != nil {
				return errors.Wrapf(err, "while deleting scenarios for runtime with id: %s", label.ObjectID)
			}
			continue
		}

		err = e.labelService.UpsertLabel(ctx, in.Tenant, rmtLabelInput)
		if err != nil {
			return errors.Wrapf(err, "while upserting scenarios label for runtime with id: %s", label.ObjectID)
		}
	}
	return nil
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

func (e *engine) removeScenario(scenarios, toRemove []string) []string {
	toRemoveSet := make(map[string]bool, 0)
	for _, elem := range toRemove {
		toRemoveSet[elem] = true
	}

	var newScenarios []string
	for _, scenario := range scenarios {
		if toRemoveSet[scenario] {
			continue
		}
		newScenarios = append(newScenarios, scenario)
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

func (e *engine) validateNoBundleInstanceAuthsExist(ctx context.Context, tenant, runtimeId, scenarioToRemove string) error {
	bndlAuths, err := e.bundleInstanceAuthService.GetForRuntimeAndAnyMatchingScenarios(ctx, runtimeId, []string{scenarioToRemove})
	if err != nil {
		return errors.Wrapf(err, "while getting existing bundle for old scenarios")
	}

	if len(bndlAuths) == 0 {
		return nil
	}

	runtime, err := e.runtimeRepo.GetByID(ctx, tenant, runtimeId)
	if err != nil {
		return errors.Wrapf(err, "while getting runtime")
	}

	authCtx := make([]*string, 0, len(bndlAuths))
	for _, auth := range bndlAuths {
		authCtx = append(authCtx, auth.Context)
	}
	return apperrors.NewScenarioUnassignWhenCredentialsExistsError(resource.Runtime, runtime.Name, authCtx)
}
