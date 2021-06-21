package systemfetcher

import (
	"context"
	"fmt"
	"sync"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore
type TenantService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

//go:generate mockery --name=SystemsService --output=automock --outpkg=automock --case=underscore
type SystemsService interface {
	CreateManyIfNotExistsWithEventualTemplate(ctx context.Context, applicationInputs []model.ApplicationRegisterInput, systemToTemplateMapping []string) error
}

//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore
type ApplicationTemplateService interface {
	GetByName(ctx context.Context, name string) (*model.ApplicationTemplate, error)
}

//go:generate mockery --name=SystemsAPIClient --output=automock --outpkg=automock --case=underscore
type SystemsAPIClient interface {
	FetchSystemsForTenant(ctx context.Context, tenant string) ([]System, error)
}

type SystemFetcher struct {
	transaction        persistence.Transactioner
	tenantService      TenantService
	systemsService     SystemsService
	systemsAPIClient   SystemsAPIClient
	appTemplateService ApplicationTemplateService

	workers chan struct{}
}

func NewSystemFetcher(tx persistence.Transactioner, ts TenantService, ss SystemsService, sac SystemsAPIClient, appTemplateService ApplicationTemplateService, fetcherParallellism int) *SystemFetcher {
	return &SystemFetcher{
		transaction:        tx,
		tenantService:      ts,
		systemsService:     ss,
		systemsAPIClient:   sac,
		appTemplateService: appTemplateService,
		workers:            make(chan struct{}, fetcherParallellism),
	}
}

type TenantSystems struct {
	tenant  *model.BusinessTenantMapping
	systems []System
}

func (s *SystemFetcher) SyncSystems(ctx context.Context) error {
	tenants, err := s.listTenants(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list tenants")
	}

	systemsQueue := make(chan TenantSystems, 100)
	wgDB := sync.WaitGroup{}
	wgDB.Add(1)
	go func() {
		defer func() {
			wgDB.Done()
		}()
		for tenantSystems := range systemsQueue {
			err = s.saveSystemsForTenant(ctx, tenantSystems.tenant, tenantSystems.systems)
			if err != nil {
				log.C(ctx).Error(errors.Wrap(err, fmt.Sprintf("failed to save systems for tenant %s", tenantSystems.tenant.ExternalTenant)))
				continue
			}

			log.C(ctx).Info(fmt.Sprintf("Successfully synced systems for tenant %s", tenantSystems.tenant.ExternalTenant))
		}
	}()

	wg := sync.WaitGroup{}
	for number, t := range tenants {
		wg.Add(1)
		s.workers <- struct{}{}
		go func(t *model.BusinessTenantMapping, number int) {
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
				systemsQueue <- TenantSystems{
					tenant:  t,
					systems: systems,
				}
			}
		}(t, number)
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

func (s *SystemFetcher) saveSystemsForTenant(ctx context.Context, tenantMapping *model.BusinessTenantMapping, systems []System) error {
	var appInputs []model.ApplicationRegisterInput
	var templateTypes []string

	tx, err := s.transaction.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer s.transaction.RollbackUnlessCommitted(ctx, tx)

	ctx = tenant.SaveToContext(ctx, tenantMapping.ID, tenantMapping.ExternalTenant)
	ctx = persistence.SaveToContext(ctx, tx)

	for _, system := range systems {
		appInput, templateType := s.convertSystemToAppRegisterInput(system)
		template, err := s.appTemplateService.GetByName(ctx, templateType)
		if err != nil && !apperrors.IsNotFoundError(err) {
			return err
		}
		templateTypes = append(templateTypes, template.ID)
		appInputs = append(appInputs, appInput)
	}

	err = s.systemsService.CreateManyIfNotExistsWithEventualTemplate(ctx, appInputs, templateTypes)
	if err != nil {
		return errors.Wrap(err, "failed to create applications")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to commit applications for tenant %s", tenantMapping.ExternalTenant))
	}

	return nil
}

func (s *SystemFetcher) convertSystemToAppRegisterInput(sc System) (model.ApplicationRegisterInput, string) {
	initStatusCond := model.ApplicationStatusConditionManaged
	baseURL := sc.BaseURL

	appRegisterInput := model.ApplicationRegisterInput{
		Name:                sc.DisplayName,
		Description:         &sc.ProductDescription,
		BaseURL:             &baseURL,
		ProviderName:        &sc.InfrastructureProvider,
		StatusCondition:     &initStatusCond,
		SystemNumber:        &sc.SystemNumber,
		Labels:              nil,
		HealthCheckURL:      nil,
		IntegrationSystemID: nil,
		OrdLabels:           nil,
		Bundles:             nil,
		Webhooks:            nil,
	}

	return appRegisterInput, sc.TemplateType
}
