package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"

	"github.com/kyma-incubator/compass/components/director/internal/secure_http"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"
)

// IsNormalizedLabel missing godoc
const IsNormalizedLabel = "isNormalized"

// RuntimeRepository missing godoc
//go:generate mockery --name=RuntimeRepository --output=automock --outpkg=automock --case=underscore
type RuntimeRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
	ListByFiltersGlobal(context.Context, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	Create(ctx context.Context, item *model.Runtime) error
	Update(ctx context.Context, item *model.Runtime) error
	UpdateTenantID(ctx context.Context, runtimeID, newTenantID string) error
	Delete(ctx context.Context, tenant, id string) error
}

// LabelRepository missing godoc
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
	DeleteByKeyNegationPattern(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labelKeyPattern string) error
}

// LabelUpsertService missing godoc
//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore
type LabelUpsertService interface {
	UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

// ScenariosService missing godoc
//go:generate mockery --name=ScenariosService --output=automock --outpkg=automock --case=underscore
type ScenariosService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error
	AddDefaultScenarioIfEnabled(ctx context.Context, labels *map[string]interface{})
}

// ScenarioAssignmentEngine missing godoc
//go:generate mockery --name=ScenarioAssignmentEngine --output=automock --outpkg=automock --case=underscore
type ScenarioAssignmentEngine interface {
	GetScenariosForSelectorLabels(ctx context.Context, inputLabels map[string]string) ([]string, error)
	MergeScenariosFromInputLabelsAndAssignments(ctx context.Context, inputLabels map[string]interface{}) ([]interface{}, error)
	MergeScenarios(baseScenarios, scenariosToDelete, scenariosToAdd []interface{}) []interface{}
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo      RuntimeRepository
	labelRepo LabelRepository
	cfg       ServiceConfig

	labelUpsertService       LabelUpsertService
	uidService               UIDService
	scenariosService         ScenariosService
	scenarioAssignmentEngine ScenarioAssignmentEngine
	caller                   *secure_http.Caller

	protectedLabelPattern string
}

// NewService missing godoc
func NewService(repo RuntimeRepository,
	labelRepo LabelRepository,
	scenariosService ScenariosService,
	labelUpsertService LabelUpsertService,
	uidService UIDService,
	scenarioAssignmentEngine ScenarioAssignmentEngine,
	cfg ServiceConfig,
	protectedLabelPattern string) *service {
	caller, _ := secure_http.NewCaller(&graphql.OAuthCredentialData{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		URL:          cfg.TokenURL,
	})
	return &service{
		repo:                     repo,
		labelRepo:                labelRepo,
		scenariosService:         scenariosService,
		labelUpsertService:       labelUpsertService,
		uidService:               uidService,
		scenarioAssignmentEngine: scenarioAssignmentEngine,
		caller:                   caller,
		cfg:                      cfg,
		protectedLabelPattern:    protectedLabelPattern,
	}

}

// List missing godoc
func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.List(ctx, rtmTenant, filter, pageSize, cursor)
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.Runtime, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	runtime, err := s.repo.GetByID(ctx, rtmTenant, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Runtime with ID %s", id)
	}

	return runtime, nil
}

// GetByTokenIssuer missing godoc
func (s *service) GetByTokenIssuer(ctx context.Context, issuer string) (*model.Runtime, error) {
	const (
		consoleURLLabelKey = "runtime_consoleUrl"
		dexSubdomain       = "dex"
		consoleSubdomain   = "console"
	)
	consoleURL := strings.Replace(issuer, dexSubdomain, consoleSubdomain, 1)

	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(consoleURLLabelKey, fmt.Sprintf(`"%s"`, consoleURL)),
	}

	runtime, err := s.repo.GetByFiltersGlobal(ctx, filters)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting the Runtime by the console URL label (%s)", consoleURL)
	}

	return runtime, nil
}

// GetByFiltersGlobal missing godoc
func (s *service) GetByFiltersGlobal(ctx context.Context, filters []*labelfilter.LabelFilter) (*model.Runtime, error) {
	runtimes, err := s.repo.GetByFiltersGlobal(ctx, filters)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtimes by filters from repo")
	}
	return runtimes, nil
}

// ListByFiltersGlobal missing godoc
func (s *service) ListByFiltersGlobal(ctx context.Context, filters []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	runtimes, err := s.repo.ListByFiltersGlobal(ctx, filters)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtimes by filters from repo")
	}
	return runtimes, nil
}

// Exist missing godoc
func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "while loading tenant from context")
	}

	exist, err := s.repo.Exists(ctx, rtmTenant, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Runtime with ID %s", id)
	}

	return exist, nil
}

// Create missing godoc
func (s *service) Create(ctx context.Context, in model.RuntimeInput) (string, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}
	id := s.uidService.Generate()
	rtm := in.ToRuntime(id, rtmTenant, time.Now(), time.Now())

	err = s.repo.Create(ctx, rtm)
	if err != nil {
		return "", errors.Wrapf(err, "while creating Runtime")
	}

	err = s.scenariosService.EnsureScenariosLabelDefinitionExists(ctx, rtmTenant)
	if err != nil {
		return "", errors.Wrapf(err, "while ensuring Label Definition with key %s exists", model.ScenariosKey)
	}

	scenarios, err := s.scenarioAssignmentEngine.MergeScenariosFromInputLabelsAndAssignments(ctx, in.Labels)
	if err != nil {
		return "", errors.Wrap(err, "while merging scenarios from input and assignments")
	}

	if len(scenarios) > 0 {
		in.Labels[model.ScenariosKey] = scenarios
	} else {
		s.scenariosService.AddDefaultScenarioIfEnabled(ctx, &in.Labels)
	}

	if in.Labels == nil || in.Labels[IsNormalizedLabel] == nil {
		if in.Labels == nil {
			in.Labels = make(map[string]interface{}, 1)
		}
		in.Labels[IsNormalizedLabel] = "true"
	}

	log.C(ctx).Debugf("Removing protected labels. Labels before: %+v", in.Labels)
	in.Labels, err = unsafeExtractUnProtectedLabels(in.Labels, s.protectedLabelPattern)
	if err != nil {
		return "", err
	}
	log.C(ctx).Debugf("Successfully stripped protected labels. Resulting labels after operation are: %+v", in.Labels)

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, rtmTenant, model.RuntimeLabelableObject, id, in.Labels)
	if err != nil {
		return id, errors.Wrapf(err, "while creating multiple labels for Runtime")
	}

	return id, nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, id string, in model.RuntimeInput) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	rtm, err := s.repo.GetByID(ctx, rtmTenant, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Runtime with id %s", id)
	}

	rtm = in.ToRuntime(id, rtm.Tenant, rtm.CreationTimestamp, time.Now())

	err = s.repo.Update(ctx, rtm)
	if err != nil {
		return errors.Wrap(err, "while updating Runtime")
	}

	if in.Labels == nil || in.Labels[IsNormalizedLabel] == nil {
		if in.Labels == nil {
			in.Labels = make(map[string]interface{}, 1)
		}
		in.Labels[IsNormalizedLabel] = "true"
	}

	log.C(ctx).Debugf("Removing protected labels. Labels before: %+v", in.Labels)
	in.Labels, err = unsafeExtractUnProtectedLabels(in.Labels, s.protectedLabelPattern)
	if err != nil {
		return err
	}
	log.C(ctx).Debugf("Successfully stripped protected labels. Resulting labels after operation are: %+v", in.Labels)

	// NOTE: The db layer does not support OR currently so multiple label patterns can't be implemented easily
	err = s.labelRepo.DeleteByKeyNegationPattern(ctx, rtmTenant, model.RuntimeLabelableObject, id, s.protectedLabelPattern)
	if err != nil {
		return errors.Wrapf(err, "while deleting all labels for Runtime")
	}

	if in.Labels == nil {
		return nil
	}

	scenarios, err := s.scenarioAssignmentEngine.MergeScenariosFromInputLabelsAndAssignments(ctx, in.Labels)
	if err != nil {
		return errors.Wrap(err, "while merging scenarios from input and assignments")
	}

	if len(scenarios) > 0 {
		in.Labels[model.ScenariosKey] = scenarios
	}

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, rtmTenant, model.RuntimeLabelableObject, id, in.Labels)
	if err != nil {
		return errors.Wrapf(err, "while creating multiple labels for Runtime")
	}

	return nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	err = s.repo.Delete(ctx, rtmTenant, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Runtime")
	}

	// All labels are deleted (cascade delete)

	return nil
}

// SetLabel missing godoc
func (s *service) SetLabel(ctx context.Context, labelInput *model.LabelInput) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	err = s.ensureRuntimeExists(ctx, rtmTenant, labelInput.ObjectID)
	if err != nil {
		return err
	}

	currentRuntimeLabels, err := s.getCurrentLabelsForRuntime(ctx, rtmTenant, labelInput.ObjectID)
	if err != nil {
		return err
	}

	newRuntimeLabels := make(map[string]interface{})
	for k, v := range currentRuntimeLabels {
		newRuntimeLabels[k] = v
	}

	newRuntimeLabels[labelInput.Key] = labelInput.Value

	err = s.upsertScenariosLabelIfShould(ctx, labelInput.ObjectID, labelInput.Key, currentRuntimeLabels, newRuntimeLabels)
	if err != nil {
		return err
	}

	protected, err := isProtected(labelInput.Key, s.protectedLabelPattern)
	if err != nil {
		return err
	}
	if protected {
		return apperrors.NewInvalidDataError("could not set protected label key %s", labelInput.Key)
	}
	if labelInput.Key != model.ScenariosKey {
		err = s.labelUpsertService.UpsertLabel(ctx, rtmTenant, labelInput)
		if err != nil {
			return errors.Wrapf(err, "while creating label for Runtime")
		}
	}

	return nil
}

// GetLabel missing godoc
func (s *service) GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	rtmExists, err := s.repo.Exists(ctx, rtmTenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Runtime existence")
	}
	if !rtmExists {
		return nil, fmt.Errorf("Runtime with ID %s doesn't exist", runtimeID)
	}

	label, err := s.labelRepo.GetByKey(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID, key)
	if err != nil {
		return nil, errors.Wrap(err, "while getting label for Runtime")
	}

	return label, nil
}

// ListLabels missing godoc
func (s *service) ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	rtmExists, err := s.repo.Exists(ctx, rtmTenant, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Runtime existence")
	}

	if !rtmExists {
		return nil, fmt.Errorf("Runtime with ID %s doesn't exist", runtimeID)
	}

	labels, err := s.labelRepo.ListForObject(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting label for Runtime")
	}

	return extractUnProtectedLabels(labels, s.protectedLabelPattern)
}

// UpdateTenantID missing godoc
func (s *service) UpdateTenantID(ctx context.Context, runtimeID, newTenantID string) error {
	if err := s.repo.UpdateTenantID(ctx, runtimeID, newTenantID); err != nil {
		return errors.Wrapf(err, "while updating tenant_id for runtime with ID %s", runtimeID)
	}
	return nil
}

// DeleteLabel missing godoc
func (s *service) DeleteLabel(ctx context.Context, runtimeID string, key string) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	err = s.ensureRuntimeExists(ctx, rtmTenant, runtimeID)
	if err != nil {
		return err
	}

	currentRuntimeLabels, err := s.getCurrentLabelsForRuntime(ctx, rtmTenant, runtimeID)
	if err != nil {
		return err
	}

	newRuntimeLabels := make(map[string]interface{})
	for k, v := range currentRuntimeLabels {
		newRuntimeLabels[k] = v
	}

	delete(newRuntimeLabels, key)

	err = s.upsertScenariosLabelIfShould(ctx, runtimeID, key, currentRuntimeLabels, newRuntimeLabels)
	if err != nil {
		return err
	}

	protected, err := isProtected(key, s.protectedLabelPattern)
	if err != nil {
		return err
	}
	if protected {
		return apperrors.NewInvalidDataError("could not delete protected label key %s", key)
	}
	if key != model.ScenariosKey {
		err = s.labelRepo.Delete(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID, key)
		if err != nil {
			return errors.Wrapf(err, "while deleting Runtime label")
		}
	}

	return nil
}

//TODO: Extract configrable values
//PrepareRuntimeForSelfRegistration executes the prerequisite calls for self-registration in case the runtime is being self-registered
func (s *service) PrepareRuntimeForSelfRegistration(ctx context.Context, in *graphql.RuntimeInput) error {
	//consumerInfo, err := consumer.LoadFromContext(ctx)
	//if err != nil {
	//	return errors.Wrapf(err, "while loading consumer")
	//}
	if _, exists := in.Labels["xsappname"]; exists /*&& consumerInfo.Flow.IsCertFlow()*/ { //this means that the runtime is being self-registered

		var body = `{
		   "xsappname": "lasmagi-jasmagi",
		 	"authorities": [
				"$XSAPPNAME.application:write",
				"$XSAPPNAME.application:read",
				"$XSAPPNAME.label_definition:write",
				"$XSAPPNAME.label_definition:read",
				"$XSAPPNAME.automatic_scenario_assignment:write",
				"$XSAPPNAME.automatic_scenario_assignment:read",
				"$XSAPPNAME.application_template:read"
		],
		   "oauth2-configuration": {
		         "credential-types": ["binding-secret"]
		   }
		}`
		url := "https://cmp.authentication.sap.hana.ondemand.com" + "/sap/rest/broker/clones?serviceinstanceid=" + "lasmagi-jasmagi" + "&subaccountid=92936559-9275-4da0-a7c8-fe4c6b8aa659"
		request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
		if err != nil {
			return errors.Wrapf(err, "while preparing request for self regisstered runtime")
		}
		request.Header.Set("Content-Type", "application/json")

		response, err := s.caller.Call(request)
		if err != nil {
			return errors.Wrapf(err, "while executing preparation of self registered runtime")
		}
		defer httputils.Close(ctx, response.Body)

		respBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return errors.Wrapf(err, "while reading response body")
		}

		if response.StatusCode != http.StatusCreated {
			return errors.New(fmt.Sprintf("recieved unexpected status %d while preparing self-registered runtime: %s", response.StatusCode, string(respBytes)))
		}

		selfRegLabelVal := gjson.GetBytes(respBytes, "appid")
		in.Labels["clone-name"] = selfRegLabelVal.Str

		log.C(ctx).Infof("successfully executed prep for self-registered runtime with label value %s", selfRegLabelVal.Str)
	}
	return nil
}

//TODO: Extract configrable values
//CleanupSelfRegisteredRuntime executes cleanup calls for self-registered runtimes
func (s *service) CleanupSelfRegisteredRuntime(ctx context.Context, selfRegisterLabelValue string) error {
	if selfRegisterLabelValue != "" {
		url := "https://cmp.authentication.sap.hana.ondemand.com" + "/sap/rest/broker/clones/" + selfRegisterLabelValue
		request, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			return errors.Wrapf(err, "while preparing request for cleanup of self-registered runtime with label value %s", selfRegisterLabelValue)
		}
		request.Header.Set("Content-Type", "application/json")
		resp, err := s.caller.Call(request)
		if err != nil {
			return errors.Wrapf(err, "while executing cleanup of self-registered runtime with label value %s", selfRegisterLabelValue)
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(fmt.Sprintf("recieved unexpected status code %d while cleaning up self-registered runtime with label value %s", resp.StatusCode, selfRegisterLabelValue))
		}
		log.C(ctx).Infof("Successfully executed clean-up self-registered runtime with label value %s", selfRegisterLabelValue)
	}
	return nil
}

func (s *service) ensureRuntimeExists(ctx context.Context, tnt string, runtimeID string) error {
	rtmExists, err := s.repo.Exists(ctx, tnt, runtimeID)
	if err != nil {
		return errors.Wrap(err, "while checking Runtime existence")
	}
	if !rtmExists {
		return fmt.Errorf("Runtime with ID %s doesn't exist", runtimeID)
	}

	return nil
}

func (s *service) upsertScenariosLabelIfShould(ctx context.Context, runtimeID string, modifiedLabelKey string, currentRuntimeLabels, newRuntimeLabels map[string]interface{}) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	finalScenarios := make([]interface{}, 0)

	if modifiedLabelKey == model.ScenariosKey {
		scenarios, err := s.scenarioAssignmentEngine.MergeScenariosFromInputLabelsAndAssignments(ctx, newRuntimeLabels)
		if err != nil {
			return errors.Wrap(err, "while merging scenarios from input and assignments")
		}

		finalScenarios = append(finalScenarios, scenarios...)
	} else {
		oldScenariosLabel, err := getScenariosLabel(currentRuntimeLabels)
		if err != nil {
			return err
		}

		previousScenariosFromAssignments, err := s.getScenariosFromAssignments(ctx, currentRuntimeLabels)
		if err != nil {
			return errors.Wrap(err, "while getting old scenarios label and scenarios from assignments")
		}

		newScenariosFromAssignments, err := s.getScenariosFromAssignments(ctx, newRuntimeLabels)
		if err != nil {
			return errors.Wrap(err, "while getting new scenarios from assignments")
		}

		finalScenarios = s.scenarioAssignmentEngine.MergeScenarios(oldScenariosLabel, previousScenariosFromAssignments, newScenariosFromAssignments)
	}

	// TODO compare finalScenarios and oldScenariosLabel to determine when to delete scenarios label
	if len(finalScenarios) == 0 {
		err := s.labelRepo.Delete(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey)
		if err != nil {
			return errors.Wrapf(err, "while deleting scenarios label from runtime with id [%s]", runtimeID)
		}
		return nil
	}

	scenariosLabelInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      finalScenarios,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	err = s.labelUpsertService.UpsertLabel(ctx, rtmTenant, scenariosLabelInput)
	if err != nil {
		return errors.Wrapf(err, "while creating scenarios label for Runtime with id [%s]", runtimeID)
	}

	return nil
}

func (s *service) getCurrentLabelsForRuntime(ctx context.Context, tenantID, runtimeID string) (map[string]interface{}, error) {
	labels, err := s.labelRepo.ListForObject(ctx, tenantID, model.RuntimeLabelableObject, runtimeID)
	if err != nil {
		return nil, err
	}

	currentLabels := make(map[string]interface{})
	for _, v := range labels {
		currentLabels[v.Key] = v.Value
	}
	return currentLabels, nil
}

func extractUnProtectedLabels(labels map[string]*model.Label, protectedLabelsKeyPattern string) (map[string]*model.Label, error) {
	result := make(map[string]*model.Label)
	for labelKey, label := range labels {
		protected, err := isProtected(labelKey, protectedLabelsKeyPattern)
		if err != nil {
			return nil, err
		}
		if !protected {
			result[labelKey] = label
		}
	}
	return result, nil
}

func unsafeExtractUnProtectedLabels(labels map[string]interface{}, protectedLabelsKeyPattern string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for labelKey, label := range labels {
		protected, err := isProtected(labelKey, protectedLabelsKeyPattern)
		if err != nil {
			return nil, err
		}
		if !protected {
			result[labelKey] = label
		}
	}
	return result, nil
}

func isProtected(labelKey string, labelKeyPattern string) (bool, error) {
	matched, err := regexp.MatchString(labelKeyPattern, labelKey)
	if err != nil {
		return false, err
	}
	return matched, nil
}

func getScenariosLabel(currentRuntimeLabels map[string]interface{}) ([]interface{}, error) {
	oldScenariosLabel, ok := currentRuntimeLabels[model.ScenariosKey]

	var oldScenariosLabelInterfaceSlice []interface{}
	if ok {
		oldScenariosLabelInterfaceSlice, ok = oldScenariosLabel.([]interface{})
		if !ok {
			return nil, apperrors.NewInternalError("value for scenarios label must be []interface{}")
		}
	}
	return oldScenariosLabelInterfaceSlice, nil
}

func (s *service) getScenariosFromAssignments(ctx context.Context, currentRuntimeLabels map[string]interface{}) ([]interface{}, error) {
	selectors := s.convertMapStringInterfaceToMapStringString(currentRuntimeLabels)

	ScenariosFromAssignments, err := s.scenarioAssignmentEngine.GetScenariosForSelectorLabels(ctx, selectors)
	if err != nil {
		return nil, errors.Wrap(err, "while getting scenarios for selector labels")
	}

	newScenariosInterfaceSlice := s.convertStringSliceToInterfaceSlice(ScenariosFromAssignments)

	return newScenariosInterfaceSlice, nil
}

func (s *service) convertMapStringInterfaceToMapStringString(in map[string]interface{}) map[string]string {
	out := make(map[string]string)

	for k, v := range in {
		val, ok := v.(string)
		if ok {
			out[k] = val
		}
	}

	return out
}

func (s *service) convertStringSliceToInterfaceSlice(in []string) []interface{} {
	out := make([]interface{}, 0)
	for _, v := range in {
		out = append(out, v)
	}

	return out
}
