package hydroform

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeout  = 20 * time.Minute
	interval = 30 * time.Second
)

const credentialsKey = "credentials"

//go:generate mockery -name=Service
type Service interface {
	ProvisionCluster(clusterConfig model.Cluster) (ClusterInfo, error)
	DeprovisionCluster(clusterConfig model.Cluster) error
}

type service struct {
	hydroformClient client.Client
	secretsClient   v1.SecretInterface
}

func NewHydroformService(hydroformClient client.Client, secretsClient v1.SecretInterface) Service {
	return &service{
		hydroformClient: hydroformClient,
		secretsClient:   secretsClient,
	}
}

type ClusterInfo struct {
	ClusterStatus types.Phase
	KubeConfig    string
	State         string
}

func (s service) ProvisionCluster(clusterData model.Cluster) (ClusterInfo, error) {
	log.Infof("Preparing config for %s runtime provisioning", clusterData.ID)
	credentialsFile, err := s.saveCredentialsToFile(clusterData.CredentialsSecretName)
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Failed to save credentials to secret for %s Runtime", clusterData.ID)
	}
	defer removeFile(credentialsFile)

	cluster, provider := clusterData.ClusterConfig.ToHydroformConfiguration(credentialsFile)

	log.Infof("Starting %s Runtime provisioning", clusterData.ID)
	cluster, err = s.hydroformClient.Provision(cluster, provider)
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Cluster %s provisioning failed", clusterData.ID)
	}

	var status *types.ClusterStatus

	//TODO Change this temporary solution when Hydroform handles Provisioning status correctly
	err = util.WaitForFunction(interval, timeout, func() (bool, error) {
		status, err = s.hydroformClient.Status(cluster, provider)
		if err != nil {
			return false, err
		}
		return status.Phase == types.Provisioned, nil
	})
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Unexpected status for %s Runtime", clusterData.ID)
	}

	log.Infof("Retrieving kubeconfig for %s Runtime", clusterData.ID)

	kubeconfig, err := s.hydroformClient.Credentials(cluster, provider)
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Failed to get kubeconfig for %s Runtime", clusterData.ID)
	}

	internalState, err := util.EncodeJson(cluster.ClusterInfo.InternalState)
	if err != nil {
		return ClusterInfo{}, errors.WithMessagef(err, "Failed to encode Terraform state for %s Runtime", clusterData.ID)
	}

	return ClusterInfo{
		ClusterStatus: status.Phase,
		State:         internalState,
		KubeConfig:    string(kubeconfig),
	}, nil
}

func (s service) DeprovisionCluster(clusterData model.Cluster) error {
	log.Infof("Preparing config for %s runtime deprovisioning", clusterData.ID)
	credentialsFile, err := s.saveCredentialsToFile(clusterData.CredentialsSecretName)
	if err != nil {
		return errors.WithMessagef(err, "Failed to save credentials to secret for %s Runtime", clusterData.ID)
	}
	defer removeFile(credentialsFile)

	cluster, provider := clusterData.ClusterConfig.ToHydroformConfiguration(credentialsFile)

	var state types.InternalState
	err = util.DecodeJson(clusterData.TerraformState, &state)
	if err != nil {
		return errors.WithMessagef(err, "Failed to decode Terraform state for %s Runtime", clusterData.ID)
	}

	cluster.ClusterInfo = &types.ClusterInfo{InternalState: &state}

	log.Infof("Starting deprovisioning of %s Runtime", clusterData.ID)
	return s.hydroformClient.Deprovision(cluster, provider)
}

func (s service) saveCredentialsToFile(secretName string) (string, error) {
	secret, err := s.secretsClient.Get(secretName, meta.GetOptions{})
	if err != nil {
		return "", errors.WithMessagef(err, "Failed to get credentials from %s secret", secretName)
	}

	bytes, ok := secret.Data[credentialsKey]
	if !ok {
		return "", errors.New("Credentials not found within the secret")
	}

	tempFile, err := ioutil.TempFile("", secretName)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create credentials file")
	}

	_, err = tempFile.Write(bytes)
	if err != nil {
		return "", errors.WithMessagef(err, "Failed to save credentials to %s file", tempFile.Name())
	}

	return tempFile.Name(), nil
}

func removeFile(fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		log.Errorf("Error while removing temporary credentials file %s: %s", fileName, err.Error())
	}
}
