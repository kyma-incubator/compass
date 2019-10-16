package hydroform

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"

	"strconv"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	hf "github.com/kyma-incubator/hydroform"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	credentialsFileFmt = "credentials-%s.yaml"
)

//go:generate mockery -name=Client
type Client interface {
	ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (ClusterInfo, error)
	DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string) error
}

type client struct {
	secrets v1.SecretInterface
}

func NewHydroformClient(secrets v1.SecretInterface) Client {
	return &client{secrets: secrets}
}

type ClusterInfo struct {
	ClusterStatus types.Phase
	KubeConfig    string
	State         string
}

func (c client) ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (ClusterInfo, error) {
	credentialsFile := generateRandomFileName()

	log.Infof("Saving credentials to file %s", credentialsFile)
	err := c.saveCredentialsToFile(secretName, credentialsFile)

	if err != nil {
		log.Errorf("Failed to save credentials to file %s: %s", credentialsFile, err.Error())
		return ClusterInfo{}, err
	}

	log.Info("Preparing config for runtime provisioning")

	cluster, provider, err := c.prepareConfig(runtimeConfig, credentialsFile)

	if err != nil {
		log.Errorf("Config preparation failed: %s", err.Error())
		return ClusterInfo{}, err
	}

	log.Infof("Config prepared - cluster: %s, provider: %s. Starting cluster provisioning", cluster, provider)

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

	log.Info("Restrieving cluster state")

	internalState, err := stateToJson(cluster.ClusterInfo.InternalState)

	if err != nil {
		log.Errorf("Failed to retrieve cluster state: %s", err.Error())
		return ClusterInfo{}, err
	}

	return ClusterInfo{
		ClusterStatus: status.Phase,
		State:         internalState,
		KubeConfig:    string(kubeconfig),
	}, nil
}

func stateToJson(state *types.InternalState) (string, error) {
	bytes, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (c client) DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string) error {
	credentialsFile := generateRandomFileName()

	c.saveCredentialsToFile(secretName, credentialsFile)

	cluster, provider, err := c.prepareConfig(runtimeConfig, credentialsFile)

	if err != nil {
		return err
	}

	return hf.Deprovision(cluster, provider)
}

func (c client) saveCredentialsToFile(secretName string, filename string) error {
	secret, err := c.secrets.Get(secretName, meta.GetOptions{})

	if err != nil {
		return err
	}

	bytes, ok := secret.Data["kubeconfig"]

	if !ok {
		return errors.New("kubeconfig not found within the secret")
	}

	err = ioutil.WriteFile(filename, bytes, os.ModePerm)

	if err != nil {
		return err
	}

	return os.Chmod(filename, os.ModePerm)
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

func generateRandomFileName() string {
	return fmt.Sprintf(credentialsFileFmt, rand.String(6))
}
