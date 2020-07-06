package metrics

import (
	"context"
	"fmt"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prometheusNamespace = "compass"
	prometheusSubsystem = "keb"

	resultFailed     float64 = 0
	resultSucceeded  float64 = 1
	resultInProgress float64 = 2
)

// OperationResultCollector provides the following metrics:
// - compass_keb_provisioning_result{"operation_id", "runtime_id", "instance_id"}
// - compass_keb_deprovisioning_result{"operation_id", "runtime_id", "instance_id"}
// These gauges show the status of the operation.
// The value of the gauge could be:
// 0 - Failed
// 1 - Succeeded
// 2 - In progress
type OperationResultCollector struct {
	provisioningResultGauge   *prometheus.GaugeVec
	deprovisioningResultGauge *prometheus.GaugeVec
}

func NewOperationResultCollector() *OperationResultCollector {
	return &OperationResultCollector{
		provisioningResultGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "provisioning_result",
			Help:      "Result of the provisioning",
		}, []string{"operation_id", "runtime_id", "instance_id"}),
		deprovisioningResultGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "deprovisioning_result",
			Help:      "Result of the deprovisioning",
		}, []string{"operation_id", "runtime_id", "instance_id"}),
	}
}

func (c *OperationResultCollector) Describe(ch chan<- *prometheus.Desc) {
	c.provisioningResultGauge.Describe(ch)
	c.deprovisioningResultGauge.Describe(ch)
}

func (c *OperationResultCollector) Collect(ch chan<- prometheus.Metric) {
	c.provisioningResultGauge.Collect(ch)
	c.deprovisioningResultGauge.Collect(ch)
}

func (c *OperationResultCollector) OnProvisioningStepProcessed(ctx context.Context, ev interface{}) error {
	stepProcessed, ok := ev.(process.ProvisioningStepProcessed)
	if !ok {
		return fmt.Errorf("expected ProvisioningStepProcessed but got %+v", ev)
	}

	var resultValue float64
	switch stepProcessed.Operation.State {
	case domain.InProgress:
		resultValue = resultInProgress
	case domain.Succeeded:
		resultValue = resultSucceeded
	case domain.Failed:
		resultValue = resultFailed
	}
	op := stepProcessed.Operation
	c.provisioningResultGauge.
		WithLabelValues(op.ID, op.RuntimeID, op.InstanceID).
		Set(resultValue)

	return nil
}

func (c *OperationResultCollector) OnDeprovisioningStepProcessed(ctx context.Context, ev interface{}) error {
	stepProcessed, ok := ev.(process.DeprovisioningStepProcessed)
	if !ok {
		return fmt.Errorf("expected DeprovisioningStepProcessed but got %+v", ev)
	}

	var resultValue float64
	switch stepProcessed.Operation.State {
	case domain.InProgress:
		resultValue = resultInProgress
	case domain.Succeeded:
		resultValue = resultSucceeded
	case domain.Failed:
		resultValue = resultFailed
	}
	op := stepProcessed.Operation
	c.deprovisioningResultGauge.
		WithLabelValues(op.ID, op.RuntimeID, op.InstanceID).
		Set(resultValue)
	return nil
}
