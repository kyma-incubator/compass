package provisioning

import (
	"fmt"
	azure "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/Azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	// the time after which the operation is marked as expired
	SetuUpBackupTimeOut = 1 * time.Hour
)

type SetupBackupStep struct {
	bucketName string
	zone string
	operationManager  *process.OperationManager
	//instanceStorage   storage.Instances
	//provisionerClient provisioner.Client
	azureClientInterface azure.AzureClientInterface
	accountProvider  hyperscaler.AccountProvider
}

func (s *SetupBackupStep) Name() string {
	return "Setup_Backup"
}

func NewSetupBackupStep(os storage.Operations, accountProvider hyperscaler.AccountProvider, azureClientInterface azure.AzureClientInterface) *SetupBackupStep {
	return &SetupBackupStep{
		operationManager:  process.NewOperationManager(os),
		accountProvider:  accountProvider,
		azureClientInterface: azureClientInterface,

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
		//TODO: Uncomment when we have azure client
		//azureCfg, err := azure.GetConfigfromHAPCredentialsAndProvisioningParams(credentials, pp, log)
		//if err != nil {
		//	log.Errorf("Unable to set the Azure config: %v", err)
		//	return operation, 5 * time.Second, nil
		//}
		//azSacClient := s.azureClientInterface.GetStorageAccountClientOrDie(azureCfg)
		//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		//defer cancel()
		//name := "foobar"
		//typeSac := "Microsoft.Storage/storageAccounts"
		//accountName := azStorage.AccountCheckNameAvailabilityParameters{
		//	Name: &name,
		//	Type: &typeSac,
		//}
		//result, err  := azSacClient.CheckAccountNameAvailability(ctx, accountName)
		//
		//if err != nil {
		//	log.Errorf("Unable to set the check if storage account name is available: %v", err)
		//	return operation, 5 * time.Second, nil
		//}
		//
		//log.Infof("result is %v", *result.NameAvailable)
		//log.Infof("result statuscode 2 is %v", result.StatusCode)


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
			Key: "configuration.provider",
			Value: "azure",
			Secret: ptr.Bool(true),
		},
	}
return backupStepOverrides
	
}
