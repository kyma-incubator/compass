package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// Pusher missing godoc
type Pusher struct {
	eventingRequestTotal *prometheus.GaugeVec
	failedTenantsSyncJob *prometheus.CounterVec
	pusher               *push.Pusher
	instanceID           uuid.UUID
}

// NewPusher missing godoc
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

// RecordEventingRequest missing godoc
func (p *Pusher) RecordEventingRequest(method string, statusCode int, desc string) {
	log.D().WithFields(logrus.Fields{
		InstanceIDKeyName: p.instanceID,
		"statusCode":      statusCode,
		"desc":            desc,
	}).Infof("Recording eventing request...")
	p.eventingRequestTotal.WithLabelValues(method, strconv.Itoa(statusCode), desc).Inc()
}

// NewPusherPerJob missing godoc
func NewPusherPerJob(jobName string, endpoint string, timeout time.Duration) *Pusher {
	failedTenantsSyncJob := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: TenantFetcherSubsystem,
		Name:      strings.ReplaceAll(strings.ToLower(jobName), "-", "_") + "_job_sync_failure_number",
		Help:      jobName + " job sync failure number",
	}, []string{"method", "code", "desc"})

	instanceID := uuid.New()
	log.D().WithField(InstanceIDKeyName, instanceID).Infof("Initializing Metrics Pusher...")

	registry := prometheus.NewRegistry()
	registry.MustRegister(failedTenantsSyncJob)
	pusher := push.New(endpoint, TenantFetcherJobName).Gatherer(registry).Client(&http.Client{
		Timeout: timeout,
	})

	return &Pusher{
		failedTenantsSyncJob: failedTenantsSyncJob,
		pusher:               pusher,
		instanceID:           instanceID,
	}
}

// RecordTenantsSyncJobFailure missing godoc
func (p *Pusher) RecordTenantsSyncJobFailure(method string, statusCode int, desc string) {
	log.D().WithFields(logrus.Fields{
		InstanceIDKeyName: p.instanceID,
		"statusCode":      statusCode,
		"desc":            desc,
	}).Infof("Recording failed tenants sync job...")
	if p.failedTenantsSyncJob != nil {
		p.failedTenantsSyncJob.WithLabelValues(method, strconv.Itoa(statusCode), desc).Inc()
	}
}

// Push missing godoc
func (p *Pusher) Push() {
	log.D().WithField(InstanceIDKeyName, p.instanceID).Info("Pushing metrics...")
	if err := p.pusher.Add(); err != nil {
		wrappedErr := errors.Wrap(err, "while pushing metrics to Pushgateway")
		log.D().WithField(InstanceIDKeyName, p.instanceID).Error(wrappedErr)
	}
}
