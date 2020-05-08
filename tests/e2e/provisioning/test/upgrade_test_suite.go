package test

import (
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/provisioner"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
	"testing"
	"time"
)

type UpgradeConfig struct {
	ManagedRuntimeComponentsYAMLPath string
	UpgradeTimeout                   time.Duration `default:"3h"`
	PreUpgradeKymaVersion            string        `default:""` // If empty default version should be used
	UpgradeKymaVersion               string        `default:""` // TODO: get latest master?
}

func WithUpgrade() options {
	return func(t *testing.T, config Config, suite *Suite) {
		upgradeSuite := newUpgradeSuite(t, config, suite)

		suite.upgradeSuite = upgradeSuite
	}
}

type UpgradeSuite struct {
	upgradeClient *provisioner.RuntimeUpgradeClient

	UpgradeTimeout time.Duration

	PreUpgradeKymaVersion string
	UpgradeKymaVersion    string
}

func newUpgradeSuite(t *testing.T, baseConfig Config, suite *Suite) *UpgradeSuite {
	cfg := &UpgradeConfig{}
	err := envconfig.InitWithPrefix(cfg, "APP")
	require.NoError(t, err)

	log := logrus.New()

	// TODO: if upgrade empty fetch lateste master?

	httpClient := newHTTPClient(baseConfig.SkipCertVerification)
	provisionerClient := provisioner.NewProvisionerClient(baseConfig.ProvisionerURL, baseConfig.TenantID, log.WithField("service", "provisioner_client"), httpClient)
	componentsProvider := provisioner.NewComponentsListProvider(cfg.ManagedRuntimeComponentsYAMLPath)

	upgradeClient := provisioner.NewRuntimeUpgradeClient(suite.directorClient, provisionerClient, componentsProvider, baseConfig.TenantID, suite.InstanceID, log.WithField("service", "upgrade_client"))

	return &UpgradeSuite{
		upgradeClient:         upgradeClient,
		UpgradeTimeout:        cfg.UpgradeTimeout,
		PreUpgradeKymaVersion: cfg.PreUpgradeKymaVersion,
		UpgradeKymaVersion:    cfg.UpgradeKymaVersion,
	}
}
