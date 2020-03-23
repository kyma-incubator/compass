# Set overrides for provisioning runtime

To provision cluster with custom overrides you can add `Secret` or `ConfigMap` with specific label.

Overrides can be global or for concrete components. In the second case, an appropriate label `component` should be used to indicate to which component it refers.

All overrides have to be in `compass-system` namespace.

Examples are given below:

1. ConfigMap with global overrides:
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

2. Secret with overrides for `core` component:
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
