# Compass installation

You can install Compass both on a cluster and on your local machine in the following modes:
- Single Kyma cluster
- Compass as a Central Management Plane

## Prerequisites for cluster installation

### Required versions

- check [here](https://github.com/kyma-incubator/compass#prerequisites) for required CLI tools versions
- Kubernetes 1.21

### Managed PostgreSQL Database

For more information about how you can use GCP managed PostgreSQL database instance with Compass, see: [Configure Managed GCP PostgreSQL](https://github.com/kyma-incubator/compass/blob/main/chart/compass/configure-managed-gcp-postgresql.md).

## Compass as a Central Management Plane

This is a multi-cluster installation mode, in which one cluster is needed with Compass. This mode allows you to integrate your Runtimes with Applications and manage them in one central place.

Compass as a Central Management Plane cluster requires minimal Kyma installation. The installation steps can vary depending on the installation environment.

### Cluster installation

> **NOTE:** During Kyma installation, Kyma version must be the same as in the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file on a specific commit.

#### Perform minimal Kyma installation

If custom domains and certificates are needed, see [Set up your custom domain TLS certificate](https://github.com/kyma-project/kyma/blob/10ae3a8acf7d57a329efa605890d11f9a9b40991/docs/03-tutorials/sec-01-tls-certificates-security.md#L1-L0) in the Kyma installation guide, and the resources in the [Certificate Management](#certificate-management) section of this document.

1. Save the following .yaml with installation overrides into a file (e.g: additionalKymaOverrides.yaml)
```yaml
istio-configuration:
   components:
      ingressGateways:
         config:
            service:
               loadBalancerIP: ${GATEWAY_IP_ADDRESS}
               type: LoadBalancer
   meshConfig:
      defaultConfig:
         holdApplicationUntilProxyStarts: true
global:
   loadBalancerIP: ${GATEWAY_IP_ADDRESS}
# uncomment below values if you want to proceed with your custom values; default domain is 'local.kyma.dev' and there is default self-signed cert and key for that domain
   #domainName: ${DOMAIN} 
   #tlsCrt: ${TLS_CERT} 
   #tlsKey: ${TLS_KEY} 
   #ingress:
      #domainName: ${DOMAIN}
      #tlsCrt: ${TLS_CERT}
      #tlsKey: ${TLS_KEY}
```
And then execute the kyma installation with the following command:

```bash
kyma deploy --source <version from ../../installation/resources/KYMA_VERSION> -c <minimal file from ../../installation/resources/kyma/kyma-components-minimal.yaml> -f <overrides file from ../../installation/resources/kyma/kyma-overrides-minimal.yaml> -f <file from above step - e.g. additionalKymaOverrides.yaml> --ci
```

#### Install Compass

> **NOTE:** If you installed Kyma on a cluster with a custom domain and certificates, you must apply that overrides to Compass as well.

The proper work of JWT token flows and Compass Cockpit require a set up and configured OpenID Connect (OIDC) Authorization Server.
The OIDC Authorization Server is needed for the support of the respective users, user groups, and scopes. The OIDC server host and client-id are specified as overrides of the Compass Helm chart. Then a set of admin scopes are granted to a user based on the groups in the id_token, those trusted groups can be configured with overrides as well.

1. Save the following .yaml with installation overrides into a file (e.g: additionalCompassOverrides.yaml)
```yaml
hydrator:
   adminGroupNames: ${ADMIN_GROUP_NAMES}
global:
   isLocalEnv: false
   migratorJob:
      pvc:
         isLocalEnv: false
   enableInternalCommunicationPolicies: false
   loadBalancerIP: 34.140.141.115
   cockpit:
      auth:
         idpHost: ${IDP_HOST}
         clientID: ${CLIENT_ID}
# uncomment below values if you want to proceed with your custom values; default domain is 'local.kyma.dev' and there is default self-signed cert and key for that domain
#   domainName: ${DOMAIN}
#   tlsCrt: ${TLS_CERT}
#   tlsKey: ${TLS_KEY}
#   ingress:
#      domainName: ${DOMAIN}
#      tlsCrt: ${TLS_CERT}
#      tlsKey: ${TLS_KEY}
```

```bash
. <script from ../../installation/scripts/install-compass.sh> --overrides-file <file from ../../installation/resources/compass-overrides-local.yaml> --overrides-file <file from above step - e.g. additionalCompassOverrides.yaml> --timeout <e.g: 30m0s>
```

### Certificate Management

In case certificate rotation is needed, you can install JetStack's [Certificate Manager](https://github.com/jetstack/cert-manager) to take care of the certificates.

The following certificates can be rotated:
* Connector intermediate certificate that is used to issue Application and Runtime client certificates.
* Istio gateway certificate for the regular HTTPS gateway.
* Istio gateway certificate for the mTLS gateway.

#### Create issuers

To issue certificates, the Certificate Manager requires a resource called issuer.

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

For more information about the different cluster issuer configurations, see: [ACME Issuer](https://cert-manager.io/docs/configuration/acme/)

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

Certificate Manager's role is to rotate certificates as defined in the resources. Make sure that the certificate as shown in the example is specified as `isCA: true`. This means that the certificate is used to issue other certificates.

#### Domain certificates

Certificate Manager can also rotate Istio gateway certificates. The following example shows such certificate:

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

### Local k3d installation

For local development, install Compass with the minimal Kyma installation on k3d from the `main` branch. To do so, run the following script:

```bash
./installation/cmd/run.sh --oidc-host {URL_TO_OIDC_SERVER} --oidc-client-id {OIDC_CLIENT_ID} --oidc-admin-group {OIDC_ADMIN_GROUP}
```

The Kyma version is read from the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file. You can override it with the following command:

```bash
./installation/cmd/run.sh --kyma-release {KYMA_VERSION} --oidc-host {URL_TO_OIDC_SERVER} --oidc-client-id {OIDC_CLIENT_ID} --oidc-admin-group {OIDC_ADMIN_GROUP}
```
You can also specify if you want the Kyma installation to contain only `minimal` components or whether you want `full` Kyma

```bash
./installation/cmd/run.sh --kyma-installation full --oidc-host {URL_TO_OIDC_SERVER} --oidc-client-id {OIDC_CLIENT_ID} --oidc-admin-group {OIDC_ADMIN_GROUP}
```

Optionally, you can use the `--dump-db` flag to populate the DB with sample data. As a result, a DB dump is downloaded from the Compass development environment and is imported into the DB during the installation of Compass. Note that you can only use this feature if you are part of the Compass contributors. Otherwise, you will not have access to the development environment, from which the data is obtained.
Note that using this flag also results in building a new `schema-migrator` image from the local files.
```bash
./installation/cmd/run.sh --dump-db --oidc-host {URL_TO_OIDC_SERVER} --oidc-client-id {OIDC_CLIENT_ID} --oidc-admin-group {OIDC_ADMIN_GROUP}
```


>**NOTE:** The versions of the components that are installed depend on their tag versions listed in the `values.yaml` file in the Helm charts.
If you want to build and deploy the local source code version of a component (for example, Director), you can run the following command in the component directory:
  ```bash
  make deploy-on-k3d
  ```

> **_NOTE:_**  
>To configure an OIDC identity provider that is required for the JWT flows, the OIDC configuration arguments (`--oidc-host`, `--oidc-client-id`, `--oidc-admin-group`) are mandatory. If they are omitted, the **run.sh** script tries to get the required values from the **~/.compass.yaml** file. To run the `run.sh` script, you need the [yq](https://mikefarah.gitbook.io/yq/) tool.
>
>The  **~/compass.yaml** file must have the following structure:
>  > idpHost: {URL_TO_OIDC_SERVER}
>  >
>  > clientID: {OIDC_CLIENT_ID}
>  >
>  > adminGroupNames: {OIDC_ADMIN_GROUPS}
>
> Note that the JWT flows work properly only when the configuration arguments are passed and the **~.compass.yaml** file exists.

## Single cluster with Compass and Runtime Agent

This is a single-tenant mode, which provides the complete cluster Kyma installation with all components, including the Runtime Agent. You can install Compass on top of it.
In this mode, the Runtime Agent is already connected to Compass. This mode facilitates various kind of testing and development.

> **NOTE:** The version of the Kyma installed on the cluster must match the Kyma version in the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file on a specific commit.

### Cluster installation

To install Compass and Runtime components on a single cluster, follow these steps:

TODO: As mentioned previously, there is no longer `cluster-installation` doc for Kyma 2.0.4. Consider whether to remove the one in the step below if no such doc is found.

1. Bear in mind the [Installation](https://github.com/kyma-project/kyma/blob/2.0.4/docs/04-operation-guides/operations/ra-01-enable-kyma-with-runtime-agent.md) configuration that enables the Runtime Agent, and then, install Kyma. 

   1. Save the following .yaml into some file (e.g: additionalKymaOverrides.yaml)
    ```yaml
    istio-configuration:
       components:
          ingressGateways:
             config:
                service:
                   loadBalancerIP: ${GATEWAY_IP_ADDRESS}
                   type: LoadBalancer
       meshConfig:
          defaultConfig:
             holdApplicationUntilProxyStarts: true
    global:
       loadBalancerIP: ${GATEWAY_IP_ADDRESS}
       disableLegacyConnectivity: true   
    # uncomment below values if you want to proceed with your custom values; default domain is 'local.kyma.dev' and there is default self-signed cert and key for that domain
       #domainName: ${DOMAIN} 
       #tlsCrt: ${TLS_CERT} 
       #tlsKey: ${TLS_KEY} 
       #ingress:
       #domainName: ${DOMAIN} 
       #tlsCrt: ${TLS_CERT} 
       #tlsKey: ${TLS_KEY}
    ```
   2. Add the `compass-runtime-agent` module in the `compass-system` Namespace to the [kyma-components-file](../../installation/resources/kyma/kyma-components-full.yaml) and save the following .yaml into some file (e.g: kyma-components-with-runtime-agent.yaml)
    ```bash
    kyma deploy --source <version from ../../installation/resources/KYMA_VERSION> -c <file from above step - e.g: kyma-components-with-runtime-agent.yaml> -f <overrides file from ../../installation/resources/kyma/kyma-overrides-full.yaml> -f <file from above step - e.g. additionalKymaOverrides.yaml> --ci
    ```

5. Install Compass using the following command:
   1. Save the following .yaml into some file (e.g: additionalCompassOverrides.yaml)
    ```yaml
    global:
       isLocalEnv: false
       migratorJob:
          pvc:
             isLocalEnv: false
       enableInternalCommunicationPolicies: false
       loadBalancerIP: ${GATEWAY_IP_ADDRESS}
    # uncomment below values if you want to proceed with your custom values; default domain is 'local.kyma.dev' and there is default self-signed cert and key for that domain
       #domainName: ${DOMAIN} 
       #tlsCrt: ${TLS_CERT} 
       #tlsKey: ${TLS_KEY} 
       #ingress:
       #domainName: ${DOMAIN} 
       #tlsCrt: ${TLS_CERT} 
       #tlsKey: ${TLS_KEY}
    ```

    ```bash
    . <script from ../../installation/scripts/install-compass.sh> --override-files <file from ../../installation/resources/compass-overrides-local.yaml> --overrides-file <file from above step - e.g. additionalCompassOverrides.yaml> --timeout <e.g: 30m0s>
    ```
TODO: in the above Compass installation step, do we want to pass `local` overrides or `benchmark` ones as `benchmark` are used in productive cluster (not local one). In addition, double-check if we want to set `global.isLocalEnv: false` in the Compass overrides (`migratorJob.pvc.isLocalEnv` must be set to false)
   
Once Compass is installed, Runtime Agent will be configured to fetch the Runtime configuration from the Compass installation within the same cluster.

### Local k3d installation

To install Compass and Runtime components on k3d, run the command below. Kyma source code will be picked up according to the KYMA_VERSION file and Compass source code will be picked up from the local sources (locally checked out branch).

```bash
./installation/cmd/run.sh --kyma-installation full --oidc-host {URL_TO_OIDC_SERVER} --oidc-client-id {OIDC_CLIENT_ID}
```

>**NOTE:** The versions of the components that are installed depend on their tag versions listed in the `values.yaml` file in the Helm charts.
If you want to build and deploy the local source code version of a component (for example, Director), you can run the following command in the component directory:
  ```bash
  make deploy-on-k3d
  ```

> **Note:** To reduce memory and CPU usage, from the `installer-cr-kyma.yaml` file, comment out the components you don't want to use, such as `monitoring`, `tracing`, `logging`, or `kiali`.
