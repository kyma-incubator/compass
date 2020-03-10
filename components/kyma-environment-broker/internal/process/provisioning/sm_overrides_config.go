package provisioning

import (
	"strings"

	"github.com/pkg/errors"
)

const ServiceManagerComponentName = "service-manager-proxy"

type ServiceManagerOverrideConfig struct {
	OverrideMode ServiceManagerOverrideMode `envconfig:"default=Never"`
	URL          string
	Password     string
	Username     string
}

type ServiceManagerOverrideMode string

const (
	SMOverrideModeAlways               ServiceManagerOverrideMode = "Always"
	SMOverrideModeWhenNotSentInRequest ServiceManagerOverrideMode = "WhenNotSentInRequest"
	SMOverrideModeNever                ServiceManagerOverrideMode = "Never"
)

func (m ServiceManagerOverrideMode) IsUnknown() bool {
	switch m {
	case SMOverrideModeAlways, SMOverrideModeWhenNotSentInRequest, SMOverrideModeNever:
		return false
	default:
		return true
	}
}

func (m ServiceManagerOverrideMode) Names() string {
	all := []string{string(SMOverrideModeAlways), string(SMOverrideModeWhenNotSentInRequest), string(SMOverrideModeNever)}
	return strings.Join(all, ",")
}

// Unmarshal provides custom parsing of service manager credential mode.
// Implements envconfig.Unmarshal interface.
func (m *ServiceManagerOverrideMode) Unmarshal(in string) error {
	*m = ServiceManagerOverrideMode(in)

	if m.IsUnknown() {
		return errors.Errorf("Unsupported override mode %q, possible values %s ", in, m.Names())
	}

	return nil
}
