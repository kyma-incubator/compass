package deprovisioning

import (
	"errors"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeprovisionClusterStep struct {
	gardenerClient GardenerClient
	dbsFactory     dbsession.Factory
	directorClient director.DirectorClient
	nextStep       model.OperationStage
	timeLimit      time.Duration
}

//go:generate mockery -name=GardenerClient
type GardenerClient interface {
	Delete(name string, options *v1.DeleteOptions) error
}

func NewDeprovisionClusterStep(gardenerClient GardenerClient, dbsFactory dbsession.Factory, directorClient director.DirectorClient, nextStep model.OperationStage, timeLimit time.Duration) *DeprovisionClusterStep {
	return &DeprovisionClusterStep{
		gardenerClient: gardenerClient,
		dbsFactory:     dbsFactory,
		directorClient: directorClient,
		nextStep:       nextStep,
		timeLimit:      timeLimit,
	}
}

func (s *DeprovisionClusterStep) Name() model.OperationStage {
	return model.DeprovisionCluster
}

func (s *DeprovisionClusterStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *DeprovisionClusterStep) Run(cluster model.Cluster, operation model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		err := errors.New("failed to read GardenerConfig")
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	err := s.deleteShoot(gardenerConfig.Name, logger)
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
}

func (s *DeprovisionClusterStep) deleteShoot(gardenerClusterName string, logger logrus.FieldLogger) error {
	err := s.gardenerClient.Delete(gardenerClusterName, &v1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
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

func (s *DeprovisionClusterStep) deleteRuntime(cluster model.Cluster, gardenerClusterName string) error {
	exists, err := s.directorClient.RuntimeExists(gardenerClusterName, cluster.Tenant)
	if err != nil {
		return fmt.Errorf("error checking Runtime exists in Director: %s", err.Error())
	}

	if !exists {
		return nil
	}

	err = s.directorClient.DeleteRuntime(cluster.ID, cluster.Tenant)
	if err != nil {
		return fmt.Errorf("error deleting Runtime form Director: %s", err.Error())
	}

	return nil
}
