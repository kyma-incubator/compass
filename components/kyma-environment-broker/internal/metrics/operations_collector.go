package metrics

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	namespace = "keb_operations"
	subsystem = "operations"
)

// OperationsStatsGetter provides metrics, which shows how many operations were done:
// - compass_keb_operations_provisioning_failed_total
// - compass_keb_operations_provisioning_in_progress_total
// - compass_keb_operations_provisioning_succeeded_total
// - compass_keb_operations_deprovisioning_failed_total
// - compass_keb_operations_deprovisioning_in_progress_total
// - compass_keb_operations_deprovisioning_secceeded_total
type OperationsStatsGetter interface {
	GetOperationStats() (internal.OperationStats, error)
}

type OperationsCollector struct {
	statsGetter OperationsStatsGetter

	provisioningInProgressDesc *prometheus.Desc
	provisioningSucceededDesc  *prometheus.Desc
	provisioningFailedDesc     *prometheus.Desc

	deprovisioningInProgressDesc *prometheus.Desc
	deprovisioningSucceededDesc  *prometheus.Desc
	deprovisioningFailedDesc     *prometheus.Desc
}

func NewOperationsCollector(statsGetter OperationsStatsGetter) *OperationsCollector {
	return &OperationsCollector{
		statsGetter: statsGetter,

		provisioningInProgressDesc: prometheus.NewDesc(
			fqName(dbmodel.OperationTypeProvision, domain.InProgress),
			"The number of provisioning operations in progress",
			[]string{},
			nil),
		provisioningFailedDesc: prometheus.NewDesc(
			fqName(dbmodel.OperationTypeProvision, domain.Failed),
			"The number of failed provisioning operations",
			[]string{},
			nil),
		provisioningSucceededDesc: prometheus.NewDesc(
			fqName(dbmodel.OperationTypeProvision, domain.Succeeded),
			"The number of succeeded provisioning operations",
			[]string{},
			nil),

		deprovisioningInProgressDesc: prometheus.NewDesc(
			fqName(dbmodel.OperationTypeDeprovision, domain.InProgress),
			"The number of deprovisioning operations in progress",
			[]string{},
			nil),
		deprovisioningFailedDesc: prometheus.NewDesc(
			fqName(dbmodel.OperationTypeDeprovision, domain.Failed),
			"The number of failed deprovisioning operations",
			[]string{},
			nil),
		deprovisioningSucceededDesc: prometheus.NewDesc(
			fqName(dbmodel.OperationTypeDeprovision, domain.Succeeded),
			"The number of succeeded deprovisioning operations",
			[]string{},
			nil),
	}
}

func fqName(operationType dbmodel.OperationType, state domain.LastOperationState) string {
	var opType string
	switch operationType {
	case dbmodel.OperationTypeProvision:
		opType = "provisioning"
	case dbmodel.OperationTypeDeprovision:
		opType = "deprovisioning"
	}

	var st string
	switch state {
	case domain.Failed:
		st = "failed"
	case domain.Succeeded:
		st = "succeeded"
	case domain.InProgress:
		st = "in_progress"
	}
	name := fmt.Sprintf("operations_%s_%s_total", opType, st)
	return prometheus.BuildFQName(prometheusNamespace, prometheusSubsystem, name)
}

func (c *OperationsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.deprovisioningFailedDesc
	ch <- c.deprovisioningSucceededDesc
	ch <- c.deprovisioningInProgressDesc
	ch <- c.provisioningFailedDesc
	ch <- c.provisioningSucceededDesc
	ch <- c.deprovisioningInProgressDesc
}

// Collect implements the prometheus.Collector interface.
func (c *OperationsCollector) Collect(ch chan<- prometheus.Metric) {
	stats, err := c.statsGetter.GetOperationStats()
	if err != nil {
		return
	}

	collect(ch,
		c.provisioningInProgressDesc,
		stats.Provisioning[domain.InProgress],
	)
	collect(ch,
		c.provisioningSucceededDesc,
		stats.Provisioning[domain.Succeeded],
	)
	collect(ch,
		c.provisioningFailedDesc,
		stats.Provisioning[domain.Failed],
	)

	collect(ch,
		c.deprovisioningInProgressDesc,
		stats.Deprovisioning[domain.InProgress],
	)
	collect(ch,
		c.deprovisioningSucceededDesc,
		stats.Deprovisioning[domain.Succeeded],
	)
	collect(ch,
		c.deprovisioningFailedDesc,
		stats.Deprovisioning[domain.Failed],
	)
}

func collect(ch chan<- prometheus.Metric, desc *prometheus.Desc, value int, labelValues ...string) {
	m, err := prometheus.NewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(value),
		labelValues...)

	if err != nil {
		logrus.Errorf("unable to register metric %s", err.Error())
		return
	}
	ch <- m
}
