# Compass installation

You can install Compass both on a cluster or on your local machine.

## Cluster installation

Follow these steps to install Compass on a cluster:

1. [Install Kyma with the Runtime Agent.](https://kyma-project.io/docs/master/components/runtime-agent#installation-installation)
2. Apply the following ConfigMap before you proceed with the Compass installation:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: compass-overrides
      namespace: compass-installer
      labels:
        installer: overrides
        component: compass
        kyma-project.io/installation: ""
    data:
          # The name of the currently used gateway
          global.istio.gateway.name: "kyma-gateway"
          # The Namespace of the currently used gateway
          global.istio.gateway.namespace: "kyma-system"
          global.connector.secrets.ca.name: "connector-service-app-ca"
          # The Namespace with a Secret that contains a certificate for the Connector Service
          global.connector.secrets.ca.namespace: "kyma-integration"
          # The parameter that enables the Compass gateway, as the default Kyma gateway is disabled in this installation mode
          gateway.gateway.enabled: "false"
          global.disableLegacyConnectivity: "true"
    ```

3. Install Compass. ​There are three possible installation options:

    | Installation option     	| Value to use in the installation command   	| Example value          	|
    |-------------------------	|-------------------	|-------------------------	|
    | From the `master` branch 	| `master`          	| `master`                	|
    | From the specific commit on the `master` branch 	| `master-{COMMIT_HASH}` 	| `master-34edf09a` 	|
    | From the specific PR       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

    Once you decide on the installation option, use this command:
    ```bash
    kubectl apply -f https://storage.cloud.google.com/kyma-development-artifacts/compass/{INSTALLATION_OPTION}/compass-installer.yaml
    ```
​
4. Check the installation progress. To do so, download the script that checks the progress of the installation:
```bash
wget https://storage.cloud.google.com/kyma-development-artifacts/compass/{INSTALLATION_OPTION}/is-installed.sh && chmod +x ./is-installed.sh
```
​
Then, use the script to check the progress of the Compass installation:
```bash
./is-installed.sh
```

## Local installation

For local development, install Compass along with the minimal Kyma installation from the `master` branch. To do so, run this script:
```bash
./installation/cmd/run.sh
```

You can also specify Kyma version, such as 1.6 or newer:
```bash
./installation/cmd/run.sh {version}
```
