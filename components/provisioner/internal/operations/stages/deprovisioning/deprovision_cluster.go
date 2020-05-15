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

type DeprovisionClusterStep struct {
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

func NewDeprovisionClusterStep(installationClient installation.Service, gardenerClient GardenerClient, dbsFactory dbsession.Factory, directorClient director.DirectorClient, nextStep model.OperationStage, timeLimit time.Duration) *DeprovisionClusterStep {
	return &DeprovisionClusterStep{
		installationClient: installationClient,
		gardenerClient:     gardenerClient,
		dbsFactory:         dbsFactory,
		directorClient:     directorClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *DeprovisionClusterStep) Name() model.OperationStage {
	return model.DeprovisionCluster
}

func (s *DeprovisionClusterStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *DeprovisionClusterStep) Run(cluster model.Cluster, operation model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	// TODO: is the case with Succeeded state possible?
	if operation.State == model.InProgress || operation.State == model.Succeeded {

		gardenerConfig, ok := cluster.GardenerConfig()
		if !ok {
			// Non recoverable error?
			return operations.StageResult{}, errors.New("failed to read GardenerConfig")
		}

		err := s.deleteShoot(cluster, logger)
		if err != nil {
			logger.Errorf("Error deleting shoot: %s", err.Error())
			return operations.StageResult{}, err
		}

		err = s.setDeprovisioningFinished(cluster, gardenerConfig.Name, operation)
		if err != nil {
			logger.Errorf("Error setting deprovisioning finished: %s", err.Error())
			return operations.StageResult{}, err
		}

		return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
	} else {
		err := fmt.Errorf("invalid state occurred while deprovisioning, last operation: %s, state: %s", operation.Type, operation.State)
		logger.Errorf("Invalid ", err.Error())
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	return operations.StageResult{Stage: s.Name(), Delay: 30 % time.Second}, nil
}

func (s *DeprovisionClusterStep) setDeprovisioningFinished(cluster model.Cluster, gardenerClusterName string, lastOp model.Operation) error {
	session, dberr := s.dbsFactory.NewSessionWithinTransaction()
	if dberr != nil {
		return fmt.Errorf("error starting db session with transaction: %s", dberr.Error())
	}
	defer session.RollbackUnlessCommitted()

	dberr = session.MarkClusterAsDeleted(cluster.ID)
	if dberr != nil {
		return fmt.Errorf("error marking cluster for deletion: %s", dberr.Error())
	}

	err := s.deleteRuntime(cluster, gardenerClusterName)
	if err != nil {
		return err
	}

	dberr = session.Commit()
	if dberr != nil {
		return fmt.Errorf("error commiting transaction: %s", dberr.Error())
	}

	return nil
}

func (s *DeprovisionClusterStep) deleteShoot(cluster model.Cluster, logger logrus.FieldLogger) error {
	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		logger.Error("Error converting to GardenerConfig")
		err := errors.New("failed to convert to GardenerConfig")
		return operations.NewNonRecoverableError(err)
	}

	// TODO: check how Delete behaves when deleting non-existent object
	_, err := s.gardenerClient.Get(gardenerConfig.Name, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return s.gardenerClient.Delete(gardenerConfig.Name, &v1.DeleteOptions{})
}

func (s *DeprovisionClusterStep) deleteRuntime(cluster model.Cluster, gardenerClusterName string) error {
	exists, err := s.directorClient.RuntimeExists(gardenerClusterName, cluster.Tenant)
	if !exists {
		return nil
	}

	err = s.directorClient.DeleteRuntime(cluster.ID, cluster.Tenant)
	if err != nil {
		return fmt.Errorf("error deleting Runtime form Director: %s", err.Error())
	}

	return nil
}
