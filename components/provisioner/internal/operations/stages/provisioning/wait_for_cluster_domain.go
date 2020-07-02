package provisioning

import (
	"fmt"
	"time"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/control-plane/components/provisioner/internal/director"
	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/internal/operations"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WaitForClusterDomainStep struct {
	gardenerClient GardenerClient
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

	shoot, err := s.gardenerClient.Get(cluster.ClusterConfig.Name, v1.GetOptions{})
	if err != nil {
		return operations.StageResult{}, err
	}

	if shoot.Spec.DNS == nil || shoot.Spec.DNS.Domain == nil {
		log.Warnf("DNS Domain is not set yet for runtime ID: %s", cluster.ID)
		return operations.StageResult{Stage: s.Name(), Delay: 5 * time.Second}, nil
	}

	// TODO: Consider updating Labels and StatusCondition separately without getting the Runtime
	//       It'll be possible after this issue implementation:
	//       - https://github.com/kyma-project/control-plane/issues/1186
	runtimeInput, err := s.prepareProvisioningUpdateRuntimeInput(cluster.ID, cluster.Tenant, shoot)
	if err != nil {
		return operations.StageResult{}, err
	}
	if err := s.directorClient.UpdateRuntime(cluster.ID, runtimeInput, cluster.Tenant); err != nil {
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
