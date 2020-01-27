# Service description

Kyma Environment Broker (KEB) is compatible with the [Open Service Broker API](https://www.openservicebrokerapi.org/) (OSBA) specification. It provides a ServiceClass that provisions Kyma Runtime on a cluster.

## Service plans

Currently, KEB ServiceClass provides one ServicePlan that allows you to provision Kyma Runtime on Azure.

| Plan name | Description |
|-----------|-------------|
| `azure` | Installs Kyma Runtime on the Azure cluster. |

## Provisioning parameters

These are the provisioning parameters for this plan:

| Parameter Name | Type | Description | Required | Default value |
|----------------|-------|-------------|:----------:|---------------|
| **name** | string | Specifies the name of the cluster. | Yes | None |
| **nodeCount** | int | Specifies the number of Nodes in a cluster. | No | `3` |
| **volumeSizeGb** | int | Specifies the size of the root volume. | No | `50` |
| **machineType** | string | Specifies the provider-specific Virtual Machine type. | No | `Standard_D2_v3` |
| **region** | string | Defines the cluster region. | No | `westeurope` |
| **zone** | string | Defines the cluster zone. | No | None |
| **autoScalerMin** | int | Specifies the minimum number of Virtual Machines to create. | No | `2` |
| **autoScalerMax** | int | Specifies the maximum number of Virtual Machines to create. | No | `4` |
| **maxSurge** | int | Specifies the maximum number of Virtual Machines that are created during an update. | No | `4` |
| **maxUnavailable** | int | Specifies the maximum number of VMs that can be unavailable during an update. | No | `1` |
| **components** | array | Defines optional components that are installed in Kyma Runtime. The possible values are `monitoring`, `kiali`, `loki`, and `jaeger`. | No | [] |
| **providerSpecificConfig.AzureConfig.VnetCidr** | string | Provides configuration variables specific for Azure. | No | `10.250.0.0/19` |
