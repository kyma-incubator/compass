package metrics

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"
)

type Pusher struct {
	eventingRequestTotal *prometheus.CounterVec
	pusher               *push.Pusher
	instanceID           uuid.UUID
}

func NewPusher(endpoint string) *Pusher {
	eventingRequestTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: TenantFetcherSubsystem,
		Name:      "eventing_request_total",
		Help:      "Total Eventing Requests",
	}, []string{"method", "code", "desc"})

	instanceID := uuid.New()
	log.WithField(InstanceIDKeyName, instanceID).Infof("Initializing Metrics Pusher...")

	registry := prometheus.NewRegistry()
	registry.MustRegister(eventingRequestTotal)
	pusher := push.New(endpoint, TenantFetcherJobName).Grouping(InstanceIDKeyName, instanceID.String()).Gatherer(registry)

	return &Pusher{
		eventingRequestTotal: eventingRequestTotal,
		pusher:               pusher,
		instanceID:           instanceID,
	}
}

func (p *Pusher) RecordEventingRequest(method string, statusCode int, desc string) {
	log.WithField(InstanceIDKeyName, p.instanceID).Infof("Recording request with status code '%d'...", statusCode)
	p.eventingRequestTotal.WithLabelValues(method, strconv.Itoa(statusCode), desc).Inc()
}

func (p *Pusher) Push() {
	log.WithField(InstanceIDKeyName, p.instanceID).Info("Pushing metrics...")
	err := p.pusher.Add()
	if err != nil {
		wrappedErr := errors.Wrap(err, "while pushing metrics to Pushgateway")
		log.WithField(InstanceIDKeyName, p.instanceID).Error(wrappedErr)
	}
}
