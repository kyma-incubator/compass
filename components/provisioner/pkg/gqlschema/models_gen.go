// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlschema

import (
	"fmt"
	"io"
	"strconv"
)

type ClusterConfig interface {
	IsClusterConfig()
}

type ProviderSpecificConfig interface {
	IsProviderSpecificConfig()
}

type AWSProviderConfig struct {
	Zone         *string `json:"zone"`
	VpcCidr      *string `json:"vpcCidr"`
	PublicCidr   *string `json:"publicCidr"`
	InternalCidr *string `json:"internalCidr"`
}

func (AWSProviderConfig) IsProviderSpecificConfig() {}

type AWSProviderConfigInput struct {
	Zone         string `json:"zone"`
	VpcCidr      string `json:"vpcCidr"`
	PublicCidr   string `json:"publicCidr"`
	InternalCidr string `json:"internalCidr"`
}

type AzureProviderConfig struct {
	VnetCidr *string `json:"vnetCidr"`
}

func (AzureProviderConfig) IsProviderSpecificConfig() {}

type AzureProviderConfigInput struct {
	VnetCidr string `json:"vnetCidr"`
}

type CleanUpRuntimeDataResult struct {
	ID      string  `json:"id"`
	Message *string `json:"message"`
}

type ClusterConfigInput struct {
	GardenerConfig *GardenerConfigInput `json:"gardenerConfig"`
	GcpConfig      *GCPConfigInput      `json:"gcpConfig"`
}

type CredentialsInput struct {
	SecretName string `json:"secretName"`
}

type Error struct {
	Message *string `json:"message"`
}

type GCPConfig struct {
	Name              *string `json:"name"`
	ProjectName       *string `json:"projectName"`
	KubernetesVersion *string `json:"kubernetesVersion"`
	NumberOfNodes     *int    `json:"numberOfNodes"`
	BootDiskSizeGb    *int    `json:"bootDiskSizeGB"`
	MachineType       *string `json:"machineType"`
	Region            *string `json:"region"`
	Zone              *string `json:"zone"`
}

func (GCPConfig) IsClusterConfig() {}

type GCPConfigInput struct {
	Name              string  `json:"name"`
	ProjectName       string  `json:"projectName"`
	KubernetesVersion string  `json:"kubernetesVersion"`
	NumberOfNodes     int     `json:"numberOfNodes"`
	BootDiskSizeGb    int     `json:"bootDiskSizeGB"`
	MachineType       string  `json:"machineType"`
	Region            string  `json:"region"`
	Zone              *string `json:"zone"`
}

type GCPProviderConfig struct {
	Zone *string `json:"zone"`
}

func (GCPProviderConfig) IsProviderSpecificConfig() {}

type GCPProviderConfigInput struct {
	Zone string `json:"zone"`
}

type GardenerConfig struct {
	Name                   *string                `json:"name"`
	ProjectName            *string                `json:"projectName"`
	KubernetesVersion      *string                `json:"kubernetesVersion"`
	NodeCount              *int                   `json:"nodeCount"`
	VolumeSizeGb           *int                   `json:"volumeSizeGB"`
	MachineType            *string                `json:"machineType"`
	Region                 *string                `json:"region"`
	Provider               *string                `json:"provider"`
	Seed                   *string                `json:"seed"`
	TargetSecret           *string                `json:"targetSecret"`
	DiskType               *string                `json:"diskType"`
	WorkerCidr             *string                `json:"workerCidr"`
	AutoScalerMin          *int                   `json:"autoScalerMin"`
	AutoScalerMax          *int                   `json:"autoScalerMax"`
	MaxSurge               *int                   `json:"maxSurge"`
	MaxUnavailable         *int                   `json:"maxUnavailable"`
	ProviderSpecificConfig ProviderSpecificConfig `json:"providerSpecificConfig"`
}

func (GardenerConfig) IsClusterConfig() {}

type GardenerConfigInput struct {
	Name                   string                 `json:"name"`
	ProjectName            string                 `json:"projectName"`
	KubernetesVersion      string                 `json:"kubernetesVersion"`
	NodeCount              int                    `json:"nodeCount"`
	VolumeSizeGb           int                    `json:"volumeSizeGB"`
	MachineType            string                 `json:"machineType"`
	Region                 string                 `json:"region"`
	Provider               string                 `json:"provider"`
	Seed                   string                 `json:"seed"`
	TargetSecret           string                 `json:"targetSecret"`
	DiskType               string                 `json:"diskType"`
	WorkerCidr             string                 `json:"workerCidr"`
	AutoScalerMin          int                    `json:"autoScalerMin"`
	AutoScalerMax          int                    `json:"autoScalerMax"`
	MaxSurge               int                    `json:"maxSurge"`
	MaxUnavailable         int                    `json:"maxUnavailable"`
	ProviderSpecificConfig *ProviderSpecificInput `json:"providerSpecificConfig"`
}

type KymaConfig struct {
	Version *string       `json:"version"`
	Modules []*KymaModule `json:"modules"`
}

type KymaConfigInput struct {
	Version string       `json:"version"`
	Modules []KymaModule `json:"modules"`
}

type OperationStatus struct {
	ID        *string        `json:"id"`
	Operation OperationType  `json:"operation"`
	State     OperationState `json:"state"`
	Message   *string        `json:"message"`
	RuntimeID *string        `json:"runtimeID"`
}

type ProviderSpecificInput struct {
	GcpConfig   *GCPProviderConfigInput   `json:"gcpConfig"`
	AzureConfig *AzureProviderConfigInput `json:"azureConfig"`
	AwsConfig   *AWSProviderConfigInput   `json:"awsConfig"`
}

type ProvisionRuntimeInput struct {
	RuntimeInput  *RuntimeInput       `json:"runtimeInput"`
	ClusterConfig *ClusterConfigInput `json:"clusterConfig"`
	Credentials   *CredentialsInput   `json:"credentials"`
	KymaConfig    *KymaConfigInput    `json:"kymaConfig"`
}

type RuntimeConfig struct {
	ClusterConfig         ClusterConfig `json:"clusterConfig"`
	CredentialsSecretName *string       `json:"credentialsSecretName"`
	KymaConfig            *KymaConfig   `json:"kymaConfig"`
	Kubeconfig            *string       `json:"kubeconfig"`
}

type RuntimeConnectionStatus struct {
	Status RuntimeAgentConnectionStatus `json:"status"`
	Errors []*Error                     `json:"errors"`
}

type RuntimeInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Labels      *string `json:"labels"`
}

type RuntimeStatus struct {
	LastOperationStatus     *OperationStatus         `json:"lastOperationStatus"`
	RuntimeConnectionStatus *RuntimeConnectionStatus `json:"runtimeConnectionStatus"`
	RuntimeConfiguration    *RuntimeConfig           `json:"runtimeConfiguration"`
}

type UpgradeClusterInput struct {
	Version string `json:"version"`
}

type UpgradeRuntimeInput struct {
	ClusterConfig *UpgradeClusterInput `json:"clusterConfig"`
	KymaConfig    *KymaConfigInput     `json:"kymaConfig"`
}

type KymaModule string

const (
	KymaModuleBackup             KymaModule = "Backup"
	KymaModuleBackupInit         KymaModule = "BackupInit"
	KymaModuleJaeger             KymaModule = "Jaeger"
	KymaModuleLogging            KymaModule = "Logging"
	KymaModuleMonitoring         KymaModule = "Monitoring"
	KymaModulePrometheusOperator KymaModule = "PrometheusOperator"
	KymaModuleKiali              KymaModule = "Kiali"
	KymaModuleKnativeBuild       KymaModule = "KnativeBuild"
)

var AllKymaModule = []KymaModule{
	KymaModuleBackup,
	KymaModuleBackupInit,
	KymaModuleJaeger,
	KymaModuleLogging,
	KymaModuleMonitoring,
	KymaModulePrometheusOperator,
	KymaModuleKiali,
	KymaModuleKnativeBuild,
}

func (e KymaModule) IsValid() bool {
	switch e {
	case KymaModuleBackup, KymaModuleBackupInit, KymaModuleJaeger, KymaModuleLogging, KymaModuleMonitoring, KymaModulePrometheusOperator, KymaModuleKiali, KymaModuleKnativeBuild:
		return true
	}
	return false
}

func (e KymaModule) String() string {
	return string(e)
}

func (e *KymaModule) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = KymaModule(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid KymaModule", str)
	}
	return nil
}

func (e KymaModule) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type OperationState string

const (
	OperationStatePending    OperationState = "Pending"
	OperationStateInProgress OperationState = "InProgress"
	OperationStateSucceeded  OperationState = "Succeeded"
	OperationStateFailed     OperationState = "Failed"
)

var AllOperationState = []OperationState{
	OperationStatePending,
	OperationStateInProgress,
	OperationStateSucceeded,
	OperationStateFailed,
}

func (e OperationState) IsValid() bool {
	switch e {
	case OperationStatePending, OperationStateInProgress, OperationStateSucceeded, OperationStateFailed:
		return true
	}
	return false
}

func (e OperationState) String() string {
	return string(e)
}

func (e *OperationState) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = OperationState(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid OperationState", str)
	}
	return nil
}

func (e OperationState) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type OperationType string

const (
	OperationTypeProvision        OperationType = "Provision"
	OperationTypeUpgrade          OperationType = "Upgrade"
	OperationTypeDeprovision      OperationType = "Deprovision"
	OperationTypeReconnectRuntime OperationType = "ReconnectRuntime"
)

var AllOperationType = []OperationType{
	OperationTypeProvision,
	OperationTypeUpgrade,
	OperationTypeDeprovision,
	OperationTypeReconnectRuntime,
}

func (e OperationType) IsValid() bool {
	switch e {
	case OperationTypeProvision, OperationTypeUpgrade, OperationTypeDeprovision, OperationTypeReconnectRuntime:
		return true
	}
	return false
}

func (e OperationType) String() string {
	return string(e)
}

func (e *OperationType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = OperationType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid OperationType", str)
	}
	return nil
}

func (e OperationType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type RuntimeAgentConnectionStatus string

const (
	RuntimeAgentConnectionStatusPending      RuntimeAgentConnectionStatus = "Pending"
	RuntimeAgentConnectionStatusConnected    RuntimeAgentConnectionStatus = "Connected"
	RuntimeAgentConnectionStatusDisconnected RuntimeAgentConnectionStatus = "Disconnected"
)

var AllRuntimeAgentConnectionStatus = []RuntimeAgentConnectionStatus{
	RuntimeAgentConnectionStatusPending,
	RuntimeAgentConnectionStatusConnected,
	RuntimeAgentConnectionStatusDisconnected,
}

func (e RuntimeAgentConnectionStatus) IsValid() bool {
	switch e {
	case RuntimeAgentConnectionStatusPending, RuntimeAgentConnectionStatusConnected, RuntimeAgentConnectionStatusDisconnected:
		return true
	}
	return false
}

func (e RuntimeAgentConnectionStatus) String() string {
	return string(e)
}

func (e *RuntimeAgentConnectionStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RuntimeAgentConnectionStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RuntimeAgentConnectionStatus", str)
	}
	return nil
}

func (e RuntimeAgentConnectionStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
