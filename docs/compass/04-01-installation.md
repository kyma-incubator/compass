# Compass installation

You can install Compass both on a cluster and on your local machine in two modes.

## Compass cluster with essential Kyma components

Compass as a central Management Plane cluster requires minimal Kyma installation. Steps to perform the installation vary depending on the installation environment.

#### Prerequisites

In case certificate rotation is needed, you can install [cert manager](https://github.com/jetstack/cert-manager) to take care of certificates.

The following certificates can be rotated:
* Connector intermediate certificate that is used to issue Application and Runtime client certificates.
* Istio gateway certificate for the regular HTTPS gateway.
* Istio gateway certificate for the MTLS gateway.

##### Create issuers

To issue certificates, the cert manager requires a resource called issuer.

Example with self provided CA certificate:

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: <name>
spec:
  ca:
    secretName: "<secret-name-containing-the-ca-cert>"
```

Example with **Let's encrypt** issuer:

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: <name>
spec:
  acme:
    email: <lets_encrypt_email>
    server: <lets_encrypt_server_url>
    privateKeySecretRef:
      name: <lets_encrypt_secret>
    solvers:
      - dns01:
          clouddns:
            project: <project_name>
            serviceAccountSecretRef:
              name: <secret_name>
              key: <secret_key>
```

For more inforamtion about the different cluster issuer configurations, see the following document: [ACME Issuer](https://cert-manager.io/docs/configuration/acme/)

##### Connector certificate

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: <name>
  namespace: <namespace>
spec:
  secretName: "<secret-to-put-the-generated-certificate>"
  duration: 3d
  renewBefore: 1d
  issuerRef:
    name: <name-of-issuer-to-issue-certificates-from>
    kind: ClusterIssuer
  commonName: Kyma
  isCA: true
  keyAlgorithm: rsa
  keySize: 4096
  usages:
    - "digital signature"
    - "key encipherment"
    - "cert sign"

```

Cert manager's role is to rotate certificates as defined in the resources. Make sure that the certificate as shown in the example is specified as `isCA: true`. This means that the certificate is used to issue other certificates.

##### Domain certificates

Cert manager can also rotate Istio gateway certificates. The following example shows such certificate:

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: <name>
  namespace: <namespace>
spec:
  secretName: "<secret-to-put-the-generated-certificate>"
  issuerRef:
    name: <name-of-issuer-to-issue-certificates-from> (this can be Let's encrypt issuer)
    kind: ClusterIssuer
  commonName: <wildcard_domain_name>
  dnsNames:
    - <wildcard_domain_name>
    - <alternative_wildcard_domain_name>
```

In this case, as this certificate is not used to issue other certificates, it is not a CA certificate. Additionally, its validity depends on the settings by the issuer (for example **Let's encrypt**).

### Cluster installation


To install Compass as central Management Plane on a cluster, follow these steps:

1. Select installation option for Compass and Kyma. ​There are three possible installation options:

   | Installation option     	| Value to use with the installation command   	| Example value          	|
   |-------------------------	|-------------------	|-------------------------	|
   | From the Compass `main` branch 	| `main`          	| `main`                	|
   | From a specific commit on the Compass `main` branch 	| `main-{COMMIT_HASH}` 	| `main-34edf09a` 	|
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

For local development, install Compass with the minimal Kyma installation on Minikube from the `main` branch. To do so, run this script:

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

Additional option is to specify if you want to populate the DB with data. That can be happen with `--dump-db` flag, using this option a DB dump will be downloaded from our dev environment. Otherwise, the database will be empty. `Note:` Keep in mind if you specified this flag then new `schema-migrator` image will be build from the local files.

```bash
./installation/cmd/run.sh --dump-db
```

## Single cluster with Compass and Runtime Agent

You can install Compass on a single cluster with all Kyma components, including the Runtime Agent. In this mode, the Runtime Agent is already connected to Compass. This mode is useful for all kind of testing and development purposes.

### Cluster installation

To install Compass and Runtime components on a single cluster, follow these steps:

1. [Install Kyma with the Runtime Agent](https://kyma-project.io/docs/main/components/runtime-agent#installation-installation).
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
   | From the `main` branch 	| `main`          	| `main`                	|
   | From a specific commit on the `main` branch 	| `main-{COMMIT_HASH}` 	| `main-34edf09a` 	|
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

To install Compass and Runtime components on Minikube, run the following command. Kyma source code will be picked up according to the KYMA_VERSION file and Compass source code will be picked up from the local sources (locally checked out branch):

```bash
./installation/cmd/run.sh --kyma-installation full
```

> **Note:** To reduce memory and CPU usage, from the `installer-cr-kyma.yaml` file, comment out the components you don't want to use, such as `monitoring`, `tracing`, `logging`, or `kiali`.
