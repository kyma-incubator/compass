package stages

import (
	"fmt"
	"time"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
)

type WaitForClusterCreationStep struct {
	gardenerClient GardenerClient
	dbsFactory     dbsession.Factory
	directorClient director.DirectorClient
	nextStep       model.OperationStage
	timeLimit      time.Duration
}

type GardenerClient interface {
	Get(name string) (gardener_types.Shoot, error)
}

func NewWaitForClusterCreationStepStep(gardenerClient GardenerClient, dbsFactory dbsession.Factory, directorClient director.DirectorClient, nextStep model.OperationStage, timeLimit time.Duration) *WaitForClusterCreationStep {
	return &WaitForClusterCreationStep{
		gardenerClient: gardenerClient,
		dbsFactory:     dbsFactory,
		directorClient: directorClient,
		nextStep:       nextStep,
		timeLimit:      timeLimit,
	}
}

func (s *WaitForClusterCreationStep) Name() model.OperationStage {
	return model.StartingInstallation
}

func (s *WaitForClusterCreationStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForClusterCreationStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		// Non recoverable error?
		return operations.StageResult{}, errors.New("failed to read GardenerConfig")
	}

	shoot, err := s.gardenerClient.Get(gardenerConfig.Name)
	if err != nil {
		return operations.StageResult{}, err
	}

	if shoot.Spec.DNS == nil || shoot.Spec.DNS.Domain == nil {
		log.Warnf("DNS Domain is not set yet for runtime ID: %s", cluster.ID)
		return operations.StageResult{Stage: s.Name(), Delay: 30 * time.Second}, nil
	}

	log.Infof("Updating Runtime in Director with Gardener labels and the status...")
	tenant, dberr := s.dbsFactory.NewReadSession().GetTenant(cluster.ID)
	if dberr != nil {
		log.Errorf("Error getting Gardener cluster by name: %s", dberr.Error())
		return operations.StageResult{}, dberr
	}

	// TODO: Consider updating Labels and StatusCondition separately without getting the Runtime
	//       It'll be possible after this issue implementation:
	//       - https://github.com/kyma-incubator/compass/issues/1186
	runtimeInput, err := s.prepareProvisioningUpdateRuntimeInput(cluster.ID, tenant, shoot)
	if err != nil {
		log.Errorf("Error preparing Runtime Input: %s", err.Error())
		return operations.StageResult{}, err
	}
	if err := s.directorClient.UpdateRuntime(cluster.ID, runtimeInput, tenant); err != nil {
		log.Errorf("Error updating Runtime in Director: %s", err.Error())
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 30 * time.Second}, nil
}

func (s *WaitForClusterCreationStep) prepareProvisioningUpdateRuntimeInput(
	runtimeId, tenant string, shoot gardener_types.Shoot) (*graphql.RuntimeInput, error) {

	runtime, err := s.directorClient.GetRuntime(runtimeId, tenant)
	if err != nil {
		return &graphql.RuntimeInput{}, errors.Wrap(err, fmt.Sprintf("failed to get Runtime by ID: %s", runtimeId))
	}

	if runtime.Labels == nil {
		runtime.Labels = graphql.Labels{}
	}
	runtime.Labels["gardenerClusterName"] = shoot.ObjectMeta.Name
	runtime.Labels["gardenerClusterDomain"] = *shoot.Spec.DNS.Domain
	statusCondition := graphql.RuntimeStatusConditionProvisioning

	runtimeInput := &graphql.RuntimeInput{
		Name:            runtime.Name,
		Description:     runtime.Description,
		Labels:          &runtime.Labels,
		StatusCondition: &statusCondition,
	}
	return runtimeInput, nil
}
