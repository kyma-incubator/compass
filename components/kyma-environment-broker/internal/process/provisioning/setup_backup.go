package provisioning

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"

	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

const (
	// the time after which the operation is marked as expired
	SetuUpBackupTimeOut = 1 * time.Hour
)

type SetupBackupStep struct {
	bucketName       string
	zone             string
	operationManager *process.OperationManager
	accountProvider  hyperscaler.AccountProvider
}

func (s *SetupBackupStep) Name() string {
	return "Setup_Backup"
}

func NewSetupBackupStep(os storage.Operations, accountProvider hyperscaler.AccountProvider) *SetupBackupStep {
	return &SetupBackupStep{
		operationManager: process.NewOperationManager(os),
		accountProvider:  accountProvider,
	}
}

func (s *SetupBackupStep) SetupBackupStep() {

}

func (s *SetupBackupStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	if time.Since(operation.UpdatedAt) > SetuUpBackupTimeOut {
		log.Infof("operation has reached the time limit: updated operation time: %s", operation.UpdatedAt)
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("operation has reached the time limit: %s", SetuUpBackupTimeOut))
	}
	log.Info("Setting Up Backup")

	pp, err := operation.GetProvisioningParameters()
	if err != nil {
		return s.operationManager.OperationFailed(operation, "invalid operation provisioning parameters")
	}

	hypType, err := getHyperscalerTypeForPlanID(pp.PlanID)

	if err != nil {
		errMsg := fmt.Sprintf("Aborting after failing to determine the type of Hyperscaler to use for planID: %s", pp.PlanID)
		return s.operationManager.OperationFailed(operation, errMsg)
	}

	// get credentials
	//credentials, err := s.accountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
	_, err = s.accountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)

	if err != nil {
		log.Errorf("Unable to retrieve Gardener Credentials from HAP lookup: %v", err)
		return operation, 5 * time.Second, nil
	}
	switch hypType {
	case "azure":
		log.Info("Setting up backup overrides")
		backupOverrides := s.setupBackUpOverride()
		operation.InputCreator.SetOverrides("backup-init", backupOverrides)
	}

	// Create bucket

	// Setup access rights
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
