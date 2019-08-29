# Introduction

One of the responsibilities of Management Plane Services is provisioning Kyma runtimes.The goal of this document is to describe requirements for Kyma provisioning operations and propose an API.

## Requirements

## Functional

### Runtime management

The following operations must be supported:

- Creating Runtime on specified infrastructure
- Upgrading Runtime
- Removing Runtimes

### Obtaining status

The following operations must be supported:

- Getting status determining state of provisioning/deprovisioning/upgrade process 
- Getting current status of Runtime 

### Runtime connection management

The following operations must be supported:

- Getting connection to provisioned clusters (kubeconfig)
- Resetting connection between Runtime and Management Plane Services so that Compass Runtime Agent installed on the Runtime establieshes a new connection

## Non-Functional

The following requiremens must be met:

- Provisioning, deprovisioning and upgrade operations are asynchronous
- API is based on GraphQL 

# API proposal

## General 

The following assumptions have been taken:

- Cluster provisioning and Kyma installation is considered as a single operation
- Provisioning, deprovisioning, and upgrade is uniquely identified by OperationID
- Runtime (Kubernetes cluster with Kyma installed) is uniquely identified by RuntimeID
- It must be possible to install minimal Kyma  (Kyma Lite) and specify additional modules
- Two types of status information are available:
  - status of asynchronous operation
  - status of Runtime comprised of:
    - status of last operation
    - status of Compass Runtime Agent Connection

## Runtime managament

### Provision runtime

Provisioning is implemented by ***provisionRuntime*** mutation. The object that must be passed to the mutation contains the following fields:

- Kyma installation settings
  - Release version
  - List of modules to be installed
- Kubernetes cluster settings
  - name
  - size
  - memory
  - region and zone
  - credentials
  -  version
  - Infrastructure provider (e.g. GKE, AKS)

Some Kubernetes cluster settings (such as size, memory, and version) are optional and default values will be used.

The mutation returns object with containing:

- Runtime identifier allowing to monitor Runtime status and obtain runtime connection
- Operation identifier allowing to monitor the operation status

### Upgrade Runtime

Upgrade is implemented by ***upgradeRuntime*** mutation. The mutation supports both Kubernetes cluster and Kyma version update. It takes the same argument as ***provisionRuntime*** and return operation identifier.

### Runtime deprovisioning

## Obtaining operation status

## Runtime connection management



### Queries



## Implementation 

# Summary

## Open questions