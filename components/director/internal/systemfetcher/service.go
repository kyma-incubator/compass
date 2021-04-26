package systemfetcher

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
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
	CreateManyIfNotExists(ctx context.Context, applicationInputs []model.ApplicationRegisterInput) error
}

//go:generate mockery --name=SystemsAPIClient --output=automock --outpkg=automock --case=underscore
type SystemsAPIClient interface {
	FetchSystemsForTenant(ctx context.Context, tenant string) []ProductInstanceExtended
}

type SystemFetcher struct {
	transaction      persistence.Transactioner
	tenantService    TenantService
	systemsService   SystemsService
	systemsAPIClient SystemsAPIClient
}

func NewSystemFetcher(tx persistence.Transactioner, ts TenantService, ss SystemsService, sac SystemsAPIClient) *SystemFetcher {
	return &SystemFetcher{
		transaction:      tx,
		tenantService:    ts,
		systemsService:   ss,
		systemsAPIClient: sac,
	}
}

func (s *SystemFetcher) SyncSystems(ctx context.Context) error {
	//TODO: Open transact here instead? So that all DB calls are in one transaction - avoid phantom DB stuff, but there's a problem that we have HTTP calls inside of the DB call

	tx, err := s.transaction.Begin()
	ctx = persistence.SaveToContext(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	tenants, err := s.tenantService.List(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve tenants")
	}

	s.transaction.RollbackUnlessCommitted(ctx, tx)

	for _, t := range tenants {
		systems := s.systemsAPIClient.FetchSystemsForTenant(ctx, t.ExternalTenant)

		var appInputs []model.ApplicationRegisterInput
		for _, s := range systems {
			sc := s

			initStatusCond := model.ApplicationStatusConditionInitial
			baseURL := sc.BaseURL
			if mainURL, ok := sc.AdditionalUrls["mainUrl"]; ok {
				baseURL = mainURL
			}

			a := model.ApplicationRegisterInput{
				Name:                sc.DisplayName,
				Description:         &sc.ProductDescription,
				BaseURL:             &baseURL,
				ProviderName:        &sc.InfrastructureProvider,
				StatusCondition:     &initStatusCond,
				Labels:              nil,
				HealthCheckURL:      nil,
				IntegrationSystemID: nil,
				OrdLabels:           nil,
				Bundles:             nil,
				Webhooks:            nil,
			}

			appInputs = append(appInputs, a)
		}

		tx, err := s.transaction.Begin()
		if err != nil {
			return errors.Wrap(err, "failed to begin transaction")
		}

		ctx := tenant.SaveToContext(ctx, t.ID, t.ExternalTenant)
		ctx = persistence.SaveToContext(ctx, tx)
		err = s.systemsService.CreateManyIfNotExists(ctx, appInputs)
		if err != nil {
			return errors.Wrap(err, "failed to create applications")
		}

		s.transaction.RollbackUnlessCommitted(ctx, tx)
	}
	return nil
}
