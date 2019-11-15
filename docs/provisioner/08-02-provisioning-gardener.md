---
title: Provision clusters on GCP, AWS, or AZURE through Gardener
type: Tutorials
---

This tutorial shows how to provision clusters with Kyma Runtimes (Runtimes) on Google Cloud Platform (GCP), Microsoft Azure, and Amazon Web Services (AWS) using [Gardener](https://dashboard.garden.canary.k8s.ondemand.com).

## Prerequisites

<div tabs name="Prerequisites" group="Provisioning-Gardener">
  <details>
  <summary>
  GCP
  </summary>
  
  - Existing project on GCP
  - Existing project on Gardener
  - Service account for GCP with the following roles:
      * Service Account Admin
      * Service Account Token Creator
      * Service Account User
      * Compute Admin
  - Key generated for your service account downloaded in the `json` format
  </details>
  
  <details>
  <summary>
  Azure
  </summary>
  
  - Existing project on Gardener
  - Valid Azure Subscription with the Contributor role and the Subscription ID 
  - Existing App registration on Azure and its following credentials:
    * Application ID (Client ID)
    * Directory ID (Tenant ID)
    * Client secret (application password)

  </details>
  
  <details>
  <summary>
  AWS
  </summary>
  
    
  </details>
</div>

## Provision Kyma Runtime through Gardener

<div tabs name="Provisioning" group="Provisioning-Gardener">
  <details>
  <summary>
  GCP
  </summary>

  To provision Kyma Runtime on GCP, follow these steps:

  1. Access your project on [Gardener](https://dashboard.garden.canary.k8s.ondemand.com).

  2. In the **Secrets** tab, add a new Google Secret for GCP. Use the `json` file with the service account key you downloaded from GCP.

  3. In the **Members** tab, create a service account for Gardener. 

  4. Download the service account configuration (`kubeconfig.yaml`) and use it to create a Secret in the `compass-system` Namespace.

  5. Make a call to the Runtime Provisioner to create a cluster on GCP.

      > **NOTE:** To access the Runtime Provisioner, make a call from another Pod in the cluster containing the Runtime Provisioner or forward the port on which the GraphQL Server is listening.
    
      > **NOTE:** The cluster name must not be longer than 21 characters.                                                                    
                                                                          
      ```graphql
      mutation { provisionRuntime(id:"61d1841b-ccb5-44ed-a9ec-45f70cd2b0d6" config: {
        clusterConfig: {
          gardenerConfig: {
            name: "{CLUSTER_NAME}" 
            projectName: "{GARDENER_PROJECT_NAME}" 
            kubernetesVersion: "1.15.4"
            diskType: "pd-standard"
            volumeSizeGB: 30
            nodeCount: 3
            machineType: "n1-standard-4"
            region: "europe-west4"
            provider: "gcp"
              seed: "gcp-eu1"
            targetSecret: "{GCP_SERVICE_ACCOUNT_KEY_SECRET_NAME}"
            workerCidr: "10.250.0.0/19"
            autoScalerMin: 2
            autoScalerMax: 4
            maxSurge: 4
            maxUnavailable: 1
            providerSpecificConfig: { 
              gcpConfig: {
                zone: "europe-west4-a"
              }
            }
          }
        }
        kymaConfig: {
          version: "1.5"
          modules: Backup
        }
        credentials: {
          secretName: "{GAREDENER_SERVICE_ACCOUNT_CONFIGURATION_SECERT_NAME}" 
        }
      }
      )}
      ```
    
      A successful call returns the ID of the provisioning operation:
    
      ```graphql
      {
        "data": {
          "provisionRuntime": "7a8dc760-812c-4a35-a5fe-656a648ee2c8"
        }
      }
      ```
    
      The operation of provisioning is asynchronous. Use the provisioning operation ID to check the Runtime Operation Status and verify that the provisioning was successful.
  </details>

  <details>
  <summary>
  Azure
  </summary>

  To provision Kyma Runtime on Azure, follow these steps:

  1. Access your project on [Gardener](https://dashboard.garden.canary.k8s.ondemand.com).

  2. In the **Secrets** tab, add a new Google Secret for Azure. Use the credentials you got from Azure.

  3. In the **Members** tab, create a service account for Gardener. 

  4. Download the service account configuration (`kubeconfig.yaml`) and use it to create a Secret in the `compass-system` Namespace.

  5. Make a call to the Runtime Provisioner to create a cluster on Azure.

      > **NOTE:** To access the Runtime Provisioner, make a call from another Pod in the cluster containing the Runtime Provisioner or forward the port on which the GraphQL Server is listening.
    
      > **NOTE:** The cluster name must not be longer than 21 characters.                                                                    
                                                                          
      ```graphql
      mutation { provisionRuntime(id:"61d1841b-ccb5-44ed-a9ec-45f70cd1b0d3" config: {
        clusterConfig: {
          gardenerConfig: {
            name: "{CLUSTER_NAME}"
            projectName: "{GARDENER_PROJECT_NAME}"
            kubernetesVersion: "1.15.4"
            diskType: "Standard_LRS"
            volumeSizeGB: 35
            nodeCount: 3
            machineType: "Standard_D2_v3"
            region: "westeurope"
            provider: "azure"
            seed: "az-eu1"
            targetSecret: "{AZURE_APP_REGISTRATION_CLIENT_SECRET}"
            workerCidr: "10.250.0.0/19"
            autoScalerMin: 2
            autoScalerMax: 4
            maxSurge: 4
            maxUnavailable: 1
            providerSpecificConfig: { 
              azureConfig: {
                vnetCidr: "10.250.0.0/19"
              }
            }
          }
        }
        kymaConfig: {
          version: "1.5"
          modules: Backup
        }
        credentials: {
          secretName: "{GARDENER_SERVICE_ACCOUNT_CONFIGURATION_SECRET_NAME}"
        }
      }
      )}
      ```
    
      A successful call returns the ID of the provisioning operation:
    
      ```graphql
      {
        "data": {
          "provisionRuntime": "af0c8122-27ee-4a36-afa5-6e26c39929f2"
        }
      }
      ```
    
      The operation of provisioning is asynchronous. Use the provisioning operation ID to check the Runtime Operation Status and verify that the provisioning was successful.
  </details>
  
  <details>
    <summary>
    AWS
    </summary>
      
    
  </details>
    
</div>

## Check the Runtime Operation Status

Make a call to the Runtime Provisioner to verify that provisioning succeeded. Pass the ID of the provisioning operation as `id`.

```graphql
query { 
  runtimeOperationStatus(id: "7a8dc760-812c-4a35-a5fe-656a648ee2c8") { 
    operation 
    state 
    message 
    runtimeID
  } 
}
```

A successful call returns a response which includes the status of the provisioning operation (`state`) and the id of the provisioned Runtime (`runtimeID`):

```graphql
{
  "data": {
    "runtimeOperationStatus": {
      "operation": "Provision",
      "state": "Succeeded",
      "message": "Operation succeeded.",
      "runtimeID": "61d1841b-ccb5-44ed-a9ec-45f70cd2b0d6"
    }
  }
}
```

The `Succeeded` status means that the provisioning was successful and the cluster was created.

If you get the `InProgress` status, it means that the provisioning has not yet finished. In that case, wait a few moments and check the status again. 

## Check the Runtime Status

Make a call to the Runtime Provisioner to check the Runtime Status. Pass the Runtime ID as `id`. 

```graphql
query { runtimeStatus(id: "61d1841b-ccb5-44ed-a9ec-45f70cd2b0d6") {
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
        "operation": "Provision",
        "state": "Succeeded",
        "runtimeID": "61d1841b-ccb5-44ed-a9ec-45f70cd2b0d6",
        "id": "7a8dc760-812c-4a35-a5fe-656a648ee2c8"
      },
      "runtimeConfiguration": {
        "kubeconfig": "{KUBECONFIG}",
        "kymaConfig": {
          "version": "1.5",
          "modules": [
            "Backup"
          ]
        },
        "clusterConfig": {
          "__typename": "GardenerConfig",
          "name": "{CLUSTER_NAME}",
          "workerCidr": "10.250.0.0/19",
          "region": "europe-west4",
          "diskType": "pd-standard",
          "maxSurge": 4,
          "nodeCount": 3,
          "volumeSizeGB": 30,
          "projectName": "{GARDENER_PROJECT_NAME}",
          "machineType": "n1-standard-4",
          "targetSecret": "{SERVICE_ACCOUNT_KEY_SECRET_NAME}",
          "autoScalerMin": 2,
          "autoScalerMax": 4,
          "provider": "{PROVIDER}",
          "maxUnavailable": 1,
          "kubernetesVersion": "1.15.4"
        }
      }
    }
  }
}
```