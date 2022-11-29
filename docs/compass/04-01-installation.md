# Compass installation

You can install Compass both on a cluster and on your local machine in the following modes:
- Single Kyma cluster
- Compass as a Central Management Plane

## Prerequisites for cluster installation

### Required versions

- Kubernetes 1.21
- For more information about the required CLI tools versions, see: [Compass Prerequisites](https://github.com/kyma-incubator/compass#prerequisites) 

### Managed PostgreSQL Database

For more information about how you can use GCP managed PostgreSQL database instance with Compass, see: [Configure Managed GCP PostgreSQL](https://github.com/kyma-incubator/compass/blob/main/chart/compass/configure-managed-gcp-postgresql.md).

## Compass as a Central Management Plane

This is a multi-cluster installation mode, in which one cluster needs to be dedicated to Compass. This mode allows you to integrate your Runtimes with Applications and manage them in one central place.

Compass as a Central Management Plane cluster requires minimal Kyma installation. The installation steps can vary depending on the installation environment.

### Cluster installation

**Security Prerequisites**

- The proper work of JWT token flows and Compass Cockpit require a set up and configured OpenID Connect (OIDC) Authorization Server.
  The OIDC Authorization Server is needed for the support of the respective users, user groups, and scopes. The OIDC server host and client-id are specified as overrides of the Compass Helm chart. Then, a set of administrator scopes are granted to a user, based on the groups in the `id_token`. Those trusted groups can be configured with overrides as well.

> **NOTE:** Compass relies on the `name` claim in the id_token. Therefore, you must configure your IDP to contain that attribute in the resulting token as this is the claim that is used for user identification.

- For internal communication between components, Compass relies on Kubernetes Service Account tokens and [Service Account Issuer Discovery](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#service-account-issuer-discovery) for the validation of the tokens.
Therefore, `serviceAccountTokenJWKS` and `serviceAccountTokenIssuer` need to be configured as overrides. This configuration could be infrastructure specific.

> For more information about GCP, see [getOpenid-configuration](https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.locations.clusters.well-known/getOpenid-configuration). An example configuration of GKE clusters looks like this:
```yaml
   kubernetes:
       serviceAccountTokenIssuer: "https://container.googleapis.com/v1/projects/${PROJECT_NAME}/locations/${REGION}/clusters/${CLUSTER_NAME}"
       serviceAccountTokenJWKS: "https://container.googleapis.com/v1/projects/${PROJECT_NAME}/locations/${REGION}/clusters/${CLUSTER_NAME}/jwks"
```

> **NOTE:** The `serviceAccountTokenIssuer` must match exactly to the value in the `iss` claim of the service account token mounted to a pod (in `/var/run/secrets/kubernetes.io/serviceaccount/token`). This value can differ from the `iss` claim of the token in the service account secret.

#### Perform minimal Kyma installation

> **NOTE:** During the installation of Compass, the installed Kyma version (as a basis to Compass) must match to the one in the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file in the specific Compass commit.

If custom domains and certificates are needed, see the [Set up your custom domain TLS certificate](https://github.com/kyma-project/kyma/blob/2.2.0/docs/03-tutorials/sec-01-tls-certificates-security.md) document in the Kyma installation guide, as well as the resources in the [Certificate Management](#certificate-management) section in this document.

Save the following .yaml code with installation overrides into a file (for example: additionalKymaOverrides.yaml)
```yaml
ory:
  global:
    domainName: ${DOMAIN} # Optional, only needed if you use custom domains below.
  oathkeeper:
    oathkeeper:
      config:
        authenticators:
          jwt:
            config:
              jwks_urls:
                -  ${IDP_JWKS_URL}
istio:
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
# If you want to proceed with your custom values, uncomment the values below; the default domain is `local.kyma.dev` and there is a default self-signed certificate and a key for that domain
   #domainName: ${DOMAIN} 
   #tlsCrt: ${TLS_CERT} 
   #tlsKey: ${TLS_KEY} 
   #ingress:
      #domainName: ${DOMAIN}
      #tlsCrt: ${TLS_CERT}
      #tlsKey: ${TLS_KEY}
```
And then start the Kyma installation by using the following command:

```bash
kyma deploy --source <version from ../../installation/resources/KYMA_VERSION> -c <minimal file from ../../installation/resources/kyma/kyma-components-minimal.yaml> -f <overrides file from ../../installation/resources/kyma/kyma-overrides-minimal.yaml> -f <file from above step - e.g. additionalKymaOverrides.yaml> --ci
```

#### Install Compass

> **NOTE:** If you installed Kyma on a cluster with a custom domain and certificates, you must apply the overrides to Compass as well.

Save the following .yaml code with installation overrides into a file (for example: additionalCompassOverrides.yaml)
```yaml
hydrator:
   adminGroupNames: ${ADMIN_GROUP_NAMES}
global:
   isLocalEnv: false
   migratorJob:
      pvc:
         storageClass: ${ANY_SUPPORTED_STORAGE_CLASS}
   kubernetes:
      serviceAccountTokenIssuer: ${TOKEN_ISSUER} # Default is https://kubernetes.default.svc.cluster.local
      serviceAccountTokenJWKS: ${JWKS_ENDPOINT} # Default is https://kubernetes.default.svc.cluster.local/openid/v1/jwks
   loadBalancerIP: ${LOAD_BALANCER_SERVICE_EXTERNAL_IP}
   cockpit:
      auth:
         idpHost: ${IDP_HOST}
         clientID: ${CLIENT_ID}
# If you want to proceed with your custom values, uncomment the values below; the default domain is `local.kyma.dev` and there is a default self-signed certificate and a key for that domain
#   domainName: ${DOMAIN}
#   tlsCrt: ${TLS_CERT}
#   tlsKey: ${TLS_KEY}
#   ingress:
#      domainName: ${DOMAIN}
#      tlsCrt: ${TLS_CERT}
#      tlsKey: ${TLS_KEY}
```

Start the Database installation by using the following command:

```bash
<script from ../../installation/scripts/install-db.sh> --overrides-file <file from ../../installation/resources/compass-overrides-local.yaml> --overrides-file <file from above step - e.g. additionalCompassOverrides.yaml> --timeout <e.g: 30m0s>
```
Once the Database is provisioned procced and start the Compass installation by using the following command:

```bash
<script from ../../installation/scripts/install-compass.sh> --overrides-file <file from ../../installation/resources/compass-overrides-local.yaml> --overrides-file <file from above step - e.g. additionalCompassOverrides.yaml> --timeout <e.g: 30m0s>
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

### Cluster installation

**Security Prerequisites**

- The proper work of JWT token flows and Compass Cockpit require a set up and configured OpenID Connect (OIDC) Authorization Server.
  The OIDC Authorization Server is needed for the support of the respective users, user groups, and scopes. The OIDC server host and client-id are specified as overrides of the Compass Helm chart. Then, a set of administrator scopes are granted to a user, based on the groups in the `id_token`. Those trusted groups can be configured with overrides as well.

> **NOTE:** Compass relies on the `name` claim in the `id_token`. Therefore, you must configure your IDP to contain that attribute in the resulting token as this is the claim that is used for user identification.

- For internal communication between components, Compass relies on Kubernetes Service Account tokens and [Service Account Issuer Discovery](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#service-account-issuer-discovery) for the validation of the tokens.
Therefore, `serviceAccountTokenJWKS` and `serviceAccountTokenIssuer` need to be configured as overrides. This configuration could be infrastructure specific.

> For more information about GCP, see [getOpenid-configuration](https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.locations.clusters.well-known/getOpenid-configuration). An example configuration of GKE clusters looks like this:
```yaml
   kubernetes:
       serviceAccountTokenIssuer: "https://container.googleapis.com/v1/projects/${PROJECT_NAME}/locations/${REGION}/clusters/${CLUSTER_NAME}"
       serviceAccountTokenJWKS: "https://container.googleapis.com/v1/projects/${PROJECT_NAME}/locations/${REGION}/clusters/${CLUSTER_NAME}/jwks"
```

> **NOTE:** The `serviceAccountTokenIssuer` must match exactly to the value in the `iss` claim of the service account token mounted to a pod (in `/var/run/secrets/kubernetes.io/serviceaccount/token`). This value can differ from the `iss` claim of the token in the service account secret.

To install the Compass and Runtime components on a single cluster, perform the following steps:

#### Kyma Prerequisite

> **NOTE:** During the installation of Kyma, the installed version must match to the one in the [`KYMA_VERSION`](../../installation/resources/KYMA_VERSION) file in the specific Compass commit.

You must have a Kyma installation with an enabled Runtime Agent. For more information, see [Enable Kyma with Runtime Agent](https://github.com/kyma-project/kyma/blob/2.2.0/docs/04-operation-guides/operations/ra-01-enable-kyma-with-runtime-agent.md). Therefore, you must add the compass-runtime-agent module in the compass-system namespace to the list of [minimal compass components file](../../installation/resources/kyma/kyma-overrides-minimal.yaml).

If custom domains and certificates are needed, see the [Set up your custom domain TLS certificate](https://github.com/kyma-project/kyma/blob/2.2.0/docs/03-tutorials/sec-01-tls-certificates-security.md) document in the Kyma installation guide, as well as the resources in the [Certificate Management](#certificate-management) section in this document.

Save the following .yaml code with installation overrides to a file (for example: additionalKymaOverrides.yaml)
```yaml
ory:
  global:
    domainName: ${DOMAIN} # Optional, only needed if you use custom domains below.
  oathkeeper:
    oathkeeper:
      config:
        authenticators:
          jwt:
            config:
              jwks_urls:
                -  ${IDP_JWKS_URL}
istio:
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
# If you want to proceed with your custom values, uncomment the values below; the default domain is `local.kyma.dev` and there is a default self-signed certificate and a key for that domain
   #domainName: ${DOMAIN} 
   #tlsCrt: ${TLS_CERT} 
   #tlsKey: ${TLS_KEY} 
   #ingress:
      #domainName: ${DOMAIN}
      #tlsCrt: ${TLS_CERT}
      #tlsKey: ${TLS_KEY}
```

And then, start the Kyma installation by using the following command:

```bash
kyma deploy --source <version from ../../installation/resources/KYMA_VERSION> -c <minimal file from ../../installation/resources/kyma/kyma-components-minimal.yaml> -f <overrides file from ../../installation/resources/kyma/kyma-overrides-minimal.yaml> -f <file from above step - e.g. additionalKymaOverrides.yaml> --ci
```

#### Install Compass

Save the following .yaml code with installation overrides into a file (for example: additionalCompassOverrides.yaml)
```yaml
hydrator:
   adminGroupNames: ${ADMIN_GROUP_NAMES}
global:
   agentPreconfiguration: true
   isLocalEnv: false
   migratorJob:
      pvc:
         storageClass: ${ANY_SUPPORTED_STORAGE_CLASS}
   kubernetes:
     serviceAccountTokenIssuer: ${TOKEN_ISSUER} # Default is https://kubernetes.default.svc.cluster.local
     serviceAccountTokenJWKS: ${JWKS_ENDPOINT} # Default is https://kubernetes.default.svc.cluster.local/openid/v1/jwks
   loadBalancerIP: ${LOAD_BALANCER_SERVICE_EXTERNAL_IP}
   cockpit:
      auth:
         idpHost: ${IDP_HOST}
         clientID: ${CLIENT_ID}
# If you want to proceed with your custom values, uncomment the values below; the default domain is `local.kyma.dev` and there is a default self-signed certificate and a key for that domain
#   domainName: ${DOMAIN}
#   tlsCrt: ${TLS_CERT}
#   tlsKey: ${TLS_KEY}
#   ingress:
#      domainName: ${DOMAIN}
#      tlsCrt: ${TLS_CERT}
#      tlsKey: ${TLS_KEY}
```
Start Database installation:
```bash
<script from ../../installation/scripts/install-db.sh> --overrides-file <file from ../../installation/resources/compass-overrides-local.yaml> --overrides-file <file from above step - e.g. additionalCompassOverrides.yaml> --timeout <e.g: 30m0s>
```
Then, install compass component:
```bash
<script from ../../installation/scripts/install-compass.sh> --overrides-file <file from ../../installation/resources/compass-overrides-local.yaml> --overrides-file <file from above step - e.g. additionalCompassOverrides.yaml> --timeout <e.g: 30m0s>
```

Once Compass is installed, the Runtime Agent will be configured to fetch the Runtime configuration from the Compass installation within the same cluster.
