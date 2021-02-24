package ord

import "fmt"

type Config struct {
	ServiceURL string `mapstructure:"service_url" description:"an url pointing to ORD service's fully qualified root address'"`
	StaticPath string `mapstructure:"static_path" description:"path to ORD service's endpoint static path'"`
}

func DefaultConfig() *Config {
	return &Config{
		ServiceURL: "https://compass-gateway.kyma.local",
		StaticPath: "/open-resource-discovery-static/v0",
	}
}

// Validate validates the server settings
func (s *Config) Validate() error {
	if len(s.ServiceURL) == 0 {
		return fmt.Errorf("validate Settings: ServiceURL missing")
	}
	if len(s.StaticPath) == 0 {
		return fmt.Errorf("validate Settings: StaticPath missing")
	}

	return nil
}
