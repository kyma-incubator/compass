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

The rollback mutation returns the [Runtime Status](../provisioner/08-04-runtime-status.md).
