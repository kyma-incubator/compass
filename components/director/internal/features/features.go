package features

// Config missing godoc
type Config struct {
	DefaultScenarioEnabled  bool     `envconfig:"default=true,APP_DEFAULT_SCENARIO_ENABLED"`
	ProtectedLabelPattern   string   `envconfig:"default=.*_defaultEventing,APP_PROTECTED_LABEL_PATTERN"`
	SecuredApplicationTypes []string `envconfig:"default=empty,APP_SECURED_APPLICATION_TYPES"`
}
