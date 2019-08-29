# Introduction

The goal of this document is to describe requirements for Provisioner component and propose an API. 

## Requirements

## Non-Functional

The basic requirements are defined as follows:

- Provisioner API is based on GraphQL. 
- Provisioner API is asynchronous.

## Functional

### Runtime management

The following operations must be supported:

- Provisioning
- Upgrade
  - Kubernetes version.
  - Kyma version.
- Deprovisioning

### Runtime Agent connection management

Resetting connection between Runtime and Management Plane Services must be supported. The goal of this functionality is to force Compass Runtime Agent installed on the Runtime to establish a new connection.

### Status information retrieval

The following operations must be supported:

- Getting status of an asynchronous operations (e.g. Runtime provisioning)
- Getting current status of a Runtime 

# API proposal

## General 

The basic assumptions for API design are as follows:

- Cluster provisioning and Kyma installation is considered atomic operation.
- Provisioning, deprovisioning, upgrade and Compass Runtime Agent reconnecting is uniquely identified by OperationID.
- Runtime is uniquely identified by RuntimeID.
- Only one asynchronous operation can be in progress on a given Runtime.  
- It must be possible to install minimal Kyma  (Kyma Lite) and specify additional modules.
- Two types of status information are available:
  - Status of asynchronous operation.
  - Status of Runtime comprised of:
    - Status of the last operation.
    - Status of Compass Runtime Agent Connection.

## Runtime managament

### Provision Runtime

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

Some Kubernetes cluster settings (such as size, memory, and version) are optional and default values are used.

The mutation returns OperationID allowing to retrieve the operation status.

### Upgrade Runtime

Upgrade is implemented by ***upgradeRuntime*** mutation. The object that must be passed to the mutation contains the following fields:

- Kyma installation settings
  - Release version
  - List of modules to be installed
- Kubernetes cluster settings
  -  version

The mutation returns OperationID allowing to retrieve the operation status.

### Deprovision Runtime

Deprovisioning is implemented by ***deprovisionRuntime*** mutation. The RuntimeID must be passed as argument. 

The mutation returns OperationID allowing to retrieve the operation status.

### Reconnecting Compass Runtime Agent

Reconnection Compass runtime Agent is implemented by ***reconnectRuntimeAgent*** mutation. The RuntimeID must be passed as argument. 

The mutation returns OperationID allowing to retrieve the operation status.

## Retrieving operation status

### Operation status

Getting operation status is implemented by ***runtimeOperationStatus*** query. The query takes Operation ID as parameter and returns object containing the following information:

- Operation type (e.g. Provisioning).
- Operation status (e.g. InProgress).
- Message.
- Error messages list.

### Current status of a Runtime

Getting current status of a Runtime is implemented by ***runtimeOperationStatus*** query. The query takes Runtime ID as a parameter and returns object containing the following information:

- Last operation status.
- Runtime connection configuration (kubeconfig).
- Runtime Agent Connection status
  - Status (connected, disconnected).
  - Errors list.