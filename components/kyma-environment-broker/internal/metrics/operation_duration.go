package metrics

import (
	"context"
	"fmt"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/prometheus/client_golang/prometheus"
)

// OperationDurationCollector provides histograms which describes the time of provisioning/deprovisioning operations:
// - compass_keb_provisioning_duration_minutes
// - compass_keb_deprovisioning_duration_minutes
type OperationDurationCollector struct {
	provisioningHistogram   prometheus.Histogram
	deprovisioningHistogram prometheus.Histogram
}

func NewOperationDurationCollector() *OperationDurationCollector {
	return &OperationDurationCollector{
		provisioningHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "provisioning_duration_minutes",
			Help:      "The time of the provisioning process",
			Buckets:   prometheus.LinearBuckets(20, 2, 40),
		}),
		deprovisioningHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "deprovisioning_duration_minutes",
			Help:      "The time of the deprovisioning process",
			Buckets:   prometheus.LinearBuckets(1, 1, 30),
		}),
	}
}

func (c *OperationDurationCollector) Describe(ch chan<- *prometheus.Desc) {
	c.provisioningHistogram.Describe(ch)
	c.deprovisioningHistogram.Describe(ch)
}

func (c *OperationDurationCollector) Collect(ch chan<- prometheus.Metric) {
	c.provisioningHistogram.Collect(ch)
	c.deprovisioningHistogram.Collect(ch)
}

func (c *OperationDurationCollector) OnProvisioningStepProcessed(ctx context.Context, ev interface{}) error {
	stepProcessed, ok := ev.(process.ProvisioningStepProcessed)
	if !ok {
		return fmt.Errorf("expected process.ProvisioningStepProcessed but got %+v", ev)
	}

	if stepProcessed.OldOperation.State == domain.InProgress && stepProcessed.Operation.State == domain.Succeeded {
		minutes := stepProcessed.Operation.UpdatedAt.Sub(stepProcessed.Operation.CreatedAt).Minutes()
		c.provisioningHistogram.Observe(minutes)
	}

	return nil
}

func (c *OperationDurationCollector) OnDeprovisioningStepProcessed(ctx context.Context, ev interface{}) error {
	stepProcessed, ok := ev.(process.DeprovisioningStepProcessed)
	if !ok {
		return fmt.Errorf("expected process.DeprovisioningStepProcessed but got %+v", ev)
	}

	if stepProcessed.OldOperation.State == domain.InProgress && stepProcessed.Operation.State == domain.Succeeded {
		minutes := stepProcessed.Operation.UpdatedAt.Sub(stepProcessed.Operation.CreatedAt).Minutes()
		c.deprovisioningHistogram.Observe(minutes)
	}

	return nil
}
