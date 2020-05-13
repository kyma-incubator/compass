package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	// active goroutines
	// active DB connections number in connection pool

	// response status codes from Tenant Fetcher (check if we could use k8s job status code) -> k8s metric

	// response status codes from Ory Hydra
	// average time of GraphQL request handling
	// average time of DB request handling

	// prometheus-postgres-exporter:

	// number of tenants
	// number of tenants that have at least one application or runtime
	// number of runtimes
	// number of applications

	dbConnections *prometheus.GaugeVec
}

const (
	InUseDBConnections = "in_use"
	IdleDBConnections  = "idle"
	MaxDBConnections   = "max"
)

type DBTransactioner interface {
	Stats() sql.DBStats
}

func NewCollector() *Collector {
	return &Collector{
		dbConnections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "director_db_connections_total",
			Help: "Open database connections for Director",
		}, []string{"type"}),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.dbConnections.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.dbConnections.Collect(ch)
}

func (c *Collector) SetDBConnectionsMetrics(stats sql.DBStats) {
	c.dbConnections.WithLabelValues(MaxDBConnections).Set(float64(stats.MaxOpenConnections))
	c.dbConnections.WithLabelValues(InUseDBConnections).Set(float64(stats.InUse))
	c.dbConnections.WithLabelValues(IdleDBConnections).Set(float64(stats.Idle))
}
