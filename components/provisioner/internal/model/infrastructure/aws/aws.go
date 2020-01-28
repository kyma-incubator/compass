package aws

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This types are copied from https://github.com/gardener/gardener-extensions/blob/master/controllers/provider-azure/pkg/apis/azure/types_infrastructure.go as it does not contain json tags

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta

	// EnableECRAccess specifies whether the IAM role policy for the worker nodes shall contain
	// permissions to access the ECR.
	// default: true
	EnableECRAccess *bool `json:"enableECRAccess"`

	// Networks is the AWS specific network configuration (VPC, subnets, etc.)
	Networks Networks `json:"networks"`
}

// Networks holds information about the Kubernetes and infrastructure networks.
type Networks struct {
	// VPC indicates whether to use an existing VPC or create a new one.
	VPC VPC `json:"vpc"`
	// Zones belonging to the same region
	Zones []Zone `json:"zones"`
}

// Zone describes the properties of a zone
type Zone struct {
	// Name is the name for this zone.
	Name string `json:"name"`
	// Internal is the private subnet range to create (used for internal load balancers).
	Internal string `json:"internal"`
	// Public is the public subnet range to create (used for bastion and load balancers).
	Public string `json:"public"`
	// Workers isis the workers subnet range to create  (used for the VMs).
	Workers string `json:"workers"`
}

// VPC contains information about the AWS VPC and some related resources.
type VPC struct {
	// ID is the VPC id.
	ID *string `json:"id"`
	// CIDR is the VPC CIDR.
	CIDR *string `json:"cidr"`
}
