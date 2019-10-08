package hydroform

import (
	"io/ioutil"

	"strconv"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	hf "github.com/kyma-incubator/hydroform"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	kubeconfig      = "kubeconfig.yaml"
	credentialsFile = "credentials.yaml"
)

type Client interface {
	ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (types.ClusterStatus, error)
	DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string) error
}

type client struct {
	secrets v1.SecretInterface
}

func NewHydroformClient(secrets v1.SecretInterface) Client {
	return &client{secrets: secrets}
}

func (c client) ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (types.ClusterStatus, error) {
	err := c.saveCredentialsToFile(secretName, credentialsFile)

	if err != nil {
		return types.ClusterStatus{}, err
	}

	cluster, provider, err := c.prepareConfig(runtimeConfig)

	if err != nil {
		return types.ClusterStatus{}, err
	}

	cluster, err = hf.Provision(cluster, provider)
	if err != nil {
		return types.ClusterStatus{}, nil
	}

	status, err := hf.Status(cluster, provider)
	if err != nil {
		return types.ClusterStatus{}, nil
	}

	content, err := hf.Credentials(cluster, provider)
	if err != nil {
		return types.ClusterStatus{}, nil
	}

	err = ioutil.WriteFile(kubeconfig, content, 0600)

	return *status, nil
}

func (c client) DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string) error {
	cluster, provider, err := c.prepareConfig(runtimeConfig)

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

	return ioutil.WriteFile(filename, bytes, 0644)
}

func (c client) prepareConfig(input model.RuntimeConfig) (*types.Cluster, *types.Provider, error) {
	gardenerConfig, ok := input.GardenerConfig()
	if ok {
		return buildConfigForGardener(gardenerConfig)
	}

	gcpConfig, ok := input.GCPConfig()
	if ok {
		return buildConfigForGCP(gcpConfig)
	}

	return nil, nil, errors.New("configuration does not match any provider profiles")
}

func buildConfigForGCP(config model.GCPConfig) (*types.Cluster, *types.Provider, error) {
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

func buildConfigForGardener(config model.GardenerConfig) (*types.Cluster, *types.Provider, error) {
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
			"disk_type":       "pd-standard",
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
