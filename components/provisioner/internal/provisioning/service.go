package provisioning

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/runtimes"
	"time"

	uuid "github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

//go:generate mockery -name=Service
type Service interface {
	ProvisionRuntime(config gqlschema.ProvisionRuntimeInput, tenant string) (*gqlschema.OperationStatus, error)
	UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(id, tenant string) (string, error)
	ReconnectRuntimeAgent(id string) (string, error)
	RuntimeStatus(id string) (*gqlschema.RuntimeStatus, error)
	RuntimeOperationStatus(id string) (*gqlschema.OperationStatus, error)
}

//go:generate mockery -name=Provisioner
type Provisioner interface {
	ProvisionCluster(cluster model.Cluster, operationId string) error
	DeprovisionCluster(cluster model.Cluster, operationId string) (model.Operation, error)
}

type service struct {
	inputConverter   InputConverter
	graphQLConverter GraphQLConverter
	directorService  director.DirectorClient

	dbSessionFactory dbsession.Factory
	provisioner      Provisioner
	uuidGenerator    uuid.UUIDGenerator
}

func NewProvisioningService(
	inputConverter InputConverter,
	graphQLConverter GraphQLConverter,
	directorService director.DirectorClient,
	factory dbsession.Factory,
	provisioner Provisioner,
	generator uuid.UUIDGenerator,
) Service {
	return &service{
		inputConverter:   inputConverter,
		graphQLConverter: graphQLConverter,
		directorService:  directorService,
		dbSessionFactory: factory,
		provisioner:      provisioner,
		uuidGenerator:    generator,
	}
}

func (r *service) ProvisionRuntime(config gqlschema.ProvisionRuntimeInput, tenant string) (*gqlschema.OperationStatus, error) {
	runtimeInput := config.RuntimeInput

	runtimeID, err := r.directorService.CreateRuntime(runtimeInput, tenant)
	if err != nil {
		return nil, fmt.Errorf("Failed to register Runtime: %s", err.Error())
	}

	cluster, err := r.inputConverter.ProvisioningInputToCluster(runtimeID, config, tenant)
	if err != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return nil, err
	}

	dbSession, err := r.dbSessionFactory.NewSessionWithinTransaction()
	if err != nil {
		return nil, fmt.Errorf("Failed to create repository: %s", err)
	}
	defer dbSession.RollbackUnlessCommitted()

	// Try to set provisioning started before triggering it (which is hard to interrupt) to verify all unique constraints
	operation, err := r.setProvisioningStarted(dbSession, runtimeID, cluster)
	if err != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return nil, err
	}

	err = r.provisioner.ProvisionCluster(cluster, operation.ID)
	if err != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return nil, fmt.Errorf("Failed to start provisioning: %s", err.Error())
	}

	err = dbSession.Commit()
	if err != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return nil, fmt.Errorf("Failed to commit transaction: %s", err.Error())
	}

	return r.graphQLConverter.OperationStatusToGQLOperationStatus(operation), nil
}

func (r *service) unregisterFailedRuntime(id, tenant string) {
	log.Infof("Unregistering failed Runtime %s...", id)
	err := r.directorService.DeleteRuntime(id, tenant)
	if err != nil {
		log.Warnf("Failed to unregister failed Runtime %s: %s", id, err.Error())
	}
}

func (r *service) DeprovisionRuntime(id, tenant string) (string, error) {
	session := r.dbSessionFactory.NewReadWriteSession()

	lastOperation, dberr := session.GetLastOperation(id)
	if dberr != nil {
		return "", fmt.Errorf("Failed to get last operation: %s", dberr.Error())
	}

	if lastOperation.State == model.InProgress {
		return "", errors.Errorf("cannot start new operation for %s Runtime while previous one is in progress", id)
	}

	cluster, dberr := session.GetCluster(id)
	if dberr != nil {
		return "", fmt.Errorf("Failed to get cluster: %s", dberr.Error())
	}

	operation, err := r.provisioner.DeprovisionCluster(cluster, r.uuidGenerator.New())
	if err != nil {
		return "", fmt.Errorf("Failed to start deprovisioning: %s", err.Error())
	}

	dberr = session.InsertOperation(operation)
	if dberr != nil {
		return "", fmt.Errorf("Failed to insert operation to database: %s", err.Error())
	}

	return operation.ID, nil
}

func (r *service) UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *service) ReconnectRuntimeAgent(id string) (string, error) {
	return "", nil
}

func (r *service) RuntimeStatus(runtimeID string) (*gqlschema.RuntimeStatus, error) {
	runtimeStatus, err := r.getRuntimeStatus(runtimeID)
	if err != nil {
		return nil, err
	}

	return r.graphQLConverter.RuntimeStatusToGraphQLStatus(runtimeStatus), nil
}

func (r *service) RuntimeOperationStatus(operationID string) (*gqlschema.OperationStatus, error) {
	readSession := r.dbSessionFactory.NewReadSession()

	operation, err := readSession.GetOperation(operationID)
	if err != nil {
		return nil, err
	}

	return r.graphQLConverter.OperationStatusToGQLOperationStatus(operation), nil
}

func (r *service) getRuntimeStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	session := r.dbSessionFactory.NewReadSession()

	operation, err := session.GetLastOperation(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	cluster, err := session.GetCluster(runtimeID)
	if err != nil {
		return model.RuntimeStatus{}, err
	}

	return model.RuntimeStatus{
		LastOperationStatus:  operation,
		RuntimeConfiguration: cluster,
	}, nil
}

func (r *service) setProvisioningStarted(dbSession dbsession.WriteSession, runtimeID string, cluster model.Cluster) (model.Operation, dberrors.Error) {
	timestamp := time.Now()

	cluster.CreationTimestamp = timestamp

	err := dbSession.InsertCluster(cluster)
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	gcpConfig, isGCP := cluster.GCPConfig()
	if isGCP {
		err = dbSession.InsertGCPConfig(gcpConfig)
		if err != nil {
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	gardenerConfig, isGardener := cluster.GardenerConfig()
	if isGardener {
		err = dbSession.InsertGardenerConfig(gardenerConfig)
		if err != nil {
			return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
		}
	}

	err = dbSession.InsertKymaConfig(cluster.KymaConfig)
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	operation, err := r.setOperationStarted(dbSession, runtimeID, model.Provision, timestamp, "Provisioning started", "Failed to set provisioning started: %s")
	if err != nil {
		return model.Operation{}, dberrors.Internal("Failed to set provisioning started: %s", err)
	}

	return operation, nil
}

func (r *service) setOperationStarted(dbSession dbsession.WriteSession, runtimeID string, operationType model.OperationType, timestamp time.Time, message string, errorMessageFmt string) (model.Operation, dberrors.Error) {
	id := r.uuidGenerator.New()

	operation := model.Operation{
		ID:             id,
		Type:           operationType,
		StartTimestamp: timestamp,
		State:          model.InProgress,
		Message:        message,
		ClusterID:      runtimeID,
	}

	err := dbSession.InsertOperation(operation)
	if err != nil {
		return model.Operation{}, dberrors.Internal(errorMessageFmt, err)
	}

	return operation, nil
}
