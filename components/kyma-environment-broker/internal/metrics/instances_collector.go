package metrics

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type InstancesStatsGetter interface {
	GetInstanceStats() (internal.InstanceStats, error)
}

type InstancesCollector struct {
	statsGetter InstancesStatsGetter

	instancesDesc        *prometheus.Desc
	instancesPerGAIDDesc *prometheus.Desc
}

func NewInstancesCollector(statsGetter InstancesStatsGetter) *InstancesCollector {
	return &InstancesCollector{
		statsGetter: statsGetter,

		instancesDesc: prometheus.NewDesc(
			prometheus.BuildFQName(prometheusNamespace, prometheusSubsystem, "instances_total"),
			"The total number of instances",
			[]string{},
			nil),
		instancesPerGAIDDesc: prometheus.NewDesc(
			prometheus.BuildFQName(prometheusNamespace, prometheusSubsystem, "global_account_id_instances_total"),
			"The total number of instances by Global Account ID",
			[]string{"global_account_id"},
			nil),
	}
}

func (c *InstancesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.instancesDesc
	ch <- c.instancesPerGAIDDesc
}

// Collect implements the prometheus.Collector interface.
func (c *InstancesCollector) Collect(ch chan<- prometheus.Metric) {
	stats, err := c.statsGetter.GetInstanceStats()
	if err != nil {
		logrus.Error(err)
		return
	}
	collect(ch, c.instancesDesc, stats.TotalNumberOfInstances)

	for globalAccountID, num := range stats.PerGlobalAccountID {
		collect(ch, c.instancesPerGAIDDesc, num, globalAccountID)
	}
}
