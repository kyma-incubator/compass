package scenarioassignment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	GetAppIdsForScenario(ctx context.Context, scenario string) ([]string, error)
	GetScenarioNamesForApplication(ctx context.Context, applicationID string) ([]string, error)
}

type engine struct {
	labelRepo              LabelRepository
	scenarioAssignmentRepo Repository
	labelService           LabelUpsertService
	appSvc                 ApplicationService
}

func NewEngine(labelService LabelUpsertService, labelRepo LabelRepository, scenarioAssignmentRepo Repository, appService ApplicationService) *engine {
	return &engine{
		labelRepo:              labelRepo,
		scenarioAssignmentRepo: scenarioAssignmentRepo,
		labelService:           labelService,
		appSvc:                 appService,
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
	labels, err := e.labelRepo.GetScenarioLabelsForRuntimes(ctx, in.Tenant, runtimesIDs)
	if err != nil {
		return errors.Wrap(err, "while fetching scenarios labels for matched runtimes")
	}

	applications, err := e.appSvc.GetAppIdsForScenario(ctx, in.ScenarioName)
	if err != nil {
		return err
	}

	for _, label := range labels { //["id1","id2"]
		var runtimeScenarios []string
		labelAsString, ok := label.Value.(string)
		if !ok {
			return errors.New("while converting label value to string array")
		}
		err := json.Unmarshal([]byte(labelAsString), &runtimeScenarios)
		if err != nil {
			return err
		}
		for _, application := range applications {
			//check whether <current_app> is in formation with <current_runtime>
			// if yes, get bundle instance auths for this formation and this runtime
			// add scenario(from ASA) to each bundle instance auth label
		}
	}

	labels = e.appendMissingScenarioLabelsForRuntimes(in.Tenant, runtimesIDs, labels)
	return e.upsertScenarios(ctx, in.Tenant, labels, in.ScenarioName, e.uniqueScenarios)
}

func (e *engine) RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	labels, err := e.labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, in.Tenant, in.Selector.Key, in.Selector.Value)
	if err != nil {
		return errors.Wrap(err, "while getting runtimes scenarios which match given selector")
	}
	for _, label := range labels { //TODO check once again
		runtimeID := label.ObjectID
		var scenarios []string
		err := json.Unmarshal([]byte(label.Value.(string)), &scenarios)
		if err != nil {
			// todo handle
		}
		if e.isAnyBundleInstanceAuthForScenariosExist(ctx, scenarios, runtimeID) {
			return errors.New("Unable to delete label .....Bundle Instance Auths should be deleted first")
		}
	}
	return e.upsertScenarios(ctx, in.Tenant, labels, in.ScenarioName, e.removeScenario)
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
	} else {
		labelInput := model.LabelInput{
			Key:        label.Key,
			Value:      scenarios,
			ObjectID:   label.ObjectID,
			ObjectType: label.ObjectType,
		}
		return e.labelService.UpsertLabel(ctx, tenantID, &labelInput)
	}
}

func (e *engine) convertInterfaceArrayToStringArray(scenarios []interface{}) ([]string, error) {
	var scenariosString []string
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
