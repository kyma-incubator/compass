# Service description

Kyma Environment Broker (KEB) is compatible with the [Open Service Broker API](https://www.openservicebrokerapi.org/) (OSBA) specification, which means that you can register a Service Class that has only the mandatory fields. However, it is recommended to provide more detailed Service Class definitions for better user experience.

## Service plans

KEB Service Class provides the following plans:

| Plan Name | Description |
|-----------|-------------|
| `azure` | Installs Kyma Runtime on the Azure cluster. |

## Provisioning parameters

This service provisions a new AWS Service Broker which provides the Amazon Web Services. The default bucket parameters provide the AWS Service Broker with default services.

These are the provisioning parameters for the given plans:


| Parameter Name | Display name | Type | Description | Required | Default value |
|----------------|-----|-------|-------------|:----------:|---------------|
| **name** | Name | string | Specifies the name of the cluster. |  |  |
| **nodeCount** | NodeCount   | int |  |  |  |
| **volumeSizeGb** | VolumeSizeGb | int |  |  |  |
| **machineType** | MachineType  | string | The possible values are `n1-standard-2`, `n1-standard-4`, `n1-standard-8`, `n1-standard-16`, `n1-standard-32`, and `n1-standard-64`. |  |  |
| **region** | Region | string | Defines the cluster region. The possible values are `westeurope`, `eastus`, `eastus2`, `centralus`, `northeurope`, `southeastasia`, `japaneast`, `westus2`, and `uksouth`. |  |  |
| **zone** | Zone | string | Defines the cluster zone. |  |  |
| **autoScalerMin** | AutoScalerMin | int |  |  |  |
| **autoScalerMax** | AutoScalerMax | int |  |  |  |
| **maxSurge** | MaxSurge | int |  |  |  |
| **maxUnavailable** | MaxUnavailable | int | Specifies the IAM Role name to provision with. It must be used in combination with **target_account_id**. |  |  |
| **components** | Components | string | Lists all the optional components. The possible values are `monitoring`, `kiali`, `loki`, and `jaeger`. |  |  |
