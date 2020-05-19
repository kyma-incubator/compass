package metrics

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"
)

type Pusher struct {
	eventingRequestTotal *prometheus.CounterVec
	pusher               *push.Pusher
}

func NewPusher(endpoint string) *Pusher {
	eventingRequestTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: TenantFetcherSubsystem,
		Name:      "eventing_request_total",
		Help:      "Total Eventing Requests",
	}, []string{"code", "method"})

	pusher := push.New(endpoint, TenantFetcherJobName).Collector(eventingRequestTotal)

	return &Pusher{
		eventingRequestTotal: eventingRequestTotal,
		pusher:               pusher,
	}
}

func (p *Pusher) RecordEventingRequest(method string, res *http.Response) {
	p.eventingRequestTotal.With(prometheus.Labels{
		"method": method,
		"code":   string(res.StatusCode),
	}).Inc()
}

func (p *Pusher) Push() {
	err := p.pusher.Push()
	if err != nil {
		wrappedErr := errors.Wrap(err, "while pushing metrics to Pushgateway")
		log.Error(wrappedErr)
	}
}
