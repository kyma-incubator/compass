package provisioning

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"time"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
azStorage "github.com/Azure/azure-sdk-for-go/storage"

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
}

func (s *SetupBackupStep) Name() string {
	return "Setup_Backup"
}

func NewSetupBackupStep(os storage.Operations, is storage.Instances, cli provisioner.Client, smOverride internal.ServiceManagerOverride) *SetupBackupStep {
	return &SetupBackupStep{
		operationManager:  process.NewOperationManager(os),
		instanceStorage:   is,
		provisionerClient: cli,
		serviceManager:    smOverride,
	}
}

func (s *SetupBackupStep) SetupBackupStep() {

}

func (s *SetupBackupStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	log.Info("Setting Up Backup")
	c:= azStorage.N
	// Create bucket

	// Setup access rights
	return operation, 5 * time.Second, nil
}
