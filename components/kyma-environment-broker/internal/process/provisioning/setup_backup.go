package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

type SetupBackupStep struct {
	bucketName       string
	zone             string
	operationManager *OperationManager
	accountProvider  hyperscaler.AccountProvider
}

func (s *SetupBackupStep) Name() string {
	return "Setup_Backup"
}

func NewSetupBackupStep(os storage.Operations) *SetupBackupStep {
	return &SetupBackupStep{
		operationManager: NewOperationManager(os),
	}
}

func (s *SetupBackupStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	log.Info("Setting Up Backup")

	log.Info("Setting up backup overrides")
	backupOverrides := s.setupBackUpOverride()
	operation.InputCreator.AppendOverrides("backup-init", backupOverrides)

	return operation, 0, nil
}

func (s *SetupBackupStep) setupBackUpOverride() []*gqlschema.ConfigEntryInput {
	backupStepOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key:    "configuration.provider",
			Value:  "azure",
			Secret: ptr.Bool(true),
		},
	}
	return backupStepOverrides

}
