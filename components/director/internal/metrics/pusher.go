package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type Pusher struct {
	eventingRequestTotal *prometheus.GaugeVec
	pusher               *push.Pusher
	instanceID           uuid.UUID
}

func NewPusher(endpoint string, timeout time.Duration) *Pusher {
	eventingRequestTotal := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: TenantFetcherSubsystem,
		Name:      "eventing_requests_total",
		Help:      "Total Eventing Requests",
	}, []string{"method", "code", "desc"})

	instanceID := uuid.New()
	log.D().WithField(InstanceIDKeyName, instanceID).Infof("Initializing Metrics Pusher...")

	registry := prometheus.NewRegistry()
	registry.MustRegister(eventingRequestTotal)
	pusher := push.New(endpoint, TenantFetcherJobName).Gatherer(registry).Client(&http.Client{
		Timeout: timeout,
	})

	return &Pusher{
		eventingRequestTotal: eventingRequestTotal,
		pusher:               pusher,
		instanceID:           instanceID,
	}
}

func (p *Pusher) RecordEventingRequest(method string, statusCode int, desc string) {
	log.D().WithFields(logrus.Fields{
		InstanceIDKeyName: p.instanceID,
		"statusCode":      statusCode,
		"desc":            desc,
	}).Infof("Recording eventing request...")
	p.eventingRequestTotal.WithLabelValues(method, strconv.Itoa(statusCode), desc).Inc()
}

func (p *Pusher) Push() {
	log.D().WithField(InstanceIDKeyName, p.instanceID).Info("Pushing metrics...")
	err := p.pusher.Add()
	if err != nil {
		wrappedErr := errors.Wrap(err, "while pushing metrics to Pushgateway")
		log.D().WithField(InstanceIDKeyName, p.instanceID).Error(wrappedErr)
	}
}
