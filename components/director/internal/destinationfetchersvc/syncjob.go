package destinationfetchersvc

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"golang.org/x/sync/semaphore"
)

// DestinationSyncer missing godoc
//
//go:generate mockery --name=DestinationSyncer --output=automock --outpkg=automock --case=underscore --disable-version-string
type DestinationSyncer interface {
	SyncTenantDestinations(ctx context.Context, tenantID string) error
	GetSubscribedTenantIDs(ctx context.Context) ([]string, error)
}

// SyncJobConfig configuration for destination sync job
type SyncJobConfig struct {
	ElectionCfg       cronjob.ElectionConfig
	JobSchedulePeriod time.Duration
	TenantSyncTimeout time.Duration
	ParallelTenants   int
}

func syncSubscribedTenantsDestinations(ctx context.Context, subscribedTenants []string, cfg SyncJobConfig,
	destinationSyncer DestinationSyncer) int {
	parallelTenantsSemaphore := semaphore.NewWeighted(int64(cfg.ParallelTenants))
	wg := sync.WaitGroup{}
	var syncedTenants uint32
	for _, tenantID := range subscribedTenants {
		wg.Add(1)
		go func(tenantID string) {
			defer wg.Done()
			if err := parallelTenantsSemaphore.Acquire(ctx, 1); err != nil {
				log.C(ctx).WithError(err).Errorf("Could not acquire semaphore")
				return
			}
			defer parallelTenantsSemaphore.Release(1)
			err := syncTenantDestinationsWithTimeout(ctx, destinationSyncer, tenantID, cfg.TenantSyncTimeout)
			if err != nil {
				log.C(ctx).WithError(err).Error()
				return
			}
			atomic.AddUint32(&syncedTenants, 1)
			currentlySynced := int(atomic.LoadUint32(&syncedTenants))
			// Log on each ParallelTenants synced to track progress
			if currentlySynced%cfg.ParallelTenants == 0 {
				log.C(ctx).Infof("%d/%d tenants have been synced", currentlySynced, len(subscribedTenants))
			}
		}(tenantID)
	}
	wg.Wait()
	return int(syncedTenants)
}

// StartDestinationFetcherSyncJob starts destination sync job and blocks
func StartDestinationFetcherSyncJob(ctx context.Context, cfg SyncJobConfig, destinationSyncer DestinationSyncer) error {
	resyncJob := cronjob.CronJob{
		Name: "DestinationFetcherSync",
		Fn: func(jobCtx context.Context) {
			subscribedTenants, err := destinationSyncer.GetSubscribedTenantIDs(jobCtx)
			if err != nil {
				log.C(jobCtx).WithError(err).Errorf("Could not fetch subscribed tenants for destination resync")
				return
			}
			if len(subscribedTenants) == 0 {
				log.C(jobCtx).Info("No subscribed tenants found. Skipping destination sync job")
				return
			}
			log.C(jobCtx).Infof("Found %d subscribed tenants. Starting destination sync...", len(subscribedTenants))
			syncedTenantsCount := syncSubscribedTenantsDestinations(jobCtx, subscribedTenants, cfg, destinationSyncer)
			log.C(jobCtx).Infof("Destination sync finished with %d/%d tenants synced",
				syncedTenantsCount, len(subscribedTenants))
		},
		SchedulePeriod: cfg.JobSchedulePeriod,
	}
	return cronjob.RunCronJob(ctx, cfg.ElectionCfg, resyncJob)
}

func syncTenantDestinationsWithTimeout(
	ctx context.Context, destinationSyncer DestinationSyncer, tenantID string, timeout time.Duration) error {
	tenantSyncTimeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		log.C(ctx).Debugf("Destinations synced for tenant '%s' for %s", tenantID, duration)
		if duration > timeout/2 {
			log.C(ctx).Warnf("Destinations synchronization for tenant '%s' took %s", tenantID, duration)
		}
	}()
	return destinationSyncer.SyncTenantDestinations(tenantSyncTimeoutCtx, tenantID)
}
