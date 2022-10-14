package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	maxErrMessageLength = 50
	errorMetricLabel    = "error"
)

// PusherConfig is used to provide configuration options for AggregationFailurePusher.
type PusherConfig struct {
	Enabled    bool
	Endpoint   string
	MetricName string
	Timeout    time.Duration
	Subsystem  string
}

// AggregationFailurePusher is used for pushing metrics to Prometheus related to failed aggregation.
type AggregationFailurePusher struct {
	aggregationFailuresCounter *prometheus.CounterVec
	pusher                     *push.Pusher
	instanceID                 uuid.UUID
}

// NewAggregationFailurePusher returns a new Prometheus metrics pusher that can be used to report aggregation failures.
func NewAggregationFailurePusher(cfg PusherConfig) AggregationFailurePusher {
	if !cfg.Enabled {
		return AggregationFailurePusher{}
	}
	instanceID := uuid.New()
	log.D().WithField(InstanceIDKeyName, instanceID).Infof("Initializing Metrics Pusher...")

	aggregationFailuresCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: cfg.Subsystem,
		Name:      cfg.MetricName,
		Help:      fmt.Sprintf("Aggregation status for %s", cfg.Subsystem),
	}, []string{errorMetricLabel})

	pusher := newPusher(cfg, aggregationFailuresCounter)
	return AggregationFailurePusher{
		aggregationFailuresCounter: aggregationFailuresCounter,
		pusher:                     pusher,
		instanceID:                 instanceID,
	}
}

// ReportAggregationFailure reports failed aggregation with the provided error.
func (p AggregationFailurePusher) ReportAggregationFailure(ctx context.Context, err error) {
	if p.pusher == nil {
		log.C(ctx).Error("Metrics pusher is not configured, skipping report...")
		return
	}

	log.C(ctx).WithFields(logrus.Fields{InstanceIDKeyName: p.instanceID}).Info("Reporting failed aggregation...")
	p.aggregationFailuresCounter.WithLabelValues(errorDescription(err)).Inc()
	p.push(ctx)
}

func (p AggregationFailurePusher) push(ctx context.Context) {
	if err := p.pusher.Add(); err != nil {
		wrappedErr := errors.Wrap(err, "while pushing metrics to Pushgateway")
		log.C(ctx).WithField(InstanceIDKeyName, p.instanceID).Error(wrappedErr)
	}
}

func newPusher(cfg PusherConfig, collectors ...prometheus.Collector) *push.Pusher {
	registry := prometheus.NewRegistry()
	for _, c := range collectors {
		registry.MustRegister(c)
	}

	return push.New(cfg.Endpoint, cfg.Subsystem).Gatherer(registry).Client(&http.Client{
		Timeout: cfg.Timeout,
	})
}

func errorDescription(err error) string {
	var e *net.OpError
	if errors.As(err, &e) && e.Err != nil {
		return e.Err.Error()
	}

	if len(err.Error()) > maxErrMessageLength {
		// not all errors are actually wrapped, sometimes the error message is just concatenated with ":"
		errParts := strings.Split(err.Error(), ":")
		return errParts[len(errParts)-1]
	}

	return err.Error()
}
