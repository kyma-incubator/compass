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

	interruptSubaccountJob := make(chan bool, 5)
	runTenantFetcherJob2(ctx, 5, true, interruptSubaccountJob)
	interruptGlobalAccountJob := make(chan bool, 5)
	runTenantFetcherJob2(ctx, 10, false, interruptGlobalAccountJob)

	time.Sleep(20 * time.Second)
	cancel()

	<-interruptSubaccountJob
	<-interruptGlobalAccountJob
}
func runTenantFetcherJob2(ctx context.Context, jobInterval int, syncSubaccount bool, interrupt chan bool) {
	tenantFetcherJobTicker := time.NewTicker(time.Duration(jobInterval) * time.Second)

	go func() {
		for {
			select {
			case <-tenantFetcherJobTicker.C:
				log.C(ctx).Infof("Scheduled %s tenant fetcher job will be executed, job interval is %d minutes", tenantType2(syncSubaccount), jobInterval)
			case <-ctx.Done():
				log.C(ctx).Errorf("Context is canceled and %s tenant fetcher job will be interrupted",
					tenantType2(syncSubaccount))
				stopTenantFetcherJobTicker2(ctx, syncSubaccount, tenantFetcherJobTicker)
				interrupt <- true
				return
			}
		}
	}()
}

func tenantType2(syncSubaccount bool) string {
	tenantType := "subaccount"
	if !syncSubaccount {
		tenantType = "global account"
	}
	return tenantType
}

func stopTenantFetcherJobTicker2(ctx context.Context, syncSubaccount bool, tenantFetcherJobTicker *time.Ticker) {
	tenantFetcherJobTicker.Stop()
	log.C(ctx).Infof("Tenant fetcher job ticker for %ss is stopped", tenantType2(syncSubaccount))
}
