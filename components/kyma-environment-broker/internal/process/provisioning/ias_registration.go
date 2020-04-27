package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/sirupsen/logrus"
)

type IASRegistrationStep struct {
	operationManager *process.ProvisionOperationManager
	bundleBuilder    ias.BundleBuilder
}

func NewIASRegistrationStep(os storage.Operations, builder ias.BundleBuilder) *IASRegistrationStep {
	return &IASRegistrationStep{
		operationManager: process.NewProvisionOperationManager(os),
		bundleBuilder:    builder,
	}
}

func (s *IASRegistrationStep) Name() string {
	return "IAS_Registration"
}

func (s *IASRegistrationStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	spb := s.bundleBuilder.NewBundle(operation.InstanceID)

	log.Info("Check if IAS ServiceProvider already exist")
	err := spb.FetchServiceProviderData()
	if err != nil {
		return s.handleError(operation, err, log, "fetching IAS ServiceProvider data failed")
	}

	if !spb.ServiceProviderExist() {
		log.Info("Create IAS ServiceProvider")
		err = spb.CreateServiceProvider()
		if err != nil {
			return s.handleError(operation, err, log, "creating IAS ServiceProvider failed")
		}
	} else {
		log.Infof("IAS ServiceProvider %q already registered", spb.ServiceProviderName())
	}

	log.Info("Configure IAS ServiceProvider")
	err = spb.ConfigureServiceProvider()
	if err != nil {
		return s.handleError(operation, err, log, "configuring IAS ServiceProvider failed")
	}

	log.Info("Generate IAS ServiceProvider Secret")
	secret, err := spb.GenerateSecret()
	if err != nil {
		return s.handleError(operation, err, log, "creating secret for IAS ServiceProvider failed")
	}

	operation.InputCreator.AppendOverrides("monitoring", []*gqlschema.ConfigEntryInput{
		{
			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_CLIENT_ID",
			Value:  secret.ClientID,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET",
			Value:  secret.ClientSecret,
			Secret: ptr.Bool(true),
		},
	})

	return operation, 0, nil
}

func (s *IASRegistrationStep) handleError(operation internal.ProvisioningOperation, err error, log logrus.FieldLogger, msg string) (internal.ProvisioningOperation, time.Duration, error) {
	log.Errorf("%s: %s", msg, err)
	switch {
	case kebError.IsTemporaryError(err):
		return s.operationManager.RetryOperation(operation, msg, 10*time.Second, time.Minute*30, log)
	default:
		return s.operationManager.OperationFailed(operation, msg)
	}
}
