# Compass installation

You can install Compass both on a cluster and on your local machine in two modes.

## Compass cluster with essential Kyma components

Compass as a central Management Plane cluster requires minimal Kyma installation. Steps to perform the installation vary depending on the installation environment.

### Cluster installation

To install Compass as central Management Plane on cluster, follow these steps:

1. Select installation option for Compass and Kyma. ​There are three possible installation options:

    | Installation option     	| Value to use with the installation command   	| Example value          	|
    |-------------------------	|-------------------	|-------------------------	|
    | From the Compass `master` branch 	| `master`          	| `master`                	|
    | From the specific commit on the Compass `master` branch 	| `master-{COMMIT_HASH}` 	| `master-34edf09a` 	|
    | From the specific PR on the Compass repository       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

    The Kyma version is read from the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file on a specific commit.

    Once you decide on the installation option, use this command:

    ```bash
    export INSTALLATION_OPTION={installationOption}
    ```

1. Prepare the cluster for custom installation. Read how to prepare the cluster [with `xip.io` domain](https://kyma-project.io/docs/#installation-install-kyma-on-a-cluster-prepare-the-cluster) or [with custom domain](https://kyma-project.io/docs/#installation-install-kyma-with-your-own-domain-prepare-the-cluster). Remember to apply all global overrides twice - in both `kyma-installer` and `compass-installer` namespaces.
1. Apply overrides using the following command from the root directory of the Compass repository:

    ```bash
    kubectl create namespace kyma-installer || true \
        && kubectl apply -f ./installation/resources/installer-overrides-compass-gateway.yaml
    ```

2. Perform minimal Kyma installation with the following command:

    ```bash
    kubectl apply -f "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/kyma-installer.yaml"
    ```
​
1. Check the Kyma installation progress. To do so, download the script and check the progress of the installation:
    ```bash
    source <(curl -s "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/is-kyma-installed.sh")
    ```

1. Perform Kyma post-installation steps for cluster [with `xip.io` domain](https://kyma-project.io/docs/#installation-install-kyma-on-a-cluster-post-installation-steps) or [with custom domain](https://kyma-project.io/docs/#installation-install-kyma-with-your-own-domain-configure-dns-for-the-cluster-load-balancer).
1. Install Compass with the following command: 

    ```bash
    kubectl apply -f "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/compass-installer.yaml"
    ```
​
1. Check the Compass installation progress. To do so, download the script and check the progress of the installation:
    ```bash
    source <(curl -s "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTI§ON}/is-installed.sh")
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

### Cluster installation

To install Compass and Runtime components in a single cluster, follow these steps:

1. [Install Kyma with the Runtime Agent](https://kyma-project.io/docs/master/components/runtime-agent#installation-installation). You can use the `installer-cr-cluster-runtime.yaml.tpl` Installation CR both for cluster and Minikube installation.

    **NOTE:** For local Minikube installation, in order to reduce memory and CPU usage, from the `installer-cr-cluster-runtime.yaml.tpl` Installation CR you can comment out the components you don't want to use, such as `monitoring`, `tracing`, `logging`, `kiali`.
    
    kyma provision minikube
    kyma install -c installer-cr-cluster-runtime.yaml.tpl -o kyma/installation/resources/installer-config-local.yaml.tpl

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

    For local installation, apply additional local overrides using the following command:

    ```bash
    MINIKUBE_IP=$(minikube ip)
    cat <<EOF | kubectl apply -f -
    $(sed -e 's/\.minikubeIP: .*/\.minikubeIP: '"${MINIKUBE_IP}"'/g' installation/resources/installer-config-local.yaml.tpl)
    EOF
    ```

1. Install Compass. ​There are three possible installation options:

    | Installation option     	| Value to use with the installation command   	| Example value          	|
    |-------------------------	|-------------------	|-------------------------	|
    | From the `master` branch 	| `master`          	| `master`                	|
    | From the specific commit on the `master` branch 	| `master-{COMMIT_HASH}` 	| `master-34edf09a` 	|
    | From the specific PR       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

    Once you decide on the installation option, use these commands:
    ```bash
    export INSTALLATION_OPTION={installationOption}
    kubectl apply -f "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/compass-installer.yaml"
    ```

1. Check the Compass installation progress. To do so, download the script and check the progress of the installation:
    ```bash
    source <(curl -s "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/is-installed.sh")
    ```

    For local machine, run the following command:
    ```bash
    sudo sh -c 'echo "\n$(minikube ip) adapter-gateway.kyma.local adapter-gateway-mtls.kyma.local compass-gateway-mtls.kyma.local compass-gateway-auth-oauth.kyma.local compass-gateway.kyma.local compass.kyma.local compass-mf.kyma.local kyma-env-broker.kyma.local" >> /etc/hosts'
    ```

Once Compass is installed, Runtime Agent will be configured to fetch configuration from the Compass installation within the same cluster.

### Local Minikube installation

To install Compass and Runtime components on a Minikube, follow these steps:

1. Use Kyma CLI to provision Minikube cluster:

    ```bash
    kyma provision minikube
    ```
1. Prepare Installation CR and overrides:
    
    ```bash
    wget https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/installer-cr-cluster-runtime.yaml.tpl
    wget https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/installer-config-local.yaml.tpl
    cat <<EOF > disable-legacy-connectivity-override.yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: disable-legacy-connectivity-override
      namespace: kyma-installer
      labels:
        installer: overrides
        kyma-project.io/installation: ""
    data:
      global.disableLegacyConnectivity: "true"
    EOF
    ```
1. In order to reduce memory and CPU usage, from the downloaded `installer-cr-cluster-runtime.yaml.tpl` file, comment out the components you don't want to use, such as `monitoring`, `tracing`, `logging`, `kiali`.
    
1. Install Kyma with the Runtime Agent using the prepared files: 

    ```bash
    kyma install -c installer-cr-cluster-runtime.yaml.tpl -o installer-config-local.yaml.tpl -o disable-legacy-connectivity-override.yaml
    ```

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
    MINIKUBE_IP=$(minikube ip)
    cat <<EOF | kubectl apply -f -
    $(sed -e 's/\.minikubeIP: .*/\.minikubeIP: '"${MINIKUBE_IP}"'/g' installation/resources/installer-config-local.yaml.tpl)
    EOF
    ```

1. Install Compass. ​There are three possible installation options:

    | Installation option     	| Value to use with the installation command   	| Example value          	|
    |-------------------------	|-------------------	|-------------------------	|
    | From the `master` branch 	| `master`          	| `master`                	|
    | From the specific commit on the `master` branch 	| `master-{COMMIT_HASH}` 	| `master-34edf09a` 	|
    | From the specific PR       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

    Once you decide on the installation option, use these commands:
    ```bash
    export INSTALLATION_OPTION={installationOption}
    kubectl apply -f "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/compass-installer.yaml"
    ```

1. Check the Compass installation progress. To do so, download the script and check the progress of the installation:
    ```bash
    source <(curl -s "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/is-installed.sh")
    ```

1. Once the Compass installation is finished, run the following command:
    
    ```bash
    sudo sh -c 'echo "\n$(minikube ip) adapter-gateway.kyma.local adapter-gateway-mtls.kyma.local compass-gateway-mtls.kyma.local compass-gateway-auth-oauth.kyma.local compass-gateway.kyma.local compass.kyma.local compass-mf.kyma.local kyma-env-broker.kyma.local" >> /etc/hosts'
    ```

Runtime Agent will be configured to fetch configuration from the Compass installation within the same cluster.
