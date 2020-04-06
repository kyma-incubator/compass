package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

type ServiceManagerOverridesStep struct {
	serviceManager   ServiceManagerOverrideConfig
	operationManager *process.ProvisionOperationManager
}

func NewServiceManagerOverridesStep(os storage.Operations, smOverride ServiceManagerOverrideConfig) *ServiceManagerOverridesStep {
	return &ServiceManagerOverridesStep{
		serviceManager:   smOverride,
		operationManager: process.NewProvisionOperationManager(os),
	}
}

func (s *ServiceManagerOverridesStep) Name() string {
	return "ServiceManagerOverrides"
}

func (s *ServiceManagerOverridesStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	pp, err := operation.GetParameters()
	if err != nil {
		log.Errorf("cannot fetch provisioning parameters from operation: %s", err)
		return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}

	ersCtx := pp.ErsContext
	var smOverrides []*gqlschema.ConfigEntryInput
	if s.shouldOverride(ersCtx.ServiceManager) {
		smOverrides = []*gqlschema.ConfigEntryInput{
			{
				Key:   "config.sm.url",
				Value: s.serviceManager.URL,
			},
			{
				Key:   "sm.user",
				Value: s.serviceManager.Username,
			},
			{
				Key:    "sm.password",
				Value:  s.serviceManager.Password,
				Secret: ptr.Bool(true),
			},
		}
	} else {
		if ersCtx.ServiceManager == nil {
			log.Errorf("Service Manager Credentials are required to be send in provisioning request (override_mode: %q)", s.serviceManager.OverrideMode)
			return s.operationManager.OperationFailed(operation, "Service Manager Credentials are required to be send in provisioning request.")
		}

		smOverrides = []*gqlschema.ConfigEntryInput{
			{
				Key:   "config.sm.url",
				Value: ersCtx.ServiceManager.URL,
			},
			{
				Key:   "sm.user",
				Value: ersCtx.ServiceManager.Credentials.BasicAuth.Username,
			},
			{
				Key:    "sm.password",
				Value:  ersCtx.ServiceManager.Credentials.BasicAuth.Password,
				Secret: ptr.Bool(true),
			},
		}
	}
	operation.InputCreator.AppendOverrides(ServiceManagerComponentName, smOverrides)

	return operation, 0, nil
}

func (s *ServiceManagerOverridesStep) shouldOverride(reqCreds *internal.ServiceManagerEntryDTO) bool {
	if s.serviceManager.OverrideMode == SMOverrideModeAlways {
		return true
	}

	if s.serviceManager.OverrideMode == SMOverrideModeWhenNotSentInRequest && reqCreds == nil {
		return true
	}

	return false
}
