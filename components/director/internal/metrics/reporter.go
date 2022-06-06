package metrics

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// MetricsReporter for tenant fetcher execution metrics
type MetricsReporter struct {
	pusher tenantfetchersvc.MetricsPusher
}

// NewMetricsReporter for tenant fetcher execution metrics
func NewMetricsReporter(pusher tenantfetchersvc.MetricsPusher) MetricsReporter {
	return MetricsReporter{
		pusher: pusher,
	}
}

// ReportFailedSync reports failed tenant fetcher job
func (r *MetricsReporter) ReportFailedSync(err error, ctx context.Context) {
	log.C(ctx).WithError(err).Errorf("Report failed job sync: %v", err)
	if err != nil {
		if r.pusher != nil {
			desc := tenantfetchersvc.GetErrorDesc(err)
			r.pusher.RecordTenantsSyncJobFailure(http.MethodGet, 0, desc)
			r.pusher.Push()
		}
	}
}
