package systemfetcher

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const (
	// LifecycleAttributeName is the lifecycle status attribute of the application in the external source response for applications retrieval.
	LifecycleAttributeName string = "lifecycleStatus"
	// LifecycleDeleted is the string matching the deleted lifecycle state of the application in the external source.
	LifecycleDeleted string = "DELETED"

	// ConcurrentDeleteOperationErrMsg is the error message returned by the Compass Director, when we try to delete an application, which is already undergoing a delete operation.
	ConcurrentDeleteOperationErrMsg = "Concurrent operation [reason=delete operation is in progress]"
	mainURLKey                      = "mainUrl"
	productIDKey                    = "productId"
	displayNameKey                  = "displayName"
	systemNumberKey                 = "systemNumber"
	additionalAttributesKey         = "additionalAttributes"
	productDescriptionKey           = "productDescription"
	infrastructureProviderKey       = "infrastructureProvider"
	additionalUrlsKey               = "additionalUrls"
	ppmsProductVersionIDKey         = "ppmsProductVersionId"
	businessTypeIDKey               = "businessTypeId"
	businessTypeDescriptionKey      = "businessTypeDescription"
)

//go:generate mockery --name=tenantService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type tenantService interface {
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

//go:generate mockery --name=systemsService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type systemsService interface {
	TrustedUpsert(ctx context.Context, in model.ApplicationRegisterInput) error
	TrustedUpsertFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) error
	GetBySystemNumber(ctx context.Context, systemNumber string) (*model.Application, error)
}

// SystemsSyncService is the service for managing systems synchronization timestamps
//
//go:generate mockery --name=SystemsSyncService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type SystemsSyncService interface {
	ListByTenant(ctx context.Context, tenant string) ([]*model.SystemSynchronizationTimestamp, error)
	Upsert(ctx context.Context, in *model.SystemSynchronizationTimestamp) error
}

//go:generate mockery --name=systemsAPIClient --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type systemsAPIClient interface {
	FetchSystemsForTenant(ctx context.Context, tenant *model.BusinessTenantMapping, systemSynchronizationTimestamps map[string]SystemSynchronizationTimestamp) ([]System, error)
}

//go:generate mockery --name=directorClient --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type directorClient interface {
	DeleteSystemAsync(ctx context.Context, id, tenant string) error
}

// Config holds the configuration available for the SystemFetcher.
type Config struct {
	DirectorGraphqlURL        string        `envconfig:"APP_DIRECTOR_GRAPHQL_URL"`
	DirectorRequestTimeout    time.Duration `envconfig:"default=30s,APP_DIRECTOR_REQUEST_TIMEOUT"`
	DirectorSkipSSLValidation bool          `envconfig:"default=false,APP_DIRECTOR_SKIP_SSL_VALIDATION"`

	EnableSystemDeletion   bool   `envconfig:"default=true,APP_ENABLE_SYSTEM_DELETION"`
	OperationalMode        string `envconfig:"APP_OPERATIONAL_MODE"`
	AsyncRequestProcessors int    `envconfig:"default=100,APP_ASYNC_REQUEST_PROCESSORS"`
	TemplatesFileLocation  string `envconfig:"optional,APP_TEMPLATES_FILE_LOCATION"`
	VerifyTenant           string `envconfig:"optional,APP_VERIFY_TENANT"`
}

// SystemFetcher is responsible for synchronizing the existing applications in Compass and a pre-defined external source.
type SystemFetcher struct {
	transaction        persistence.Transactioner
	tenantService      tenantService
	systemsService     systemsService
	systemsSyncService SystemsSyncService
	templateRenderer   TemplateRenderer
	systemsAPIClient   systemsAPIClient
	directorClient     directorClient

	config Config
}

// NewSystemFetcher returns a new SystemFetcher.
func NewSystemFetcher(tx persistence.Transactioner, ts tenantService, ss systemsService, sSync SystemsSyncService, tr TemplateRenderer, sac systemsAPIClient, directorClient directorClient, config Config) *SystemFetcher {
	return &SystemFetcher{
		transaction:        tx,
		tenantService:      ts,
		systemsService:     ss,
		systemsSyncService: sSync,
		templateRenderer:   tr,
		systemsAPIClient:   sac,
		directorClient:     directorClient,
		config:             config,
	}
}

type tenantSystems struct {
	tenant        *model.BusinessTenantMapping
	systems       []System
	syncTimestamp time.Time
}

// SetTemplateRenderer replaces the current template renderer
func (s *SystemFetcher) SetTemplateRenderer(templateRenderer TemplateRenderer) {
	s.templateRenderer = templateRenderer
}

// ProcessTenant performs resync of systems for provided tenant
func (s *SystemFetcher) ProcessTenant(ctx context.Context, tenantID string) error {
	log.C(ctx).Infof("Processing systems for tenant %s", tenantID)
	systemSynchronizationTimestamps, err := s.loadSystemsSynchronizationTimestampsForTenant(ctx, tenantID)
	if err != nil {
		return errors.Wrap(err, "failed while loading systems synchronization timestamps")
	}

	tenant, err := s.getTenantByID(ctx, tenantID)
	if err != nil {
		return errors.Wrapf(err, "failed while loading tenant with ID %s", tenantID)
	}

	if tenant == nil {
		// Stop processing early.
		log.C(ctx).Warnf("Cannot find tenant with ID %s", tenantID)
		return nil
	}

	systems, err := s.systemsAPIClient.FetchSystemsForTenant(ctx, tenant, systemSynchronizationTimestamps)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch systems for tenant %s of type %s", tenant.ExternalTenant, tenant.Type)
	}

	log.C(ctx).Infof("Found %d systems for tenant %s of type %s", len(systems), tenant.ExternalTenant, tenant.Type)
	if len(s.config.VerifyTenant) > 0 {
		log.C(ctx).Infof("Systems: %#v", systems)
	}

	if err := s.processSystemsForTenant(ctx, tenant, systems); err != nil {
		return errors.Wrapf(err, "failed to save systems for tenant %s", tenantID)
	}

	currentTime := time.Now()
	for _, i := range systems {
		currentTimestamp := SystemSynchronizationTimestamp{
			ID:                uuid.NewString(),
			LastSyncTimestamp: currentTime,
		}

		systemPayload, err := json.Marshal(i.SystemPayload)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal a system payload for tenant %s", tenantID)
		}
		productID := gjson.GetBytes(systemPayload, productIDKey).String()

		if v, ok1 := systemSynchronizationTimestamps[productID]; ok1 {
			currentTimestamp.ID = v.ID
		}

		systemSynchronizationTimestamps[productID] = currentTimestamp
	}

	err = s.UpsertSystemsSyncTimestampsForTenant(ctx, tenantID, systemSynchronizationTimestamps)
	if err != nil {
		// Not a breaking case - exit without error.
		log.C(ctx).Warnf("Failed to upsert timestamps for synced systems for tenant %s", tenantID)
	}

	log.C(ctx).Infof("Successfully synced systems for tenant %s", tenantID)
	return nil
}

func (s *SystemFetcher) loadSystemsSynchronizationTimestampsForTenant(ctx context.Context, tenant string) (map[string]SystemSynchronizationTimestamp, error) {
	systemSynchronizationTimestamps := make(map[string]SystemSynchronizationTimestamp, 0)

	tx, err := s.transaction.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer s.transaction.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	syncTimestamps, err := s.systemsSyncService.ListByTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}

	for _, s := range syncTimestamps {
		currentTimestamp := SystemSynchronizationTimestamp{
			ID:                s.ID,
			LastSyncTimestamp: s.LastSyncTimestamp,
		}

		systemSynchronizationTimestamps[s.ProductID] = currentTimestamp
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return systemSynchronizationTimestamps, nil
}

func (s *SystemFetcher) getTenantByID(ctx context.Context, tenantID string) (*model.BusinessTenantMapping, error) {
	tx, err := s.transaction.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer s.transaction.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tenant, err := s.tenantService.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return tenant, nil
}

// UpsertSystemsSyncTimestampsForTenant updates the synchronization timestamps of the systems for each tenant or creates new ones if they don't exist in the database
func (s *SystemFetcher) UpsertSystemsSyncTimestampsForTenant(ctx context.Context, tenant string, timestamps map[string]SystemSynchronizationTimestamp) error {
	tx, err := s.transaction.Begin()
	if err != nil {
		return errors.Wrap(err, "Error while beginning transaction")
	}
	defer s.transaction.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = s.upsertSystemsSyncTimestampsForTenantInternal(ctx, tenant, timestamps)
	if err != nil {
		return errors.Wrapf(err, "failed to upsert systems sync timestamps for tenant %s", tenant)
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (s *SystemFetcher) upsertSystemsSyncTimestampsForTenantInternal(ctx context.Context, tenant string, timestamps map[string]SystemSynchronizationTimestamp) error {
	for productID, timestamp := range timestamps {
		in := &model.SystemSynchronizationTimestamp{
			ID:                timestamp.ID,
			TenantID:          tenant,
			ProductID:         productID,
			LastSyncTimestamp: timestamp.LastSyncTimestamp,
		}

		err := s.systemsSyncService.Upsert(ctx, in)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SystemFetcher) processSystemsForTenant(ctx context.Context, tenantMapping *model.BusinessTenantMapping, systems []System) error {
	log.C(ctx).Infof("Saving %d systems for tenant %s", len(systems), tenantMapping.Name)

	for _, system := range systems {
		err := func() error {
			tx, err := s.transaction.Begin()
			if err != nil {
				return errors.Wrap(err, "failed to begin transaction")
			}
			ctx = tenant.SaveToContext(ctx, tenantMapping.ID, tenantMapping.ExternalTenant)
			ctx = persistence.SaveToContext(ctx, tx)
			defer s.transaction.RollbackUnlessCommitted(ctx, tx)

			systemPayload, err := json.Marshal(system.SystemPayload)
			if err != nil {
				log.C(ctx).Error(errors.Wrapf(err, "failed to marshal a system payload for tenant %s", tenantMapping.ExternalTenant))
				return err
			}
			displayName := gjson.GetBytes(systemPayload, displayNameKey).String()
			systemNumber := gjson.GetBytes(systemPayload, systemNumberKey).String()
			lifecycleStatus := gjson.GetBytes(systemPayload, additionalAttributesKey+"."+LifecycleAttributeName).String()

			log.C(ctx).Infof("Getting system by name %s and system number %s", displayName, systemNumber)

			system.StatusCondition = model.ApplicationStatusConditionInitial
			app, err := s.systemsService.GetBySystemNumber(ctx, systemNumber)
			if err != nil {
				if !apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("Could not get system with name %s and system number %s", displayName, systemNumber)
					return nil
				}
			}

			if lifecycleStatus == LifecycleDeleted && s.config.EnableSystemDeletion {
				if app == nil {
					log.C(ctx).Warnf("System with system number %s is not present. Skipping deletion.", systemNumber)
					return nil
				}

				if !app.Ready && !app.GetDeletedAt().IsZero() {
					log.C(ctx).Infof("System with id %s is currently being deleted", app.ID)
					return nil
				}
				if err := s.directorClient.DeleteSystemAsync(ctx, app.ID, tenantMapping.ID); err != nil {
					if strings.Contains(err.Error(), ConcurrentDeleteOperationErrMsg) {
						log.C(ctx).Warnf("Delete operation is in progress for system with id %s", app.ID)
					} else {
						log.C(ctx).WithError(err).Errorf("Could not delete system with id %s", app.ID)
					}
					return nil
				}
				log.C(ctx).Infof("Started asynchronously delete for system with id %s", app.ID)
				return nil
			}

			if app != nil && app.Status != nil {
				system.StatusCondition = app.Status.Condition
			}

			appInput, err := s.convertSystemToAppRegisterInput(ctx, system)
			if err != nil {
				return err
			}

			log.C(ctx).Infof("Started processing tenant business type for system with system number %s", systemNumber)
			appInput = s.enrichAppInputLabelsWithTenantBusinessType(systemPayload, appInput)

			if err = s.systemsService.TrustedUpsertFromTemplate(ctx, appInput.ApplicationRegisterInput, &appInput.TemplateID); err != nil {
				return errors.Wrap(err, "while upserting application")
			}

			if err = tx.Commit(); err != nil {
				return errors.Wrapf(err, "failed to commit applications for tenant %s", tenantMapping.ExternalTenant)
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemFetcher) convertSystemToAppRegisterInput(ctx context.Context, sc System) (*model.ApplicationRegisterInputWithTemplate, error) {
	input, err := s.templateRenderer.ApplicationRegisterInputFromTemplate(ctx, sc)
	if err != nil {
		return nil, err
	}

	return &model.ApplicationRegisterInputWithTemplate{
		ApplicationRegisterInput: *input,
		TemplateID:               sc.TemplateID,
	}, nil
}

func (s *SystemFetcher) enrichAppInputLabelsWithTenantBusinessType(systemPayload []byte, appInput *model.ApplicationRegisterInputWithTemplate) *model.ApplicationRegisterInputWithTemplate {
	businessTypeID := gjson.GetBytes(systemPayload, businessTypeIDKey).String()
	businessTypeDescription := gjson.GetBytes(systemPayload, businessTypeDescriptionKey).String()
	if businessTypeID != "" && businessTypeDescription != "" {
		if len(appInput.Labels) == 0 {
			appInput.Labels = map[string]interface{}{}
		}

		appInput.Labels[application.TenantBusinessTypeCodeLabelKey] = businessTypeID
		appInput.Labels[application.TenantBusinessTypeNameLabelKey] = businessTypeDescription
	}

	return appInput
}
