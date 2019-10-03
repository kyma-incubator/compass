package model

type InfrastructureProvider int

const (
	GCP      InfrastructureProvider = iota
	AKS      InfrastructureProvider = iota
	Gardener InfrastructureProvider = iota
)

type ClusterConfig struct {
	Name                   string
	NodeCount              string // Should'n there be an Int in the schema?
	Memory                 string
	ComputeZone            string
	Version                string
	InfrastructureProvider InfrastructureProvider
	ProviderConfig         interface{}
}

type KymaModule int

const (
	BackupModule             KymaModule = iota
	BackupInitModule         KymaModule = iota
	JaegerModule             KymaModule = iota
	LoggingModule            KymaModule = iota
	MonitoringModule         KymaModule = iota
	PrometheusOperatorModule KymaModule = iota
	KialiModule              KymaModule = iota
	KnativeBuildModule       KymaModule = iota
)

type KymaConfig struct {
	Version string
	Modules []KymaModule
}

type OperationState int

const (
	Pending    OperationState = iota
	InProgress OperationState = iota
	Succeeded  OperationState = iota
	Failed     OperationState = iota
)

type OperationType int

const (
	Provision        OperationType = iota
	Upgrade          OperationType = iota
	Deprovision      OperationType = iota
	ReconnectRuntime OperationType = iota
)

type Operation struct {
	Operation OperationType
	State     OperationState
	Message   string
	RuntimeID string // ???
	Errors    []string
}

type Error struct {
	Message string
}

type AdditionalProperties map[string]interface{}

type GardenerProviderConfig struct {
	TargetProvider       string
	TargetSecret         string
	AutoScalerMin        string
	AutoScalerMax        string
	MaxSurge             int
	MaxUnavailable       int
	AdditionalProperties AdditionalProperties
}

type GCPProviderConfig struct {
	AdditionalProperties AdditionalProperties
}

type AKSProviderConfig struct {
	AdditionalProperties AdditionalProperties
}
