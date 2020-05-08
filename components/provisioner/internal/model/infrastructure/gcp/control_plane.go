package gcp

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// This types are copied from https://github.com/gardener/gardener-extension-provider-gcp/blob/master/pkg/apis/gcp/types_controlplane.go

// ControlPlaneConfig contains configuration settings for the control plane.
type ControlPlaneConfig struct {
	metav1.TypeMeta

	// Zones are the GCP zones.
	Zone string `json:"zone"`

	// CloudControllerManager contains configuration settings for the cloud-controller-manager.
	CloudControllerManager *CloudControllerManagerConfig `json:"cloudControllerManager,omitempty"`
}

// CloudControllerManagerConfig contains configuration settings for the cloud-controller-manager.
type CloudControllerManagerConfig struct {
	// FeatureGates contains information about enabled feature gates.
	FeatureGates map[string]bool
}
