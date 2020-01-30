package broker

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	kymaServiceID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"
)

//go:generate mockery -name=OptionalComponentNamesProvider -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=InputBuilderForPlan -output=automock -outpkg=automock -case=underscore

type (
	DirectorClient interface {
		GetConsoleURL(accountID, runtimeID string) (string, error)
	}

	StructDumper interface {
		Dump(value ...interface{})
	}
)

var planIDsMapping = map[string]string{
	"azure": azurePlanID,
	"gcp":   gcpPlanID,
}

// Config represents configuration for broker
type Config struct {
	EnablePlans EnablePlans `envconfig:"default=azure"`
}

// EnablePlans defines the plans that should be available for provisioning
type EnablePlans []string

// Unmarshal provides custom parsing of Log Level.
// Implements envconfig.Unmarshal interface.
func (m *EnablePlans) Unmarshal(in string) error {
	plans := strings.Split(in, ",")
	for _, name := range plans {
		if _, exists := planIDsMapping[name]; !exists {
			return errors.Errorf("unrecognized %v plan name ", name)
		}
	}

	*m = plans
	return nil
}
