# Introduction

One of the responsibilities of Management Planse Services is provisioning Kyma runtimes.The goal of this document is to describe requirements for Kyma provisioning operations and propose an API.

## Requirements

## Functional

### Runtime management

The following operations must be supported:

- Creating Kyma Runtime on specified infrastructure
- Upgrading Kyma Runtime
- Removing Kyma Runtimes

### Obtaining operation status

The following operations must be supported:

- Getting runtime management status determining state of provisioning/deprovisioning/upgrading process 
- Getting Runtime Connection status determining whether Runtime is properly connected to Management Plane Services 

### Runtime connection management

The following operations must be supported:

- Getting connection to provisioned clusters (kubeconfig)
- Resetting connection between Runtime and Management Plane Services so that Compass Runtime Agent installed on the Runtime establieshes a new connection

## Non-Functional

The following requiremens must be met:

- Operations in API are asynchronous
- API is based on GraphQL 

# API proposal

## General assumptions

## Design

### Mutations

### Queries



## 

# Summary