package bundleinstanceauth

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore
type Repository interface {
	Create(ctx context.Context, item *model.BundleInstanceAuth) error
	GetByID(ctx context.Context, tenantID string, id string) (*model.BundleInstanceAuth, error)
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.BundleInstanceAuth, error)
	ListByBundleID(ctx context.Context, tenantID string, bundleID string) ([]*model.BundleInstanceAuth, error)
	ListByRuntimeID(ctx context.Context, tenantID string, runtimeID string) ([]*model.BundleInstanceAuth, error)
	Update(ctx context.Context, item *model.BundleInstanceAuth) error
	Delete(ctx context.Context, tenantID string, id string) error
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type ScenarioService interface {
	GetScenarioNamesForApplication(ctx context.Context, applicationID string) ([]string, error)
	GetScenarioNamesForRuntime(ctx context.Context, runtimeID string) ([]string, error)
}

//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore
type LabelService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
	UpsertScenarios(ctx context.Context, tenantID string, labels []model.Label, newScenarios []string, mergeFn func(scenarios []string, diffScenario string) []string) error
}

type LabelRepository interface {
	GetBundleInstanceAuthsScenarioLabels(ctx context.Context, appId, runtimeId string) ([]model.Label, error)
}

type RuntimeIdsForScenariosSupplier func(ctx context.Context, tenant string, scenarios []string) ([]string, error)
type AppIdsForScenariosSupplier func(ctx context.Context, tenant string, scenarios []string) ([]string, error)

type service struct {
	repo         Repository
	uidService   UIDService
	timestampGen timestamp.Generator
	bundleSvc    BundleService
	scenarioSvc  ScenarioService
	labelService LabelService
	labelRepo    LabelRepository
}

func NewService(repo Repository, uidService UIDService, bundleService BundleService, scenarioSvc ScenarioService, labelSvc LabelService, labelRepo LabelRepository) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator(),
		bundleSvc:    bundleService,
		scenarioSvc:  scenarioSvc,
		labelService: labelSvc,
		labelRepo:    labelRepo,
	}
}

func (s *service) Create(ctx context.Context, bundleID string, in model.BundleInstanceAuthRequestInput, defaultAuth *model.Auth, requestInputSchema *string) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	log.C(ctx).Debugf("Validating BundleInstanceAuth request input for Bundle with id %s", bundleID)
	err = s.validateInputParamsAgainstSchema(in.InputParams, requestInputSchema)
	if err != nil {
		return "", errors.Wrapf(err, "while validating BundleInstanceAuth request input for Bundle with id %s", bundleID)
	}

	con, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	var runtimeID *string
	if con.ConsumerType == consumer.Runtime {
		runtimeID = &con.ConsumerID
	}

	id := s.uidService.Generate()
	log.C(ctx).Debugf("ID %s generated for BundleInstanceAuth for Bundle with id %s", id, bundleID)
	bndlInstAuth := in.ToBundleInstanceAuth(id, bundleID, tnt, defaultAuth, nil, runtimeID, nil)

	err = s.setCreationStatusFromAuth(ctx, &bndlInstAuth, defaultAuth)
	if err != nil {
		return "", errors.Wrapf(err, "while setting creation status for BundleInstanceAuth with id %s", id)
	}

	err = s.repo.Create(ctx, &bndlInstAuth)
	if err != nil {
		return "", errors.Wrapf(err, "while creating BundleInstanceAuth with id %s for Bundle with id %s", id, bundleID)
	}

	applicationID, err := s.bundleSvc.GetByApplicationID(ctx, tnt, bundleID)
	if err != nil {
		return "", err //todo
	}
	sceanriosForApp, err := s.scenarioSvc.GetScenarioNamesForApplication(ctx, applicationID)
	if err != nil {
		return "", err //todo
	}
	sceanriosForRuntime, err := s.scenarioSvc.GetScenarioNamesForRuntime(ctx, *runtimeID)
	if err != nil {
		return "", err
	}

	appScenarios := make(map[string]bool)
	var commonScenarios []string
	for _, elem := range sceanriosForApp {
		appScenarios[elem] = true
	}
	for _, elem := range sceanriosForRuntime {
		if appScenarios[elem] {
			commonScenarios = append(commonScenarios, elem)
		}
	}

	authLabel := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      commonScenarios,
		ObjectID:   id,
		ObjectType: model.BundleInstanceAuthObject,
	}

	if err = s.labelService.UpsertLabel(ctx, tnt, authLabel); err != nil {
		return "", errors.Wrap(err, "while creating bundle instance auth scenario label")
	}

	return id, nil
}

func (s *service) AssociateBundleInstanceAuthForNewApplicationScenarios(ctx context.Context, existingScenarios, inputScenarios []string, appId string, runtIdsForScenario RuntimeIdsForScenariosSupplier) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	scToRemove := getScenariosToRemove(existingScenarios, inputScenarios)

	authExist, err := s.IsAnyExistForAppAndScenario(ctx, scToRemove, appId)
	if err != nil {
		return err
	}

	if authExist {
		return errors.New("Unable to delete label .....Bundle Instance Auths should be deleted first")
	}

	scToAdd := getScenariosToAdd(existingScenarios, inputScenarios)
	scenariosToKeep := GetScenariosToKeep(existingScenarios, inputScenarios)
	commonRuntimes := getCommonRuntimes(ctx, appTenant, scenariosToKeep, scToAdd, runtIdsForScenario) // runtimeID -> [scenario1,scenario2..]

	for runtimeID, scenarios := range commonRuntimes {
		bundleInstanceAuthsLabels, err := s.labelRepo.GetBundleInstanceAuthsScenarioLabels(ctx, appId, runtimeID)
		if err != nil {
			return err
		}

		assocAuthLabels := make([]model.Label, 0)
		for _, currentLabel := range bundleInstanceAuthsLabels {
			if currentLabel.Key == model.ScenariosKey {
				//TODO: apply in SQL query
				assocAuthLabels = append(assocAuthLabels, currentLabel)
			}
		}

		if err := s.labelService.UpsertScenarios(ctx, appTenant, assocAuthLabels, scenarios, label.UniqueScenarios); err != nil {
			return errors.Wrap(err, fmt.Sprintf("while associating scenario: '%s' to all bundle_instance_auths for appId: %s and runtimeId: %s", scenarios, appId, runtimeID))
		}
	}

	return nil
}

func (s *service) AssociateBundleInstanceAuthForNewRuntimeScenarios(ctx context.Context, existingScenarios, inputScenarios []string, runtimeId string, appIdsForScenario AppIdsForScenariosSupplier) error {
	//TODO: unify with AssociateBundleInstanceAuthForNewApplicationScenarios
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	scToRemove := getScenariosToRemove(existingScenarios, inputScenarios)

	authExist, err := s.IsAnyExistForRuntimeAndScenario(ctx, scToRemove, runtimeId)
	if err != nil {
		return err
	}

	if authExist {
		return errors.New("Unable to delete label .....Bundle Instance Auths should be deleted first")
	}

	scToAdd := getScenariosToAdd(existingScenarios, inputScenarios)
	scenariosToKeep := GetScenariosToKeep(existingScenarios, inputScenarios)
	commonApplications := s.getCommonApplications(ctx, appTenant, scenariosToKeep, scToAdd, appIdsForScenario)

	for appId, scenarios := range commonApplications {
		bundleInstanceAuthsLabels, err := s.labelRepo.GetBundleInstanceAuthsScenarioLabels(ctx, appId, runtimeId)
		if err != nil {
			return err
		}

		assocAuthLabels := make([]model.Label, 0)
		for _, currentLabel := range bundleInstanceAuthsLabels {
			if currentLabel.Key == model.ScenariosKey {
				//TODO: apply in SQL query
				assocAuthLabels = append(assocAuthLabels, currentLabel)
			}
		}

		if err := s.labelService.UpsertScenarios(ctx, appTenant, assocAuthLabels, scenarios, label.UniqueScenarios); err != nil {
			return errors.Wrap(err, fmt.Sprintf("while associating scenario: '%s' to all bundle_instance_auths for appId: %s and runtimeId: %s", scenarios, appId, runtimeId))
		}
	}

	return nil
}

func (s *service) associateBundleInstanceAuthForNewObject(ctx context.Context, existingScenarios, inputScenarios []string, runtimeId string, commonRelatedObject func(ctx context.Context, tenant string, scenariosToKeep, scenariosToAdd []string) map[string]string) error {
	//TODO: common implementation
	return nil
}

func (s *service) Get(ctx context.Context, id string) (*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting BundleInstanceAuth with id %s", id)
	}

	return instanceAuth, nil
}

func (s *service) GetForBundle(ctx context.Context, id string, bundleID string) (*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bndl, err := s.repo.GetForBundle(ctx, tnt, id, bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle Instance Auth with ID: [%s]", id)
	}

	return bndl, nil
}

func (s *service) List(ctx context.Context, bundleID string) ([]*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bndlInstanceAuths, err := s.repo.ListByBundleID(ctx, tnt, bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Bundle Instance Auths")
	}

	return bndlInstanceAuths, nil
}

func (s *service) ListByRuntimeID(ctx context.Context, runtimeID string) ([]*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bndlInstanceAuths, err := s.repo.ListByRuntimeID(ctx, tnt, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Bundle Instance Auths")
	}

	return bndlInstanceAuths, nil
}

func (s *service) Update(ctx context.Context, instanceAuth *model.BundleInstanceAuth) error {
	err := s.repo.Update(ctx, instanceAuth)
	if err != nil {
		return errors.Wrap(err, "while updating Bundle Instance Auths")
	}

	return nil
}

func (s *service) SetAuth(ctx context.Context, id string, in model.BundleInstanceAuthSetInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting BundleInstanceAuth with id %s", id)
	}
	if instanceAuth == nil {
		return errors.Errorf("BundleInstanceAuth with id %s not found", id)
	}

	if instanceAuth.Status == nil || instanceAuth.Status.Condition != model.BundleInstanceAuthStatusConditionPending {
		return apperrors.NewInvalidOperationError("auth can be set only on BundleInstanceAuths in PENDING state")
	}

	err = s.setUpdateAuthAndStatus(ctx, instanceAuth, in)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, instanceAuth)
	if err != nil {
		return errors.Wrapf(err, "while updating BundleInstanceAuth with ID %s", id)
	}
	return nil
}

func (s *service) RequestDeletion(ctx context.Context, instanceAuth *model.BundleInstanceAuth, defaultBundleInstanceAuth *model.Auth) (bool, error) {
	if instanceAuth == nil {
		return false, apperrors.NewInternalError("BundleInstanceAuth is required to request its deletion")
	}

	if defaultBundleInstanceAuth == nil {
		log.C(ctx).Debugf("Default credentials for BundleInstanceAuth with id %s are not provided.", instanceAuth.ID)

		err := instanceAuth.SetDefaultStatus(model.BundleInstanceAuthStatusConditionUnused, s.timestampGen())
		if err != nil {
			return false, errors.Wrapf(err, "while setting status of BundleInstanceAuth with id %s to '%s'", instanceAuth.ID, model.BundleInstanceAuthStatusConditionUnused)
		}
		log.C(ctx).Infof("Status for BundleInstanceAuth with id %s set to '%s'. Credentials are ready for being deleted by Application or Integration System.", instanceAuth.ID, model.BundleInstanceAuthStatusConditionUnused)

		err = s.repo.Update(ctx, instanceAuth)
		if err != nil {
			return false, errors.Wrapf(err, "while updating BundleInstanceAuth with id %s", instanceAuth.ID)
		}

		return false, nil
	}

	log.C(ctx).Debugf("Default credentials for BundleInstanceAuth with id %s are provided.", instanceAuth.ID)
	err := s.Delete(ctx, instanceAuth.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	log.C(ctx).Debugf("Deleting BundleInstanceAuth entity with id %s in db", id)
	err = s.repo.Delete(ctx, tnt, id)

	return errors.Wrapf(err, "while deleting BundleInstanceAuth with id %s", id)
}

func (s *service) setUpdateAuthAndStatus(ctx context.Context, instanceAuth *model.BundleInstanceAuth, in model.BundleInstanceAuthSetInput) error {
	if instanceAuth == nil {
		return nil
	}

	ts := s.timestampGen()

	instanceAuth.Auth = in.Auth.ToAuth()
	instanceAuth.Status = in.Status.ToBundleInstanceAuthStatus(ts)

	// Input validation ensures that status can be nil only when auth was provided, so we can assume SUCCEEDED status
	if instanceAuth.Status == nil {
		log.C(ctx).Infof("Updating the status of BundleInstanceAuth with id %s to '%s'", instanceAuth.ID, model.BundleInstanceAuthStatusConditionSucceeded)
		err := instanceAuth.SetDefaultStatus(model.BundleInstanceAuthStatusConditionSucceeded, ts)
		if err != nil {
			return errors.Wrapf(err, "while setting status '%s' to BundleInstanceAuth with id %s", model.BundleInstanceAuthStatusConditionSucceeded, instanceAuth.ID)
		}
	}

	return nil
}

func (s *service) setCreationStatusFromAuth(ctx context.Context, instanceAuth *model.BundleInstanceAuth, defaultAuth *model.Auth) error {
	if instanceAuth == nil {
		return nil
	}

	var condition model.BundleInstanceAuthStatusCondition
	if defaultAuth != nil {
		log.C(ctx).Infof("Default credentials for BundleInstanceAuth with id %s from Bundle with id %s are provided. Setting creation status to '%s'", instanceAuth.ID, instanceAuth.BundleID, model.BundleInstanceAuthStatusConditionSucceeded)
		condition = model.BundleInstanceAuthStatusConditionSucceeded
	} else {
		log.C(ctx).Infof("Default credentials for BundleInstanceAuth with id %s from Bundle with id %s are not provided. Setting creation status to '%s'", instanceAuth.ID, instanceAuth.BundleID, model.BundleInstanceAuthStatusConditionPending)
		condition = model.BundleInstanceAuthStatusConditionPending
	}

	err := instanceAuth.SetDefaultStatus(condition, s.timestampGen())
	return errors.Wrapf(err, "while setting default status for BundleInstanceAuth with id %s", instanceAuth.ID)
}

func (s *service) validateInputParamsAgainstSchema(inputParams *string, schema *string) error {
	if schema == nil {
		return nil
	}
	if inputParams == nil {
		return apperrors.NewInvalidDataError("json schema for input parameters was defined for the bundle but no input parameters were provided")
	}

	validator, err := jsonschema.NewValidatorFromStringSchema(*schema)
	if err != nil {
		return errors.Wrapf(err, "while creating JSON Schema validator for schema %+s", *schema)
	}

	result, err := validator.ValidateString(*inputParams)
	if err != nil {
		return errors.Wrapf(err, "while validating value %s against JSON Schema: %s", *inputParams, *schema)
	}
	if !result.Valid {
		return errors.Wrapf(result.Error, "while validating value %s against JSON Schema: %s", *inputParams, *schema)
	}

	return nil
}

func (s *service) IsAnyExistForAppAndScenario(ctx context.Context, scenarios []string, appId string) (bool, error) {
	for _, scenario := range scenarios {
		exist, err := s.existForAppAndScenario(ctx, scenario, appId)

		if err != nil {
			return false, err
		}

		if exist {
			return true, nil
		}
	}
	return false, nil
}

func (s *service) IsAnyExistForRuntimeAndScenario(ctx context.Context, scenarios []string, runtimeId string) (bool, error) {
	for _, scenario := range scenarios {
		exist, err := s.existForRuntimeAndScenario(ctx, scenario, runtimeId)

		if err != nil {
			return false, err
		}

		if exist {
			return true, nil
		}
	}

	return false, nil
}

func (s *service) existForAppAndScenario(ctx context.Context, scenario, appId string) (bool, error) {
	persist, _ := persistence.FromCtx(ctx)

	var count int
	query := "SELECT 1 FROM bundle_instance_auths_with_labels WHERE json_build_array($1::text)::jsonb <@ bundle_instance_auths_with_labels.value AND app_id=$2 AND status_condition='SUCCEEDED'"
	err := persist.Get(&count, query, scenario, appId)
	if err != nil {
		mappedErr := persistence.MapSQLError(ctx, err, resource.Label, resource.List, "while fetching list of objects from '%s' table", "bundle_instance_auths_with_labels")
		if apperrors.IsNotFoundError(mappedErr) {
			return false, nil
		}
		return false, mappedErr
	}

	return count != 0, nil
}

func (s *service) existForRuntimeAndScenario(ctx context.Context, scenario, runtimeId string) (bool, error) {
	persist, _ := persistence.FromCtx(ctx)

	var count int
	query := "SELECT 1 FROM labels INNER JOIN bundle_instance_auths ON labels.bundle_instance_auth_id = bundle_instance_auths.id WHERE json_build_array($1::text)::jsonb <@ labels.value AND bundle_instance_auths.runtime_id=$2 AND bundle_instance_auths.status_condition='SUCCEEDED'"
	err := persist.Get(&count, query, scenario, runtimeId)
	if err != nil {
		mappedErr := persistence.MapSQLError(ctx, err, resource.Label, resource.List, "while fetching list of objects from '%s' table", "bundle_instance_auths_with_labels")
		if apperrors.IsNotFoundError(mappedErr) {
			return false, nil
		}
		return false, mappedErr
	}

	return count != 0, err
}

func getCommonRuntimes(ctx context.Context, tenant string, scenariosToKeep, scenariosToAdd []string, supplier RuntimeIdsForScenariosSupplier) map[string][]string {
	runtimeNamesForScenariosToKeep, err := supplier(ctx, tenant, scenariosToKeep)
	if err != nil {
		// TODO HANDLE ERROR
	}

	commonRuntimesScenarios := make(map[string][]string)
	for _, scenario := range scenariosToAdd {
		runtimeIDs, err := supplier(ctx, tenant, []string{scenario})
		if err != nil {
			// todo handle
		}
		for _, runtime := range runtimeIDs {
			if contains(runtimeNamesForScenariosToKeep, runtime) {
				commonRuntimesScenarios[runtime] = append(commonRuntimesScenarios[runtime], scenario)
			}
		}
	}
	return commonRuntimesScenarios
}

func (s *service) getCommonApplications(ctx context.Context, tenant string, scenariosToKeep, scenariosToAdd []string, supplier AppIdsForScenariosSupplier) map[string][]string {
	//TODO: Skip db queryies if scenariosToAdd is empty
	appIdsForScenariosToKeep, err := supplier(ctx, tenant, scenariosToKeep)
	if err != nil {
		// TODO HANDLE ERROR
	}

	commonApplicationScenarios := make(map[string][]string)
	for _, scenario := range scenariosToAdd {
		appIds, err := supplier(ctx, tenant, []string{scenario})
		if err != nil {
			// todo handle
		}
		for _, appId := range appIds {
			if contains(appIdsForScenariosToKeep, appId) {
				commonApplicationScenarios[appId] = append(commonApplicationScenarios[appId], scenario)
			}
		}
	}
	return commonApplicationScenarios
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getScenariosToRemove(existing, new []string) []string {
	newScenariosMap := make(map[string]bool, 0)
	for _, scenario := range new {
		newScenariosMap[scenario] = true
	}

	result := make([]string, 0)
	for _, scenario := range existing {
		if _, ok := newScenariosMap[scenario]; !ok {
			result = append(result, scenario)
		}
	}
	return result
}

func getScenariosToAdd(existing, new []string) []string {
	existingScenarioMap := make(map[string]bool, 0)
	for _, scenario := range existing {
		existingScenarioMap[scenario] = true
	}

	result := make([]string, 0)
	for _, scenario := range new {
		if _, ok := existingScenarioMap[scenario]; !ok {
			result = append(result, scenario)
		}
	}
	return result
}

func GetScenariosToKeep(existing []string, input []string) []string {
	existingScenarioMap := make(map[string]bool, 0)
	for _, scenario := range existing {
		existingScenarioMap[scenario] = true
	}

	result := make([]string, 0)
	for _, scenario := range input {
		if _, ok := existingScenarioMap[scenario]; ok {
			result = append(result, scenario)
		}
	}
	return result
}
