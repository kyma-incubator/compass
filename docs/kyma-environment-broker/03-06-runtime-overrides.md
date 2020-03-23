# Set overrides for Kyma Runtime

You can set overrides to customize your Kyma Runtime. To provision a cluster with custom overrides, add a Secret or a ConfigMap with a specific label. Kyma Environment Broker uses this Secret and/or ConfigMap to prepare a request to the Runtime Provisioner.

Overrides can be either global or specified for a given component. In the second case, use the `component: {"COMPONENT_NAME"}` label to indicate the component. Create all overrides in the `compass-system` Namespace.

See the examples:

- ConfigMap with global overrides:
    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      labels:
        provisioning-runtime-override: "true"
      name: global-overrides
      namespace: compass-system
    data:
      global.enableAPIPackages: "true"
    ```  

- Secret with overrides for the `core` component:
    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      labels:
        component: "core"
        provisioning-runtime-override: "true"
      name: core-overrides
      namespace: compass-system
    data:
      database.password: YWRtaW4xMjMK
    ```  
