package metrics

import (
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
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

func (r *MetricsReporter) ReportFailedSync(err error) {
	if err != nil {
		if r.pusher != nil {
			desc := tenantfetcher.GetErrorDesc(err)
			r.pusher.RecordTenantsSyncJobFailure(http.MethodGet, 0, desc)
			r.pusher.Push()
		}
	}
}
