# Introduction

One of the responsibilities of Management Plane Services is provisioning Kyma Runtimes being Kubernetes clusters with Kyma installed.

The goal of this document is to describe requirements for Kyma Runtime provisioning operations and propose an API.

## Requirements

## Non-Functional

The following requiremens must be met:

- API is based on GraphQL 
- API works in an asynchronous manner

## Functional

### Runtime management

The following operations must be supported:

- Provisioning Runtime on specified infrastructure
- Upgrading Runtime
  - Kubernetes cluster version upgrade
  - Kyma version upgrade
- Deprovisioning Runtimes

### Runtime Agent connection management

Resetting connection between Runtime and Management Plane Services must be supported. The goal of this functionality is to force Compass Runtime Agent installed on the Runtime to establish a new connection.

### Status information retrieval

The following operations must be supported:

- Getting status of an asynchronous operations (e.g. Runtime provisioning)
- Getting current status of a Runtime 

# API proposal

## General 

The following assumptions have been taken:

- Cluster provisioning and Kyma installation is considered as a single operation
- Provisioning, deprovisioning, upgrade and Compass Runtime Agent reconnecting is uniquely identified by OperationID
- Runtime is uniquely identified by RuntimeID
- Only one asynchronous operation can be in progress on given RuntimeID. 
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

The mutation returns Operation identifier allowing to monitor the operation status.

### Upgrade Runtime

Upgrade is implemented by ***upgradeRuntime*** mutation. The object that must be passed to the mutation contains the following fields:

- Kyma installation settings
  - Release version
  - List of modules to be installed
- Kubernetes cluster settings
  -  version

### Runtime deprovisioning

Deprovisioning is implemented by ***deprovisionRuntime*** mutation. The RuntimeID must be passed as argument. The mutation returns Operation identifier allowing to monitor the operation status. 

## Runtime connection management

## Retrieving operation status

### Asynchronous operation status

Getting current Runtime status is implemented by ***runtimeOperationStatus*** query. The query takes Operation ID as parameter and returns object containing the following information:

- Operation type (e.g. Provisioning)
- Operation status (e.g. InProgress)
- Message
- Error messages list

### Current status of a Runtime

Getting current status of a Runtime is implemented by ***runtimeOperationStatus*** query. The query takes Runtime ID as a parameter and returns object containing the following information:

- last operation status
- Runtime connection configuration (kubeconfig)
- Runtime Agent Connection status
  - Status (connected, disconnected)
  - Errors list

# Summary

## Open questions