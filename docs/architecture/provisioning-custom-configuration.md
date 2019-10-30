# Custom configuration of provisioned Runtimes

## Introduction

There is a need for configuring Kyma Runtimes provisioned by the Runtime Provisioner.
For now only option is to configure Runtime manually after provisioning is finished and Kubeconfig is fetched.


## Solution

The proposed solution is to provide ability to configure the Installer overrides as a part of provisioning API.

Changes to the GraphQL schema:
```graphql
scalar Overrides # map[string]string

type KymaConfig {
    version: String
    modules: [KymaModule]
    installationOverrides: InstallationOverrides
}

input KymaConfigInput {
    version: String!
    modules: [KymaModule!]
    installationOverrides: [InstallationOverridesInput]
}

type InstallationOverrides {
    configOverrides: [ComponentOverride]
}

input InstallationOverridesInput {
    configOverrides: [ComponentOverrideInput]
    secretOverrides: [ComponentOverrideInput]
}

type ComponentOverride {
    component: String 
    overrides: Overrides
}

input ComponentOverrideInput {
    # If not specified will be used for all components
    component: String 
    overrides: Overrides
}
```

### Assumption

The Provisioner API is accessed by the administrator or other intermediary Service and the generic configuration mechanisms are not available directly to the cluster user.

### Mechanism

The proposed solution leverages the mechanism of Installer.
More information on how the overrides work can be found [here](https://kyma-project.io/docs/#configuration-helm-overrides-for-kyma-installation)

### Overrides configuration

Every `ComponentOverrideInput` defines a single Kubernetes resource which stores the provided values.
The resource can be either a ConfigMap or a Secret.

Overrides configured as the `configOverrides` in `InstallationOverridesInput` are created as a ConfigMap on a created Cluster and stored in the database.
Overrides configured as the `secretOverrides` are mapped to the Secret and either encrypted before storing in database or not stored at all.


### Possible extension

In a presented format the Provisioner API provides ability to configure Kyma during the cluster provisioning.
The same mechanism could be used to updated existing cluster thanks to such functionality in the Installer component. 


### Pros
- The solution is simple from API and Installation standpoint
- It leverages the well established mechanism of Installer
- Together with Module definition it gives ability to configure Kyma (open source stuff) in pretty much any way possible.
- Does not uses anything external of Kyma (like Helm etc.)
- All current use cases can be covered with the following approach

### Cons
- No ability to extend with custom modules outside of Kyma

