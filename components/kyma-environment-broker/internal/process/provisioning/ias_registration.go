package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

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
	for spID := range ias.ServiceProviderInputs {
		spb, err := s.bundleBuilder.NewBundle(operation.InstanceID, spID)
		if err != nil {
			return s.handleError(operation, err, log, "failed to create new ServiceProvider Bundle")
		}

		log.Infof("Check if IAS ServiceProvider %q already exist", spb.ServiceProviderName())
		err = spb.FetchServiceProviderData()
		if err != nil {
			return s.handleError(operation, err, log, "fetching IAS ServiceProvider data failed")
		}

		if !spb.ServiceProviderExist() {
			log.Infof("Create IAS ServiceProvider %q", spb.ServiceProviderName())
			err = spb.CreateServiceProvider()
			if err != nil {
				return s.handleError(operation, err, log, "creating IAS ServiceProvider failed")
			}
		} else {
			log.Infof("IAS ServiceProvider %q already registered", spb.ServiceProviderName())
		}

		log.Infof("Configure IAS ServiceProvider %q", spb.ServiceProviderName())
		err = spb.ConfigureServiceProvider()
		if err != nil {
			return s.handleError(operation, err, log, "configuring IAS ServiceProvider failed")
		}

		componentName, overrides := spb.GetProvisioningOverrides()
		if componentName != "" {
			operation.InputCreator.AppendOverrides(componentName, overrides)
		}
	}

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
