# Custom configuration of provisioned Runtimes

## Introduction

There is a need for configuring Kyma Runtimes provisioned by the Runtime Provisioner.
For now only option is to configure Runtime manually after provisioning is finished and Kubeconfig is fetched.


## Solution

The proposed solution is to provide ability to provide the configuration for components, which then will be used as Installer overrides.

Changes to the GraphQL schema:
```graphql
type Configuration {
    key: String!
    value: String!
    secret: Boolean
} 

input ConfigurationInput {
    key: String!
    value: String!
    secret: Boolean
}

type ComponentConfiguration {
    component: KymaComponent
    configuration: [Configuration]
}

input ComponentConfigurationInput {
    component: KymaComponent!
    configuration: [ConfigurationInput]
}

type KymaConfig {
    version: String
    components: [ComponentConfiguration]
    configuration: [Configuration]
}

input KymaConfigInput {
    version: String!
    components: [ComponentConfigurationInput]!
    configuration: [ConfigurationInput] 
}

```

### Assumption

The Provisioner API is accessed by the administrator or other intermediary Service and the generic configuration mechanisms are not available directly to the cluster user.

### Mechanism

The proposed solution leverages the mechanism of Installer.
More information on how the overrides work can be found [here](https://kyma-project.io/docs/#configuration-helm-overrides-for-kyma-installation)

### Components configuration

The configuration can be provided for every Kyma component as a list of key-value pairs in `ComponentConfigurationInput`.
The additional `secret` flag specifies if the value is confidential.
The `configuration` in `KymaConfigInput` is a configuration not specific for any component.

All configurations for component not marked as a `secret` are then saved to the ConfigMap on a created Cluster and marked as Installation overrides. They are also stored in the database.
The confidential configuration is stored in Secret on a created cluster and encrypted before storing in database. The examples of such sensitive data could be: certificate's private keys, Minio Gateway credentials or Dex configuration containing some secrets. 


### Pros
- The solution is simple from API and Installation standpoint
- It leverages the well established mechanism of Installer
- Together with component definition it gives ability to configure Kyma (open source stuff) in pretty much any way possible.
- Does not uses anything external of Kyma (like Helm etc.)
- All current use cases can be covered with the following approach

### Cons
- No ability to extend with custom components outside of Kyma
- One ConfigMap and Secret per component


### Example

The example mutation shows how to configure Minio with Azure Blob Storage Gateway mode, using the proposed mechanism:

```
mutation {
    result: provisionRuntime(id: "39e201ed-ec5b-4a2e-87cf-2395ebb0e9e7", config: {
      clusterConfig: {
          gcpConfig: {
            name: "test"
            projectName: "test"
            kubernetesVersion: "1.15"
            numberOfNodes: 2
            bootDiskSize: "30GB"
            machineType: "ns"
            region: "mordor"
          }
      }
      credentials: {
        secretName: "test"
      }
      kymaConfig: {
        version: "1.6"
        components: [
            {
                component: "assetstore"
                configuration: [
                   {
                       key: "minio.persistence.enabled"
                       value: "false"
                   }
                   {
                       key: "minio.azuregateway.enabled"
                       value: "true"
                   }
                   {
                       key: "minio.DeploymentUpdate.type"
                       value: "RollingUpdate"
                   }
                   {
                       key: "minio.DeploymentUpdate.maxSurge"
                       value: "0"
                   }
                   {
                        key: "minio.DeploymentUpdate.maxUnavailable"
                        value: "50%"
                   }
                   {
                        key: "minio.accessKey"
                        value: "azure-account"
                        secret: true
                   }
                   {
                        key: "minio.secretKey"
                        value: "secret-key"
                        secret: true
                   }
                ]
            }
            {
                component: "core"
            }
        ]
      }
    })
}
``` 
