package bundleinstanceauth

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
	ExistForAppAndScenario(ctx context.Context, tenant, scenario, appId string) (bool, error)
	ExistForRuntimeAndScenario(ctx context.Context, tenant, scenario, runtimeId string) (bool, error)
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type ScenarioService interface {
	GetScenarioNamesForApplication(ctx context.Context, applicationID string) ([]string, error)
	GetScenarioNamesForRuntime(ctx context.Context, runtimeID string) ([]string, error)
	GetRuntimeScenarioLabelsForAnyMatchingScenario(ctx context.Context, scenarios []string) ([]model.Label, error)
	GetApplicationScenarioLabelsForAnyMatchingScenario(ctx context.Context, scenarios []string) ([]model.Label, error)
}

//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore
type LabelService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
	UpsertScenarios(ctx context.Context, tenantID string, labels []model.Label, newScenarios []string, mergeFn func(scenarios []string, diffScenario string) []string) error
}

type LabelRepository interface {
	GetBundleInstanceAuthsScenarioLabels(ctx context.Context, appId, runtimeId string) ([]model.Label, error)
}

type scenarioReAssociator struct {
	isAnyBundleInstanceAuthExist         func(ctx context.Context, scenarios []string, objectId string) (bool, error)
	relatedObjectScenarioLabels          func(ctx context.Context, scenarios []string) ([]model.Label, error)
	getBundleInstanceAuthsScenarioLabels func(relatedObjId string) ([]model.Label, error)
	labelService                         LabelService
	labelRepo                            LabelRepository
}

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

	if con.ConsumerType == consumer.Runtime {
		err := s.createInitialBundleInstanceAuthScenarioAssociation(ctx, tnt, id, bundleID, con.ConsumerID)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}

func (s *service) AssociateBundleInstanceAuthForNewApplicationScenarios(ctx context.Context, existingScenarios, inputScenarios []string, appId string) error {
	assoc := &scenarioReAssociator{
		isAnyBundleInstanceAuthExist: s.IsAnyExistForAppAndScenario,
		relatedObjectScenarioLabels:  s.scenarioSvc.GetRuntimeScenarioLabelsForAnyMatchingScenario,
		getBundleInstanceAuthsScenarioLabels: func(runtimeId string) ([]model.Label, error) {
			return s.labelRepo.GetBundleInstanceAuthsScenarioLabels(ctx, appId, runtimeId)
		},
		labelService: s.labelService,
		//labelRepo:    s.labelRepo,
	}

	return assoc.associateBundleInstanceAuthForNewObjectScenarios(ctx, existingScenarios, inputScenarios, appId)
}

func (s *service) AssociateBundleInstanceAuthForNewRuntimeScenarios(ctx context.Context, existingScenarios, inputScenarios []string, runtimeId string) error {
	assoc := &scenarioReAssociator{
		isAnyBundleInstanceAuthExist: s.IsAnyExistForRuntimeAndScenario,
		relatedObjectScenarioLabels:  s.scenarioSvc.GetApplicationScenarioLabelsForAnyMatchingScenario,
		getBundleInstanceAuthsScenarioLabels: func(appId string) ([]model.Label, error) {
			return s.labelRepo.GetBundleInstanceAuthsScenarioLabels(ctx, appId, runtimeId)
		},
		labelService: s.labelService,
		//labelRepo:    s.labelRepo,
	}

	return assoc.associateBundleInstanceAuthForNewObjectScenarios(ctx, existingScenarios, inputScenarios, runtimeId)
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

func (s *service) IsAnyExistForAppAndScenario(ctx context.Context, scenarios []string, appId string) (bool, error) {
	return s.isAnyExistForScenario(ctx, scenarios, func(tenant, currScenario string) (bool, error) {
		return s.repo.ExistForAppAndScenario(ctx, tenant, currScenario, appId)
	})
}

func (s *service) IsAnyExistForRuntimeAndScenario(ctx context.Context, scenarios []string, runtimeId string) (bool, error) {
	return s.isAnyExistForScenario(ctx, scenarios, func(tenant, currScenario string) (bool, error) {
		return s.repo.ExistForRuntimeAndScenario(ctx, tenant, currScenario, runtimeId)
	})
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

func (s *service) createInitialBundleInstanceAuthScenarioAssociation(ctx context.Context, tnt, bundleInstanceAuthId, bundleID, runtimeId string) error {
	applicationID, err := s.bundleSvc.GetByApplicationID(ctx, tnt, bundleID)
	if err != nil {
		return errors.Wrap(err, "While getting application id")
	}
	scenariosForApp, err := s.scenarioSvc.GetScenarioNamesForApplication(ctx, applicationID)
	if err != nil {
		return errors.Wrapf(err, "while fetching scenario names for application: %s", applicationID)
	}

	scenariosForRuntime, err := s.scenarioSvc.GetScenarioNamesForRuntime(ctx, runtimeId)
	if err != nil {
		return errors.Wrapf(err, "white fetching scenario names for runtime: %s", runtimeId)
	}

	commonScenarios := str.IntersectSlice(scenariosForApp, scenariosForRuntime)

	authLabel := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      commonScenarios,
		ObjectID:   bundleInstanceAuthId,
		ObjectType: model.BundleInstanceAuthObject,
	}

	if err = s.labelService.UpsertLabel(ctx, tnt, authLabel); err != nil {
		return errors.Wrap(err, "while creating bundle instance auth scenario label")
	}

	return nil
}

func (s service) isAnyExistForScenario(ctx context.Context, scenarios []string, existFn func(tenant, scenario string) (bool, error)) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, err
	}

	for _, scenario := range scenarios {
		exist, err := existFn(tnt, scenario)

		if err != nil {
			return false, err
		}

		if exist {
			return true, nil
		}
	}
	return false, nil
}

func (sa *scenarioReAssociator) getCommonRelatedObjects(ctx context.Context, scenariosToKeep, scenariosToAdd []string) (map[string][]string, error) {
	relatedObjScenarioLabels, err := sa.relatedObjectScenarioLabels(ctx, scenariosToKeep)
	if err != nil {
		return nil, err
	}

	relatedObjIdsPerScenario := make(map[string]map[string]bool, 0)

	for _, lbl := range relatedObjScenarioLabels {
		scenariosSlice, err := label.GetScenariosAsStringSlice(lbl)
		if err != nil {
			return nil, err
		}

		for _, scenario := range scenariosSlice {
			if _, ok := relatedObjIdsPerScenario[scenario]; !ok {
				relatedObjIdsPerScenario[scenario] = make(map[string]bool, 0)
			}

			relatedObjIdsPerScenario[scenario][lbl.ObjectID] = true
		}
	}

	newScenariosPerRelatedObjId := make(map[string][]string, 0)
	for _, scenario := range scenariosToAdd {
		if _, ok := relatedObjIdsPerScenario[scenario]; !ok {
			continue
		}

		for id := range relatedObjIdsPerScenario[scenario] {
			newScenariosPerRelatedObjId[id] = append(newScenariosPerRelatedObjId[id], scenario)
		}
	}

	return newScenariosPerRelatedObjId, nil
}

func (sa *scenarioReAssociator) associateBundleInstanceAuthForNewObjectScenarios(ctx context.Context, existingScenarios, inputScenarios []string, objId string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	scenariosToRemove := str.SubstractSlice(existingScenarios, inputScenarios)

	authExist, err := sa.isAnyBundleInstanceAuthExist(ctx, scenariosToRemove, objId)
	if err != nil {
		return err
	}

	if authExist {
		return errors.New("Unable to delete label. Bundle Instance Auths should be deleted first")
	}

	scToAdd := str.SubstractSlice(inputScenarios, existingScenarios)
	scenariosToKeep := str.IntersectSlice(existingScenarios, inputScenarios)

	commonRelatedObjects, err := sa.getCommonRelatedObjects(ctx, scenariosToKeep, scToAdd)
	if err != nil {
		return err
	}

	for relatedObjId, scenarios := range commonRelatedObjects {
		bundleInstanceAuthsLabels, err := sa.getBundleInstanceAuthsScenarioLabels(relatedObjId)
		if err != nil {
			return err
		}

		if err := sa.labelService.UpsertScenarios(ctx, tnt, bundleInstanceAuthsLabels, scenarios, label.UniqueScenarios); err != nil {
			authIds := make([]string, 0, len(bundleInstanceAuthsLabels))
			for _, authLabel := range bundleInstanceAuthsLabels {
				authIds = append(authIds, authLabel.ID)
			}

			return errors.Wrap(err, fmt.Sprintf("while associating scenarios: '%s' to all bundle_instance_auths: %s", scenarios))
		}
	}

	return nil
}
