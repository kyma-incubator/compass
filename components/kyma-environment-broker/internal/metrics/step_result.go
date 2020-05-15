package metrics

import (
	"context"
	"fmt"

	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/prometheus/client_golang/prometheus"
)

type StepResultCollector struct {
	provisioningResultGauge   *prometheus.GaugeVec
	deprovisioningResultGauge *prometheus.GaugeVec
}

func NewStepResultCollector() *StepResultCollector {
	return &StepResultCollector{
		provisioningResultGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "provisioning_step_result",
			Help:      "Result of the provisioning step",
		}, []string{"operation_id", "runtime_id", "instance_id", "step_name"}),
		deprovisioningResultGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "deprovisioning_step_result",
			Help:      "Result of the deprovisioning step",
		}, []string{"operation_id", "runtime_id", "instance_id", "step_name"}),
	}
}

func (c *StepResultCollector) Describe(ch chan<- *prometheus.Desc) {
	c.provisioningResultGauge.Describe(ch)
	c.deprovisioningResultGauge.Describe(ch)
}

func (c *StepResultCollector) Collect(ch chan<- prometheus.Metric) {
	c.provisioningResultGauge.Collect(ch)
	c.deprovisioningResultGauge.Collect(ch)
}

func (c *StepResultCollector) OnProvisioningStepProcessed(ctx context.Context, ev interface{}) error {
	stepProcessed, ok := ev.(process.ProvisioningStepProcessed)
	if !ok {
		return fmt.Errorf("expected ProvisioningStepProcessed but got %+v", ev)
	}

	var resultValue float64
	switch {
	case stepProcessed.Operation.State == domain.Succeeded:
		resultValue = resultSucceeded
	case stepProcessed.When > 0 && stepProcessed.Error == nil:
		resultValue = resultInProgress
	case stepProcessed.When == 0 && stepProcessed.Error == nil:
		resultValue = resultSucceeded
	case stepProcessed.Error != nil:
		resultValue = resultFailed
	}
	c.provisioningResultGauge.WithLabelValues(
		stepProcessed.Operation.ID,
		stepProcessed.Operation.RuntimeID,
		stepProcessed.Operation.InstanceID,
		stepProcessed.StepName).Set(resultValue)

	return nil
}

func (c *StepResultCollector) OnDeprovisioningStepProcessed(ctx context.Context, ev interface{}) error {
	stepProcessed, ok := ev.(process.DeprovisioningStepProcessed)
	if !ok {
		return fmt.Errorf("expected DeprovisioningStepProcessed but got %+v", ev)
	}

	var resultValue float64
	switch {
	case stepProcessed.When > 0 && stepProcessed.Error == nil:
		resultValue = resultInProgress
	case stepProcessed.When == 0 && stepProcessed.Error == nil:
		resultValue = resultSucceeded
	case stepProcessed.Error != nil:
		resultValue = resultFailed
	}

	// Create_Runtime step always returns operation, 1 second, nil if everything is ok
	// this code is a workaround and should be removed when the step engine is refactored
	if stepProcessed.StepName == "Create_Runtime" && stepProcessed.When == time.Second {
		resultValue = resultSucceeded
	}

	c.deprovisioningResultGauge.WithLabelValues(
		stepProcessed.Operation.ID,
		stepProcessed.Operation.RuntimeID,
		stepProcessed.Operation.InstanceID,
		stepProcessed.StepName).Set(resultValue)
	return nil
}
