package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

type LoggingOverrides struct {
	operationManager *process.OperationManager
	accountProvider  hyperscaler.AccountProvider
}

func (s *LoggingOverrides) Name() string {
	return "Logging_Overrides"
}

func NewLoggingOverrides(os storage.Operations) *LoggingOverrides {
	return &LoggingOverrides{
		operationManager: process.NewOperationManager(os),
	}
}

func (s *LoggingOverrides) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	log.Info("Setting Up Overrides for logging")

	loggingOverrides := s.setupLoggingOverrides()
	operation.InputCreator.AppendOverrides("logging", loggingOverrides)

	return operation, 0, nil
}

func (s *LoggingOverrides) setupLoggingOverrides() []*gqlschema.ConfigEntryInput {
	loggingStepOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key:    "conf.Input.Kubernetes_loki.exclude.namespaces",
			Value:  "kube-node-lease,kube-public,kube-system,kyma-system,istio-system,kyma-installer,kyma-integration,knative-serving,knative-eventing",
			Secret: ptr.Bool(true),
		},
	}
	return loggingStepOverrides

}
