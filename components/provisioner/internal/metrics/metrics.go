package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	prometheusNamespace = "compass"
	prometheusSubsystem = "provisioner"
)

func Register(opsStatsGetter OperationsStatsGetter) error {
	err := prometheus.Register(NewInProgressOperationsCollector(opsStatsGetter))
	if err != nil {
		return err
	}

	return nil
}
