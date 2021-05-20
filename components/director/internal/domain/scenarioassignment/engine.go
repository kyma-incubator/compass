package scenarioassignment

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	app "github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore
type LabelRepository interface {
	GetRuntimeScenariosWhereLabelsMatchSelector(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]model.Label, error)
	GetRuntimesIDsByStringLabel(ctx context.Context, tenantID, selectorKey, selectorValue string) ([]string, error)
	GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error)
	ListForObjectTypeByScenario(ctx context.Context, tenant string, objectType model.LabelableObject, scenario string) ([]model.Label, error)
	ListScenariosForBundleInstanceAuthsByAppAndRuntimeIdAndCommonScenarios(ctx context.Context, tenant string, appId, runtimeId string, scenarios []string) ([]model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	Upsert(ctx context.Context, label *model.Label) error
}

//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore
type LabelUpsertService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
	UpsertScenarios(ctx context.Context, tenantID string, labels []model.Label, newScenario string, mergeFn func(scenarios []string, diffScenario string) []string) error
}

type engine struct {
	labelRepo              LabelRepository
	scenarioAssignmentRepo Repository
	labelService           LabelUpsertService
}

func NewEngine(labelService LabelUpsertService, labelRepo LabelRepository, scenarioAssignmentRepo Repository) *engine {
	return &engine{
		labelRepo:              labelRepo,
		scenarioAssignmentRepo: scenarioAssignmentRepo,
		labelService:           labelService,
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

	err = e.addNewScenarioToExistingBundleInstanceAuthFromMatchedApplication(ctx, runtimeLabels, in.Tenant, in.ScenarioName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("while adding new scenario to already existing bundle instance auth between app and runtime matched by scenario: "))
	}

	runtimeLabels = e.appendMissingScenarioLabelsForRuntimes(in.Tenant, runtimesIDs, runtimeLabels)
	return e.labelService.UpsertScenarios(ctx, in.Tenant, runtimeLabels, in.ScenarioName, label.UniqueScenarios)
}

func (e *engine) addNewScenarioToExistingBundleInstanceAuthFromMatchedApplication(ctx context.Context, runtimeLabels []model.Label, tenant, scenario string) error {
	appLabels, err := e.labelRepo.ListForObjectTypeByScenario(ctx, tenant, model.ApplicationLabelableObject, scenario)
	if err != nil {
		return errors.Wrap(err, "while fetching scenarios labels for application")
	}

	authLabels, err := e.getBundleInstanceAuthLabelsByCommonAppAndRuntimeScenarios(ctx, runtimeLabels, appLabels, tenant)
	if err != nil {
		return errors.Wrap(err, "while getting bundle instance auth labels by common application and runtime scenarios")
	}

	err = e.labelService.UpsertScenarios(ctx, tenant, authLabels, scenario, label.UniqueScenarios)
	if err != nil {
		return errors.Wrap(err, "while adding scenario to bundle instance auth labels")
	}

	return nil
}

func (e *engine) getBundleInstanceAuthLabelsByCommonAppAndRuntimeScenarios(ctx context.Context, runtimeLabels, appLabels []model.Label, tenant string) ([]model.Label, error) {
	authsResult := make([]model.Label, 0)

	for _, runtimeLabel := range runtimeLabels { //["id1","id2"]
		runtimeScenarios, err := e.parseLabelScenariosToSlice(runtimeLabel)
		if err != nil {
			return nil, errors.Wrap(err, "while parsing runtime label value")
		}

		for _, appLabel := range appLabels {
			appScenarios, err := e.parseLabelScenariosToSlice(appLabel)
			if err != nil {
				return nil, errors.Wrap(err, "while parsing application label value")
			}

			commonScenarios := app.GetScenariosToKeep(runtimeScenarios, appScenarios)
			bundleInstanceAuthLabels, err := e.labelRepo.ListScenariosForBundleInstanceAuthsByAppAndRuntimeIdAndCommonScenarios(ctx, tenant, appLabel.ObjectID, runtimeLabel.ObjectID, commonScenarios)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("while getting scenario labels for bundle instance auths for application id: '%s' and runtime id '%s'", appLabel.ObjectID, runtimeLabel.ObjectID))
			}

			authsResult = append(authsResult, bundleInstanceAuthLabels...)
		}
	}

	return authsResult, nil
}

func (e *engine) parseLabelScenariosToSlice(label model.Label) ([]string, error) {
	runtimeLabels, ok := label.Value.([]interface{})
	if !ok {
		return nil, errors.New("while converting label value to []interface{}")
	}

	convertedScenarios, err := str.InterfaceSliceToStringSlice(runtimeLabels)
	if err != nil {
		return nil, apperrors.NewInternalError("scenario value is not a string")
	}

	return convertedScenarios, nil
}

func (e *engine) RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	labels, err := e.labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, in.Tenant, in.Selector.Key, in.Selector.Value)
	if err != nil {
		return errors.Wrap(err, "while getting runtimes scenarios which match given selector")
	}
	for _, label := range labels { //TODO check once again
		runtimeID := label.ObjectID
		scenarios, err := e.parseLabelScenariosToSlice(label)
		if err != nil {
			return errors.Wrap(err, "while parsing runtime label value")
		}

		if e.isAnyBundleInstanceAuthForScenariosExist(ctx, scenarios, runtimeID) {
			return errors.New("Unable to delete label .....Bundle Instance Auths should be deleted first")
		}
	}
	return e.labelService.UpsertScenarios(ctx, in.Tenant, labels, in.ScenarioName, e.removeScenario)
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

func (e *engine) isAnyBundleInstanceAuthForScenariosExist(ctx context.Context, scenarios []string, runtimeId string) bool {
	for _, scenario := range scenarios {
		if e.isBundleInstanceAuthForScenarioExist(ctx, scenario, runtimeId) {
			return true
		}
	}
	return false
}

func (e *engine) isBundleInstanceAuthForScenarioExist(ctx context.Context, scenario, runtimeId string) bool {
	persist, _ := persistence.FromCtx(ctx)

	var count int
	query := "SELECT 1 FROM labels INNER JOIN bundle_instance_auths ON labels.bundle_instance_auth_id = bundle_instance_auths.id WHERE json_build_array($1::text)::jsonb <@ labels.value AND bundle_instance_auths.runtime_id=$2 AND bundle_instance_auths.status_condition='SUCCEEDED'"
	err := persist.Get(&count, query, scenario, runtimeId)
	if err != nil {
		return false
	}

	return count != 0
}
