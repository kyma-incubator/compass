package destinationfetchersvc

import (
	"context"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"golang.org/x/sync/semaphore"
)

//go:generate mockery --name=DestinationSyncer --output=automock --outpkg=automock --case=underscore --disable-version-string
// DestinationSyncer missing godoc
type DestinationSyncer interface {
	SyncTenantDestinations(ctx context.Context, tenantID string) error
	GetSubscribedTenantIDs(ctx context.Context) ([]string, error)
}

// SyncJobConfig configuration for destination sync job
type SyncJobConfig struct {
	ElectionCfg       cronjob.ElectionConfig
	JobSchedulePeriod time.Duration
	TenantSyncTimeout time.Duration
	ParallelTenants   int64
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
				log.C(jobCtx).Info("No subscribed tenants found. Skipping sync job")
				return
			}
			log.C(jobCtx).Infof("Found %d subscribed tenants. Starting sync...", len(subscribedTenants))
			sem := semaphore.NewWeighted(cfg.ParallelTenants)
			wg := &sync.WaitGroup{}
			for _, tenantID := range subscribedTenants {
				wg.Add(1)
				go func(tenantID string) {
					defer wg.Done()
					if err := sem.Acquire(jobCtx, 1); err != nil {
						log.C(jobCtx).WithError(err).Errorf("Could not acquire semaphor")
						return
					}
					defer sem.Release(1)
					syncTenantDestinations(jobCtx, destinationSyncer, tenantID, cfg.TenantSyncTimeout)
				}(tenantID)
			}
			wg.Wait()
		},
		SchedulePeriod: cfg.JobSchedulePeriod,
	}
	return cronjob.RunCronJob(ctx, cfg.ElectionCfg, resyncJob)
}

func syncTenantDestinations(
	ctx context.Context, destinationSyncer DestinationSyncer, tenantID string, timeout time.Duration) {
	tenantSyncTimeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := destinationSyncer.SyncTenantDestinations(tenantSyncTimeoutCtx, tenantID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Could not resync destinations for tenant %s", tenantID)
	} else {
		log.C(ctx).WithError(err).Debugf("Resynced destinations for tenant %s", tenantID)
	}
}
