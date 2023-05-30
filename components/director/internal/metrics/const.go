package metrics

const (
	// Namespace missing godoc
	Namespace = "compass"
	// DirectorSubsystem missing godoc
	DirectorSubsystem = "director"
	// TenantFetcherSubsystem missing godoc
	TenantFetcherSubsystem = "tenantfetcher"
	// TenantFetcherJobName missing godoc
	TenantFetcherJobName = TenantFetcherSubsystem
	// InstanceIDKeyName missing godoc
	InstanceIDKeyName = "instance"
	// OrdAggregatorSubsystem is used in the metrics pusher configuration as value for key 'subsystem'
	OrdAggregatorSubsystem = "ordaggregator"
	// ErrorMetricLabel is the error label used by metrics config for creating CounterVec for Prometheus
	ErrorMetricLabel = "error"
	// ResourceIDMetricLabel is an additional label used by ord aggregator for creating CounterVec for Prometheus
	ResourceIDMetricLabel   = "resource_id"
	ResourceTypeMetricLabel = "resource_type"
	// CorrelationIDMetricLabel is an additional label used by ord aggregator for creating CounterVec for Prometheus
	CorrelationIDMetricLabel = "x_request_id"
)
