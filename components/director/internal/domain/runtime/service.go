package runtime

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"
)

// IsNormalizedLabel represents the label that is used to mark a runtime as normalized
const IsNormalizedLabel = "isNormalized"

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
	ListByFiltersGlobal(context.Context, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	Create(ctx context.Context, tenant string, item *model.Runtime) error
	Update(ctx context.Context, tenant string, item *model.Runtime) error
	ListAll(context.Context, string, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	Delete(ctx context.Context, tenant, id string) error
	GetByFilters(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
	DeleteByKeyNegationPattern(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labelKeyPattern string) error
}

//go:generate mockery --exported --name=labelUpsertService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelUpsertService interface {
	UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

//go:generate mockery --exported --name=scenariosService --output=automock --outpkg=automock --case=underscore --disable-version-string
type scenariosService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error
	AddDefaultScenarioIfEnabled(ctx context.Context, tenant string, labels *map[string]interface{})
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantService interface {
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uidService interface {
	Generate() string
}

type service struct {
	repo      runtimeRepository
	labelRepo labelRepository

	labelUpsertService    labelUpsertService
	uidService            uidService
	scenariosService      scenariosService
	formationService      formationService
	tenantSvc             tenantService
	webhookService        WebhookService
	runtimeContextService RuntimeContextService

	protectedLabelPattern     string
	immutableLabelPattern     string
	runtimeTypeLabelKey       string
	kymaRuntimeTypeLabelValue string
}

// NewService missing godoc
func NewService(repo runtimeRepository,
	labelRepo labelRepository,
	scenariosService scenariosService,
	labelUpsertService labelUpsertService,
	uidService uidService,
	formationService formationService,
	tenantService tenantService,
	webhookService WebhookService,
	runtimeContextService RuntimeContextService,
	protectedLabelPattern, immutableLabelPattern, runtimeTypeLabelKey, kymaRuntimeTypeLabelValue string) *service {
	return &service{
		repo:                      repo,
		labelRepo:                 labelRepo,
		scenariosService:          scenariosService,
		labelUpsertService:        labelUpsertService,
		uidService:                uidService,
		formationService:          formationService,
		tenantSvc:                 tenantService,
		webhookService:            webhookService,
		runtimeContextService:     runtimeContextService,
		protectedLabelPattern:     protectedLabelPattern,
		immutableLabelPattern:     immutableLabelPattern,
		runtimeTypeLabelKey:       runtimeTypeLabelKey,
		kymaRuntimeTypeLabelValue: kymaRuntimeTypeLabelValue,
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

// GetByFilters retrieves model.Runtime matching on the given label filters
func (s *service) GetByFilters(ctx context.Context, filters []*labelfilter.LabelFilter) (*model.Runtime, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	runtime, err := s.repo.GetByFilters(ctx, rtmTenant, filters)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime by filters from repo")
	}
	return runtime, nil
}

// ListByFiltersGlobal missing godoc
func (s *service) ListByFiltersGlobal(ctx context.Context, filters []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	runtimes, err := s.repo.ListByFiltersGlobal(ctx, filters)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtimes by filters from repo")
	}
	return runtimes, nil
}

// ListByFilters lists all runtimes in a given tenant that match given label filter.
func (s *service) ListByFilters(ctx context.Context, filters []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	runtimes, err := s.repo.ListAll(ctx, rtmTenant, filters)
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

// Create creates a runtime in a given tenant.
// If the runtime has a global_subaccount_id label which value is a valid external subaccount from our DB and a child of the caller tenant. The subaccount is used to register the runtime.
// After successful registration, the ASAs in the parent of the caller tenant are processed to add all matching scenarios for the runtime in the parent tenant.
func (s *service) Create(ctx context.Context, in model.RuntimeRegisterInput) (string, error) {
	labels := make(map[string]interface{})
	id := s.uidService.Generate()
	return id, s.CreateWithMandatoryLabels(ctx, in, id, labels)
}

// CreateWithMandatoryLabels creates a runtime in a given tenant and also adds mandatory labels to it.
func (s *service) CreateWithMandatoryLabels(ctx context.Context, in model.RuntimeRegisterInput, id string, mandatoryLabels map[string]interface{}) error {
	if saVal, ok := in.Labels[scenarioassignment.SubaccountIDKey]; ok { // TODO: <backwards-compatibility>: Should be deleted once the provisioner start creating runtimes in a subaccount
		tnt, err := s.extractTenantFromSubaccountLabel(ctx, saVal)
		if err != nil {
			return err
		}
		ctx = tenant.SaveToContext(ctx, tnt.ID, tnt.ExternalTenant)
	}

	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	rtm := in.ToRuntime(id, time.Now(), time.Now())

	if err = s.repo.Create(ctx, rtmTenant, rtm); err != nil {
		return errors.Wrapf(err, "while creating Runtime")
	}

	s.scenariosService.AddDefaultScenarioIfEnabled(ctx, rtmTenant, &in.Labels)

	if in.Labels == nil || in.Labels[IsNormalizedLabel] == nil {
		if in.Labels == nil {
			in.Labels = make(map[string]interface{}, 1)
		}
		in.Labels[IsNormalizedLabel] = "true"
	}

	log.C(ctx).Debugf("Removing protected labels. Labels before: %+v", in.Labels)
	if in.Labels, err = s.UnsafeExtractModifiableLabels(in.Labels); err != nil {
		return err
	}
	log.C(ctx).Debugf("Successfully stripped protected labels. Resulting labels after operation are: %+v", in.Labels)

	for key, value := range mandatoryLabels {
		in.Labels[key] = value
	}

	var scenariosToAssign interface{} = nil
	if _, areScenariosInLabels := in.Labels[model.ScenariosKey]; areScenariosInLabels {
		scenariosToAssign = in.Labels[model.ScenariosKey]
		delete(in.Labels, model.ScenariosKey)
	}

	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading consumer")
	}

	if consumerInfo.ConsumerType == consumer.IntegrationSystem {
		in.Labels[s.runtimeTypeLabelKey] = s.kymaRuntimeTypeLabelValue
	}

	if err = s.labelUpsertService.UpsertMultipleLabels(ctx, rtmTenant, model.RuntimeLabelableObject, id, in.Labels); err != nil {
		return errors.Wrapf(err, "while creating multiple labels for Runtime")
	}

	if scenariosToAssign != nil {
		if err := s.assignRuntimeScenarios(ctx, rtmTenant, id, scenariosToAssign); err != nil {
			return err
		}
	}

	for _, w := range in.Webhooks {
		if _, err = s.webhookService.Create(ctx, rtm.ID, *w, model.RuntimeWebhookReference); err != nil {
			return errors.Wrap(err, "while Creating Webhook for Runtime")
		}
	}

	// The runtime is created successfully, however there can be ASAs in the parent that should be processed.
	tnt, err := s.tenantSvc.GetTenantByID(ctx, rtmTenant)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant with id %s", rtmTenant)
	}

	if len(tnt.Parent) == 0 {
		return nil
	}

	ctxWithParentTenant := tenant.SaveToContext(ctx, tnt.Parent, "")

	mergedScenarios, err := s.formationService.MergeScenariosFromInputLabelsAndAssignments(ctxWithParentTenant, map[string]interface{}{}, id)
	if err != nil {
		return errors.Wrap(err, "while merging scenarios from input and assignments")
	}

	if err := s.assignRuntimeScenarios(ctxWithParentTenant, tnt.Parent, id, mergedScenarios); err != nil {
		return errors.Wrapf(err, "while assigning merged formations")
	}

	return nil
}

// Update updates Runtime and its labels
func (s *service) Update(ctx context.Context, id string, in model.RuntimeUpdateInput) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	rtm, err := s.repo.GetByID(ctx, rtmTenant, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Runtime with id %s", id)
	}

	rtm = in.ToRuntime(id, rtm.CreationTimestamp, time.Now())

	if err = s.repo.Update(ctx, rtmTenant, rtm); err != nil {
		return errors.Wrap(err, "while updating Runtime")
	}

	if in.Labels == nil || in.Labels[IsNormalizedLabel] == nil {
		if in.Labels == nil {
			in.Labels = make(map[string]interface{}, 1)
		}
		in.Labels[IsNormalizedLabel] = "true"
	}

	if err := s.updateScenariosLabel(ctx, rtmTenant, id, in.Labels); err != nil {
		return errors.Wrap(err, "while updating scenarios label")
	}
	delete(in.Labels, model.ScenariosKey)

	log.C(ctx).Debugf("Removing protected labels. Labels before: %+v", in.Labels)
	if in.Labels, err = s.UnsafeExtractModifiableLabels(in.Labels); err != nil {
		return err
	}
	log.C(ctx).Debugf("Successfully stripped protected labels. Resulting labels after operation are: %+v", in.Labels)

	unmodifiablePattern := s.protectedLabelPattern + "|^" + model.ScenariosKey + "$" + "|" + s.immutableLabelPattern
	// NOTE: The db layer does not support OR currently so multiple label patterns can't be implemented easily
	if err = s.labelRepo.DeleteByKeyNegationPattern(ctx, rtmTenant, model.RuntimeLabelableObject, id, unmodifiablePattern); err != nil {
		return errors.Wrapf(err, "while deleting all labels for Runtime")
	}

	if err = s.labelUpsertService.UpsertMultipleLabels(ctx, rtmTenant, model.RuntimeLabelableObject, id, in.Labels); err != nil {
		return errors.Wrapf(err, "while creating multiple labels for Runtime")
	}

	return nil
}

// Delete deletes all RuntimeContexts associated with the runtime with ID `id` and then deletes the runtime and its labels
func (s *service) Delete(ctx context.Context, id string) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	runtimeContexts, err := s.runtimeContextService.ListAllForRuntime(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while listing runtimeContexts for runtime with ID %q", id)
	}

	for _, rc := range runtimeContexts {
		if err = s.runtimeContextService.Delete(ctx, rc.ID); err != nil {
			return errors.Wrapf(err, "while deleting runtimeContext with ID %q", rc.ID)
		}
	}

	if err = s.unassignRuntimeScenarios(ctx, rtmTenant, id); err != nil {
		return err
	}

	if err = s.repo.Delete(ctx, rtmTenant, id); err != nil {
		return errors.Wrapf(err, "while deleting Runtime")
	}

	// All labels are deleted (cascade delete)

	return nil
}

// SetLabel sets Runtime label from a given input
func (s *service) SetLabel(ctx context.Context, labelInput *model.LabelInput) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	if err = s.ensureRuntimeExists(ctx, rtmTenant, labelInput.ObjectID); err != nil {
		return err
	}

	if modifiable, err := isLabelModifiable(labelInput.Key, s.protectedLabelPattern, s.immutableLabelPattern); err != nil {
		return err
	} else if !modifiable {
		return apperrors.NewInvalidDataError("could not set unmodifiable label with key %s", labelInput.Key)
	}

	id := labelInput.ObjectID

	if labelInput.Key == model.ScenariosKey {
		if err := s.updateScenariosLabel(ctx, rtmTenant, id, map[string]interface{}{model.ScenariosKey: labelInput.Value}); err != nil {
			return errors.Wrap(err, "while updating scenarios label")
		}
	} else {
		if err = s.labelUpsertService.UpsertLabel(ctx, rtmTenant, labelInput); err != nil {
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

// DeleteLabel deletes Runtime label from a given label key
func (s *service) DeleteLabel(ctx context.Context, runtimeID string, key string) error {
	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	if err = s.ensureRuntimeExists(ctx, rtmTenant, runtimeID); err != nil {
		return err
	}

	if modifiable, err := isLabelModifiable(key, s.protectedLabelPattern, s.immutableLabelPattern); err != nil {
		return err
	} else if !modifiable {
		return apperrors.NewInvalidDataError("could not delete unmodifiable label with key %s", key)
	}

	if key == model.ScenariosKey {
		if err := s.unassignRuntimeScenarios(ctx, rtmTenant, runtimeID); err != nil {
			return err
		}
	} else {
		if err = s.labelRepo.Delete(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID, key); err != nil {
			return errors.Wrapf(err, "while deleting Runtime label")
		}
	}

	return nil
}

// UnsafeExtractModifiableLabels returns all labels except the protected and immutable labels
func (s *service) UnsafeExtractModifiableLabels(labels map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for labelKey, lbl := range labels {
		modifiable, err := isLabelModifiable(labelKey, s.protectedLabelPattern, s.immutableLabelPattern)
		if err != nil {
			return nil, err
		}
		if modifiable {
			result[labelKey] = lbl
		}
	}
	return result, nil
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

func (s *service) assignRuntimeScenarios(ctx context.Context, rtmTenant, id string, scenarios interface{}) error {
	scenariosStr, err := label.ValueToStringsSlice(scenarios)
	if err != nil {
		return errors.Wrapf(err, "while converting scenarios: %+v to slice of strings", scenarios)
	}

	for _, scenario := range scenariosStr {
		if _, err = s.formationService.AssignFormation(ctx, rtmTenant, id, graphql.FormationObjectTypeRuntime, model.Formation{Name: scenario}); err != nil {
			return errors.Wrapf(err, "while assigning formation %q from runtime with ID %q", scenario, id)
		}
	}

	return nil
}

func (s *service) unassignRuntimeScenarios(ctx context.Context, rtmTenant, runtimeID string) error {
	currentRuntimeLabels, err := s.getCurrentLabelsForRuntime(ctx, rtmTenant, runtimeID)
	if err != nil {
		return err
	}

	if scenarios, areScenariosInLabels := currentRuntimeLabels[model.ScenariosKey]; areScenariosInLabels {
		scenariosStr, err := label.ValueToStringsSlice(scenarios)
		if err != nil {
			return errors.Wrapf(err, "while converting scenarios: %+v to slice of strings", scenarios)
		}

		for _, scenario := range scenariosStr {
			if _, err = s.formationService.UnassignFormation(ctx, rtmTenant, runtimeID, graphql.FormationObjectTypeRuntime, model.Formation{Name: scenario}); err != nil {
				return errors.Wrapf(err, "while unassigning formation %q from runtime with ID %q", scenario, runtimeID)
			}
		}
	}

	return nil
}

func (s *service) updateScenariosLabel(ctx context.Context, rtmTenant, rtmID string, inputLabels map[string]interface{}) error {
	mergedScenarios, err := s.formationService.MergeScenariosFromInputLabelsAndAssignments(ctx, inputLabels, rtmID)
	if err != nil {
		return errors.Wrap(err, "while merging scenarios from input and assignments")
	}

	mergedScenariosSlice, err := label.ValueToStringsSlice(mergedScenarios)
	if err != nil {
		return errors.Wrapf(err, "while converting merged scenarios: %+v to slice of strings", mergedScenarios)
	}

	currentRuntimeLabels, err := s.getCurrentLabelsForRuntime(ctx, rtmTenant, rtmID)
	if err != nil {
		return err
	}

	currentScenarios, areCurrentScenariosInLabels := currentRuntimeLabels[model.ScenariosKey]
	if !areCurrentScenariosInLabels {
		currentScenarios = []interface{}{}
	}

	currentScenariosSlice, err := label.ValueToStringsSlice(currentScenarios)
	if err != nil {
		return errors.Wrapf(err, "while converting current runtime scenarios: %+v to slice of strings", currentScenarios)
	}

	currentScenariosMap := make(map[string]struct{}, len(currentScenariosSlice))
	for _, s := range currentScenariosSlice {
		currentScenariosMap[s] = struct{}{}
	}

	mergedScenariosMap := make(map[string]struct{}, len(mergedScenariosSlice))
	for _, s := range mergedScenariosSlice {
		mergedScenariosMap[s] = struct{}{}
	}

	for scenario := range currentScenariosMap {
		if _, found := mergedScenariosMap[scenario]; !found {
			if _, err = s.formationService.UnassignFormation(ctx, rtmTenant, rtmID, graphql.FormationObjectTypeRuntime, model.Formation{Name: scenario}); err != nil {
				return errors.Wrapf(err, "while unassigning formation %q from runtime with ID %q", scenario, rtmID)
			}
		}
	}

	for scenario := range mergedScenariosMap {
		if _, found := currentScenariosMap[scenario]; !found {
			if _, err = s.formationService.AssignFormation(ctx, rtmTenant, rtmID, graphql.FormationObjectTypeRuntime, model.Formation{Name: scenario}); err != nil {
				return errors.Wrapf(err, "while assigning formation %q from runtime with ID %q", scenario, rtmID)
			}
		}
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

func (s *service) extractTenantFromSubaccountLabel(ctx context.Context, value interface{}) (*model.BusinessTenantMapping, error) {
	callingTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	sa, err := convertLabelValue(value)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s label", scenarioassignment.SubaccountIDKey)
	}

	log.C(ctx).Infof("Runtime registered by tenant %s with %s label with value %s. Will proceed with the subaccount as tenant...", callingTenant, scenarioassignment.SubaccountIDKey, sa)

	tnt, err := s.tenantSvc.GetTenantByExternalID(ctx, sa)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting tenant %s", sa)
	}

	if callingTenant != tnt.ID && callingTenant != tnt.Parent {
		log.C(ctx).Errorf("Caller tenant %s is not parent of the subaccount %s in the %s label", callingTenant, sa, scenarioassignment.SubaccountIDKey)
		return nil, apperrors.NewInvalidOperationError(fmt.Sprintf("Tenant provided in %s label should be child of the caller tenant", scenarioassignment.SubaccountIDKey))
	}
	return tnt, nil
}

func extractUnProtectedLabels(labels map[string]*model.Label, protectedLabelsKeyPattern string) (map[string]*model.Label, error) {
	result := make(map[string]*model.Label)
	for labelKey, label := range labels {
		protected, err := regexp.MatchString(protectedLabelsKeyPattern, labelKey)
		if err != nil {
			return nil, err
		}
		if !protected {
			result[labelKey] = label
		}
	}
	return result, nil
}

func isLabelModifiable(labelKey, protectedLabelsKeyPattern, immutableLabelsKeyPattern string) (bool, error) {
	protected, err := regexp.MatchString(protectedLabelsKeyPattern, labelKey)
	if err != nil {
		return false, err
	}
	immutable, err := regexp.MatchString(immutableLabelsKeyPattern, labelKey)
	if err != nil {
		return false, err
	}
	return !protected && !immutable, err
}

func convertLabelValue(value interface{}) (string, error) {
	values, err := label.ValueToStringsSlice(value)
	if err != nil {
		result := str.CastOrEmpty(value)
		if len(result) == 0 {
			return "", errors.New("cannot cast label value: expected []string or string")
		}
		return result, nil
	}
	if len(values) != 1 {
		return "", errors.New("expected single value for label")
	}
	return values[0], nil
}
