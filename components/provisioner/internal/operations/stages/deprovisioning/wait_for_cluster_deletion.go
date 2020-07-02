package deprovisioning

import (
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

type WaitForClusterDeletionStep struct {
	gardenerClient GardenerClient
	dbsFactory     dbsession.Factory
	directorClient director.DirectorClient
	nextStep       model.OperationStage
	timeLimit      time.Duration
}

func NewWaitForClusterDeletionStep(gardenerClient GardenerClient, dbsFactory dbsession.Factory, directorClient director.DirectorClient, nextStep model.OperationStage, timeLimit time.Duration) *WaitForClusterDeletionStep {
	return &WaitForClusterDeletionStep{
		gardenerClient: gardenerClient,
		dbsFactory:     dbsFactory,
		directorClient: directorClient,
		nextStep:       nextStep,
		timeLimit:      timeLimit,
	}
}

func (s *WaitForClusterDeletionStep) Name() model.OperationStage {
	return model.WaitForClusterDeletion
}

func (s *WaitForClusterDeletionStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForClusterDeletionStep) Run(cluster model.Cluster, operation model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	shootExists, err := s.shootExists(cluster.ClusterConfig.Name, logger)
	if err != nil {
		return operations.StageResult{}, err
	}

	if shootExists {
		return operations.StageResult{Stage: s.Name(), Delay: 20 * time.Second}, nil
	}

	err = s.setDeprovisioningFinished(cluster, operation)
	if err != nil {
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}

func (s *WaitForClusterDeletionStep) shootExists(gardenerClusterName string, logger logrus.FieldLogger) (bool, error) {
	_, err := s.gardenerClient.Get(gardenerClusterName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
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

	err := s.deleteRuntime(cluster)
	if err != nil {
		return err
	}

	dberr = session.Commit()
	if dberr != nil {
		return fmt.Errorf("error commiting transaction: %s", dberr.Error())
	}

	return nil
}

func (s *WaitForClusterDeletionStep) deleteRuntime(cluster model.Cluster) error {
	exists, err := s.directorClient.RuntimeExists(cluster.ID, cluster.Tenant)
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
