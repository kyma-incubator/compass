# Compass installation

You can install Compass both on a cluster and on your local machine in two modes - single Kyma cluster and Compass as a Central Management Plane.

The Compass version that will be installed can be one of the following:
 | Installation option     	| Value to use with the installation command   	| Example value          	|
   |-------------------------	|-------------------	|-------------------------	|
   | From the Compass `main` branch 	| `main`          	| `main`                	|
   | From a specific commit on the Compass `main` branch 	| `main-{COMMIT_HASH}` 	| `main-34edf09a` 	|
   | From a specific PR on the Compass repository       	| `PR-{PR_NUMBER}`         	| `PR-1420`     	|

   The Kyma version is read from the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file on a specific commit.

## Prerequisites for cluster installation

### Managed PostgreSQL Database

You can check [this guide](https://github.com/kyma-incubator/compass/blob/main/chart/compass/configure-managed-gcp-postgresql.md) for detailed steps on how you can use GCP managed PostgreSQL database instance with Compass.

### Custom domain
Check the Kyma installation guide for using [custom domain](https://github.com/kyma-project/kyma/blob/1.24.11/docs/kyma/04-04-use-your-own-domain.md), and the resources for [certificate management](#certificate-management) in this document.

 > **NOTE:** If you installed Kyma on a cluster with a custom domain, remember to apply global overrides to the `compass-installer` Namespace as well. To do that, run this command (you should have `yq` installed prior to that), or manually replicate the overrides ConfigMap from `kyma-installer` to `compass-installer` Namespace:
    
  ```bash
  kubectl get configmap -n kyma-installer {OVERRIDE_NAME} -o yaml \
  | yq eval 'del(.metadata.resourceVersion, .metadata.uid, .metadata.annotations, .metadata.creationTimestamp, .metadata.selfLink, .metadata.managedFields, .metadata.namespace)' -
  ```

### Certificate Management

In case certificate rotation is needed, you can install JetStack's [Certificate Manager](https://github.com/jetstack/cert-manager) to take care of the certificates.

The following certificates can be rotated:
* Connector intermediate certificate that is used to issue Application and Runtime client certificates.
* Istio gateway certificate for the regular HTTPS gateway.
* Istio gateway certificate for the mTLS gateway.

#### Create issuers

To issue certificates, the Cert Manager requires a resource called issuer.

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

For more information about the different cluster issuer configurations, see the following document: [ACME Issuer](https://cert-manager.io/docs/configuration/acme/)

#### Connector certificate

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

Cert Manager's role is to rotate certificates as defined in the resources. Make sure that the certificate as shown in the example is specified as `isCA: true`. This means that the certificate is used to issue other certificates.

#### Domain certificates

Cert Manager can also rotate Istio gateway certificates. The following example shows such certificate:

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

## Compass as a Central Management Plane

This is a multi-cluster mode in which you need one cluster with Compass. This mode allows you to integrate your Runtimes with Applications and manage them in one central place.

Compass as a Central Management Plane cluster requires minimal Kyma installation. Steps to perform the installation vary depending on the installation environment.

### Cluster installation

1. Select installation option for Compass and Kyma. ​Then use this command:
    ```bash
    export INSTALLATION_OPTION={CHOSEN_INSTALLATION_OPTION_HERE}
    ```
1. Prepare the cluster for custom Kyma installation - check the prerequisites from above and the ones from the [Kyma documentation](https://github.com/kyma-project/kyma/blob/1.24.11/docs/kyma/04-03-cluster-installation.md) depending on infrastructure your provider. 

1. Apply overrides using the following command from the root directory of the Compass repository:
    ```bash
    kubectl create namespace kyma-installer || true \
        && kubectl apply -f ./installation/resources/kyma/installer-cr-kyma-minimal.yaml
    ```
    >**NOTE:** Remember to apply all global overrides in both the `kyma-installer` and `compass-installer` Namespaces _before_ starting the respective installations.

1. Perform minimal Kyma installation with the following command:
    ```bash
    kubectl apply -f "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/kyma-installer.yaml"
    ```

1. Check the Kyma installation progress. To do so, download the script and check the progress of the installation:
    ```bash
    source <(curl -s "https://storage.googleapis.com/kyma-development-artifacts/compass/${INSTALLATION_OPTION}/is-kyma-installed.sh")
    ```

1. Perform Kyma [post-installation steps](https://github.com/kyma-project/kyma/blob/1.24.11/docs/kyma/04-03-cluster-installation.md#post-installation-steps) if required.

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

Optionally, you can use the `--dump-db` flag to populate the DB with sample data. As a result, a DB dump is downloaded from the Compass development environment and is imported into the DB during the installation of Compass. Keep in mind, you can only use this feature if you are part of the Compass contributors, otherwise you won't have access to the development environment from which the data is obtained.
Note that using this flag results in building a new `schema-migrator` image from the local files.
```bash
./installation/cmd/run.sh --dump-db
```

>**NOTE:** The versions of the components that will be installed depends on their tag versions listed in the `values.yaml` file in the Helm charts.
If you want to build and deploy the local source code version of a component (e.g. Director), you can run the following command in the component directory:
  ```bash
  make deploy-on-minikube
  ```

## Single cluster with Compass and Runtime Agent

This is a single-tenant mode which provides the complete cluster Kyma installation with all components, including the Runtime Agent. You can install Compass top of that.
In this mode, the Runtime Agent is already connected to Compass. This mode is useful for all kind of testing and development purposes.

### Cluster installation

To install Compass and Runtime components on a single cluster, follow these steps:

1. Apply [this configuration](https://github.com/kyma-project/kyma/blob/1.24.11/docs/runtime-agent/04-01-installation-modes.md) that will enable the Runtime Agent and [install Kyma](https://github.com/kyma-project/kyma/blob/1.24.11/docs/kyma/04-03-cluster-installation.md). 
1. Apply the required overrides configuration in Compass, that will enable the automatic registration of the Kyma runtime into Compass:
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
      global.agentPreconfiguration: "true"
    EOF
    ```

1. Choose an installation option from the ones listed at the beginning of this document and install Compass:
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

To install Compass and Runtime components on Minikube, run the command below. Kyma source code will be picked up according to the KYMA_VERSION file and Compass source code will be picked up from the local sources (locally checked out branch).

```bash
./installation/cmd/run.sh --kyma-installation full
```

>**NOTE:** The versions of the components that will be installed depends on their tag versions listed in the `values.yaml` file in the Helm charts.
If you want to build and deploy the local source code version of a component (e.g. Director), you can run the following command in the component directory:
  ```bash
  make deploy-on-minikube
  ```

> **Note:** To reduce memory and CPU usage, from the `installer-cr-kyma.yaml` file, comment out the components you don't want to use, such as `monitoring`, `tracing`, `logging`, or `kiali`.
