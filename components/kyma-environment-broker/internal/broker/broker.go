package broker

import (
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pkg/errors"
)

const (
	kymaServiceID = "47c9dcbf-ff30-448e-ab36-d3bad66ba281"

	// time delay after which the instance becomes obsolete in the process of polling for last operation
	delayInstanceTime = 3 * time.Hour
)

//go:generate mockery -name=OptionalComponentNamesProvider -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=InputBuilderForPlan -output=automock -outpkg=automock -case=underscore

// OptionalComponentNamesProvider provides optional components names
type OptionalComponentNamesProvider interface {
	GetAllOptionalComponentsNames() []string
}

type DirectorClient interface {
	GetConsoleURL(accountID, runtimeID string) (string, error)
}

type StructDumper interface {
	Dump(value ...interface{})
}

// ProvisioningConfig holds all configurations connected with Provisioner API
type ProvisioningConfig struct {
	URL                 string
	SecretName          string
	GCPSecretName       string
	AzureSecretName     string
	AWSSecretName       string
	GardenerProjectName string
}

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

// KymaEnvBroker implements the Kyma Environment Broker
type KymaEnvBroker struct {
	dumper             StructDumper
	provisioningCfg    ProvisioningConfig
	provisionerClient  provisioner.Client
	instancesStorage   storage.Instances
	builderFactory     InputBuilderForPlan
	DirectorClient     DirectorClient
	optionalComponents OptionalComponentNamesProvider
	enabledPlanIDs     map[string]struct{}
}

func New(cfg Config, pCli provisioner.Client, dCli DirectorClient, provisioningCfg ProvisioningConfig, instStorage storage.Instances, optComponentsSvc OptionalComponentNamesProvider,
	builderFactory InputBuilderForPlan, dumper StructDumper) (*KymaEnvBroker, error) {

	enabledPlanIDs := map[string]struct{}{}
	for _, planName := range cfg.EnablePlans {
		id := planIDsMapping[planName]
		enabledPlanIDs[id] = struct{}{}
	}

	return &KymaEnvBroker{
		provisionerClient:  pCli,
		DirectorClient:     dCli,
		dumper:             dumper,
		provisioningCfg:    provisioningCfg,
		instancesStorage:   instStorage,
		enabledPlanIDs:     enabledPlanIDs,
		builderFactory:     builderFactory,
		optionalComponents: optComponentsSvc,
	}, nil
}
