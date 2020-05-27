package metrics

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"strings"
)

//go:generate mockery -name=OperationsStatsGetter
type OperationsStatsGetter interface {
	InProgressOperationsCount() (model.OperationsCount, dberrors.Error)
}

type InProgressOperationsCollector struct {
	statsGetter OperationsStatsGetter

	provisioningDesc   *prometheus.Desc
	deprovisioningDesc *prometheus.Desc
	upgradeDesc        *prometheus.Desc

	log logrus.FieldLogger
}

func NewInProgressOperationsCollector(statsGetter OperationsStatsGetter) *InProgressOperationsCollector {
	return &InProgressOperationsCollector{
		statsGetter: statsGetter,

		provisioningDesc: prometheus.NewDesc(
			buildFQName(model.Provision),
			"The number of provisioning operations in progress",
			[]string{},
			nil),
		deprovisioningDesc: prometheus.NewDesc(
			buildFQName(model.Deprovision),
			"The number of deprovisioning operations in progress",
			[]string{},
			nil),
		upgradeDesc: prometheus.NewDesc(
			buildFQName(model.Upgrade),
			"The number of upgrade operations in progress",
			[]string{},
			nil),

		log: logrus.WithField("collector", "in-progress-operations"),
	}
}

func (c *InProgressOperationsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.provisioningDesc
	ch <- c.deprovisioningDesc
	ch <- c.upgradeDesc
}

func (c *InProgressOperationsCollector) Collect(ch chan<- prometheus.Metric) {
	inProgressOpsCounts, err := c.statsGetter.InProgressOperationsCount()
	if err != nil {
		c.log.Errorf("failed to get in progress operation while collecting metrics: %s", err.Error())
		return
	}

	fmt.Println("COUNT: ", inProgressOpsCounts)

	for a, b := range inProgressOpsCounts.Count {
		fmt.Println(a, "  ", b)

	}

	c.newMeasure(ch,
		c.provisioningDesc,
		inProgressOpsCounts.Count[model.Provision],
	)
	c.newMeasure(ch,
		c.deprovisioningDesc,
		inProgressOpsCounts.Count[model.Deprovision],
	)
	c.newMeasure(ch,
		c.upgradeDesc,
		inProgressOpsCounts.Count[model.Upgrade],
	)
}

func (c *InProgressOperationsCollector) newMeasure(ch chan<- prometheus.Metric, desc *prometheus.Desc, value int, labelValues ...string) {
	m, err := prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(value),
		labelValues...)
	if err != nil {
		c.log.Errorf("unable to register metric %s", err.Error())
		return
	}
	ch <- m
}

func buildFQName(operationType model.OperationType) string {
	operation := strings.ToLower(string(operationType))
	name := fmt.Sprintf("in_progress_%s_operations_total", operation)
	return prometheus.BuildFQName(prometheusNamespace, prometheusSubsystem, name)
}
