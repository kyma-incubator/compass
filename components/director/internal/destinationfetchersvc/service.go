package destinationfetchersvc

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

const regionLabelKey = "region"

//go:generate mockery --name=UUIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
// UUIDService missing godoc
type UUIDService interface {
	Generate() string
}

//go:generate mockery --name=DestinationRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
// DestinationRepo missing godoc
type DestinationRepo interface {
	Upsert(ctx context.Context, in model.DestinationInput, id, tenantID, bundleID, revision string) error
	DeleteOld(ctx context.Context, latestRevision, tenantID string) error
}

//go:generate mockery --name=LabelRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
// LabelRepo missing godoc
type LabelRepo interface {
	GetSubdomainLabelForSubscribedRuntime(ctx context.Context, tenantID string) (*model.Label, error)
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

//go:generate mockery --name=BundleRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
// BundleRepo missing godoc
type BundleRepo interface {
	ListByDestination(ctx context.Context, tenantID string, destination model.DestinationInput) ([]*model.Bundle, error)
}

//go:generate mockery --name=TenantRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
// TenantRepo missing godoc
type TenantRepo interface {
	ListBySubscribedRuntimes(ctx context.Context) ([]*model.BusinessTenantMapping, error)
}

// DestinationService missing godoc
type DestinationService struct {
	Transactioner      persistence.Transactioner
	UUIDSvc            UUIDService
	DestinationRepo    DestinationRepo
	BundleRepo         BundleRepo
	LabelRepo          LabelRepo
	DestinationsConfig config.DestinationsConfig
	APIConfig          DestinationServiceAPIConfig
	TenantRepo         TenantRepo
}

// GetSubscribedTenantIDs returns subscribed tenants
func (d *DestinationService) GetSubscribedTenantIDs(ctx context.Context) ([]string, error) {
	tenants, err := d.getSubscribedTenants(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to get subscribed tenants")
		return nil, err
	}
	tenantIDs := make([]string, 0, len(tenants))
	for _, tenant := range tenants {
		tenantIDs = append(tenantIDs, tenant.ID)
	}
	return tenantIDs, nil
}

func (d *DestinationService) getSubscribedTenants(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	var tenants []*model.BusinessTenantMapping
	transactionError := d.transaction(ctx, func(ctxWithTransact context.Context) error {
		var err error
		tenants, err = d.TenantRepo.ListBySubscribedRuntimes(ctxWithTransact)
		if err != nil {
			log.C(ctxWithTransact).WithError(err).Error("An error occurred while getting subscribed tenants")
			return err
		}
		return nil
	})

	return tenants, transactionError
}

func (d *DestinationService) generateClientBySubdomainLabel(ctx context.Context, subdomainLabel *model.Label) (*Client, error) {
	regionLabel, err := d.getRegionLabel(ctx, *subdomainLabel.Tenant)
	if err != nil {
		return nil, err
	}

	subdomain, ok := subdomainLabel.Value.(string)
	if !ok {
		log.C(ctx).Errorf("cannot cast label value as a string")
		return nil, errors.New("cannot cast label value as a string")
	}

	region, ok := regionLabel.Value.(string)
	if !ok {
		log.C(ctx).Errorf("cannot cast label value as a string")
		return nil, errors.New("cannot cast label value as a string")
	}

	instanceConfig, ok := d.DestinationsConfig.RegionToInstanceConfig[region]
	if !ok {
		log.C(ctx).Errorf("No destination instance credentials found for region '%s'", region)
		return nil, errors.New(fmt.Sprintf("No destination instance credentials found for region '%s'", region))
	}

	client, err := NewClient(instanceConfig, d.APIConfig, subdomain)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to create Destination API client")
		return nil, err
	}

	return client, nil
}

// SyncTenantDestinations syncs destinations for a given tenant
func (d *DestinationService) SyncTenantDestinations(ctx context.Context, tenantID string) error {
	log.C(ctx).Infof("Starting sync of destinations for tenant '%s'", tenantID)

	subdomainLabel, err := d.getSubscribedSubdomainLabel(ctx, tenantID)
	if err != nil {
		return err
	}

	client, err := d.generateClientBySubdomainLabel(ctx, subdomainLabel)
	if err != nil {
		return err
	}

	revision := d.UUIDSvc.Generate()
	err = d.walkthroughPages(ctx, client, tenantID, func(destinations []destinationFromService) error {
		return d.mapDestinationsToTenant(ctx, tenantID, revision, destinations)
	})
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to sync destinations for tenant '%s'", tenantID)
		return err
	}

	if err := d.deleteMissingDestinations(ctx, revision, *subdomainLabel.Tenant); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to delete missing destinations for tenant '%s'", tenantID)
		return err
	}
	log.C(ctx).Infof("Finished sync of destinations for tenant '%s'", tenantID)

	return nil
}

func (d *DestinationService) transaction(ctx context.Context, dbCalls func(ctxWithTransact context.Context) error) error {
	tx, err := d.Transactioner.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to begin db transaction")
		return err
	}
	ctx = persistence.SaveToContext(ctx, tx)
	defer d.Transactioner.RollbackUnlessCommitted(ctx, tx)

	if err := dbCalls(ctx); err != nil {
		log.C(ctx).WithError(err).Error("Failed to execute database calls")
		return err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("Failed to commit database transaction")
		return err
	}
	return nil
}

func (d *DestinationService) deleteMissingDestinations(ctx context.Context, revision, tenant string) error {
	return d.transaction(ctx, func(ctxWithTransact context.Context) error {
		if err := d.DestinationRepo.DeleteOld(ctxWithTransact, revision, tenant); err != nil {
			log.C(ctxWithTransact).WithError(err).Errorf("Failed to delete removed destinations for tenant '%s'", tenant)
		}
		return nil
	})
}

func (d *DestinationService) mapDestinationsToTenant(ctx context.Context, tenant, revision string, destinations []destinationFromService) error {
	return d.transaction(ctx, func(ctxWithTransact context.Context) error {
		for _, destinationFromService := range destinations {
			destination, err := destinationFromService.ToModel()
			if err != nil {
				// Log on info as there could be many destinations that should not be gathered
				log.C(ctxWithTransact).WithError(err).Infof("Destination '%s' from tenant with id '%s' could not be processed",
					destinationFromService.Name, tenant)
				continue
			}
			bundles, err := d.BundleRepo.ListByDestination(ctxWithTransact, tenant, destination)

			if err != nil {
				log.C(ctxWithTransact).WithError(err).Errorf(
					"Failed to fetch bundle for system '%s', url '%s', correlation id '%s', tenant id '%s'",
					destination.XSystemTenantName, destination.URL, destination.XCorrelationID, tenant)
				return err
			}

			bundlesCount := len(bundles)
			if bundlesCount == 0 {
				log.C(ctxWithTransact).Infof("No bundles found for system '%s', url '%s', correlation id '%s'",
					destination.XSystemTenantName, destination.URL, destination.XCorrelationID)
				continue
			}

			log.C(ctxWithTransact).Infof("Found %d bundles for system '%s', url '%s', correlation id '%s'",
				bundlesCount, destination.XSystemTenantName, destination.URL, destination.XCorrelationID)

			for _, bundle := range bundles {
				id := d.UUIDSvc.Generate()
				if err := d.DestinationRepo.Upsert(ctxWithTransact, destination, id, tenant, bundle.ID, revision); err != nil {
					log.C(ctxWithTransact).WithError(err).Errorf(
						"Failed to insert destination with name '%s' for bundle '%s' and tenant '%s' to DB",
						destination.Name, bundle.ID, tenant)
					return err
				}
			}
		}
		return nil
	})
}

type processFunc func([]destinationFromService) error

func (d *DestinationService) walkthroughPages(
	ctx context.Context, client *Client, tenantID string, process processFunc) error {
	hasMorePages := true
	logPageCount := sync.Once{}
	for page := 1; hasMorePages; page++ {
		pageString := strconv.Itoa(page)
		resp, err := client.FetchTenantDestinationsPage(ctx, pageString)
		if err != nil {
			return errors.Wrap(err, "failed to fetch destinations page")
		}

		if err := process(resp.destinations); err != nil {
			return errors.Wrap(err, "failed to process destinations page")
		}

		hasMorePages = pageString != resp.pageCount
		logPageCount.Do(func() {
			log.C(ctx).Infof("Found %s pages of destinations in tenant '%s'", resp.pageCount, tenantID)
		})
	}
	return nil
}

// FetchDestinationsSensitiveData returns sensitive data of destinations for a given tenant
func (d *DestinationService) FetchDestinationsSensitiveData(ctx context.Context, tenantID string, destinationNames []string) ([]byte, error) {
	subdomainLabel, err := d.getSubscribedSubdomainLabel(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	client, err := d.generateClientBySubdomainLabel(ctx, subdomainLabel)
	if err != nil {
		return nil, err
	}

	nameCount := len(destinationNames)
	results := make([][]byte, nameCount)
	weighted := semaphore.NewWeighted(d.APIConfig.GoroutineLimit)
	resChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		for _, destination := range destinationNames {
			if err := weighted.Acquire(ctx, 1); err != nil {
				log.C(ctx).WithError(err).
					Errorf("Could not acquire semaphore. Destination %s will not be fetched", destination)
				return
			}
			go fetchDestination(ctx, destination, weighted, client, resChan, errChan)
		}
	}()

	for i := 0; i < nameCount; {
		select {
		case err := <-errChan:
			return nil, err
		case res := <-resChan:
			results[i] = res
			i++
		}
	}

	combinedInfoJSON := bytes.Join(results, []byte(","))
	combinedInfoJSON = append(combinedInfoJSON, '}', '}')

	return append([]byte(`{ "destinations": {`), combinedInfoJSON...), nil
}

func fetchDestination(ctx context.Context, destinationName string, weighted *semaphore.Weighted,
	client *Client, resChan chan []byte, errChan chan error) {
	log.C(ctx).Infof("Fetching data for destination: %s \n", destinationName)
	defer weighted.Release(1)
	result, err := client.FetchDestinationSensitiveData(ctx, destinationName)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to fetch data for destination %s", destinationName)
		errChan <- err
		return
	}

	result = append([]byte("\""+destinationName+"\":"), result...)

	resChan <- result
}

func (d *DestinationService) getSubscribedSubdomainLabel(ctx context.Context, tenantID string) (*model.Label, error) {
	var label *model.Label
	transactionErr := d.transaction(ctx, func(ctxWithTransact context.Context) error {
		var err error
		label, err = d.LabelRepo.GetSubdomainLabelForSubscribedRuntime(ctxWithTransact, tenantID)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				log.C(ctxWithTransact).Errorf("No subscribed subdomain found for tenant '%s'", tenantID)
				return apperrors.NewNotFoundErrorWithMessage(resource.Label, tenantID, fmt.Sprintf("tenant %s not found", tenantID))
			}
			log.C(ctxWithTransact).WithError(err).Errorf("Failed to get subdomain for tenant '%s' from db", tenantID)
			return err
		}
		return nil
	})

	return label, transactionErr
}

func (d *DestinationService) getRegionLabel(ctx context.Context, tenantID string) (*model.Label, error) {
	var region *model.Label

	transactionErr := d.transaction(ctx, func(ctxWithTransact context.Context) error {
		var err error
		region, err = d.LabelRepo.GetByKey(ctxWithTransact, tenantID, model.TenantLabelableObject, tenantID, regionLabelKey)
		if err != nil {
			log.C(ctxWithTransact).WithError(err).Errorf("Failed to fetch region for tenant '%s'", tenantID)
			return err
		}

		return nil
	})

	return region, transactionErr
}
