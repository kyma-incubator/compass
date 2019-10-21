package hydroform

import (
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"strconv"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	hf "github.com/kyma-incubator/hydroform"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

//go:generate mockery -name=Service
type Service interface {
	ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (ClusterInfo, error)
	DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string, terraformState string) error
}

type client struct {
	secrets v1.SecretInterface
}

func NewHydroformClient(secrets v1.SecretInterface) Service {
	return &client{secrets: secrets}
}

type ClusterInfo struct {
	ClusterStatus types.Phase
	KubeConfig    string
	State         string
}

func (c client) ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (ClusterInfo, error) {
	credentialsFileName, err := c.saveCredentialsToFile(secretName)
	if err != nil {
		return ClusterInfo{}, err
	}
	defer removeFile(credentialsFileName)

	log.Info("Preparing config for runtime provisioning")
	cluster, provider, err := c.prepareConfig(runtimeConfig, credentialsFileName)
	if err != nil {
		log.Errorf("Config preparation failed: %s", err.Error())
		return ClusterInfo{}, err
	}

	log.Infof("Config prepared - cluster: %+v, provider: %+v. Starting cluster provisioning", cluster, provider)

	cluster, err = hf.Provision(cluster, provider)
	if err != nil {
		log.Errorf("Cluster provisioning failed: %s", err.Error())
		return ClusterInfo{}, err
	}

	status, err := hf.Status(cluster, provider)
	if err != nil {
		return ClusterInfo{}, err
	}

	log.Info("Retrieving kubeconfig")

	kubeconfig, err := hf.Credentials(cluster, provider)
	if err != nil {
		log.Errorf("Failed to get kubeconfig: %s", err.Error())
		return ClusterInfo{}, err
	}

	log.Info("Retrieving cluster state")

	internalState, err := stateToJson(cluster.ClusterInfo.InternalState)

	if err != nil {
		log.Errorf("Failed to retrieve cluster state: %s", err.Error())
		return ClusterInfo{}, err
	}

	log.Infof("Cluster state: %+v", internalState)

	return ClusterInfo{
		ClusterStatus: status.Phase,
		State:         internalState,
		KubeConfig:    string(kubeconfig),
	}, nil
}

func (c client) DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string, terraformState string) error {
	credentialsFileName, err := c.saveCredentialsToFile(secretName)
	if err != nil {
		return err
	}
	defer removeFile(credentialsFileName)

	log.Info("Preparing config for runtime deprovisioning")
	cluster, provider, err := c.prepareConfig(runtimeConfig, credentialsFileName)

	if err != nil {
		log.Errorf("Config preparation failed: %s", err.Error())
		return err
	}

	state, err := jsonToState(terraformState)

	if err != nil {
		log.Errorf("Config preparation failed: %s", err.Error())
		return err
	}

	cluster.ClusterInfo = &types.ClusterInfo{InternalState: state}

	log.Infof("Config prepared - cluster: %+v, provider: %+v. Starting cluster deprovisioning", cluster, provider)
	return hf.Deprovision(cluster, provider)
}

func jsonToState(state string) (*types.InternalState, error) {
	var terraformState types.InternalState

	err := json.Unmarshal([]byte(state), &terraformState)

	if err != nil {
		return &types.InternalState{}, err
	}

	return &terraformState, nil
}

func stateToJson(state *types.InternalState) (string, error) {
	bytes, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (c client) saveCredentialsToFile(secretName string) (string, error) {
	secret, err := c.secrets.Get(secretName, meta.GetOptions{})
	if err != nil {
		return "", errors.WithMessagef(err, "Failed to get credentials from %s secret", secretName)
	}

	bytes, ok := secret.Data["kubeconfig"]

	if !ok {
		return "", errors.New("kubeconfig not found within the secret")
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

func (c client) prepareConfig(input model.RuntimeConfig, credentialsFile string) (*types.Cluster, *types.Provider, error) {
	gardenerConfig, ok := input.GardenerConfig()
	if ok {
		return buildConfigForGardener(gardenerConfig, credentialsFile)
	}

	gcpConfig, ok := input.GCPConfig()
	if ok {
		return buildConfigForGCP(gcpConfig, credentialsFile)
	}

	return nil, nil, errors.New("configuration does not match any provider profiles")
}

func buildConfigForGCP(config model.GCPConfig, credentialsFile string) (*types.Cluster, *types.Provider, error) {
	diskSize, err := strconv.Atoi(config.BootDiskSize)

	if err != nil {
		return &types.Cluster{}, &types.Provider{}, err
	}

	cluster := &types.Cluster{
		KubernetesVersion: config.KubernetesVersion,
		Name:              config.Name,
		DiskSizeGB:        diskSize,
		NodeCount:         config.NumberOfNodes,
		Location:          config.Region,
		MachineType:       config.MachineType,
	}

	provider := &types.Provider{
		Type:                types.GCP,
		ProjectName:         config.ProjectName,
		CredentialsFilePath: credentialsFile,
	}
	return cluster, provider, nil
}

func buildConfigForGardener(config model.GardenerConfig, credentialsFile string) (*types.Cluster, *types.Provider, error) {
	diskSize, err := strconv.Atoi(config.VolumeSize)

	if err != nil {
		return &types.Cluster{}, &types.Provider{}, err
	}

	cluster := &types.Cluster{
		KubernetesVersion: config.KubernetesVersion,
		Name:              config.Name,
		DiskSizeGB:        diskSize,
		NodeCount:         config.NodeCount,
		Location:          config.Region,
		MachineType:       config.MachineType,
	}

	provider := &types.Provider{
		Type:                types.Gardener,
		ProjectName:         config.ProjectName,
		CredentialsFilePath: credentialsFile,
		CustomConfigurations: map[string]interface{}{
			"target_provider": config.TargetProvider,
			"target_secret":   config.TargetSecret,
			"disk_type":       config.DiskType,
			"zone":            config.Zone,
			"cidr":            config.Cidr,
			"autoscaler_min":  config.AutoScalerMin,
			"autoscaler_max":  config.AutoScalerMax,
			"max_surge":       config.MaxSurge,
			"max_unavailable": config.MaxUnavailable,
		},
	}
	return cluster, provider, nil
}
