package main

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		log.C(ctx).Info("Server shutdowned")
	}()

	interrupt := make(chan bool, 1)
	runTenantFetcherJob2(ctx, 1, interrupt)

	//time.Sleep(3 * time.Second)
	//cancel()

	<-interrupt
}
func runTenantFetcherJob2(ctx context.Context, jobInterval int, interrupt chan bool) {
	tenantFetcherJobTicker := time.NewTicker(time.Duration(jobInterval) * time.Second)

	go func() {
		for {
			select {
			case <-tenantFetcherJobTicker.C:
				log.C(ctx).Infof("Scheduled sync of tenants to be executed, job interval is %d minutes", jobInterval)
			case <-ctx.Done():
				log.C(ctx).Error("Context is canceled and tenant fetcher job will be interrupted")
				stopTenantFetcherJobTicker2(ctx, tenantFetcherJobTicker)
				interrupt <- true
				return
			}
		}
	}()
}

func stopTenantFetcherJobTicker2(ctx context.Context, tenantFetcherJobTicker *time.Ticker) {
	tenantFetcherJobTicker.Stop()
	log.C(ctx).Info("Tenant fetcher job ticker is stopped")
}
