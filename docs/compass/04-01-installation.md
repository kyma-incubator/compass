# Compass installation

You can install Compass both on a cluster and on your local machine in two modes.

## Compass cluster with essential Kyma components

Compass as a central Management Plane cluster requires minimal Kyma installation. Steps to perform the installation vary depending on the installation environment.

### Cluster installation

To install Compass as central Management Plane on a cluster, follow these steps:

1. Select installation option for Compass and Kyma. ​There are three possible installation options:

    | Installation option     	| Value to use with the installation command   	| Example value          	|
    |-------------------------	|-------------------	|-------------------------	|
    | From the Compass `master` branch 	| `master`          	| `master`                	|
    | From a specific commit on the Compass `master` branch 	| `master-{COMMIT_HASH}` 	| `master-34edf09a` 	|
    | From a specific PR on the Compass repository       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

    The Kyma version is read from the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file on a specific commit.

    Once you decide on the installation option, use this command:

    ```bash
    export INSTALLATION_OPTION={CHOSEN_INSTALLATION_OPTION_HERE}
    ```
1. Prepare the cluster for custom installation. Read how to prepare the cluster [with the `xip.io` domain](https://kyma-project.io/docs/#installation-install-kyma-on-a-cluster-prepare-the-cluster) or [with a custom domain](https://kyma-project.io/docs/#installation-install-kyma-with-your-own-domain-prepare-the-cluster). Remember to apply all global overrides in both the `kyma-installer` and `compass-installer` Namespaces.
1. Apply overrides using the following command from the root directory of the Compass repository:

    ```bash
    kubectl create namespace kyma-installer || true \
        && kubectl apply -f ./installation/resources/installer-overrides-compass-gateway.yaml
    ```
1. Perform minimal Kyma installation with the following command:

    ```bash
    kubectl apply -f "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/kyma-installer.yaml"
    ```
1. Check the Kyma installation progress. To do so, download the script and check the progress of the installation:
    
    ```bash
    source <(curl -s "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/is-kyma-installed.sh")
    ```
1. Perform Kyma post-installation steps for a cluster [with the `xip.io` domain](https://kyma-project.io/docs/#installation-install-kyma-on-a-cluster-post-installation-steps) or [with a custom domain](https://kyma-project.io/docs/#installation-install-kyma-with-your-own-domain-configure-dns-for-the-cluster-load-balancer).
1. Install Compass with the following command: 

    ```bash
    kubectl apply -f "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/compass-installer.yaml"
    ```​   
1. Check the Compass installation progress. To do so, download the script and check the progress of the installation:
   
     ```bash
    source <(curl -s "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/is-installed.sh")
    ```

### Local Minikube installation

For local development, install Compass with the minimal Kyma installation on Minikube from the `master` branch. To do so, run this script:

```bash
./installation/cmd/run.sh
```

The Kyma version is read from the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file. You can override it with the following command:

```bash
./installation/cmd/run.sh --kyma-release {KYMA_VERSION}
```
You can also specify if you want the Kyma installation to contain only `minimal` components or whether you want `full` Kyma

```bash
./installation/cmd/run.sh --kyma-installation full
```

## Single cluster with Compass and Runtime Agent

You can install Compass on a single cluster with all Kyma components, including the Runtime Agent. In this mode, the Runtime Agent is already connected to Compass. This mode is useful for all kind of testing and development purposes.

### Cluster installation

To install Compass and Runtime components on a single cluster, follow these steps:

1. [Install Kyma with the Runtime Agent](https://kyma-project.io/docs/master/components/runtime-agent#installation-installation).
1. Apply the required overrides using the following command:

    ```bash
    kubectl create namespace compass-installer || true && cat <<EOF | kubectl apply -f -
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
          global.agentPreconfiguration: "true"
    EOF
    ```

    > **NOTE:** If you installed Kyma on a cluster with a custom domain, remember to apply global overrides to the `compass-installer` Namespace as well. To do that, run this command:

    ```bash
    kubectl get configmap -n kyma-installer {OVERRIDE_NAME} -oyaml --export | kubectl apply -n compass-installer -f -
    ```
1. Install Compass. ​There are three possible installation options:

    | Installation option     	| Value to use with the installation command   	| Example value          	|
    |-------------------------	|-------------------	|-------------------------	|
    | From the `master` branch 	| `master`          	| `master`                	|
    | From a specific commit on the `master` branch 	| `master-{COMMIT_HASH}` 	| `master-34edf09a` 	|
    | From a specific PR       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

    Once you decide on the installation option, use these commands:
    
    ```bash
    export INSTALLATION_OPTION={CHOSEN_INSTALLATION_OPTION}
    kubectl apply -f "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/compass-installer.yaml"
    ```
1. Check the Compass installation progress. To do so, download the script and check the progress of the installation:
    
    ```bash
    source <(curl -s "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/is-installed.sh")
    ```

Once Compass is installed, Runtime Agent will be configured to fetch the Runtime configuration from the Compass installation within the same cluster.

### Local Minikube installation

To install Compass and Runtime components on Minikube, run the following command. Kyma source code will be picked up according to KYMA_VERSION file and compass source code will be picked up from local sources (locally checked out branch):

```bash
./installation/cmd/run.sh --kyma-installation full
```

> **Note:** In order to reduce memory and CPU usage, from the `installer-cr-kyma.yaml` file, comment out the components you don't want to use, such as `monitoring`, `tracing`, `logging`, and `kiali`.