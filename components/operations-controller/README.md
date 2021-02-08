# Operations Controller

## Overview

The Operations Controller component provides a [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) project 
that describes an `Operation` CustomResourceDefinition. 

The CRD definition can be found at `api/v1alpha1/operation_types.go`. 

The file that is generated from the CRD definition can be found at `config/crd/bases/operations.compass_operations.yaml`.

A CRD sample can be found at `config/samples/operations_v1alpha1_operation.yaml`.

Furthermore, this component also provides a Kubernetes controller for this CRD, it's located at `controllers/operation_controller.go`.

## Prerequisites

- Docker
- Kubernetes CLI
- Kubernetes cluster

## Installation and Usage

To install the CRD into a Kubernetes cluster, run `make install`. To revert the CRD installation, run `make uninstall`.

In order to deploy the Kubernetes controller for the Operation CRD onto a Kubernetes cluster, run `make deploy`. 

## Development

### CRD

If there are changes to the CRD definition, run `make manifests` in order to regenerate the files that define the CRD.
Afterwards, those files should be copied to this component's Helm chart which is located at 
`compass/chart/compass/charts/operations-controller`. To accomplish this, run `make copy-crds-to-chart`.

### Controller

If we make changes to the controller, we have two options:
1. Deploy the new version of the controller using `make install` which will install the controller onto the cluster
   that is described by your `~/.kube/config` file.
2. Create a Docker image for the new version of the controller and deploy it manually on a Kubernetes cluster.
