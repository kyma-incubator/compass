package destinationfetchersvc_test

import (
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestDestinationSyncJob(t *testing.T) {
	const testTimeout = time.Second * 5

	var (
		tenantsToResync = []string{"t1", "t2", "t3"}

		cfg = destinationfetchersvc.SyncJobConfig{
			ParallelTenants:   2,
			JobSchedulePeriod: time.Minute,
			ElectionCfg: cronjob.ElectionConfig{
				ElectionEnabled: false,
			},
		}
		expectedError = errors.New("expected")

		cancelCtxAfterAllDoneReceived = func(done <-chan struct{}, doneCount int, cancel context.CancelFunc) {
			go func() {
				defer cancel()
				for i := 0; i < doneCount; i++ {
					select {
					case <-done:
					case <-time.After(testTimeout):
						t.Errorf("Test timed out - not all tenants re-synced")
						return
					}
				}
			}()
		}
	)

	t.Run("Should re-sync all tenants", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{}, len(tenantsToResync))
		addDone := func(args mock.Arguments) {
			done <- struct{}{}
		}
		cancelCtxAfterAllDoneReceived(done, len(tenantsToResync), cancel)

		destinationSyncer := &automock.DestinationSyncer{}
		destinationSyncer.Mock.On("SyncTenantDestinations",
			mock.Anything, tenantsToResync[0]).Return(nil).Run(addDone)
		destinationSyncer.Mock.On("SyncTenantDestinations",
			mock.Anything, tenantsToResync[1]).Return(nil).Run(addDone)
		destinationSyncer.Mock.On("SyncTenantDestinations",
			mock.Anything, tenantsToResync[2]).Return(nil).Run(addDone)
		destinationSyncer.Mock.On("GetSubscribedTenantIDs", mock.Anything).
			Return(tenantsToResync, nil)

		err := destinationfetchersvc.StartDestinationFetcherSyncJob(ctx, cfg, destinationSyncer)
		assert.Nil(t, err)
		destinationSyncer.Mock.AssertNumberOfCalls(t, "GetSubscribedTenantIDs", 1)
		destinationSyncer.Mock.AssertNumberOfCalls(t, "SyncTenantDestinations", len(tenantsToResync))
		destinationSyncer.AssertExpectations(t)
	})

	t.Run("Should not fail on one tenant re-sync failure", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{}, len(tenantsToResync))
		addDone := func(args mock.Arguments) {
			done <- struct{}{}
		}
		cancelCtxAfterAllDoneReceived(done, len(tenantsToResync), cancel)

		destinationSyncer := &automock.DestinationSyncer{}
		destinationSyncer.Mock.On("SyncTenantDestinations",
			mock.Anything, tenantsToResync[0]).Return(nil).Run(addDone)
		destinationSyncer.Mock.On("SyncTenantDestinations",
			mock.Anything, tenantsToResync[1]).Return(expectedError).Run(addDone)
		destinationSyncer.Mock.On("SyncTenantDestinations",
			mock.Anything, tenantsToResync[2]).Return(nil).Run(addDone)
		destinationSyncer.Mock.On("GetSubscribedTenantIDs", mock.Anything).
			Return(tenantsToResync, nil)

		err := destinationfetchersvc.StartDestinationFetcherSyncJob(ctx, cfg, destinationSyncer)
		assert.Nil(t, err)
		destinationSyncer.Mock.AssertNumberOfCalls(t, "GetSubscribedTenantIDs", 1)
		destinationSyncer.Mock.AssertNumberOfCalls(t, "SyncTenantDestinations", len(tenantsToResync))
		destinationSyncer.AssertExpectations(t)
	})

	t.Run("Should not re-sync if subscribed tenants could not be fetched", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		destinationSyncer := &automock.DestinationSyncer{}
		destinationSyncer.Mock.On("GetSubscribedTenantIDs", mock.Anything).
			Return(nil, expectedError).Run(func(args mock.Arguments) { cancel() })

		err := destinationfetchersvc.StartDestinationFetcherSyncJob(ctx, cfg, destinationSyncer)
		assert.Nil(t, err)
		destinationSyncer.Mock.AssertNumberOfCalls(t, "GetSubscribedTenantIDs", 1)
		destinationSyncer.AssertExpectations(t)
	})

	t.Run("Should not re-sync if there are no subscribed tenants", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		destinationSyncer := &automock.DestinationSyncer{}
		destinationSyncer.Mock.On("GetSubscribedTenantIDs", mock.Anything).
			Return(nil, nil).Run(func(args mock.Arguments) { cancel() })

		err := destinationfetchersvc.StartDestinationFetcherSyncJob(ctx, cfg, destinationSyncer)
		assert.Nil(t, err)
		destinationSyncer.Mock.AssertNumberOfCalls(t, "GetSubscribedTenantIDs", 1)
		destinationSyncer.AssertExpectations(t)
	})
}
