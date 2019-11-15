---
title: Provision clusters on Google Cloud Platform (GCP)
type: Tutorials
---

This tutorial shows how to provision clusters with Kyma Runtimes (Runtimes) on Google Cloud Platform (GCP).

## Prerequisites

- Existing project on GCP
- Service account on GCP with the following roles:
    * Compute Admin
    * Kubernetes Engine Admin
    * Kubernetes Engine Cluster Admin
    * Service Account User
- Key generated for your service account downloaded in the `json` format
- Secret from the service account key created in the `compass-system` Namespace

## Provision Kyma Runtime on GCP

To provision Kyma Runtime, make a call to the Runtime Provisioner with this example mutation:

> **NOTE:** To access the Runtime Provisioner, make a call from another Pod in the cluster containing the Runtime Provisioner or forward the port on which the GraphQL Server is listening.

> **NOTE:** The cluster name must start with a lowercase letter followed by up to 39 lowercase letters, numbers, or hyphens, and cannot end with a hyphen.

```graphql
mutation { 
  provisionRuntime(
    id:"309051b6-0bac-44c8-8bae-3fc59c12bb5c" 
    config: {
      clusterConfig: {
        gcpConfig: {
          name: "{CLUSTER_NAME}"
          projectName: "{GCP_PROJECT_NAME}"
          kubernetesVersion: "1.13"
          bootDiskSizeGB: 30
          numberOfNodes: 1
          machineType: "n1-standard-4"
          region: "europe-west3-a"
         }
      }
      kymaConfig: {
        version: "1.5"
        modules: Backup
      }
      credentials: {
        secretName: "{SECRET_NAME}"
      }
    }
  )
}
```

A successful call returns the ID of the provisioning operation:

```graphql
{
  "data": {
    "provisionRuntime": "e9c9ed2d-2a3c-4802-a9b9-16d599dafd25"
  }
}
```

The operation of provisioning is asynchronous. Use the provisioning operation ID to check the Runtime Operation Status and verify that the provisioning was successful. 

## Check the Runtime Status

Make a call to the Runtime Provisioner to check the Runtime status. Pass the Runtime ID as `id`. 

```graphql
query { runtimeStatus(id: "309051b6-0bac-44c8-8bae-3fc59c12bb5c") {
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
        bootDiskSizeGB name numberOfNodes kubernetesVersion projectName machineType zone region }
      ... on GardenerConfig { name workerCidr region diskType maxSurge nodeCount volumeSizeGB projectName machineType targetSecret autoScalerMin autoScalerMax provider maxUnavailable kubernetesVersion 
        } 
      } 
    } 
  } 
}
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
        "operation": "Provision",
        "state": "Succeeded",
        "runtimeID": "309051b6-0bac-44c8-8bae-3fc59c12bb5c",
        "id": "e9c9ed2d-2a3c-4802-a9b9-16d599dafd25"
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