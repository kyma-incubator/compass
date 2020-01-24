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
| **nodeCount** | NodeCount   | int | Specifies the amount of Nodes in a cluster. |  |  |
| **volumeSizeGb** | VolumeSizeGb | int |  |  |  |
| **machineType** | MachineType  | string | The possible values are `n1-standard-2`, `n1-standard-4`, `n1-standard-8`, `n1-standard-16`, `n1-standard-32`, and `n1-standard-64`. |  |  |
| **region** | Region | string | Defines the cluster region. The possible values are `westeurope`, `eastus`, `eastus2`, `centralus`, `northeurope`, `southeastasia`, `japaneast`, `westus2`, and `uksouth`. |  |  |
| **zone** | Zone | string | Defines the cluster zone. |  |  |
| **autoScalerMin** | AutoScalerMin | int |  |  |  |
| **autoScalerMax** | AutoScalerMax | int |  |  |  |
| **maxSurge** | MaxSurge | int |  |  |  |
| **maxUnavailable** | MaxUnavailable | int | Specifies the IAM Role name to provision with. It must be used in combination with **target_account_id**. |  |  |
| **components** | Components | array | Specifies components that you want to be installed on Kyma Runtime. The possible values are `monitoring`, `kiali`, `loki`, and `jaeger`. |  |  |
| **ProviderSpecificConfig.AzureConfig.VnetCidr** |  | | Provides configuration variables specific for Azure. | | `10.250.0.0/19` |
