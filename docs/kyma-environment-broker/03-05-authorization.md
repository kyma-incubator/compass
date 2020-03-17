# Authorization

Kyma Environment Broker provides two ways to authorize users:
- Basic authorization
- OAuth2 authorization

## Basic authorization

To access the Kyma Environment Broker endpoints with the Basic authorization enabled, specify the `Authorization: Basic` token header:

```
Authorization: Basic {BASE64_ENCODED_CREDENTIALS}
```

>**NOTE**: This implementation is currently being replaced by the Oauth2 implementation and will be deprecated soon.

## OAuth2 authorization

The Kyma Environment Broker allows to authorize users using the OAuth2 authorization. It is using the [ApiRule](https://github.com/kyma-project/kyma/blob/master/docs/api-gateway-v2/06-01-apirule.md) to provide a [VirtualService](https://istio.io/docs/reference/config/networking/virtual-service/) and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR.
To authorize with the Kyma Environment Broker, use an OAuth2 client registered through the [Hydra Maester controller](https://github.com/ory/k8s/blob/master/docs/helm/hydra-maester.md).

To access the Kyma Environment Broker endpoints with the OAuth2 authorization enabled, use the `/oauth` prefix. For example:

```
/oauth/v2/catalog
```

You must also specify the `Authorization: Bearer` token header:

```
Authorization: Bearer {ACCESS_TOKEN}
```

### Access token

Follow these steps to obtain a new access token:

1. Export these values as environment variables:

  - The name of your client and the Secret which stores the client credentials:

    ```shell
    export CLIENT_NAME={YOUR_CLIENT_NAME}
    ```

  - The Namespace in which you want to create the client and the Secret that stores its credentials:

    ```shell
    export CLIENT_NAMESPACE={YOUR_CLIENT_NAMESPACE}
    ```

  - The domain of your cluster:

    ```shell
    export DOMAIN={CLUSTER_DOMAIN}
    ```

2. Create an OAuth2 client:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: hydra.ory.sh/v1alpha1
kind: OAuth2Client
metadata:
  name: $CLIENT_NAME
  namespace: $CLIENT_NAMESPACE
spec:
  grantTypes:
    - "client_credentials"
  scope: "broker:write"
  secretName: $CLIENT_NAME
EOF
```

3. Export the credentials of the created client as environment variables. Run:

```shell
export CLIENT_ID="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_id}' | base64 --decode)"
export CLIENT_SECRET="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_secret}' | base64 --decode)"
```

4. Encode your client credentials and export them as an environment variable:

```shell
export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
```

5. Get the access token:
```shell
curl -ik -X POST "https://oauth2.$DOMAIN/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=broker:write"
```
