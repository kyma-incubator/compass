package metrics

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"
)

type MetricsReporter struct {
	pusher tenantfetcher.MetricsPusher
}

// NewMetricsReporter missing godoc
func NewMetricsReporter(pusher tenantfetcher.MetricsPusher) MetricsReporter {
	return MetricsReporter{
		pusher: pusher,
	}
}

func (r *MetricsReporter) ReportFailedSync(err error, ctx context.Context) {
	log.C(ctx).WithError(err).Errorf("Report failed job sync: %v", err)
	if err != nil {
		if r.pusher != nil {
			desc := tenantfetcher.GetErrorDesc(err)
			r.pusher.RecordTenantsSyncJobFailure(http.MethodGet, 0, desc)
			r.pusher.Push()
		}
	}
}
