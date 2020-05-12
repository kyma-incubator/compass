package metrics

import "github.com/prometheus/client_golang/prometheus"

type ControllerBusinessMetric struct {
}

type Collector struct {
	// active goroutines

	// response status codes from Tenant Fetcher (check if we could use k8s job status code) -> k8s metric

	// active DB connections number in connection pool
	// response status codes from Ory Hydra
	// average time of GraphQL request handling
	// average time of DB request handling

	// prometheus-postgres-exporter:

	// number of tenants
	// number of tenants that have at least one application or runtime
	// number of runtimes
	// number of applications

}

func NewCollector() *Collector {
	return &Collector{
		//applications: prometheus.NewGaugeVec(prometheus.GaugeOpts{
		//	Name: "director__runtime_reconcile_queue_length",
		//	Help: "Length of reconcile queue per controller",
		//}, []string{"director"}),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {

}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {

}
