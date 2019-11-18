package configuration

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const credentialsKey = "credentials"

//go:generate mockery -name=ConfigBuilderFactory
type ConfigBuilderFactory interface {
	NewProvisioningBuilder(provisionInput gqlschema.ProvisionRuntimeInput) ConfigBuilder
	NewDeprovisioningBuilder(runtimeConfig model.RuntimeConfig) ConfigBuilder
}

type configBuilderFactory struct {
	secrets v1.SecretInterface
}

func NewConfigBuilderFactory(secrets v1.SecretInterface) ConfigBuilderFactory {
	return &configBuilderFactory{
		secrets: secrets,
	}
}

func (cf configBuilderFactory) NewProvisioningBuilder(provisionInput gqlschema.ProvisionRuntimeInput) ConfigBuilder {
	return &provisioningBuilder{
		secrets:               cf.secrets,
		provisionRuntimeInput: provisionInput,
	}
}

func (cf configBuilderFactory) NewDeprovisioningBuilder(runtimeConfig model.RuntimeConfig) ConfigBuilder {
	return &deprovisioningBuilder{
		secrets:       cf.secrets,
		runtimeConfig: runtimeConfig,
	}
}

//go:generate mockery -name=ConfigBuilder
type ConfigBuilder interface {
	Create() (*types.Cluster, *types.Provider, error)
	CleanUp()
}

type provisioningBuilder struct {
	filename              string
	provisionRuntimeInput gqlschema.ProvisionRuntimeInput
	secrets               v1.SecretInterface
}

func (pb *provisioningBuilder) Create() (*types.Cluster, *types.Provider, error) {
	credentialsFileName, err := saveCredentialsToFile(pb.secrets, pb.provisionRuntimeInput.Credentials.SecretName)
	if err != nil {
		return nil, nil, err
	}

	pb.filename = credentialsFileName

	return pb.prepareConfig()
}

func (pb *provisioningBuilder) CleanUp() {
	removeFile(pb.filename)
}

func (pb provisioningBuilder) prepareConfig() (*types.Cluster, *types.Provider, error) {
	config := pb.provisionRuntimeInput.ClusterConfig

	if config.GardenerConfig != nil {
		return pb.buildConfigForGardener(config.GardenerConfig)
	}

	if config.GcpConfig != nil {
		return pb.buildConfigForGCP(config.GcpConfig)
	}

	return nil, nil, errors.New("Configuration does not match any provider profile")
}

func (pb *provisioningBuilder) buildConfigForGardener(input *gqlschema.GardenerConfigInput) (*types.Cluster, *types.Provider, error) {
	cluster := &types.Cluster{
		KubernetesVersion: input.KubernetesVersion,
		Name:              input.Name,
		DiskSizeGB:        input.VolumeSizeGb,
		NodeCount:         input.NodeCount,
		Location:          input.Region,
		MachineType:       input.MachineType,
	}

	customConfiguration := map[string]interface{}{
		"target_provider": input.Provider,
		"target_seed":     input.Seed,
		"target_secret":   input.TargetSecret,
		"disk_type":       input.DiskType,
		"workercidr":      input.WorkerCidr,
		"autoscaler_min":  input.AutoScalerMin,
		"autoscaler_max":  input.AutoScalerMax,
		"max_surge":       input.MaxSurge,
		"max_unavailable": input.MaxUnavailable,
	}

	err := addProviderSpecificConfigProvisioning(customConfiguration, input.ProviderSpecificConfig)

	if err != nil {
		return &types.Cluster{}, &types.Provider{}, err
	}

	provider := &types.Provider{
		Type:                 types.Gardener,
		ProjectName:          input.ProjectName,
		CredentialsFilePath:  pb.filename,
		CustomConfigurations: customConfiguration,
	}
	return cluster, provider, nil
}

func addProviderSpecificConfigProvisioning(customConfiguration map[string]interface{}, input *gqlschema.ProviderSpecificInput) error {
	if input.GcpConfig != nil {
		customConfiguration["zone"] = input.GcpConfig.Zone
		return nil
	}

	if input.AzureConfig != nil {
		customConfiguration["vnetcidr"] = input.AzureConfig.VnetCidr
		return nil
	}

	if input.AwsConfig != nil {
		customConfiguration["zone"] = input.AwsConfig.Zone
		customConfiguration["internalscidr"] = input.AwsConfig.InternalCidr
		customConfiguration["vpccidr"] = input.AwsConfig.VpcCidr
		customConfiguration["publicscidr"] = input.AwsConfig.PublicCidr
		return nil
	}

	return errors.New("provider specific configuration does not match any provider profile")
}

func (pb *provisioningBuilder) buildConfigForGCP(input *gqlschema.GCPConfigInput) (*types.Cluster, *types.Provider, error) {
	cluster := &types.Cluster{
		KubernetesVersion: input.KubernetesVersion,
		Name:              input.Name,
		DiskSizeGB:        input.BootDiskSizeGb,
		NodeCount:         input.NumberOfNodes,
		Location:          input.Region,
		MachineType:       input.MachineType,
	}

	provider := &types.Provider{
		Type:                types.GCP,
		ProjectName:         input.ProjectName,
		CredentialsFilePath: pb.filename,
	}
	return cluster, provider, nil
}

type deprovisioningBuilder struct {
	filename      string
	runtimeConfig model.RuntimeConfig
	secrets       v1.SecretInterface
}

func (db *deprovisioningBuilder) Create() (*types.Cluster, *types.Provider, error) {
	credentialsFileName, err := saveCredentialsToFile(db.secrets, db.runtimeConfig.CredentialsSecretName)
	if err != nil {
		return nil, nil, err
	}

	db.filename = credentialsFileName

	return db.prepareConfig()
}

func (db *deprovisioningBuilder) CleanUp() {
	removeFile(db.filename)
}

func (db *deprovisioningBuilder) prepareConfig() (*types.Cluster, *types.Provider, error) {
	gardenerConfig, ok := db.runtimeConfig.GardenerConfig()
	if ok {
		return db.buildConfigForGardener(gardenerConfig)
	}

	gcpConfig, ok := db.runtimeConfig.GCPConfig()
	if ok {
		return db.buildConfigForGCP(gcpConfig)
	}

	return nil, nil, errors.New("Configuration does not match any provider profile")
}

func (db *deprovisioningBuilder) buildConfigForGCP(config model.GCPConfig) (*types.Cluster, *types.Provider, error) {
	cluster := &types.Cluster{
		KubernetesVersion: config.KubernetesVersion,
		Name:              config.Name,
		DiskSizeGB:        config.BootDiskSizeGB,
		NodeCount:         config.NumberOfNodes,
		Location:          config.Region,
		MachineType:       config.MachineType,
	}

	provider := &types.Provider{
		Type:                types.GCP,
		ProjectName:         config.ProjectName,
		CredentialsFilePath: db.filename,
	}
	return cluster, provider, nil
}

func (db *deprovisioningBuilder) buildConfigForGardener(config model.GardenerConfig) (*types.Cluster, *types.Provider, error) {
	cluster := &types.Cluster{
		KubernetesVersion: config.KubernetesVersion,
		Name:              config.Name,
		DiskSizeGB:        config.VolumeSizeGB,
		NodeCount:         config.NodeCount,
		Location:          config.Region,
		MachineType:       config.MachineType,
	}

	customConfiguration := map[string]interface{}{
		"target_provider": config.Provider,
		"target_seed":     config.Seed,
		"target_secret":   config.TargetSecret,
		"disk_type":       config.DiskType,
		"workercidr":      config.WorkerCidr,
		"autoscaler_min":  config.AutoScalerMin,
		"autoscaler_max":  config.AutoScalerMax,
		"max_surge":       config.MaxSurge,
		"max_unavailable": config.MaxUnavailable,
	}

	err := addProviderSpecificConfigDeprovisioning(customConfiguration, config.ProviderSpecificConfig)

	if err != nil {
		return &types.Cluster{}, &types.Provider{}, err
	}

	provider := &types.Provider{
		Type:                 types.Gardener,
		ProjectName:          config.ProjectName,
		CredentialsFilePath:  db.filename,
		CustomConfigurations: customConfiguration,
	}
	return cluster, provider, nil
}

func addProviderSpecificConfigDeprovisioning(customConfiguration map[string]interface{}, providerSpecificConfig string) error {
	var gcpProviderConfig gqlschema.GCPProviderConfig
	err := util.DecodeJson(providerSpecificConfig, &gcpProviderConfig)
	if err == nil {
		customConfiguration["zone"] = *gcpProviderConfig.Zone
		return nil
	}

	var azureProviderConfig gqlschema.AzureProviderConfig
	err = util.DecodeJson(providerSpecificConfig, &azureProviderConfig)
	if err == nil {
		customConfiguration["vnetcidr"] = *azureProviderConfig.VnetCidr
		return nil
	}

	var awsProviderConfig gqlschema.AWSProviderConfig
	err = util.DecodeJson(providerSpecificConfig, &awsProviderConfig)
	if err == nil {
		customConfiguration["zone"] = *awsProviderConfig.Zone
		customConfiguration["internalscidr"] = *awsProviderConfig.InternalCidr
		customConfiguration["vpccidr"] = *awsProviderConfig.VpcCidr
		customConfiguration["publicscidr"] = *awsProviderConfig.PublicCidr
		return nil
	}

	return errors.New("provider specific configuration does not match any provider profile")
}

func saveCredentialsToFile(secrets v1.SecretInterface, secretName string) (string, error) {
	secret, err := secrets.Get(secretName, meta.GetOptions{})
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
