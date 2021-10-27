package features

// Config missing godoc
type Config struct {
	DefaultScenarioEnabled bool   `envconfig:"default=true,APP_DEFAULT_SCENARIO_ENABLED"`
	ProtectedLabelPattern  string `envconfig:"default=.*_defaultEventing|^consumer_subaccount_ids$"`
}
