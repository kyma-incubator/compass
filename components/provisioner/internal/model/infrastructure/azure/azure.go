package azure

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// This types are copied from https://github.com/gardener/gardener-extensions/blob/master/controllers/provider-azure/pkg/apis/azure/types_infrastructure.go as it does not contain json tags

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta
	// ResourceGroup is azure resource group
	ResourceGroup *ResourceGroup `json:"resourceGroup"`
	// Networks is the network configuration (VNets, subnets, etc.)
	Networks NetworkConfig `json:"networks"`
	// Zoned indicates whether the cluster uses zones
	Zoned bool `json:"zoned"`
}

// ResourceGroup is azure resource group
type ResourceGroup struct {
	// Name is the name of the resource group
	Name string `json:"name"`
}

// NetworkConfig holds information about the Kubernetes and infrastructure networks.
type NetworkConfig struct {
	// VNet indicates whether to use an existing VNet or create a new one.
	VNet VNet `json:"vnet"`
	// Workers is the worker subnet range to create (used for the VMs).
	Workers string `json:"workers"`
	// ServiceEndpoints is a list of Azure ServiceEndpoints which should be associated with the worker subnet.
	ServiceEndpoints []string `json:"serviceEndpoints"`
}

// VNet contains information about the VNet and some related resources.
type VNet struct {
	// Name is the VNet name.
	Name *string `json:"name"`
	// ResourceGroup is the resource group where the existing vNet belongs to.
	ResourceGroup *string `json:"resourceGroup"`
	// CIDR is the VNet CIDR
	CIDR *string `json:"cidr"`
}

// VNetStatus contains the VNet name.
type VNetStatus struct {
	// Name is the VNet name.
	Name string `json:"name"`
	// ResourceGroup is the resource group where the existing vNet belongs to.
	ResourceGroup *string `json:"resourceGroup"`
}
