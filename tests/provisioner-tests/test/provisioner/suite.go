package provisioner

import (
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/provisioner"
)

type TestSuite struct {
	ProvisionerClient provisioner.Client
}

func NewTestSuite(config testkit.TestConfig) (*TestSuite, error) {

	provisionerClient := provisioner.NewProvisionerClient(config.InternalProvisionerUrl, config.QueryLogging)

	return &TestSuite{
		ProvisionerClient: provisionerClient,
	}, nil
}
