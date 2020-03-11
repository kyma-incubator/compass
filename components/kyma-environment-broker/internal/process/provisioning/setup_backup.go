package provisioning

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure/auth"

	azStorage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-04-01/storage"
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
	//serviceManager    internal.ServiceManagerOverride
	accountProvider  hyperscaler.AccountProvider
}

func (s *SetupBackupStep) Name() string {
	return "Setup_Backup"
}

func NewSetupBackupStep(os storage.Operations, accountProvider hyperscaler.AccountProvider) *SetupBackupStep {
	return &SetupBackupStep{
		operationManager:  process.NewOperationManager(os),
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
	creds, err := s.accountProvider.GardenerCredentials(hypType, pp.ErsContext.GlobalAccountID)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to fetch credentials for Global Account ID: %v",pp.ErsContext.GlobalAccountID)
		return s.operationManager.OperationFailed(operation, errMsg)
	}
	switch hypType {
	case "azure":
		err := backupSetupAz(creds.CredentialData, log)
		if err != nil {
			log.Info(err.Error())
			return operation, 2 * time.Minute, nil
		}
		operation.InputCreator.SetOverrides("setup_backup", s.setupBackUpOverride(pp.ErsContext))
	}

	fmt.Println(creds.CredentialData)


	// Create bucket

	// Setup access rights
	return operation, 0, nil
}

func (s *SetupBackupStep) setupBackUpOverride(ersContext internal.ERSContext) []*gqlschema.ConfigEntryInput {
	backupStepOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key: "configuration.provider",
			Value: "azure",
		},
	}
return backupStepOverrides
	
}

func getStorageAccountsClient(creds map[string][]byte) (azStorage.AccountsClient, error) {

	certificateAuthorizer := auth.NewClientCredentialsConfig(string(creds["clientSecret"]), string(creds["clientID"]), string(creds["tenantID"]))
	authorizerToken, err := certificateAuthorizer.Authorizer()
	if err != nil {
		return azStorage.AccountsClient{},  errors.New("unable to authenticate to Hyperscaler")
	}

	storageAccountsClient := azStorage.NewAccountsClient(string(creds["SubscriptionID"]) )
	storageAccountsClient.Authorizer = authorizerToken
	err = storageAccountsClient.AddToUserAgent("backup-setup")
	if err != nil {
		return  azStorage.AccountsClient{},  errors.New("unable to add useragent to storage client")
	}
	return storageAccountsClient, nil
}

func backupSetupAz(cred map[string][]byte,  log logrus.FieldLogger) error{

	storageAccountsClient, err := getStorageAccountsClient(cred)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	l, err := storageAccountsClient.List(ctx)
	if err != nil {
		return err
	}
	log.Infof("Storage account list: %v",l)
	return  nil
}
