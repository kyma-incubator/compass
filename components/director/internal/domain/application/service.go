package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/imdario/mergo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

const (
	intSysKey            = "integrationSystemID"
	nameKey              = "name"
	sccLabelKey          = "scc"
	managedKey           = "managed"
	subaccountKey        = "Subaccount"
	locationIDKey        = "LocationID"
	urlSuffixToBeTrimmed = "/"
)

type repoCreatorFunc func(ctx context.Context, tenant string, application *model.Application) error
type repoUpserterFunc func(ctx context.Context, tenant string, application *model.Application) (string, error)

// ApplicationRepository missing godoc
//go:generate mockery --name=ApplicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	GetByIDForUpdate(ctx context.Context, tenant, id string) (*model.Application, error)
	GetGlobalByID(ctx context.Context, id string) (*model.Application, error)
	GetByNameAndSystemNumber(ctx context.Context, tenant, name, systemNumber string) (*model.Application, error)
	GetByFilter(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Application, error)
	List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationPage, error)
	ListAll(ctx context.Context, tenant string) ([]*model.Application, error)
	ListAllByFilter(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Application, error)
	ListGlobal(ctx context.Context, pageSize int, cursor string) (*model.ApplicationPage, error)
	ListByScenarios(ctx context.Context, tenantID uuid.UUID, scenarios []string, pageSize int, cursor string, hidingSelectors map[string][]string) (*model.ApplicationPage, error)
	Create(ctx context.Context, tenant string, item *model.Application) error
	Update(ctx context.Context, tenant string, item *model.Application) error
	Upsert(ctx context.Context, tenant string, model *model.Application) (string, error)
	TrustedUpsert(ctx context.Context, tenant string, model *model.Application) (string, error)
	TechnicalUpdate(ctx context.Context, item *model.Application) error
	Delete(ctx context.Context, tenant, id string) error
	DeleteGlobal(ctx context.Context, id string) error
}

// LabelRepository missing godoc
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	ListGlobalByKey(ctx context.Context, key string) ([]*model.Label, error)
	ListGlobalByKeyAndObjects(ctx context.Context, objectType model.LabelableObject, objectIDs []string, key string) ([]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
}

// WebhookRepository missing godoc
//go:generate mockery --name=WebhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookRepository interface {
	CreateMany(ctx context.Context, tenant string, items []*model.Webhook) error
}

// FormationService missing godoc
//go:generate mockery --name=FormationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationService interface {
	AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
	UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
}

// RuntimeRepository missing godoc
//go:generate mockery --name=RuntimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListAll(ctx context.Context, tenantID string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
}

// IntegrationSystemRepository missing godoc
//go:generate mockery --name=IntegrationSystemRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationSystemRepository interface {
	Exists(ctx context.Context, id string) (bool, error)
}

// LabelUpsertService missing godoc
//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelUpsertService interface {
	UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

// ScenariosService missing godoc
//go:generate mockery --name=ScenariosService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ScenariosService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error
	AddDefaultScenarioIfEnabled(ctx context.Context, tenant string, labels *map[string]interface{})
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// ApplicationHideCfgProvider missing godoc
//go:generate mockery --name=ApplicationHideCfgProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationHideCfgProvider interface {
	GetApplicationHideSelectors() (map[string][]string, error)
}

type service struct {
	appNameNormalizer  normalizer.Normalizator
	appHideCfgProvider ApplicationHideCfgProvider

	appRepo       ApplicationRepository
	webhookRepo   WebhookRepository
	labelRepo     LabelRepository
	runtimeRepo   RuntimeRepository
	intSystemRepo IntegrationSystemRepository

	labelUpsertService LabelUpsertService
	scenariosService   ScenariosService
	uidService         UIDService
	bndlService        BundleService
	timestampGen       timestamp.Generator
	formationService   FormationService
}

// NewService missing godoc
func NewService(appNameNormalizer normalizer.Normalizator, appHideCfgProvider ApplicationHideCfgProvider, app ApplicationRepository, webhook WebhookRepository, runtimeRepo RuntimeRepository, labelRepo LabelRepository, intSystemRepo IntegrationSystemRepository, labelUpsertService LabelUpsertService, scenariosService ScenariosService, bndlService BundleService, uidService UIDService, formationService FormationService) *service {
	return &service{
		appNameNormalizer:  appNameNormalizer,
		appHideCfgProvider: appHideCfgProvider,
		appRepo:            app,
		webhookRepo:        webhook,
		runtimeRepo:        runtimeRepo,
		labelRepo:          labelRepo,
		intSystemRepo:      intSystemRepo,
		labelUpsertService: labelUpsertService,
		scenariosService:   scenariosService,
		bndlService:        bndlService,
		uidService:         uidService,
		timestampGen:       timestamp.DefaultGenerator,
		formationService:   formationService,
	}
}

// List missing godoc
func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationPage, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.appRepo.List(ctx, appTenant, filter, pageSize, cursor)
}

// ListGlobal missing godoc
func (s *service) ListGlobal(ctx context.Context, pageSize int, cursor string) (*model.ApplicationPage, error) {
	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.appRepo.ListGlobal(ctx, pageSize, cursor)
}

// ListByRuntimeID missing godoc
func (s *service) ListByRuntimeID(ctx context.Context, runtimeID uuid.UUID, pageSize int, cursor string) (*model.ApplicationPage, error) {
	tenantID, err := tenant.LoadFromContext(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, apperrors.NewInvalidDataError("tenantID is not UUID")
	}

	exist, err := s.runtimeRepo.Exists(ctx, tenantID, runtimeID.String())
	if err != nil {
		return nil, errors.Wrap(err, "while checking if runtime exits")
	}

	if !exist {
		return nil, apperrors.NewInvalidDataError("runtime does not exist")
	}

	scenariosLabel, err := s.labelRepo.GetByKey(ctx, tenantID, model.RuntimeLabelableObject, runtimeID.String(), model.ScenariosKey)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return &model.ApplicationPage{
				Data:       []*model.Application{},
				PageInfo:   &pagination.Page{},
				TotalCount: 0,
			}, nil
		}
		return nil, errors.Wrap(err, "while getting scenarios for runtime")
	}

	scenarios, err := label.ValueToStringsSlice(scenariosLabel.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while converting scenarios labels")
	}
	if len(scenarios) == 0 {
		return &model.ApplicationPage{
			Data:       []*model.Application{},
			TotalCount: 0,
			PageInfo: &pagination.Page{
				StartCursor: "",
				EndCursor:   "",
				HasNextPage: false,
			},
		}, nil
	}

	hidingSelectors, err := s.appHideCfgProvider.GetApplicationHideSelectors()
	if err != nil {
		return nil, errors.Wrap(err, "while getting application hide selectors from config")
	}

	return s.appRepo.ListByScenarios(ctx, tenantUUID, scenarios, pageSize, cursor, hidingSelectors)
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.Application, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	app, err := s.appRepo.GetByID(ctx, appTenant, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application with id %s", id)
	}

	return app, nil
}

// GetForUpdate returns an application retrieved globally (without tenant required in the context)
func (s *service) GetForUpdate(ctx context.Context, id string) (*model.Application, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}
	app, err := s.appRepo.GetByIDForUpdate(ctx, appTenant, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application with id %s", id)
	}

	return app, nil
}

// GetByNameAndSystemNumber missing godoc
func (s *service) GetByNameAndSystemNumber(ctx context.Context, name, systemNumber string) (*model.Application, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	app, err := s.appRepo.GetByNameAndSystemNumber(ctx, appTenant, name, systemNumber)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application with name %s and system number %s", name, systemNumber)
	}

	return app, nil
}

// Exist missing godoc
func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "while loading tenant from context")
	}

	exist, err := s.appRepo.Exists(ctx, appTenant, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Application with ID %s", id)
	}

	return exist, nil
}

// Create missing godoc
func (s *service) Create(ctx context.Context, in model.ApplicationRegisterInput) (string, error) {
	creator := func(ctx context.Context, tenant string, application *model.Application) (err error) {
		if err = s.appRepo.Create(ctx, tenant, application); err != nil {
			return errors.Wrapf(err, "while creating Application with name %s", application.Name)
		}
		return
	}

	return s.genericCreate(ctx, in, creator)
}

// GetSccSystem retrieves an application with label key "scc" and value that matches specified subaccount, location id and virtual host
func (s *service) GetSccSystem(ctx context.Context, sccSubaccount, locationID, virtualHost string) (*model.Application, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	sccLabel := struct {
		Host       string `json:"Host"`
		Subaccount string `json:"Subaccount"`
		LocationID string `json:"LocationID"`
	}{
		virtualHost, sccSubaccount, locationID,
	}
	marshal, err := json.Marshal(sccLabel)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshaling sccLabel with subaccount: %s, locationId: %s and virtualHost: %s", appTenant, locationID, virtualHost)
	}

	filter := labelfilter.NewForKeyWithQuery(sccLabelKey, string(marshal))

	app, err := s.appRepo.GetByFilter(ctx, appTenant, []*labelfilter.LabelFilter{filter})
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application with subaccount: %s, locationId: %s and virtualHost: %s", appTenant, locationID, virtualHost)
	}

	return app, nil
}

// ListBySCC retrieves all applications with label matching the specified filter
func (s *service) ListBySCC(ctx context.Context, filter *labelfilter.LabelFilter) ([]*model.ApplicationWithLabel, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	apps, err := s.appRepo.ListAllByFilter(ctx, appTenant, []*labelfilter.LabelFilter{filter})
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Applications by filters: %v", filter)
	}

	if len(apps) == 0 {
		return []*model.ApplicationWithLabel{}, nil
	}

	appIDs := make([]string, 0, len(apps))
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
	}

	labels, err := s.labelRepo.ListGlobalByKeyAndObjects(ctx, model.ApplicationLabelableObject, appIDs, sccLabelKey)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting labels with key scc for applications with IDs: %v", appIDs)
	}

	appIDToLabel := make(map[string]*model.Label, len(labels))
	for _, l := range labels {
		appIDToLabel[l.ObjectID] = l
	}

	appsWithLabel := make([]*model.ApplicationWithLabel, 0, len(apps))
	for _, app := range apps {
		appWithLabel := &model.ApplicationWithLabel{
			App:      app,
			SccLabel: appIDToLabel[app.ID],
		}
		appsWithLabel = append(appsWithLabel, appWithLabel)
	}

	return appsWithLabel, nil
}

// ListSCCs retrieves all SCCs
func (s *service) ListSCCs(ctx context.Context) ([]*model.SccMetadata, error) {
	labels, err := s.labelRepo.ListGlobalByKey(ctx, sccLabelKey)
	if err != nil {
		return nil, errors.Wrap(err, "while getting SCCs by label key: scc")
	}
	sccs := make([]*model.SccMetadata, 0, len(labels))
	for _, sccLabel := range labels {
		v, ok := sccLabel.Value.(map[string]interface{})
		if !ok {
			return nil, errors.New("Label value is not of type map[string]interface{}")
		}

		scc := &model.SccMetadata{
			Subaccount: v[subaccountKey].(string),
			LocationID: v[locationIDKey].(string),
		}

		sccs = append(sccs, scc)
	}
	return sccs, nil
}

// CreateFromTemplate missing godoc
func (s *service) CreateFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) (string, error) {
	creator := func(ctx context.Context, tenant string, application *model.Application) (err error) {
		application.ApplicationTemplateID = appTemplateID
		if err = s.appRepo.Create(ctx, tenant, application); err != nil {
			return errors.Wrapf(err, "while creating Application with name %s from template", application.Name)
		}
		return
	}

	return s.genericCreate(ctx, in, creator)
}

// CreateManyIfNotExistsWithEventualTemplate missing godoc
func (s *service) CreateManyIfNotExistsWithEventualTemplate(ctx context.Context, applicationInputs []model.ApplicationRegisterInputWithTemplate) error {
	appsToAdd, err := s.filterUniqueNonExistingApplications(ctx, applicationInputs)
	if err != nil {
		return errors.Wrap(err, "while filtering unique and non-existing applications")
	}
	log.C(ctx).Infof("Will create %d systems", len(appsToAdd))
	for _, a := range appsToAdd {
		if a.TemplateID == "" {
			_, err = s.Create(ctx, a.ApplicationRegisterInput)
			if err != nil {
				return errors.Wrap(err, "while creating application")
			}
			continue
		}
		_, err = s.CreateFromTemplate(ctx, a.ApplicationRegisterInput, &a.TemplateID)
		if err != nil {
			return errors.Wrap(err, "while creating application")
		}
	}

	return nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	exists, err := s.ensureIntSysExists(ctx, in.IntegrationSystemID)
	if err != nil {
		return errors.Wrap(err, "while validating Integration System ID")
	}

	if !exists {
		return apperrors.NewNotFoundError(resource.IntegrationSystem, *in.IntegrationSystemID)
	}

	app, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Application with id %s", id)
	}

	app.SetFromUpdateInput(in, s.timestampGen())

	if err = s.appRepo.Update(ctx, appTenant, app); err != nil {
		return errors.Wrapf(err, "while updating Application with id %s", id)
	}

	if in.IntegrationSystemID != nil {
		intSysLabel := createLabel(intSysKey, *in.IntegrationSystemID, id)
		err = s.SetLabel(ctx, intSysLabel)
		if err != nil {
			return errors.Wrapf(err, "while setting the integration system label for %s with id %s", intSysLabel.ObjectType, intSysLabel.ObjectID)
		}
		log.C(ctx).Debugf("Successfully set Label for %s with id %s", intSysLabel.ObjectType, intSysLabel.ObjectID)
	}

	label := createLabel(nameKey, s.appNameNormalizer.Normalize(app.Name), app.ID)
	err = s.SetLabel(ctx, label)
	if err != nil {
		return errors.Wrap(err, "while setting application name label")
	}
	log.C(ctx).Debugf("Successfully set Label for Application with id %s", app.ID)
	return nil
}

// Upsert persists application or update it if it already exists
func (s *service) Upsert(ctx context.Context, in model.ApplicationRegisterInput) error {
	tenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	upserterFunc := func(ctx context.Context, tenant string, application *model.Application) (string, error) {
		id, err := s.appRepo.Upsert(ctx, tenant, application)
		if err != nil {
			return "", errors.Wrapf(err, "while upserting Application with name %s", application.Name)
		}
		return id, nil
	}

	return s.genericUpsert(ctx, tenant, in, upserterFunc)
}

// UpdateBaseURL Gets application by ID. If the application does not have a BaseURL set, the API TargetURL is parsed and set as BaseURL
func (s *service) UpdateBaseURL(ctx context.Context, appID, targetURL string) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	app, err := s.Get(ctx, appID)
	if err != nil {
		return err
	}

	if app.BaseURL != nil && len(*app.BaseURL) > 0 {
		log.C(ctx).Infof("BaseURL for Application %s already exists. Will not update it.", appID)
		return nil
	}

	log.C(ctx).Infof("BaseURL for Application %s does not exist. Will update it.", appID)

	parsedTargetURL, err := url.Parse(targetURL)
	if err != nil {
		return errors.Wrapf(err, "while parsing targetURL")
	}

	app.BaseURL = str.Ptr(fmt.Sprintf("%s://%s", parsedTargetURL.Scheme, parsedTargetURL.Host))

	return s.appRepo.Update(ctx, appTenant, app)
}

// TrustedUpsert persists application or update it if it already exists ignoring tenant isolation
func (s *service) TrustedUpsert(ctx context.Context, in model.ApplicationRegisterInput) error {
	tenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	upserterFunc := func(ctx context.Context, tenant string, application *model.Application) (string, error) {
		id, err := s.appRepo.TrustedUpsert(ctx, tenant, application)
		if err != nil {
			return "", errors.Wrapf(err, "while upserting Application with name %s", application.Name)
		}
		return id, nil
	}

	return s.genericUpsert(ctx, tenant, in, upserterFunc)
}

// TrustedUpsertFromTemplate persists application from template id or update it if it already exists ignoring tenant isolation
func (s *service) TrustedUpsertFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) error {
	tenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	upserterFunc := func(ctx context.Context, tenant string, application *model.Application) (string, error) {
		application.ApplicationTemplateID = appTemplateID
		id, err := s.appRepo.TrustedUpsert(ctx, tenant, application)
		if err != nil {
			return "", errors.Wrapf(err, "while upserting Application with name %s from template", application.Name)
		}
		return id, nil
	}

	return s.genericUpsert(ctx, tenant, in, upserterFunc)
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	if err := s.ensureApplicationNotPartOfScenarioWithRuntime(ctx, appTenant, id); err != nil {
		return err
	}

	err = s.appRepo.Delete(ctx, appTenant, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Application with id %s", id)
	}

	return nil
}

// Unpair Checks if the given application is in a scenario with a runtime. Fails if it is.
// When the operation mode is sync, it sets the status condition to model.ApplicationStatusConditionInitial and does a db update, otherwise it only makes an "empty" db update.
func (s *service) Unpair(ctx context.Context, id string) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	if err := s.ensureApplicationNotPartOfScenarioWithRuntime(ctx, appTenant, id); err != nil {
		return err
	}

	app, err := s.appRepo.GetByID(ctx, appTenant, id)
	if err != nil {
		return err
	}

	if opMode := operation.ModeFromCtx(ctx); opMode == graphql.OperationModeSync {
		app.Status = &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
			Timestamp: s.timestampGen(),
		}
	}

	if err = s.appRepo.Update(ctx, appTenant, app); err != nil {
		return err
	}

	return nil
}

// SetLabel missing godoc
func (s *service) SetLabel(ctx context.Context, labelInput *model.LabelInput) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	appExists, err := s.appRepo.Exists(ctx, appTenant, labelInput.ObjectID)
	if err != nil {
		return errors.Wrap(err, "while checking Application existence")
	}
	if !appExists {
		return apperrors.NewNotFoundError(resource.Application, labelInput.ObjectID)
	}

	if labelInput.Key == model.ScenariosKey {
		inputLabels, err := ParseFormationsFromLabelValue(labelInput.Value)
		if err != nil {
			return errors.Wrapf(err, "while parsing formations from input label value")
		}

		inputFormationsMap := make(map[string]struct{}, len(inputLabels))
		for _, f := range inputLabels {
			inputFormationsMap[f] = struct{}{}
		}

		storedLabel, err := s.labelRepo.GetByKey(ctx, appTenant, model.ApplicationLabelableObject, labelInput.ObjectID, model.ScenariosKey)
		storedLabels := make([]string, 0)
		if err != nil && apperrors.ErrorCode(err) != apperrors.NotFound {
			return errors.Wrapf(err, "while getting label with id %s", labelInput.ObjectID)
		} else if err == nil {
			storedLabels, err = ParseFormationsFromLabelValue(storedLabel.Value)
			if err != nil {
				return errors.Wrapf(err, "while getting label with id %s", labelInput.ObjectID)
			}
		}

		storedFormationsMap := make(map[string]struct{}, len(storedLabels))
		for _, f := range storedLabels {
			storedFormationsMap[f] = struct{}{}
		}
		for _, f := range inputLabels {
			if _, ok := storedFormationsMap[f]; !ok {
				if _, err = s.formationService.AssignFormation(ctx, appTenant, labelInput.ObjectID, graphql.FormationObjectTypeApplication, model.Formation{Name: f}); err != nil {
					return errors.Wrapf(err, "while assigning formation with name %s", f)
				}
			}
		}

		for _, f := range storedLabels {
			if _, ok := inputFormationsMap[f]; !ok {
				if _, err = s.formationService.UnassignFormation(ctx, appTenant, labelInput.ObjectID, graphql.FormationObjectTypeApplication, model.Formation{Name: f}); err != nil {
					return errors.Wrapf(err, "while deleting formation with name %s", f)
				}
			}
		}
		return nil
	}

	err = s.labelUpsertService.UpsertLabel(ctx, appTenant, labelInput)
	if err != nil {
		return errors.Wrapf(err, "while creating label for Application")
	}

	return nil
}

// GetLabel missing godoc
func (s *service) GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	appExists, err := s.appRepo.Exists(ctx, appTenant, applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Application existence")
	}
	if !appExists {
		return nil, fmt.Errorf("application with ID %s doesn't exist", applicationID)
	}

	label, err := s.labelRepo.GetByKey(ctx, appTenant, model.ApplicationLabelableObject, applicationID, key)
	if err != nil {
		return nil, errors.Wrap(err, "while getting label for Application")
	}

	return label, nil
}

// ListLabels missing godoc
func (s *service) ListLabels(ctx context.Context, applicationID string) (map[string]*model.Label, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	appExists, err := s.appRepo.Exists(ctx, appTenant, applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Application existence")
	}

	if !appExists {
		return nil, fmt.Errorf("application with ID %s doesn't exist", applicationID)
	}

	labels, err := s.labelRepo.ListForObject(ctx, appTenant, model.ApplicationLabelableObject, applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting label for Application")
	}

	return labels, nil
}

// DeleteLabel missing godoc
func (s *service) DeleteLabel(ctx context.Context, applicationID string, key string, labelValue interface{}) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	appExists, err := s.appRepo.Exists(ctx, appTenant, applicationID)
	if err != nil {
		return errors.Wrap(err, "while checking Application existence")
	}
	if !appExists {
		return fmt.Errorf("application with ID %s doesn't exist", applicationID)
	}

	if key == model.ScenariosKey {
		scenarios, ok := labelValue.([]interface{})
		if !ok {
			return fmt.Errorf("while converting scenarios %s", labelValue)
		}
		for _, scenario := range scenarios {
			scenarioString, ok := scenario.(string)
			if !ok {
				return errors.New("Label value for scenarios key is not of type string")
			}

			if _, err = s.formationService.UnassignFormation(ctx, appTenant, applicationID, graphql.FormationObjectTypeApplication, model.Formation{Name: scenarioString}); err != nil {
				return errors.Wrapf(err, "while unassigning formation %s from application with id %s", scenarioString, applicationID)
			}
		}
		return nil
	}

	err = s.labelRepo.Delete(ctx, appTenant, model.ApplicationLabelableObject, applicationID, key)
	if err != nil {
		return errors.Wrapf(err, "while deleting Application label")
	}

	return nil
}

// Merge merges properties from Source Application into Destination Application, provided that the Destination's
// Application does not have a value set for a given property. Then the Source Application is being deleted.
func (s *service) Merge(ctx context.Context, destID, srcID string) (*model.Application, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	destApp, err := s.Get(ctx, destID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting destination application")
	}

	srcApp, err := s.Get(ctx, srcID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting source application")
	}

	destAppLabels, err := s.labelRepo.ListForObject(ctx, appTenant, model.ApplicationLabelableObject, destID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting labels for Application with id %s", destID)
	}

	srcAppLabels, err := s.labelRepo.ListForObject(ctx, appTenant, model.ApplicationLabelableObject, srcID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting labels for Application with id %s", srcID)
	}

	if destAppLabels == nil {
		destAppLabels = make(map[string]*model.Label)
	}

	if srcAppLabels == nil {
		srcAppLabels = make(map[string]*model.Label)
	}

	srcBaseURL := strings.TrimSuffix(str.PtrStrToStr(srcApp.BaseURL), urlSuffixToBeTrimmed)
	destBaseURL := strings.TrimSuffix(str.PtrStrToStr(destApp.BaseURL), urlSuffixToBeTrimmed)
	if len(srcBaseURL) == 0 || len(destBaseURL) == 0 || srcBaseURL != destBaseURL {
		return nil, errors.Errorf("BaseURL for applications %s and %s are not the same. Destination app BaseURL: %s. Source app BaseURL: %s", destID, srcID, destBaseURL, srcBaseURL)
	}

	srcTemplateID := str.PtrStrToStr(srcApp.ApplicationTemplateID)
	destTemplateID := str.PtrStrToStr(destApp.ApplicationTemplateID)
	if len(srcTemplateID) == 0 || len(destTemplateID) == 0 || srcTemplateID != destTemplateID {
		return nil, errors.Errorf("Application templates are not the same. Destination app template: %s. Source app template: %s", destTemplateID, srcTemplateID)
	}

	if srcApp.Status == nil {
		return nil, errors.Errorf("Could not determine status of source application with id %s", srcID)
	}

	if srcApp.Status.Condition != model.ApplicationStatusConditionInitial {
		return nil, errors.Errorf("Cannot merge application with id %s, because it is in a %s status", srcID, model.ApplicationStatusConditionConnected)
	}

	log.C(ctx).Infof("Merging applications with ids %s and %s", destID, srcID)
	if err := mergo.Merge(destApp, *srcApp); err != nil {
		return nil, errors.Wrapf(err, "while trying to merge applications with ids %s and %s", destID, srcID)
	}

	log.C(ctx).Infof("Merging labels for applications with ids %s and %s", destID, srcID)
	destAppLabelsMerged, err := s.handleMergeLabels(ctx, srcAppLabels, destAppLabels)
	if err != nil {
		return nil, errors.Wrapf(err, "while trying to merge labels for applications with ids %s and %s", destID, srcID)
	}

	log.C(ctx).Infof("Deleting source application with id %s", srcID)
	if err := s.Delete(ctx, srcID); err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Updating destination app with id %s", srcID)
	if err := s.appRepo.Update(ctx, appTenant, destApp); err != nil {
		return nil, err
	}

	if err := s.labelUpsertService.UpsertMultipleLabels(ctx, appTenant, model.ApplicationLabelableObject, destID, destAppLabelsMerged); err != nil {
		return nil, err
	}

	return s.appRepo.GetByID(ctx, appTenant, destID)
}

// handleMergeLabels merges source labels into destination labels. model.ScenariosKey is merged manually as well due to limitation
// of the lib that is used. The last manually merged label is managedKey which is updated only if the destination or
// source label have a value "true"
func (s *service) handleMergeLabels(ctx context.Context, srcAppLabels, destAppLabels map[string]*model.Label) (map[string]interface{}, error) {
	srcScenarios, ok := srcAppLabels[model.ScenariosKey]
	if !ok {
		log.C(ctx).Infof("No %q label found in source object.", model.ScenariosKey)
		srcScenarios = &model.Label{Value: make([]interface{}, 0)}
	}

	destScenarios, ok := destAppLabels[model.ScenariosKey]
	if !ok {
		log.C(ctx).Infof("No %q label found in destination object.", model.ScenariosKey)
		destScenarios = &model.Label{Value: make([]interface{}, 0)}
	}

	srcScenariosStrSlice, err := label.ValueToStringsSlice(srcScenarios.Value)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting source application labels to string slice")
	}

	destScenariosStrSlice, err := label.ValueToStringsSlice(destScenarios.Value)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting destination application labels to string slice")
	}

	for _, srcScenario := range srcScenariosStrSlice {
		if !str.ContainsInSlice(destScenariosStrSlice, srcScenario) {
			destScenariosStrSlice = append(destScenariosStrSlice, srcScenario)
		}
	}

	if err := mergo.Merge(&destAppLabels, srcAppLabels); err != nil {
		return nil, errors.Wrapf(err, "while trying to merge labels")
	}

	destAppLabels[model.ScenariosKey].Value = destScenariosStrSlice

	srcLabelManaged, ok := srcAppLabels[managedKey]
	if !ok {
		log.C(ctx).Infof("No %q label found in source object.", managedKey)
		srcLabelManaged = &model.Label{Value: "false"}
	}

	srcLabelManagedValue, err := str.CastToBool(srcLabelManaged.Value)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s value for source label with ID: %s", managedKey, srcAppLabels[managedKey].ID)
	}

	destLabelManaged, ok := destAppLabels[managedKey]
	if !ok {
		log.C(ctx).Infof("No %q label found in destination object.", managedKey)
		destLabelManaged = &model.Label{Value: "false"}
	}

	destLabelManagedValue, err := str.CastToBool(destLabelManaged.Value)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s value for destination label with ID: %s", managedKey, destAppLabels[managedKey].ID)
	}

	if destLabelManagedValue || srcLabelManagedValue {
		destAppLabels[managedKey].Value = "true"
	}

	conv := make(map[string]interface{}, len(destAppLabels))
	for key, val := range destAppLabels {
		conv[key] = val.Value
	}

	return conv, nil
}

// ensureApplicationNotPartOfScenarioWithRuntime Checks if an application has scenarios associated with it. if a runtime is part of any scenario, then the application is considered being used by that runtime.
func (s *service) ensureApplicationNotPartOfScenarioWithRuntime(ctx context.Context, tenant, appID string) error {
	scenarios, err := s.getScenarioNamesForApplication(ctx, appID)
	if err != nil {
		return err
	}

	validScenarios := removeDefaultScenario(scenarios)
	if len(validScenarios) > 0 {
		runtimes, err := s.getRuntimeNamesForScenarios(ctx, tenant, validScenarios)
		if err != nil {
			return err
		}

		if len(runtimes) > 0 {
			application, err := s.appRepo.GetByID(ctx, tenant, appID)
			if err != nil {
				return errors.Wrapf(err, "while getting application with id %s", appID)
			}
			msg := fmt.Sprintf("System %s is still used and cannot be deleted. Unassign the system from the following formations first: %s. Then, unassign the system from the following runtimes, too: %s", application.Name, strings.Join(validScenarios, ", "), strings.Join(runtimes, ", "))
			return apperrors.NewInvalidOperationError(msg)
		}

		return nil
	}

	return nil
}

func (s *service) createRelatedResources(ctx context.Context, in model.ApplicationRegisterInput, tenant string, applicationID string) error {
	var err error
	webhooks := make([]*model.Webhook, 0, len(in.Webhooks))
	for _, item := range in.Webhooks {
		webhooks = append(webhooks, item.ToWebhook(s.uidService.Generate(), applicationID, model.ApplicationWebhookReference))
	}
	if err = s.webhookRepo.CreateMany(ctx, tenant, webhooks); err != nil {
		return errors.Wrapf(err, "while creating Webhooks for application")
	}

	return nil
}

func (s *service) genericCreate(ctx context.Context, in model.ApplicationRegisterInput, repoCreatorFunc repoCreatorFunc) (string, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}
	log.C(ctx).Debugf("Loaded Application Tenant %s from context", appTenant)

	applications, err := s.appRepo.ListAll(ctx, appTenant)
	if err != nil {
		return "", err
	}

	normalizedName := s.appNameNormalizer.Normalize(in.Name)
	for _, app := range applications {
		if normalizedName == s.appNameNormalizer.Normalize(app.Name) && in.SystemNumber == app.SystemNumber {
			return "", apperrors.NewNotUniqueNameError(resource.Application)
		}
	}

	exists, err := s.ensureIntSysExists(ctx, in.IntegrationSystemID)
	if err != nil {
		return "", errors.Wrap(err, "while ensuring integration system exists")
	}

	if !exists {
		return "", apperrors.NewNotFoundError(resource.IntegrationSystem, *in.IntegrationSystemID)
	}

	id := s.uidService.Generate()
	log.C(ctx).Debugf("ID %s generated for Application with name %s", id, in.Name)

	app := in.ToApplication(s.timestampGen(), id)

	if err = repoCreatorFunc(ctx, appTenant, app); err != nil {
		return "", err
	}

	s.scenariosService.AddDefaultScenarioIfEnabled(ctx, appTenant, &in.Labels)

	if in.Labels == nil {
		in.Labels = map[string]interface{}{}
	}
	in.Labels[intSysKey] = ""
	if in.IntegrationSystemID != nil {
		in.Labels[intSysKey] = *in.IntegrationSystemID
	}
	in.Labels[nameKey] = normalizedName

	if scenarioLabel, ok := in.Labels[model.ScenariosKey]; ok {
		scenariosAsInterfaces, ok := scenarioLabel.([]interface{})
		if !ok {
			return "", errors.New("Label value for scenarios key is not of type []interface{}")
		}
		for _, scenario := range scenariosAsInterfaces {
			scenarioString, ok := scenario.(string)
			if !ok {
				return "", errors.New("Label value for scenarios key is not of type string")
			}
			if _, err := s.formationService.AssignFormation(ctx, appTenant, id, graphql.FormationObjectTypeApplication, model.Formation{Name: scenarioString}); err != nil {
				return "", errors.Wrapf(err, "while assigning formation %s to application with ID %s", scenarioString, id)
			}
		}
		delete(in.Labels, model.ScenariosKey)
	}

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, appTenant, model.ApplicationLabelableObject, id, in.Labels)
	if err != nil {
		return id, errors.Wrapf(err, "while creating multiple labels for Application with id %s", id)
	}

	err = s.createRelatedResources(ctx, in, appTenant, app.ID)
	if err != nil {
		return "", errors.Wrapf(err, "while creating related resources for Application with id %s", id)
	}

	if in.Bundles != nil {
		if err = s.bndlService.CreateMultiple(ctx, id, in.Bundles); err != nil {
			return "", errors.Wrapf(err, "while creating related Bundle resources for Application with id %s", id)
		}
	}

	return id, nil
}

func (s *service) filterUniqueNonExistingApplications(ctx context.Context, applicationInputs []model.ApplicationRegisterInputWithTemplate) ([]model.ApplicationRegisterInputWithTemplate, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	allApps, err := s.appRepo.ListAll(ctx, appTenant)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing all applications for tenant %s", appTenant)
	}
	log.C(ctx).Debugf("Found %d existing systems", len(allApps))

	type key struct {
		name         string
		systemNumber string
	}

	uniqueNonExistingApps := make(map[key]int)
	keys := make([]key, 0)
	for index, ai := range applicationInputs {
		alreadyExits := false
		systemNumber := ""
		if ai.SystemNumber != nil {
			systemNumber = *ai.SystemNumber
		}
		aiKey := key{
			name:         ai.Name,
			systemNumber: systemNumber,
		}

		if _, found := uniqueNonExistingApps[aiKey]; found {
			continue
		}

		for _, a := range allApps {
			bothSystemsAreWithoutSystemNumber := (ai.SystemNumber == nil && a.SystemNumber == nil)
			bothSystemsHaveSystemNumber := (ai.SystemNumber != nil && a.SystemNumber != nil && *(ai.SystemNumber) == *(a.SystemNumber))
			if ai.Name == a.Name && (bothSystemsAreWithoutSystemNumber || bothSystemsHaveSystemNumber) {
				alreadyExits = true
				break
			}
		}

		if !alreadyExits {
			uniqueNonExistingApps[aiKey] = index
			keys = append(keys, aiKey)
		}
	}

	result := make([]model.ApplicationRegisterInputWithTemplate, 0, len(uniqueNonExistingApps))
	for _, key := range keys {
		appInputIndex := uniqueNonExistingApps[key]
		result = append(result, applicationInputs[appInputIndex])
	}

	return result, nil
}

func createLabel(key string, value string, objectID string) *model.LabelInput {
	return &model.LabelInput{
		Key:        key,
		Value:      value,
		ObjectID:   objectID,
		ObjectType: model.ApplicationLabelableObject,
	}
}

func (s *service) ensureIntSysExists(ctx context.Context, id *string) (bool, error) {
	if id == nil {
		return true, nil
	}

	log.C(ctx).Infof("Ensuring Integration System with id %s exists", *id)
	exists, err := s.intSystemRepo.Exists(ctx, *id)
	if err != nil {
		return false, err
	}

	if !exists {
		log.C(ctx).Infof("Integration System with id %s does not exist", *id)
		return false, nil
	}
	log.C(ctx).Infof("Integration System with id %s exists", *id)
	return true, nil
}

func (s *service) getScenarioNamesForApplication(ctx context.Context, applicationID string) ([]string, error) {
	log.C(ctx).Infof("Getting scenarios for application with id %s", applicationID)

	applicationLabel, err := s.GetLabel(ctx, applicationID, model.ScenariosKey)
	if err != nil {
		if apperrors.ErrorCode(err) == apperrors.NotFound {
			log.C(ctx).Infof("No scenarios found for application")
			return nil, nil
		}
		return nil, err
	}

	scenarios, err := label.ValueToStringsSlice(applicationLabel.Value)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing application label values")
	}

	return scenarios, nil
}

func (s *service) getRuntimeNamesForScenarios(ctx context.Context, tenant string, scenarios []string) ([]string, error) {
	scenariosQuery := eventing.BuildQueryForScenarios(scenarios)
	runtimeScenariosFilter := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(model.ScenariosKey, scenariosQuery)}

	log.C(ctx).Debugf("Listing runtimes matching the query %s", scenariosQuery)
	runtimes, err := s.runtimeRepo.ListAll(ctx, tenant, runtimeScenariosFilter)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtimes")
	}

	runtimesNames := make([]string, 0, len(runtimes))
	for _, r := range runtimes {
		runtimesNames = append(runtimesNames, r.Name)
	}

	return runtimesNames, nil
}

func removeDefaultScenario(scenarios []string) []string {
	defaultScenarioIndex := -1
	for idx, scenario := range scenarios {
		if scenario == model.ScenariosDefaultValue[0] {
			defaultScenarioIndex = idx
			break
		}
	}

	if defaultScenarioIndex >= 0 {
		return append(scenarios[:defaultScenarioIndex], scenarios[defaultScenarioIndex+1:]...)
	}

	return scenarios
}

func (s *service) genericUpsert(ctx context.Context, appTenant string, in model.ApplicationRegisterInput, repoUpserterFunc repoUpserterFunc) error {
	exists, err := s.ensureIntSysExists(ctx, in.IntegrationSystemID)
	if err != nil {
		return errors.Wrap(err, "while validating Integration System ID")
	}

	if !exists {
		return apperrors.NewNotFoundError(resource.IntegrationSystem, *in.IntegrationSystemID)
	}

	id := s.uidService.Generate()
	log.C(ctx).Debugf("ID %s generated for Application with name %s", id, in.Name)
	app := in.ToApplication(s.timestampGen(), id)

	id, err = repoUpserterFunc(ctx, appTenant, app)
	if err != nil {
		return errors.Wrap(err, "while upserting application")
	}

	app.ID = id

	s.scenariosService.AddDefaultScenarioIfEnabled(ctx, appTenant, &in.Labels)

	if in.Labels == nil {
		in.Labels = map[string]interface{}{}
	}
	in.Labels[intSysKey] = ""
	if in.IntegrationSystemID != nil {
		in.Labels[intSysKey] = *in.IntegrationSystemID
	}
	in.Labels[nameKey] = s.appNameNormalizer.Normalize(app.Name)

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, appTenant, model.ApplicationLabelableObject, id, in.Labels)
	if err != nil {
		return errors.Wrapf(err, "while creating multiple labels for Application with id %s", id)
	}

	return nil
}

// ParseFormationsFromLabelValue returns available scenarios from the provided schema
func ParseFormationsFromLabelValue(label interface{}) ([]string, error) {
	b, err := json.Marshal(label)
	if err != nil {
		return nil, errors.Wrapf(err, "while label value schema")
	}
	var f []string
	if err = json.Unmarshal(b, &f); err != nil {
		return nil, errors.Wrapf(err, "while label value schema to %T", f)
	}
	return f, nil
}
