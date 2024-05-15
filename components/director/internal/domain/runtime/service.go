package runtime

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"k8s.io/utils/strings/slices"

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

const (
	// IsNormalizedLabel represents the label that is used to mark a runtime as normalized
	IsNormalizedLabel = "isNormalized"

	// RegionLabelKey is the key of the tenant label for region.
	RegionLabelKey = "region"
)

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

//go:generate mockery --exported --name=labelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelService interface {
	UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantService interface {
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uidService interface {
	Generate() string
}

type service struct {
	repo      runtimeRepository
	labelRepo labelRepository

	labelService          labelService
	uidService            uidService
	formationService      formationService
	tenantSvc             tenantService
	webhookService        WebhookService
	runtimeContextService RuntimeContextService

	protectedLabelPattern         string
	immutableLabelPattern         string
	runtimeTypeLabelKey           string
	kymaRuntimeTypeLabelValue     string
	kymaApplicationNamespaceValue string

	kymaAdapterWebhookMode           string
	kymaAdapterWebhookType           string
	kymaAdapterWebhookURLTemplate    string
	kymaAdapterWebhookInputTemplate  string
	kymaAdapterWebhookHeaderTemplate string
	kymaAdapterWebhookOutputTemplate string
}

// NewService missing godoc
func NewService(repo runtimeRepository,
	labelRepo labelRepository,
	labelService labelService,
	uidService uidService,
	formationService formationService,
	tenantService tenantService,
	webhookService WebhookService,
	runtimeContextService RuntimeContextService,
	protectedLabelPattern, immutableLabelPattern, runtimeTypeLabelKey, kymaRuntimeTypeLabelValue, kymaApplicationNamespaceValue,
	kymaAdapterWebhookMode, kymaAdapterWebhookType, kymaAdapterWebhookURLTemplate, kymaAdapterWebhookInputTemplate, kymaAdapterWebhookHeaderTemplate, kymaAdapterWebhookOutputTemplate string) *service {
	return &service{
		repo:                             repo,
		labelRepo:                        labelRepo,
		labelService:                     labelService,
		uidService:                       uidService,
		formationService:                 formationService,
		tenantSvc:                        tenantService,
		webhookService:                   webhookService,
		runtimeContextService:            runtimeContextService,
		protectedLabelPattern:            protectedLabelPattern,
		immutableLabelPattern:            immutableLabelPattern,
		runtimeTypeLabelKey:              runtimeTypeLabelKey,
		kymaRuntimeTypeLabelValue:        kymaRuntimeTypeLabelValue,
		kymaApplicationNamespaceValue:    kymaApplicationNamespaceValue,
		kymaAdapterWebhookMode:           kymaAdapterWebhookMode,
		kymaAdapterWebhookType:           kymaAdapterWebhookType,
		kymaAdapterWebhookURLTemplate:    kymaAdapterWebhookURLTemplate,
		kymaAdapterWebhookInputTemplate:  kymaAdapterWebhookInputTemplate,
		kymaAdapterWebhookHeaderTemplate: kymaAdapterWebhookHeaderTemplate,
		kymaAdapterWebhookOutputTemplate: kymaAdapterWebhookOutputTemplate,
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
	var subaccountTnt string
	if saVal, ok := in.Labels[scenarioassignment.SubaccountIDKey]; ok { // TODO: <backwards-compatibility>: Should be deleted once the provisioner start creating runtimes in a subaccount
		tnt, err := s.extractTenantFromSubaccountLabel(ctx, saVal)
		if err != nil {
			return err
		}
		subaccountTnt = tnt.ID
		ctx = tenant.SaveToContext(ctx, tnt.ID, tnt.ExternalTenant)
	}

	rtmTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading consumer")
	}

	isConsumerIntegrationSystem := consumerInfo.Type == consumer.IntegrationSystem
	if isConsumerIntegrationSystem {
		in.ApplicationNamespace = &s.kymaApplicationNamespaceValue
	}

	if _, areScenariosInLabels := in.Labels[model.ScenariosKey]; areScenariosInLabels {
		return errors.Errorf("label with key %s cannot be set explicitly", model.ScenariosKey)
	}

	rtm := in.ToRuntime(id, time.Now(), time.Now())

	if err = s.repo.Create(ctx, rtmTenant, rtm); err != nil {
		return errors.Wrapf(err, "while creating Runtime")
	}

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

	if isConsumerIntegrationSystem {
		in.Labels[s.runtimeTypeLabelKey] = s.kymaRuntimeTypeLabelValue

		region, err := s.extractRegionFromSubaccountTenant(ctx, subaccountTnt)
		if err != nil {
			return err
		}
		in.Labels[RegionLabelKey] = region

		webhookMode := model.WebhookMode(s.kymaAdapterWebhookMode)

		webhook := &model.WebhookInput{
			Mode: &webhookMode,
			Type: model.WebhookType(s.kymaAdapterWebhookType),
			Auth: &model.AuthInput{
				AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
			},
			URLTemplate:    &s.kymaAdapterWebhookURLTemplate,
			InputTemplate:  &s.kymaAdapterWebhookInputTemplate,
			HeaderTemplate: &s.kymaAdapterWebhookHeaderTemplate,
			OutputTemplate: &s.kymaAdapterWebhookOutputTemplate,
		}
		in.Webhooks = append(in.Webhooks, webhook)
	}

	if err = s.labelService.UpsertMultipleLabels(ctx, rtmTenant, model.RuntimeLabelableObject, id, in.Labels); err != nil {
		return errors.Wrapf(err, "while creating multiple labels for Runtime")
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

	for _, parentTenantID := range tnt.Parents {
		ctxWithParentTenant := tenant.SaveToContext(ctx, parentTenantID, "")

		scenariosToAssign, err := s.formationService.GetScenariosFromMatchingASAs(ctxWithParentTenant, id, graphql.FormationObjectTypeRuntime)
		if err != nil {
			return errors.Wrap(err, "while merging scenarios from input and assignments")
		}

		if err := s.assignRuntimeScenarios(ctxWithParentTenant, parentTenantID, id, scenariosToAssign); err != nil {
			return errors.Wrapf(err, "while assigning merged formations")
		}
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

	if _, areScenariosInLabels := in.Labels[model.ScenariosKey]; areScenariosInLabels {
		return errors.Errorf("label with key %s cannot be set explicitly", model.ScenariosKey)
	}

	rtm.SetFromUpdateInput(in, id, rtm.CreationTimestamp, time.Now())

	if err = s.repo.Update(ctx, rtmTenant, rtm); err != nil {
		return errors.Wrap(err, "while updating Runtime")
	}

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

	unmodifiablePattern := s.protectedLabelPattern + "|" + s.immutableLabelPattern
	// NOTE: The db layer does not support OR currently so multiple label patterns can't be implemented easily
	if err = s.labelRepo.DeleteByKeyNegationPattern(ctx, rtmTenant, model.RuntimeLabelableObject, id, unmodifiablePattern); err != nil {
		return errors.Wrapf(err, "while deleting all labels for Runtime")
	}

	if err = s.labelService.UpsertMultipleLabels(ctx, rtmTenant, model.RuntimeLabelableObject, id, in.Labels); err != nil {
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

	if labelInput.Key == model.ScenariosKey {
		return errors.Errorf("label with key %s cannot be set explicitly", model.ScenariosKey)
	}

	if err = s.labelService.UpsertLabel(ctx, rtmTenant, labelInput); err != nil {
		return errors.Wrapf(err, "while creating label for Runtime")
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
		return errors.Errorf("label with key %s cannot be deleted explicitly", model.ScenariosKey)
	}

	if err = s.labelRepo.Delete(ctx, rtmTenant, model.RuntimeLabelableObject, runtimeID, key); err != nil {
		return errors.Wrapf(err, "while deleting Runtime label")
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

func (s *service) assignRuntimeScenarios(ctx context.Context, rtmTenant, id string, scenarios []string) error {
	for _, scenario := range scenarios {
		if _, err := s.formationService.AssignFormation(ctx, rtmTenant, id, graphql.FormationObjectTypeRuntime, model.Formation{Name: scenario}); err != nil {
			return errors.Wrapf(err, "while assigning formation %q from runtime with ID %q", scenario, id)
		}
	}

	return nil
}

func (s *service) unassignRuntimeScenarios(ctx context.Context, rtmTenant, runtimeID string) error {
	formations, err := s.formationService.ListFormationsForObject(ctx, runtimeID)
	if err != nil {
		return errors.Wrapf(err, "while listing formations for runtime with ID %q", runtimeID)
	}

	for _, formation := range formations {
		if _, err = s.formationService.UnassignFormation(ctx, rtmTenant, runtimeID, graphql.FormationObjectTypeRuntime, model.Formation{Name: formation.Name}); err != nil {
			return errors.Wrapf(err, "while unassigning formation %q from runtime with ID %q", formation.Name, runtimeID)
		}
	}

	return nil
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

	if callingTenant != tnt.ID && !slices.Contains(tnt.Parents, callingTenant) {
		log.C(ctx).Errorf("Caller tenant %s is not parent of the subaccount %s in the %s label", callingTenant, sa, scenarioassignment.SubaccountIDKey)
		return nil, apperrors.NewInvalidOperationError(fmt.Sprintf("Tenant provided in %s label should be child of the caller tenant", scenarioassignment.SubaccountIDKey))
	}
	return tnt, nil
}

func (s *service) extractRegionFromSubaccountTenant(ctx context.Context, subaccountTnt string) (string, error) {
	if subaccountTnt == "" {
		return "", nil
	}

	regionLabel, err := s.labelService.GetByKey(ctx, subaccountTnt, model.TenantLabelableObject, subaccountTnt, RegionLabelKey)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return "", errors.Wrapf(err, "while getting label %q for %q with id %q", RegionLabelKey, model.TenantLabelableObject, subaccountTnt)
		}
	}

	regionValue := ""
	if regionLabel != nil && regionLabel.Value != nil {
		if regionLabelValue, ok := regionLabel.Value.(string); ok {
			regionValue = regionLabelValue
		}
	}

	return regionValue, nil
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
