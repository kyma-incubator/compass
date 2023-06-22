package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"

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
	ListByType(ctx context.Context, tenantType tenantEntity.Type) ([]*model.BusinessTenantMapping, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
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
	List(ctx context.Context) ([]*model.SystemSynchronizationTimestamp, error)
	Upsert(ctx context.Context, in *model.SystemSynchronizationTimestamp) error
}

//go:generate mockery --name=tenantBusinessTypeService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type tenantBusinessTypeService interface {
	Create(ctx context.Context, in *model.TenantBusinessTypeInput) (string, error)
	GetByID(ctx context.Context, id string) (*model.TenantBusinessType, error)
	ListAll(ctx context.Context) ([]*model.TenantBusinessType, error)
}

//go:generate mockery --name=systemsAPIClient --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type systemsAPIClient interface {
	FetchSystemsForTenant(ctx context.Context, tenant string, mutex *sync.Mutex) ([]System, error)
}

//go:generate mockery --name=directorClient --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type directorClient interface {
	DeleteSystemAsync(ctx context.Context, id, tenant string) error
}

//go:generate mockery --name=templateRenderer --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type templateRenderer interface {
	ApplicationRegisterInputFromTemplate(ctx context.Context, sc System) (*model.ApplicationRegisterInput, error)
}

// Config holds the configuration available for the SystemFetcher.
type Config struct {
	SystemsQueueSize          int           `envconfig:"default=100,APP_SYSTEM_INFORMATION_QUEUE_SIZE"`
	FetcherParallellism       int           `envconfig:"default=30,APP_SYSTEM_INFORMATION_PARALLELLISM"`
	DirectorGraphqlURL        string        `envconfig:"APP_DIRECTOR_GRAPHQL_URL"`
	DirectorRequestTimeout    time.Duration `envconfig:"default=30s,APP_DIRECTOR_REQUEST_TIMEOUT"`
	DirectorSkipSSLValidation bool          `envconfig:"default=false,APP_DIRECTOR_SKIP_SSL_VALIDATION"`

	EnableSystemDeletion  bool   `envconfig:"default=true,APP_ENABLE_SYSTEM_DELETION"`
	OperationalMode       string `envconfig:"APP_OPERATIONAL_MODE"`
	TemplatesFileLocation string `envconfig:"optional,APP_TEMPLATES_FILE_LOCATION"`
	VerifyTenant          string `envconfig:"optional,APP_VERIFY_TENANT"`
}

// SystemFetcher is responsible for synchronizing the existing applications in Compass and a pre-defined external source.
type SystemFetcher struct {
	transaction        persistence.Transactioner
	tenantService      tenantService
	systemsService     systemsService
	systemsSyncService SystemsSyncService
	tbtService         tenantBusinessTypeService
	templateRenderer   templateRenderer
	systemsAPIClient   systemsAPIClient
	directorClient     directorClient

	config  Config
	workers chan struct{}
}

// NewSystemFetcher returns a new SystemFetcher.
func NewSystemFetcher(tx persistence.Transactioner, ts tenantService, ss systemsService, sSync SystemsSyncService, tbts tenantBusinessTypeService, tr templateRenderer, sac systemsAPIClient, directorClient directorClient, config Config) *SystemFetcher {
	return &SystemFetcher{
		transaction:        tx,
		tenantService:      ts,
		systemsService:     ss,
		systemsSyncService: sSync,
		tbtService:         tbts,
		templateRenderer:   tr,
		systemsAPIClient:   sac,
		directorClient:     directorClient,
		workers:            make(chan struct{}, config.FetcherParallellism),
		config:             config,
	}
}

type tenantSystems struct {
	tenant  *model.BusinessTenantMapping
	systems []System
}

func splitBusinessTenantMappingsToChunks(slice []*model.BusinessTenantMapping, chunkSize int) [][]*model.BusinessTenantMapping {
	var chunks [][]*model.BusinessTenantMapping
	for {
		if len(slice) == 0 {
			break
		}

		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}

	return chunks
}

// SyncSystems synchronizes applications between Compass and external source. It deletes the applications with deleted state in the external source from Compass,
// and creates any new applications present in the external source.
func (s *SystemFetcher) SyncSystems(ctx context.Context) error {
	tenants, err := s.listTenants(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list tenants")
	}

	tenantBusinessTypes, err := s.getTenantBusinessTypes(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get tenant business types")
	}

	systemsQueue := make(chan tenantSystems, s.config.SystemsQueueSize)
	wgDB := sync.WaitGroup{}
	wgDB.Add(1)
	var mutex sync.Mutex
	go func() {
		defer func() {
			wgDB.Done()
		}()
		for tenantSystems := range systemsQueue {
			entry := log.C(ctx)
			entry = entry.WithField(log.FieldRequestID, uuid.New().String())
			ctx = log.ContextWithLogger(ctx, entry)

			if err = s.processSystemsForTenant(ctx, tenantSystems.tenant, tenantSystems.systems, tenantBusinessTypes); err != nil {
				log.C(ctx).Error(errors.Wrap(err, fmt.Sprintf("failed to save systems for tenant %s", tenantSystems.tenant.ExternalTenant)))
				continue
			}

			mutex.Lock()
			if SystemSynchronizationTimestamps == nil {
				SystemSynchronizationTimestamps = make(map[string]map[string]SystemSynchronizationTimestamp, 0)
			}

			for _, i := range tenantSystems.systems {
				currentTenant := tenantSystems.tenant.ExternalTenant
				currentTimestamp := SystemSynchronizationTimestamp{
					ID:                uuid.NewString(),
					LastSyncTimestamp: time.Now().UTC(),
				}

				if _, ok := SystemSynchronizationTimestamps[currentTenant]; !ok {
					SystemSynchronizationTimestamps[currentTenant] = make(map[string]SystemSynchronizationTimestamp, 0)
				}

				systemPayload, err := json.Marshal(i.SystemPayload)
				if err != nil {
					log.C(ctx).Error(errors.Wrapf(err, "failed to marshal a system payload for tenant %s", tenantSystems.tenant.ExternalTenant))
					return
				}
				productID := gjson.GetBytes(systemPayload, productIDKey).String()

				if v, ok1 := SystemSynchronizationTimestamps[currentTenant][productID]; ok1 {
					currentTimestamp.ID = v.ID
				}

				SystemSynchronizationTimestamps[currentTenant][productID] = currentTimestamp
			}
			mutex.Unlock()

			log.C(ctx).Info(fmt.Sprintf("Successfully synced systems for tenant %s", tenantSystems.tenant.ExternalTenant))
		}
	}()

	chunks := splitBusinessTenantMappingsToChunks(tenants, 15)

	for _, chunk := range chunks {
		time.Sleep(time.Second * 1)

		wg := sync.WaitGroup{}
		for _, t := range chunk {
			wg.Add(1)
			s.workers <- struct{}{}
			go func(t *model.BusinessTenantMapping) {
				defer func() {
					wg.Done()
					<-s.workers
				}()
				systems, err := s.systemsAPIClient.FetchSystemsForTenant(ctx, t.ExternalTenant, &mutex)
				if err != nil {
					log.C(ctx).Error(errors.Wrap(err, fmt.Sprintf("failed to fetch systems for tenant %s", t.ExternalTenant)))
					return
				}

				log.C(ctx).Infof("found %d systems for tenant %s", len(systems), t.ExternalTenant)
				if len(s.config.VerifyTenant) > 0 {
					log.C(ctx).Infof("systems: %#v", systems)
				}

				if len(systems) > 0 {
					systemsQueue <- tenantSystems{
						tenant:  t,
						systems: systems,
					}
				}
			}(t)
		}

		wg.Wait()
	}
	close(systemsQueue)
	wgDB.Wait()

	return nil
}

// UpsertSystemsSyncTimestamps updates the synchronization timestamps of the systems for each tenant or creates new ones if they don't exist in the database
func (s *SystemFetcher) UpsertSystemsSyncTimestamps(ctx context.Context, transact persistence.Transactioner) error {
	tx, err := transact.Begin()
	if err != nil {
		return errors.Wrap(err, "Error while beginning transaction")
	}
	defer transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	for tnt, v := range SystemSynchronizationTimestamps {
		err := s.upsertSystemsSyncTimestampsForTenant(ctx, tnt, v)
		if err != nil {
			return errors.Wrapf(err, "failed to upsert systems sync timestamps for tenant %s", tnt)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (s *SystemFetcher) upsertSystemsSyncTimestampsForTenant(ctx context.Context, tenant string, timestamps map[string]SystemSynchronizationTimestamp) error {
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

func (s *SystemFetcher) listTenants(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	tx, err := s.transaction.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer s.transaction.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var tenants []*model.BusinessTenantMapping
	if len(s.config.VerifyTenant) > 0 {
		singleTenant, err := s.tenantService.GetTenantByExternalID(ctx, s.config.VerifyTenant)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to retrieve tenant %s", s.config.VerifyTenant)
		}
		tenants = append(tenants, singleTenant)
	} else {
		tenants, err = s.tenantService.ListByType(ctx, tenantEntity.Account)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve tenants")
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit while retrieving tenants")
	}

	return tenants, nil
}

func (s *SystemFetcher) processSystemsForTenant(ctx context.Context, tenantMapping *model.BusinessTenantMapping, systems []System, tenantBusinessTypes map[string]*model.TenantBusinessType) error {
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

			log.C(ctx).Infof("Started processing tenant business type for system with system number %s", systemNumber)
			tenantBusinessType, err := s.processSystemTenantBusinessType(ctx, systemPayload, tenantBusinessTypes)
			if err != nil {
				return err
			}

			appInput, err := s.convertSystemToAppRegisterInput(ctx, system)
			if err != nil {
				return err
			}
			if tenantBusinessType != nil {
				appInput.TenantBusinessTypeID = &tenantBusinessType.ID
			}

			if appInput.TemplateID == "" {
				if err = s.systemsService.TrustedUpsert(ctx, appInput.ApplicationRegisterInput); err != nil {
					return errors.Wrap(err, "while upserting application")
				}
			} else {
				if err = s.systemsService.TrustedUpsertFromTemplate(ctx, appInput.ApplicationRegisterInput, &appInput.TemplateID); err != nil {
					return errors.Wrap(err, "while upserting application")
				}
			}

			if err = tx.Commit(); err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to commit applications for tenant %s", tenantMapping.ExternalTenant))
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
	input, err := s.appRegisterInput(ctx, sc)
	if err != nil {
		return nil, err
	}

	return &model.ApplicationRegisterInputWithTemplate{
		ApplicationRegisterInput: *input,
		TemplateID:               sc.TemplateID,
	}, nil
}

func (s *SystemFetcher) appRegisterInput(ctx context.Context, sc System) (*model.ApplicationRegisterInput, error) {
	if len(sc.TemplateID) > 0 {
		return s.templateRenderer.ApplicationRegisterInputFromTemplate(ctx, sc)
	}

	payload, err := json.Marshal(sc.SystemPayload)
	if err != nil {
		return nil, err
	}

	return &model.ApplicationRegisterInput{
		Name:            gjson.GetBytes(payload, displayNameKey).String(),
		Description:     str.Ptr(gjson.GetBytes(payload, productDescriptionKey).String()),
		StatusCondition: &sc.StatusCondition,
		ProviderName:    str.Ptr(gjson.GetBytes(payload, infrastructureProviderKey).String()),
		BaseURL:         str.Ptr(gjson.GetBytes(payload, additionalUrlsKey+"."+mainURLKey).String()),
		SystemNumber:    str.Ptr(gjson.GetBytes(payload, systemNumberKey).String()),
		Labels: map[string]interface{}{
			"managed":              "true",
			"productId":            str.Ptr(gjson.GetBytes(payload, productIDKey).String()),
			"ppmsProductVersionId": str.Ptr(gjson.GetBytes(payload, ppmsProductVersionIDKey).String()),
		},
	}, nil
}

func (s *SystemFetcher) getTenantBusinessTypes(ctx context.Context) (map[string]*model.TenantBusinessType, error) {
	tx, err := s.transaction.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer s.transaction.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tenantBusinessTypes, err := s.tbtService.ListAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve tenant business types")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit while retrieving tenant business types")
	}

	tbtMap := make(map[string]*model.TenantBusinessType, 0)
	for _, tbt := range tenantBusinessTypes {
		tbtMap[tbt.Code] = tbt
	}

	return tbtMap, nil
}

func (s *SystemFetcher) processSystemTenantBusinessType(ctx context.Context, systemPayload []byte, tenantBusinessTypes map[string]*model.TenantBusinessType) (*model.TenantBusinessType, error) {
	businessTypeID := gjson.GetBytes(systemPayload, businessTypeIDKey).String()
	businessTypeDescription := gjson.GetBytes(systemPayload, businessTypeDescriptionKey).String()
	tbt, exists := tenantBusinessTypes[businessTypeID]
	if businessTypeID != "" && businessTypeDescription != "" {
		if !exists {
			log.C(ctx).Infof("Creating tenant business type with code: %q", businessTypeID)
			createdTbtID, err := s.tbtService.Create(ctx, &model.TenantBusinessTypeInput{Code: businessTypeID, Name: businessTypeDescription})
			if err != nil {
				return nil, err
			}
			createdTbt, err := s.tbtService.GetByID(ctx, createdTbtID)
			if err != nil {
				return nil, err
			}
			tenantBusinessTypes[createdTbt.Code] = createdTbt
			return createdTbt, nil
		}
	}
	return tbt, nil
}
