# Compass installation

You can install Compass both on a cluster and on your local machine in two modes.

## Compass cluster with essential Kyma components

Compass as a central Management Plane cluster requires minimal Kyma installation. Steps to perform the installation vary depending on the installation environment.

### Production cluster installation

To install Compass as central Management Plane on cluster, follow these steps:

1. Perform minimal Kyma installation. To do so, run the following script:

    ```bash
    ./installation/scripts/install-minimal-kyma.sh
    ```

    The Kyma version is read from [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file. You can override it with the following command:

    ```bash
    ./installation/scripts/install-kyma-essential.sh {KYMA_VERSION}
    ```

2. Install Compass. ​There are three possible installation options:

    | Installation option     	| Value to use with the installation command   	| Example value          	|
    |-------------------------	|-------------------	|-------------------------	|
    | From the `master` branch 	| `master`          	| `master`                	|
    | From the specific commit on the `master` branch 	| `master-{COMMIT_HASH}` 	| `master-34edf09a` 	|
    | From the specific PR       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

    Once you decide on the installation option, use this command:
    ```bash
    kubectl apply -f https://storage.cloud.google.com/kyma-development-artifacts/compass/{INSTALLATION_OPTION}/compass-installer.yaml
    ```
​
1. Check the installation progress. To do so, download the script that checks the progress of the installation:
```bash
wget https://storage.cloud.google.com/kyma-development-artifacts/compass/{INSTALLATION_OPTION}/is-installed.sh && chmod +x ./is-installed.sh
```
​
Then, use the script to check the progress of the Compass installation:
```bash
./is-installed.sh
```

### Local Minikube installation

For local development, install Compass from the `master` branch, along with the minimal Kyma installation on Minikube. To do so, run this script:

```bash
./installation/cmd/run.sh
```

The Kyma version is read from [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file. You can override it with the following command:

```bash
./installation/cmd/run.sh {KYMA_VERSION}
```

## Single cluster with Compass and Runtime Agent

You can install Compass on a single cluster with all Kyma components, including the Runtime Agent. In this mode, the Runtime Agent is already connected to Compass. This mode is useful for all kind of testing and development purposes.

> **NOTE:** This mode is not supported on the local Minikube installation.

To install Compass and Runtime components in a single cluster, follow these steps:

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

    | Installation option     	| Value to use with the installation command   	| Example value          	|
    |-------------------------	|-------------------	|-------------------------	|
    | From the `master` branch 	| `master`          	| `master`                	|
    | From the specific commit on the `master` branch 	| `master-{COMMIT_HASH}` 	| `master-34edf09a` 	|
    | From the specific PR       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

    Once you decide on the installation option, use this command:
    ```bash
    kubectl apply -f https://storage.cloud.google.com/kyma-development-artifacts/compass/{INSTALLATION_OPTION}/compass-installer.yaml
    ```
​
1. Check the installation progress. To do so, download the script that checks the progress of the installation:
```bash
wget https://storage.cloud.google.com/kyma-development-artifacts/compass/{INSTALLATION_OPTION}/is-installed.sh && chmod +x ./is-installed.sh
```
​
Then, use the script to check the progress of the Compass installation:
```bash
./is-installed.sh
```

Once Compass is installed, Runtime Agent will be configured to fetch configuration from the Compass installation within the same cluster.
