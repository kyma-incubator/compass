package gardener

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
)

func NewProvisioner(
	namespace string,
	shootClient gardener_apis.ShootInterface,
	factory dbsession.Factory) *GardenerProvisioner {
	return &GardenerProvisioner{
		namespace:        namespace,
		shootClient:      shootClient,
		dbSessionFactory: factory,
	}
}

type GardenerProvisioner struct {
	namespace        string
	shootClient      gardener_apis.ShootInterface
	dbSessionFactory dbsession.Factory
	directorService  director.DirectorClient
}

func (g *GardenerProvisioner) ProvisionCluster(cluster model.Cluster, operationId string) error {
	shootTemplate, err := cluster.ClusterConfig.ToShootTemplate(g.namespace, cluster.Tenant, cluster.SubAccountId)
	if err != nil {
		return fmt.Errorf("failed to convert cluster config to Shoot template")
	}

	annotate(shootTemplate, operationIdAnnotation, operationId)
	annotate(shootTemplate, provisioningAnnotation, Provisioning.String())
	annotate(shootTemplate, runtimeIdAnnotation, cluster.ID)

	_, err = g.shootClient.Create(shootTemplate)
	if err != nil {
		return fmt.Errorf("error creating Shoot for %s cluster: %s", cluster.ID, err.Error())
	}

	return nil
}

func (g *GardenerProvisioner) DeprovisionCluster(cluster model.Cluster, operationId string) (model.Operation, error) {
	session := g.dbSessionFactory.NewWriteSession()

	gardenerCfg, ok := cluster.GardenerConfig()
	if !ok {
		return model.Operation{}, fmt.Errorf("cluster does not have Gardener configuration")
	}

	shoot, err := g.shootClient.Get(gardenerCfg.Name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			message := fmt.Sprintf("Cluster %s does not exist. Nothing to deprovision.", cluster.ID)

			dberr := session.MarkClusterAsDeleted(cluster.ID)
			if dberr != nil {
				return newDeprovisionOperation(operationId, cluster.ID, message, model.Failed, model.FinishedStage, time.Now()), dberr
			}
			return newDeprovisionOperation(operationId, cluster.ID, message, model.Succeeded, model.FinishedStage, time.Now()), nil
		}
	}

	if shoot.DeletionTimestamp != nil {
		message := fmt.Sprintf("Cluster %s already %s scheduled for deletion.", gardenerCfg.Name, cluster.ID)
		return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, model.DeprovisioningStage, shoot.DeletionTimestamp.Time), nil
	}

	deletionTime := time.Now()

	// TODO: consider adding some annotation and uninstall before deleting shoot
	annotate(shoot, provisioningAnnotation, Deprovisioning.String())
	annotate(shoot, operationIdAnnotation, operationId)
	AnnotateWithConfirmDeletion(shoot)
	err = UpdateAndDeleteShoot(g.shootClient, shoot)
	if err != nil {
		return model.Operation{}, fmt.Errorf("error scheduling shoot %s for deletion: %s", shoot.Name, err.Error())
	}

	message := fmt.Sprintf("Deprovisioning started")
	return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, model.DeprovisioningStage, deletionTime), nil
}

func newDeprovisionOperation(id, runtimeId, message string, state model.OperationState, stage model.OperationStage, startTime time.Time) model.Operation {
	return model.Operation{
		ID:             id,
		Type:           model.Deprovision,
		StartTimestamp: startTime,
		State:          state,
		Stage:          stage,
		Message:        message,
		ClusterID:      runtimeId,
	}
}
