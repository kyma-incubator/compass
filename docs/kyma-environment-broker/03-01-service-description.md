# Service description

Kyma Environment Broker (KEB) is compatible with the [Open Service Broker API](https://www.openservicebrokerapi.org/) (OSBA) specification. It provides a ServiceClass that provisions Kyma Runtime on a cluster.

## Service plans

Currently, KEB ServiceClass provides one ServicePlan that allows you to provision Kyma Runtime on Azure.

| Plan name | Description |
|-----------|-------------|
| `azure` | Installs Kyma Runtime on the Azure cluster. |

## Provisioning parameters

These are the provisioning parameters for this plan:

| Parameter Name | Display name | Type | Description | Required | Default value |
|----------------|-----|-------|-------------|:----------:|---------------|
| **name** | Name | string | Specifies the name of the cluster. |  |  |
| **nodeCount** | NodeCount   | int | Specifies the number of Nodes in a cluster. |  | `3` |
| **volumeSizeGb** | VolumeSizeGb | int |  |  | `50` |
| **machineType** | MachineType  | string | Specifies the Virtual Machine type. The possible values are `n1-standard-2`, `n1-standard-4`, `n1-standard-8`, `n1-standard-16`, `n1-standard-32`, and `n1-standard-64`. |  | `Standard_D2_v3` |
| **region** | Region | string | Defines the cluster region. The possible values are `westeurope`, `eastus`, `eastus2`, `centralus`, `northeurope`, `southeastasia`, `japaneast`, `westus2`, and `uksouth`. |  | `westeurope` |
| **zone** | Zone | string | Defines the cluster zone. |  |  |
| **autoScalerMin** | AutoScalerMin | int | Specifies the minimum number of Virtual Machines to create. |  | `2` |
| **autoScalerMax** | AutoScalerMax | int | Specifies the maximum number of Virtual Machines to create. |  | `4` |
| **maxSurge** | MaxSurge | int | Specifies the maximum number of Virtual Machines that are created during an update. |  | `4` |
| **maxUnavailable** | MaxUnavailable | int | Specifies the maximum number of VMs that can be unavailable during an update. |  | `1` |
| **components** | Components | array | Defines optional components that are installed in Kyma Runtime. The possible values are `monitoring`, `kiali`, `loki`, and `jaeger`. |  |  |
| **providerSpecificConfig.AzureConfig.VnetCidr** | ProviderSpecificConfig | | Provides configuration variables specific for Azure. | | `10.250.0.0/19` |
