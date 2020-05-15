package metrics

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "keb_operations"
	subsystem = "operations"
)

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
	return fmt.Sprintf("%s_%s_operations_%s_%s_total", prometheusNamespace, prometheusSubsystem, opType, st)
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

	ch <- prometheus.MustNewConstMetric(
		c.provisioningInProgressDesc,
		prometheus.GaugeValue,
		float64(stats.Provisioning[domain.InProgress]),
	)
	ch <- prometheus.MustNewConstMetric(
		c.provisioningSucceededDesc,
		prometheus.GaugeValue,
		float64(stats.Provisioning[domain.Succeeded]),
	)
	ch <- prometheus.MustNewConstMetric(
		c.provisioningFailedDesc,
		prometheus.GaugeValue,
		float64(stats.Provisioning[domain.Failed]),
	)

	ch <- prometheus.MustNewConstMetric(
		c.deprovisioningInProgressDesc,
		prometheus.GaugeValue,
		float64(stats.Deprovisioning[domain.InProgress]),
	)
	ch <- prometheus.MustNewConstMetric(
		c.deprovisioningSucceededDesc,
		prometheus.GaugeValue,
		float64(stats.Deprovisioning[domain.Succeeded]),
	)
	ch <- prometheus.MustNewConstMetric(
		c.deprovisioningFailedDesc,
		prometheus.GaugeValue,
		float64(stats.Deprovisioning[domain.Failed]),
	)
}
