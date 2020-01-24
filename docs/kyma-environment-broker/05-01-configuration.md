# Configuration

To configure the Kyma Environment Broker sub-chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.


>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* {[Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides)}
>* {[Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)}

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Name | Description | Default value |
|-----|---------|:--------:|
| **APP_PORT** | The port on which the HTTP server listens. | `8080` |
| **APP_PROVISIONING_URL** | Specifies an URL to the provisioner API. | None |
| **APP_PROVISIONING_SECRET_NAME** | Specifies the name of the Secret which holds credentials to the Provisioner API. | None |
| **APP_PROVISIONING_GARDENER_PROJECT_NAME** | Defines the used Gardener project name. | `true` |
| **APP_PROVISIONING_GCP_SECRET_NAME** | Defines the name of the Secret which holds credentials to GCP. | None |
| **APP_PROVISIONING_AWS_SECRET_NAME** | Defines the name of the Secret which holds credentials to AWS. | None |
| **APP_PROVISIONING_AZURE_SECRET_NAME** | Defines the name of the Secret which holds credentials to Azure. | None |
| **APP_AUTH_USERNAME** | Specifies the Kyma Environment Service Broker authentication username. | None |
| **APP_AUTH_PASSWORD** | Specifies the Kyma Environment Service Broker authentication password. | None |
