---
title: Deprovision clusters
type: Tutorials
---

This tutorial shows how to deprovision clusters with Kyma Runtimes (Runtimes).

## Deprovision Kyma Runtime

> **NOTE:** To access the Runtime Provisioner, make a call from another Pod in the cluster containing the Runtime Provisioner or forward the port on which the GraphQL Server is listening.

  To deprovision a Runtime, make a call to the Runtime Provisioner with a mutation like this:  
  
  ```graphql
  mutation { deprovisionRuntime(id: "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3") }
  ```

  A successful call returns the ID of the deprovisioning operation:

  ```graphql
  {
    "data": {
      "deprovisionRuntime": "c7e6727f-16b5-4748-ac95-197d8f79d094"
    }
  }
  ```

  The operation of deprovisioning is asynchronous. Use the deprovisioning operation ID to check the Runtime Operation Status and verify that the deprovisioning was successful.

## Check the Runtime Operation Status

Make a call to the Runtime Provisioner to verify that deprovisioning succeeded. Pass the ID of the deprovisioning operation as `id`.

```graphql
query { 
  runtimeOperationStatus(id: "c7e6727f-16b5-4748-ac95-197d8f79d094") { 
    operation 
    state 
    message 
    runtimeID
  } 
}
```

A successful call returns a response which includes the status of the deprovisioning operation (`state`) and the id of the deprovisioned Runtime (`runtimeID`):

```graphql
{
  "data": {
    "runtimeOperationStatus": {
      "operation": "Deprovision",
      "state": "Succeeded",
      "message": "Operation succeeded.",
      "runtimeID": "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3"
    }
  }
}
```

The `Succeeded` status means that the deprovisioning was successful and the cluster was deleted.

If you get the `InProgress` status, it means that the deprovisioning has not yet finished. In that case, wait a few moments and check the status again. 

## Check the Runtime Status

Make a call to the Runtime Provisioner to check the Runtime status. Pass the Runtime ID as `id`. 

```graphql
query { runtimeStatus(id: "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3") {
  runtimeConnectionStatus {
    status errors {
      message
    } 
  } 
  lastOperationStatus {
    message operation state runtimeID id
  } 
  runtimeConfiguration {
    kubeconfig kymaConfig {
      version modules 
    } clusterConfig {
      __typename ... on GCPConfig {
        bootDiskSizeGB name numberOfNodes kubernetesVersion projectName machineType zone region 
      }
      ... on GardenerConfig { name workerCidr region diskType maxSurge nodeCount volumeSizeGB projectName machineType targetSecret autoScalerMin autoScalerMax provider maxUnavailable kubernetesVersion 
      } 
    } 
  } 
}}
```

An example response for a successful request looks like this:

```graphql
{
  "data": {
    "runtimeStatus": {
      "runtimeConnectionStatus": {
        "status": "Pending",
        "errors": null
      },
      "lastOperationStatus": {
        "message": "Operation succeeded.",
        "operation": "Deprovision",
        "state": "Succeeded",
        "runtimeID": "61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3",
        "id": "c7e6727f-16b5-4748-ac95-197d8f79d094"
      },
      "runtimeConfiguration": {
        "kubeconfig": "{KUBECONFIG}",
        "kymaConfig": {
          "version": "1.5",
          "modules": [
            "Backup"
          ]
        },
        "clusterConfig": {CLUSTER_CONFIG}
      }
    }
  }
}
```