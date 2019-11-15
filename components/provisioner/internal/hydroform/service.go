package hydroform

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"io/ioutil"
	"os"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

const credentialsKey = "credentials"

//go:generate mockery -name=Service
type Service interface {
	ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (ClusterInfo, error)
	DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string, terraformState string) error
}

type service struct {
	secrets v1.SecretInterface
	client  client.Client
}

func NewHydroformService(secrets v1.SecretInterface, client client.Client) Service {
	return &service{
		secrets: secrets,
		client:  client,
	}
}

type ClusterInfo struct {
	ClusterStatus types.Phase
	KubeConfig    string
	State         string
}

func (s service) ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (ClusterInfo, error) {
	credentialsFileName, err := s.saveCredentialsToFile(secretName)
	if err != nil {
		return ClusterInfo{}, err
	}
	defer removeFile(credentialsFileName)

	log.Info("Preparing config for runtime provisioning")
	cluster, provider, err := prepareConfig(runtimeConfig, credentialsFileName)
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

func (s service) DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string, terraformStateJson string) error {
	credentialsFileName, err := s.saveCredentialsToFile(secretName)
	if err != nil {
		return err
	}
	defer removeFile(credentialsFileName)

	log.Info("Preparing config for runtime deprovisioning")
	cluster, provider, err := prepareConfig(runtimeConfig, credentialsFileName)

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

func (s service) saveCredentialsToFile(secretName string) (string, error) {
	secret, err := s.secrets.Get(secretName, meta.GetOptions{})
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
