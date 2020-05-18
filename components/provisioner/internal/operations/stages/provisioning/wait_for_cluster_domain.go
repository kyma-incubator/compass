package provisioning

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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WaitForClusterDomainStep struct {
	gardenerClient GardenerClient
	dbsFactory     dbsession.Factory
	directorClient director.DirectorClient
	nextStep       model.OperationStage
	timeLimit      time.Duration
}

//go:generate mockery -name=GardenerClient
type GardenerClient interface {
	Get(name string, options v1.GetOptions) (*gardener_types.Shoot, error)
}

func NewWaitForClusterDomainStep(gardenerClient GardenerClient, directorClient director.DirectorClient, nextStep model.OperationStage, timeLimit time.Duration) *WaitForClusterDomainStep {
	return &WaitForClusterDomainStep{
		gardenerClient: gardenerClient,
		directorClient: directorClient,
		nextStep:       nextStep,
		timeLimit:      timeLimit,
	}
}

func (s *WaitForClusterDomainStep) Name() model.OperationStage {
	return model.WaitingForClusterDomain
}

func (s *WaitForClusterDomainStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForClusterDomainStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		log.Error("Error converting to GardenerConfig")
		err := errors.New("failed to convert to GardenerConfig")
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	shoot, err := s.gardenerClient.Get(gardenerConfig.Name, v1.GetOptions{})
	if err != nil {
		log.Errorf("Error getting Gardener Shoot: %s", err.Error())
		return operations.StageResult{}, err
	}

	if shoot.Spec.DNS == nil || shoot.Spec.DNS.Domain == nil {
		log.Warnf("DNS Domain is not set yet for runtime ID: %s", cluster.ID)
		return operations.StageResult{Stage: s.Name(), Delay: 5 * time.Second}, nil
	}

	// TODO: Consider updating Labels and StatusCondition separately without getting the Runtime
	//       It'll be possible after this issue implementation:
	//       - https://github.com/kyma-incubator/compass/issues/1186
	runtimeInput, err := s.prepareProvisioningUpdateRuntimeInput(cluster.ID, cluster.Tenant, shoot)
	if err != nil {
		log.Errorf("Error preparing Runtime Input: %s", err.Error())
		return operations.StageResult{}, err
	}
	if err := s.directorClient.UpdateRuntime(cluster.ID, runtimeInput, cluster.Tenant); err != nil {
		log.Errorf("Error updating Runtime in Director: %s", err.Error())
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}

func (s *WaitForClusterDomainStep) prepareProvisioningUpdateRuntimeInput(runtimeId, tenant string, shoot *gardener_types.Shoot) (*graphql.RuntimeInput, error) {

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
