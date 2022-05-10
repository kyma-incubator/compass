package metrics

// Config configures the behaviour of the metrics collector.
type Config struct {
	Address                       string   `envconfig:"APP_METRICS_ADDRESS"`
	EnableClientIDInstrumentation bool     `envconfig:"default=true,APP_METRICS_ENABLE_CLIENT_ID_INSTRUMENTATION"`
	CensoredFlows                 []string `envconfig:"optional,APP_METRICS_CENSORED_FLOWS"`
}
