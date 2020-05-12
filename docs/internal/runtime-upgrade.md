# Runtime upgrade and rollback with the Runtime Provisioner

This document describes the Runtime upgrade and rollback procedures for Kyma.

## Upgrade the Runtime

The Runtime Provisioner allows you to upgrade your Kyma Runtime using the GraphQL API. To upgrade the Runtime to the 1.11.0 version with the standard Runtime components, use the following mutation:

```graphql
mutation {
  upgradeRuntime(id: "{RUNTIME_ID}", config: {
    kymaConfig: { 
        version: "1.11.0",
        configuration: [
            {
                key: "a.config.key"
                value: "a.config.value"
            }
        ]
        components: [
          {
            component: "cluster-essentials"
            namespace: "kyma-system"
          }
          {
            component: "testing"
            namespace: "kyma-system"
          }
          {
            component: "istio"
            namespace: "istio-system"
          }
          {
            component: "xip-patch"
            namespace: "kyma-installer"
          }
          {
            component: "istio-kyma-patch"
            namespace: "istio-system"
          }
          {
            component: "knative-serving-init"
            namespace: "knative-serving"
          }
          {
            component: "knative-serving"
            namespace: "knative-serving"
          }
          {
            component: "knative-eventing"
            namespace: "knative-eventing"
          }
          {
            component: "dex"
            namespace: "kyma-system"
          }
          {
            component: "ory"
            namespace: "kyma-system"
          }
          {
            component: "api-gateway"
            namespace: "kyma-system"
          }
          {
            component: "rafter"
            namespace: "kyma-system"
          }
          {
            component: "service-catalog"
            namespace: "kyma-system"
          }
          {
            component: "service-catalog-addons"
            namespace: "kyma-system"
          }
          {
            component: "helm-broker"
            namespace: "kyma-system"
          }
          {
            component: "nats-streaming"
            namespace: "natss"
          }
          {
            component: "core"
            namespace: "kyma-system"
          }
          {
            component: "permission-controller"
            namespace: "kyma-system"
          }
          {
            component: "apiserver-proxy"
            namespace: "kyma-system"
          }
          {
            component: "iam-kubeconfig-service"
            namespace: "kyma-system"
          }    
          {
            component: "knative-provisioner-natss"
            namespace: "knative-eventing"
          }    
          {
            component: "event-bus"
            namespace: "kyma-system"
          }    
          {
            component: "event-sources"
            namespace: "kyma-system"
          }    
          {
            component: "application-connector"
            namespace: "kyma-integration"
          }  
          {
            component: "compass-runtime-agent"
            namespace: "compass-system"
          }  
        ]
  }
  }
  ) {
    id
    operation
    runtimeID
  }
}
```

The Kyma configuration passed to the upgrade mutation overrides the previous configuration. The components and parts of the configuration not specified in the new version are removed.

> **CAUTION:** Before upgrading your Kyma deployment, you must perform the migration steps described in the Migration Guide matching the release you're upgrading to, if provided. If you upgrade to the new release without performing these steps, you can compromise the functionality of your cluster or make it unusable. To find the Migration Guide, go to `https://github.com/kyma-project/kyma/tree/release-{RELEASE_TAG_TO_UPGRADE_TO}/docs` and check if it contains the `migration-guides` directory. If it does, find the Migration Guide inside. If it doesn't, no steps need to be performed. For example, find the Migration Guide for Kyma 1.11 [here](https://github.com/kyma-project/kyma/blob/release-1.11/docs/migration-guides/1.10-1.11.md). Alternatively, find the link to the Migration Guide, if provided, in appropriate Release Notes on the [Kyma blog](https://kyma-project.io/blog/).

Similarly to provisioning, the upgrade operation is asynchronous and returns the [Runtime Operation Status](../provisioner/08-03-runtime-operation-status.md).

## Roll back the latest upgrade

The Runtime Provisioner API allows you to roll back the Kyma configuration to the one from before the latest upgrade. This mutation does not affect the cluster itself in any way, it only affects the database. In the case a cluster is rolled back manually, for example restored from a backup, a rollback on the database must be performed to keep the data in sync. To roll back an upgrade, the upgrade must be the last operation performed by the Runtime Provisioner on the given Runtime. In other words, you cannot perform a rollback if other operations have been performed on the Runtime after the upgrade. 

To roll back the Kyma configuration, use the following mutation:

```graphql
mutation { 
    rollBackUpgradeOperation(id: "{RUNTIME_ID}") {
        runtimeConnectionStatus {
            status errors {
            message
            } 
        } 
        lastOperationStatus {
            message operation state runtimeID id
        } 
        runtimeConfiguration {
            kubeconfig
            kymaConfig {
                version
                configuration {
                    key
                    value
                    secret
                }
                components {
                    component
                    namespace
                    configuration {
                        key
                        value
                        secret
                    }
                    sourceURL
                }
            }
            clusterConfig {
            __typename ... on GCPConfig { 
                bootDiskSizeGB 
                name 
                numberOfNodes 
                kubernetesVersion 
                projectName 
                machineType 
                zone 
                region 
            }
            ... on GardenerConfig { 
                name 
                workerCidr 
                region 
                diskType 
                maxSurge 
                volumeSizeGB
                machineType 
                targetSecret 
                autoScalerMin 
                autoScalerMax 
                provider 
                maxUnavailable 
                kubernetesVersion 
                providerSpecificConfig { 
                __typename ... on GCPProviderConfig { zones } 
                    ... on AzureProviderConfig {vnetCidr}
                ... on AWSProviderConfig {zone internalCidr vpcCidr publicCidr}      
            }
            } 
            } 
        } 
    } 
}
```

> **CAUTION:** Before rolling back your Kyma deployment, you must revert the migration steps described in the Migration Guide matching the release you upgraded to, if provided. If you roll back to the previous release without reverting these steps, you can compromise the functionality of your cluster or make it unusable. To find the Migration Guide, go to `https://github.com/kyma-project/kyma/tree/release-{RELEASE_TAG_YOU_UPGRADED_TO}/docs` and check if it contains the `migration-guides` directory. If it does, find the Migration Guide inside. If it doesn't, no steps need to be performed. For example, find the Migration Guide for Kyma 1.11 [here](https://github.com/kyma-project/kyma/blob/release-1.11/docs/migration-guides/1.10-1.11.md). Alternatively, find the link to the Migration Guide, if provided, in appropriate Release Notes on the [Kyma blog](https://kyma-project.io/blog/).

The rollback mutation returns the [Runtime Status](../provisioner/08-04-runtime-status.md).
