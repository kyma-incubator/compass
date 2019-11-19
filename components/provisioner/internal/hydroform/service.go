package hydroform

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/configuration"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Service
type Service interface {
	ProvisionCluster(builder configuration.Builder) (ClusterInfo, error)
	DeprovisionCluster(builder configuration.Builder, terraformState string) error
}

type service struct {
	client client.Client
}

func NewHydroformService(client client.Client) Service {
	return &service{
		client: client,
	}
}

type ClusterInfo struct {
	ClusterStatus types.Phase
	KubeConfig    string
	State         string
}

func (s service) ProvisionCluster(builder configuration.Builder) (ClusterInfo, error) {
	log.Info("Preparing config for runtime provisioning")
	cluster, provider, err := builder.Create()
	defer builder.CleanUp()
	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "Config preparation failed")
	}

	log.Infof("Starting cluster provisioning")

	cluster, err = s.client.Provision(cluster, provider)
	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "Cluster provisioning failed")
	}

	status, err := s.client.Status(cluster, provider)
	if err != nil {
		return ClusterInfo{}, err
	}

	log.Info("Retrieving kubeconfig")

	kubeconfig, err := s.client.Credentials(cluster, provider)
	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "Failed to get kubeconfig")
	}

	log.Info("Retrieving cluster state")

	internalState, err := util.EncodeJson(cluster.ClusterInfo.InternalState)

	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "Failed to retrieve cluster state")
	}

	log.Infof("Cluster state: %+v", internalState)

	return ClusterInfo{
		ClusterStatus: status.Phase,
		State:         internalState,
		KubeConfig:    string(kubeconfig),
	}, nil
}

func (s service) DeprovisionCluster(builder configuration.Builder, terraformStateJson string) error {
	log.Info("Preparing config for runtime deprovisioning")
	cluster, provider, err := builder.Create()

	if err != nil {
		return errors.Wrap(err, "Config preparation failed")
	}

	var state types.InternalState

	err = util.DecodeJson(terraformStateJson, &state)

	if err != nil {
		return errors.Wrap(err, "Config preparation failed")
	}

	cluster.ClusterInfo = &types.ClusterInfo{InternalState: &state}

	log.Infof("Starting cluster deprovisioning")
	return s.client.Deprovision(cluster, provider)
}
