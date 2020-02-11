package hydroform

import (
	"bytes"
	"time"

	"github.com/hashicorp/terraform/states/statefile"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client"
	log "github.com/sirupsen/logrus"

	"github.com/avast/retry-go"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
)

const (
	clusterCreationTimeout = 40 * time.Minute
)

const credentialsKey = "credentials"

//go:generate mockery -name=Service
type Service interface {
	ProvisionCluster(clusterConfig model.Cluster) (ClusterInfo, error)
	DeprovisionCluster(clusterConfig model.Cluster) error
}

type service struct {
	hydroformClient        client.Client
	gardenerKubeconfigPath string
}

func NewHydroformService(hydroformClient client.Client, gardenerKubeconfigPath string) Service {
	return &service{
		hydroformClient:        hydroformClient,
		gardenerKubeconfigPath: gardenerKubeconfigPath,
	}
}

type ClusterInfo struct {
	ClusterStatus types.Phase
	KubeConfig    string
	State         []byte
}

func (s service) ProvisionCluster(clusterData model.Cluster) (ClusterInfo, error) {
	log.Infof("Preparing config for %s Runtime provisioning", clusterData.ID)

	cluster, provider, err := clusterData.ClusterConfig.ToHydroformConfiguration(s.gardenerKubeconfigPath)
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Failed to convert  Provider config to Hydroform config for %s Runtime: %s", clusterData.ID, err.Error())
	}

	log.Infof("Starting %s Runtime provisioning", clusterData.ID)
	err = retry.Do(
		func() error {
			cluster, err = s.hydroformClient.Provision(cluster, provider, types.WithTimeouts(&types.Timeouts{Create: clusterCreationTimeout}))
			return err
		},
		retry.Attempts(3))

	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Cluster %s provisioning failed", clusterData.ID)
	}

	log.Infof("Cluster provisioned for %s Runtime", clusterData.ID)

	status, err := s.hydroformClient.Status(cluster, provider)
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Failed to get cluster status for %s Runtime", clusterData.ID)
	}
	if status.Phase != types.Provisioned {
		return ClusterInfo{}, errors.Errorf("Unexpected cluster status for %s Runtime, status: %s", clusterData.ID, status.Phase)
	}

	log.Infof("Retrieving kubeconfig for %s Runtime", clusterData.ID)

	kubeconfig, err := s.hydroformClient.Credentials(cluster, provider)
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Failed to get kubeconfig for %s Runtime", clusterData.ID)
	}

	var buffer bytes.Buffer
	err = statefile.Write(cluster.ClusterInfo.InternalState.TerraformState, &buffer)
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Failed to write Terraform state file for %s Runtime", clusterData.ID)
	}

	return ClusterInfo{
		ClusterStatus: status.Phase,
		State:         buffer.Bytes(),
		KubeConfig:    string(kubeconfig),
	}, nil
}

func (s service) DeprovisionCluster(clusterData model.Cluster) error {
	log.Infof("Preparing config for %s Runtime deprovisioning", clusterData.ID)

	cluster, provider, err := clusterData.ClusterConfig.ToHydroformConfiguration(s.gardenerKubeconfigPath)
	if err != nil {
		return errors.WithMessagef(err, "Failed to convert Provider config to Hydroform %s Runtime: %s", clusterData.ID, err.Error())
	}

	reader := bytes.NewReader(clusterData.TerraformState)
	stateFile, err := statefile.Read(reader)
	if err != nil {
		return errors.WithMessagef(err, "Failed to read Terraform state from file for %s Runtime", clusterData.ID)
	}

	internalState := types.InternalState{TerraformState: stateFile}
	cluster.ClusterInfo = &types.ClusterInfo{InternalState: &internalState}

	log.Infof("Starting deprovisioning of %s Runtime", clusterData.ID)
	return s.hydroformClient.Deprovision(cluster, provider)
}
