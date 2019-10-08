package hydroform

import (
	"io/ioutil"

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
	ProvisionCluster(input model.ClusterConfig) (types.ClusterStatus, error)
	DeprovisionCluster(input model.ClusterConfig) error
}

type client struct {
	secrets v1.SecretInterface
}

func NewHydroformClient(secrets v1.SecretInterface) Client {
	return &client{secrets: secrets}
}

func (c client) ProvisionCluster(input model.ClusterConfig) (types.ClusterStatus, error) {
	cluster, provider, err := c.prepareConfig(input)

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

func (c client) DeprovisionCluster(input model.ClusterConfig) error {
	cluster, provider, err := c.prepareConfig(input)

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

func (c client) prepareConfig(input model.ClusterConfig) (*types.Cluster, *types.Provider, error) {

	// TODO Adjust to the new data model
	return nil, nil, nil
	//diskSize, err := strconv.Atoi(input.DiskSize)
	//
	//if err != nil {
	//	return &types.Cluster{}, &types.Provider{}, err
	//}
	//
	//cluster := &types.Cluster{
	//	KubernetesVersion: input.Version,
	//	Name:              input.Name,
	//	DiskSizeGB:        diskSize,
	//	NodeCount:         input.NodeCount,
	//	Location:          input.Region,
	//	MachineType:       input.MachineType,
	//}
	//
	//providerConfig, ok := input.ProviderConfig.(gqlschema.ProviderConfigInput)
	//
	//if !ok {
	//	return &types.Cluster{}, &types.Provider{}, errors.New("ProviderConfig is not ProviderConfigInput type")
	//}
	//
	//gardenerConfig := providerConfig.GardenerProviderConfig
	//
	//err = c.saveCredentialsToFile(input.Credentials, credentialsFile)
	//
	//if err != nil {
	//	return &types.Cluster{}, &types.Provider{}, err
	//}
	//
	//provider := &types.Provider{
	//	Type:                types.Gardener,
	//	ProjectName:         gardenerConfig.ProjectName,
	//	CredentialsFilePath: credentialsFile,
	//	CustomConfigurations: map[string]interface{}{
	//		"target_provider": gardenerConfig.TargetProvider,
	//		"target_secret":   gardenerConfig.TargetSecret,
	//		"disk_type":       "pd-standard",
	//		"zone":            gardenerConfig.ComputeZone,
	//		"cidr":            "10.250.0.0/19",
	//		"autoscaler_min":  gardenerConfig.AutoScalerMin,
	//		"autoscaler_max":  gardenerConfig.AutoScalerMax,
	//		"max_surge":       gardenerConfig.MaxSurge,
	//		"max_unavailable": gardenerConfig.MaxUnavailable,
	//	},
	//}
	//return cluster, provider, nil
}
