package systemfetcher

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

const (
	// LifecycleAttributeName is the lifecycle status attribute of the application in the external source response for applications retrieval.
	LifecycleAttributeName string = "lifecycleStatus"
	// LifecycleDeleted is the string matching the deleted lifecycle state of the application in the external source.
	LifecycleDeleted string = "DELETED"

	// ConcurrentDeleteOperationErrMsg is the error message returned by the Compass Director, when we try to delete an application, which is already undergoing a delete operation.
	ConcurrentDeleteOperationErrMsg = "Concurrent operation [reason=delete operation is in progress]"
)

//go:generate mockery --name=tenantService --output=automock --outpkg=automock --case=underscore --exported=true
type tenantService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

//go:generate mockery --name=systemsService --output=automock --outpkg=automock --case=underscore --exported=true
type systemsService interface {
	CreateManyIfNotExistsWithEventualTemplate(ctx context.Context, applicationInputs []model.ApplicationRegisterInputWithTemplate) error
	GetByNameAndSystemNumber(ctx context.Context, name, systemNumber string) (*model.Application, error)
}

//go:generate mockery --name=systemsAPIClient --output=automock --outpkg=automock --case=underscore --exported=true
type systemsAPIClient interface {
	FetchSystemsForTenant(ctx context.Context, tenant string) ([]System, error)
}

//go:generate mockery --name=directorClient --output=automock --outpkg=automock --case=underscore --exported=true
type directorClient interface {
	DeleteSystemAsync(ctx context.Context, id, tenant string) error
}

//go:generate mockery --name=applicationTemplateService --output=automock --outpkg=automock --case=underscore --exported=true
type applicationTemplateService interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
}

//go:generate mockery --name=applicationConverter --output=automock --outpkg=automock --case=underscore --exported=true
type applicationConverter interface {
	CreateInputJSONToModel(ctx context.Context, in string) (model.ApplicationRegisterInput, error)
}

// Config holds the configuration available for the SystemFetcher.
type Config struct {
	SystemsQueueSize          int           `envconfig:"default=100,APP_SYSTEM_INFORMATION_QUEUE_SIZE"`
	FetcherParallellism       int           `envconfig:"default=30,APP_SYSTEM_INFORMATION_PARALLELLISM"`
	DirectorGraphqlURL        string        `envconfig:"APP_DIRECTOR_GRAPHQL_URL"`
	DirectorRequestTimeout    time.Duration `envconfig:"default=30s,APP_DIRECTOR_REQUEST_TIMEOUT"`
	DirectorSkipSSLValidation bool          `envconfig:"default=false,APP_DIRECTOR_SKIP_SSL_VALIDATION"`

	EnableSystemDeletion bool `envconfig:"default=true,APP_ENABLE_SYSTEM_DELETION"`
}

// SystemFetcher is responsible for synchronizing the existing applications in Compass and a pre-defined external source.
type SystemFetcher struct {
	transaction        persistence.Transactioner
	tenantService      tenantService
	systemsService     systemsService
	appTemplateService applicationTemplateService
	appConverter       applicationConverter
	systemsAPIClient   systemsAPIClient
	directorClient     directorClient

	config  Config
	workers chan struct{}
}

// NewSystemFetcher returns a new SystemFetcher.
func NewSystemFetcher(tx persistence.Transactioner, ts tenantService, ss systemsService, ats applicationTemplateService, ac applicationConverter, sac systemsAPIClient, directorClient directorClient, config Config) *SystemFetcher {
	return &SystemFetcher{
		transaction:        tx,
		tenantService:      ts,
		systemsService:     ss,
		appTemplateService: ats,
		appConverter:       ac,
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

// SyncSystems synchronizes applications between Compass and external source. It deletes the applications with deleted state in the external source from Compass,
// and creates any new applications present in the external source.
func (s *SystemFetcher) SyncSystems(ctx context.Context) error {
	tenants, err := s.listTenants(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list tenants")
	}

	systemsQueue := make(chan tenantSystems, s.config.SystemsQueueSize)
	wgDB := sync.WaitGroup{}
	wgDB.Add(1)
	go func() {
		defer func() {
			wgDB.Done()
		}()
		for tenantSystems := range systemsQueue {
			err = s.processSystemsForTenant(ctx, tenantSystems.tenant, tenantSystems.systems)
			if err != nil {
				log.C(ctx).Error(errors.Wrap(err, fmt.Sprintf("failed to save systems for tenant %s", tenantSystems.tenant.ExternalTenant)))
				continue
			}

			log.C(ctx).Info(fmt.Sprintf("Successfully synced systems for tenant %s", tenantSystems.tenant.ExternalTenant))
		}
	}()

	wg := sync.WaitGroup{}
	for _, t := range tenants {
		wg.Add(1)
		s.workers <- struct{}{}
		go func(t *model.BusinessTenantMapping) {
			defer func() {
				wg.Done()
				<-s.workers
			}()
			systems, err := s.systemsAPIClient.FetchSystemsForTenant(ctx, t.ExternalTenant)
			if err != nil {
				log.C(ctx).Error(errors.Wrap(err, fmt.Sprintf("failed to fetch systems for tenant %s", t.ExternalTenant)))
				return
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
	close(systemsQueue)
	wgDB.Wait()

	return nil
}

func (s *SystemFetcher) listTenants(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	tx, err := s.transaction.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer s.transaction.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tenants, err := s.tenantService.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve tenants")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit while retrieving tenants")
	}

	return tenants, nil
}

func (s *SystemFetcher) processSystemsForTenant(ctx context.Context, tenantMapping *model.BusinessTenantMapping, systems []System) error {
	log.C(ctx).Infof("Saving %d systems for tenant %s", len(systems), tenantMapping.Name)
	appInputs := make([]model.ApplicationRegisterInputWithTemplate, 0, len(systems))

	tx, err := s.transaction.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer s.transaction.RollbackUnlessCommitted(ctx, tx)

	ctx = tenant.SaveToContext(ctx, tenantMapping.ID, tenantMapping.ExternalTenant)
	ctx = persistence.SaveToContext(ctx, tx)

	for _, system := range systems {
		if system.AdditionalAttributes[LifecycleAttributeName] == LifecycleDeleted && s.config.EnableSystemDeletion {
			log.C(ctx).Infof("Getting system by name %s and system number %s", system.DisplayName, system.SystemNumber)
			app, err := s.systemsService.GetByNameAndSystemNumber(ctx, system.DisplayName, system.SystemNumber)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("Could not get system with name %s and system number %s", system.DisplayName, system.SystemNumber)
				continue
			}
			if !app.Ready && !app.GetDeletedAt().IsZero() {
				log.C(ctx).Infof("System with id %s is currently being deleted", app.ID)
				continue
			}
			if err := s.directorClient.DeleteSystemAsync(ctx, app.ID, tenantMapping.ID); err != nil {
				if strings.Contains(err.Error(), ConcurrentDeleteOperationErrMsg) {
					log.C(ctx).Warnf("Delete operation is in progress for system with id %s", app.ID)
				} else {
					log.C(ctx).WithError(err).Errorf("Could not delete system with id %s", app.ID)
				}
				continue
			}
			log.C(ctx).Infof("Started asynchronously delete for system with id %s", app.ID)
			continue
		}

		appInput, err := s.convertSystemToAppRegisterInput(ctx, system)
		if err != nil {
			return err
		}
		appInputs = append(appInputs, *appInput)
	}

	if len(appInputs) > 0 {
		err = s.systemsService.CreateManyIfNotExistsWithEventualTemplate(ctx, appInputs)
		if err != nil {
			return errors.Wrap(err, "failed to create applications")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to commit applications for tenant %s", tenantMapping.ExternalTenant))
	}

	return nil
}

func (s *SystemFetcher) convertSystemToAppRegisterInput(ctx context.Context, sc System) (*model.ApplicationRegisterInputWithTemplate, error) {
	if len(sc.TemplateID) > 0 {
		return s.appRegisterInputFromTemplate(ctx, sc)
	}

	appRegisterInput := model.ApplicationRegisterInput{
		Name:         sc.DisplayName,
		Description:  str.Ptr(sc.ProductDescription),
		SystemNumber: str.Ptr(sc.SystemNumber),
		BaseURL:      str.Ptr(sc.BaseURL),
		ProviderName: str.Ptr(sc.InfrastructureProvider),
	}

	return enrichAppRegisterInput(appRegisterInput, sc), nil
}

func (s *SystemFetcher) appRegisterInputFromTemplate(ctx context.Context, sc System) (*model.ApplicationRegisterInputWithTemplate, error) {
	appTemplate, err := s.appTemplateService.Get(ctx, sc.TemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application template with ID %s", sc.TemplateID)
	}

	inputValues := getTemplateInputs(appTemplate, sc)
	appRegisterInputJSON, err := s.appTemplateService.PrepareApplicationCreateInputJSON(appTemplate, inputValues)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing ApplicationRegisterInput JSON from Application Template with name %s", appTemplate.Name)
	}

	appRegisterInput, err := s.appConverter.CreateInputJSONToModel(ctx, appRegisterInputJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing ApplicationRegisterInput model from Application Template with name %s", appTemplate.Name)
	}

	return enrichAppRegisterInput(appRegisterInput, sc), nil
}

func enrichAppRegisterInput(input model.ApplicationRegisterInput, sc System) *model.ApplicationRegisterInputWithTemplate {
	initStatusCond := model.ApplicationStatusConditionInitial
	input.StatusCondition = &initStatusCond

	if len(input.Labels) == 0 {
		input.Labels = make(map[string]interface{}, 1)
	}
	input.Labels["managed"] = "true"

	return &model.ApplicationRegisterInputWithTemplate{
		ApplicationRegisterInput: input,
		TemplateID:               sc.TemplateID,
	}
}

func getTemplateInputs(t *model.ApplicationTemplate, s System) model.ApplicationFromTemplateInputValues {
	inputValues := model.ApplicationFromTemplateInputValues{}
	for _, p := range t.Placeholders {
		inputValues = append(inputValues, &model.ApplicationTemplateValueInput{
			Placeholder: p.Name,
			Value:       getPlaceholderValue(p, s),
		})
	}

	return inputValues
}

func getPlaceholderValue(placeholder model.ApplicationTemplatePlaceholder, system System) string {
	if placeholder.AppRegisterInputPlaceholderValue == nil {
		return ""
	}

	switch *placeholder.AppRegisterInputPlaceholderValue {
	case model.ApplicationRegisterInputPlaceholderValueName:
		return system.DisplayName
	case model.ApplicationRegisterInputPlaceholderValueDescription:
		return system.ProductDescription
	case model.ApplicationRegisterInputPlaceholderValueBaseURL:
		return system.BaseURL
	case model.ApplicationRegisterInputPlaceholderValueProviderName:
		return system.InfrastructureProvider
	case model.ApplicationRegisterInputPlaceholderValueSystemNumber:
		return system.SystemNumber
	}

	return ""
}
