package metrics

import (
	"strconv"

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
	}, []string{"method", "code", "desc"})

	registry := prometheus.NewRegistry()
	registry.MustRegister(eventingRequestTotal)
	pusher := push.New(endpoint, TenantFetcherJobName).Gatherer(registry)

	return &Pusher{
		eventingRequestTotal: eventingRequestTotal,
		pusher:               pusher,
	}
}

func (p *Pusher) RecordEventingRequest(method string, statusCode int, desc string) {
	p.eventingRequestTotal.WithLabelValues(method, strconv.Itoa(statusCode), desc).Inc()
}

func (p *Pusher) Push() {
	err := p.pusher.Add()
	if err != nil {
		wrappedErr := errors.Wrap(err, "while pushing metrics to Pushgateway")
		log.Error(wrappedErr)
	}
}
