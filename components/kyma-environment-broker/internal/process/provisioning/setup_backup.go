package provisioning

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
	"time"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	azStorage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"

)

const (
	// the time after which the operation is marked as expired
	CSetuUpBackup = 1 * time.Hour
)

type SetupBackupStep struct {
	bucketName string
	zone string
	operationManager  *process.OperationManager
	instanceStorage   storage.Instances
	provisionerClient provisioner.Client
	serviceManager    internal.ServiceManagerOverride
	accountProvider  hyperscaler.AccountProvider
}

func (s *SetupBackupStep) Name() string {
	return "Setup_Backup"
}

func NewSetupBackupStep(os storage.Operations, is storage.Instances, cli provisioner.Client, smOverride internal.ServiceManagerOverride, accountProvider hyperscaler.AccountProvider) *SetupBackupStep {
	return &SetupBackupStep{
		operationManager:  process.NewOperationManager(os),
		instanceStorage:   is,
		provisionerClient: cli,
		serviceManager:    smOverride,
		accountProvider:  accountProvider,
	}
}

func (s *SetupBackupStep) SetupBackupStep() {

}

func (s *SetupBackupStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
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
	creds, err := s.accountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to fetch credentials for Global Account ID: %v",pp.ErsContext.GlobalAccountID)
		return s.operationManager.OperationFailed(operation, errMsg)
	}
	switch hypType {
	case "azure": backupSetupAz(creds.CredentialData, log)
	}
	fmt.Println(creds.CredentialData)


	// Create bucket

	// Setup access rights
	return operation, 5 * time.Second, nil
}

func getStorageAccountsClient(creds map[string][]byte) azStorage.AccountsClient {

	storageAccountsClient := azStorage.NewAccountsClient(string(creds["subscriptionID"]))
	auth, _ := iam.GetResourceManagementAuthorizer()
	storageAccountsClient.Authorizer = auth
	storageAccountsClient.AddToUserAgent("setup-backup")
	return storageAccountsClient
}

func backupSetupAz(cred map[string][]byte,  log logrus.FieldLogger) {

	storageAccountsClient := getStorageAccountsClient(cred)
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	log.Infof("Storage account list: %v",storageAccountsClient.List(ctx))
}
