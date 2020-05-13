package deprovisioning

import (
	"errors"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WaitForClusterDeletionStep struct {
	installationClient installation.Service
	gardenerClient     GardenerClient
	dbsFactory         dbsession.Factory
	directorClient     director.DirectorClient
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

type GardenerClient interface {
	Get(name string, options v1.GetOptions) (*gardener_types.Shoot, error)
	Delete(name string, options *v1.DeleteOptions) error
}

func NewWaitForClusterDeletionStep(installationClient installation.Service, gardenerClient GardenerClient, dbsFactory dbsession.Factory, directorClient director.DirectorClient, nextStep model.OperationStage, timeLimit time.Duration) *WaitForClusterDeletionStep {
	return &WaitForClusterDeletionStep{
		installationClient: installationClient,
		gardenerClient:     gardenerClient,
		dbsFactory:         dbsFactory,
		directorClient:     directorClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *WaitForClusterDeletionStep) Name() model.OperationStage {
	return model.StartingInstallation
}

func (s *WaitForClusterDeletionStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForClusterDeletionStep) Run(cluster model.Cluster, operation model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	if operation.Type == model.Deprovision {
		if operation.State == model.InProgress || operation.State == model.Succeeded {

			err := s.deleteShoot(cluster)
			if err != nil {
				logger.Errorf("error deleting shoot: %s", err.Error())
				return operations.StageResult{}, err
			}

			err = s.setDeprovisioningFinished(cluster, operation)
			if err != nil {
				logger.Errorf("error setting deprovisioning finished: %s", err.Error())
				return operations.StageResult{}, err
			}

			return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
		} else {
			err := fmt.Errorf("Error: Invalid state. deprovisioning in progress, last operation: %s, state: %s", operation.Type, operation.State)
			logger.Errorf(err.Error())
			return operations.StageResult{}, operations.NewNonRecoverableError(err)
		}
	}

	return operations.StageResult{Stage: s.Name(), Delay: 30 % time.Second}, nil
}

func (s *WaitForClusterDeletionStep) setDeprovisioningFinished(cluster model.Cluster, lastOp model.Operation) error {
	session, dberr := s.dbsFactory.NewSessionWithinTransaction()
	if dberr != nil {
		return fmt.Errorf("error starting db session with transaction: %s", dberr.Error())
	}
	defer session.RollbackUnlessCommitted()

	dberr = session.MarkClusterAsDeleted(cluster.ID)
	if dberr != nil {
		return fmt.Errorf("error marking cluster for deletion: %s", dberr.Error())
	}

	dberr = session.TransitionOperation(lastOp.ID, "Deprovisioning finished.", model.FinishedStage, time.Now())
	if dberr != nil {
		return fmt.Errorf("error trainsitioning deprovision operation state %s: %s", lastOp.ID, dberr.Error())
	}

	dberr = session.UpdateOperationState(lastOp.ID, "Operation succeeded.", model.Succeeded, time.Now())
	if dberr != nil {
		return fmt.Errorf("error setting deprovisioning operation %s as succeeded: %s", lastOp.ID, dberr.Error())
	}

	err := s.directorClient.DeleteRuntime(cluster.ID, cluster.Tenant)
	if err != nil {
		return fmt.Errorf("error deleting Runtime form Director: %s", err.Error())
	}

	dberr = session.Commit()
	if dberr != nil {
		return fmt.Errorf("error commiting transaction: %s", dberr.Error())
	}

	return nil
}

func (s *WaitForClusterDeletionStep) deleteShoot(cluster model.Cluster) error {
	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		// Non recoverable error?
		return errors.New("failed to read GardenerConfig")
	}

	// TODO: check how Delete bahaves when deleting non-existent object
	_, err := s.gardenerClient.Get(gardenerConfig.Name, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return s.gardenerClient.Delete(gardenerConfig.Name, &v1.DeleteOptions{})
}
