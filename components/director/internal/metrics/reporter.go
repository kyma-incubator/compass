package metrics

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// MetricsReporter for tenant fetcher execution metrics
type MetricsReporter struct {
	pusher tenantfetcher.MetricsPusher
}

// NewMetricsReporter for tenant fetcher execution metrics
func NewMetricsReporter(pusher tenantfetcher.MetricsPusher) MetricsReporter {
	return MetricsReporter{
		pusher: pusher,
	}
}

// ReportFailedSync reports failed tenant fetcher job
func (r *MetricsReporter) ReportFailedSync(err error, ctx context.Context) {
	log.C(ctx).WithError(err).Errorf("Reporting failed job sync with error: %v", err)
	if r.pusher == nil {
		log.C(ctx).Error("Failed to report job sync failure: metrics pusher is not configured")
		return
	}

	desc := tenantfetcher.GetErrorDesc(err)
	r.pusher.RecordTenantsSyncJobFailure(http.MethodGet, 0, desc)
	r.pusher.Push()
}
