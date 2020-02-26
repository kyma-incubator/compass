# [POC] Provision Azure EventHubs Namespace

## Overview

The purpose of this POC is to be able to provision Azure EventHubs Namespace using go binary that can be integrated with CI/CD and cluster installers.
 
> Note: The sourcecode is inspired by the Azure go-sdk [samples](https://github.com/Azure-Samples/azure-sdk-for-go-samples).

## Scope

The scope of the POC:

- Provision an Azure EventHubs Namespace inside a new or existing Azure Resource Group.
- Prepare a Kubernetes secret for the provisioned Azure EventHubs Namespace.

## Prerequisites

- Have an Azure account with the right permissions to create resources.

## Verification

Use the following steps for verification:

- Update the `test.env` file with your Azure account details.

- Run the `verify.sh` file which will build and run a go binary that will execute the following flow in order:

  - Create new or update existing Azure Resource Group.
  - Create new or update existing Azure EventHubs Namespace.
  - Get the proper Access Keys for the Azure EventHubs Namespace.
  - Prepare a Kubernetes secret for the Azure EventHubs Namespace.

## Next steps

- The created secret should be persisted to the desired Kubernetes cluster for the Eventing flow to work from that cluster to the provisioned Azure Event Hubs Namespace. 
