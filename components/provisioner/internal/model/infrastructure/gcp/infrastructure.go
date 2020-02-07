package gcp

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This types are copied from https://github.com/gardener/gardener-extensions/blob/master/controllers/provider-gcp/pkg/apis/gcp/types_infrastructure.go as it does not contain json tags

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta

	// Networks is the network configuration (VPC, subnets, etc.)
	Networks NetworkConfig `json:"networks"`
}

// NetworkConfig holds information about the Kubernetes and infrastructure networks.
type NetworkConfig struct {
	// VPC indicates whether to use an existing VPC or create a new one.
	VPC *VPC `json:"vpc,omitempty"`
	// CloudNAT contains configation about the the CloudNAT resource
	CloudNAT *CloudNAT `json:"cloudNat,omitempty"`
	// Internal is a private subnet (used for internal load balancers).
	Internal *string `json:"internal,omitempty"`
	// Worker is the worker subnet range to create (used for the VMs).
	// Deprecated - use `workers` instead.
	Worker string `json:"worker"`
	// Workers is the worker subnet range to create (used for the VMs).
	Workers *string `json:"workers,omitempty"`
	// FlowLogs contains the flow log configuration for the subnet.
	FlowLogs *FlowLogs `json:"flowLogs,omitempty"`
}

// VPC contains information about the VPC and some related resources.
type VPC struct {
	// Name is the VPC name.
	Name string `json:"name"`
	// CloudRouter indicates whether to use an existing CloudRouter or create a new one
	CloudRouter *CloudRouter `json:"cloudRouter,omitempty"`
}

// CloudRouter contains information about the the CloudRouter configuration
type CloudRouter struct {
	// Name is the CloudRouter name.
	Name string `json:"name"`
}

// CloudNAT contains information about the the CloudNAT configuration
type CloudNAT struct {
	// MinPortsPerVM is the minimum number of ports allocated to a VM in the NAT config.
	// The default value is 2048 ports.
	MinPortsPerVM *int32 `json:"minPortsPerVM,omitempty"`
}

// FlowLogs contains the configuration options for the vpc flow logs.
type FlowLogs struct {
	// AggregationInterval for collecting flow logs.
	AggregationInterval *string `json:"aggregationInterval,omitempty"`
	// FlowSampling sets the sampling rate of VPC flow logs within the subnetwork where 1.0 means all collected logs are reported and 0.0 means no logs are reported.
	FlowSampling *float32 `json:"flowSampling,omitempty"`
	// Metadata configures whether metadata fields should be added to the reported VPC flow logs.
	Metadata *string `json:"metadata,omitempty"`
}
