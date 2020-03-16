# Authorization

The Kyna Environment Broker provides the following ways to authorize the broker's user.

## Basic auth

The basic auth implementation is currently being replaced by the Oauth2 implementation and will be deprecated soon.
To access the Kyma Environment Broker endpoints with basic authorization enabled, you must specify an Authorization Basic token header:

```
Authorization: Basic KgEY6JRtWtjZYHQCK5Hb4wlmDqwto2zj7fCCEwn3Vo.eUysIrnvvUW6Vz8cXiMeBRH0EGG4BK_r5rA1X2C2GRE
```

## Oauth2

The Kyma Environment Broker allows to authorize users using the Oauth authentication. It is using the [ApiRule](https://github.com/kyma-project/kyma/blob/master/docs/api-gateway-v2/06-01-apirule.md) to provide a VirtualService and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR.
To authorize with the Kyma Environment Broker, use an OAuth2 client registered through the Hydra Maester controller. Look for more information [here](https://github.com/kyma-project/kyma/blob/master/docs/api-gateway-v2/08-01-exposesecure.md#register-an-oauth2-client-and-get-tokens).
The example OAuth2 Client definition for KEB looks as follows:
```yaml
apiVersion: hydra.ory.sh/v1alpha1
kind: OAuth2Client
metadata:
  name: ${CLIENT_NAME}
spec:
  grantTypes:
    - "client_credentials"
  scope: "broker:write"
  secretName: ${SECRET_NAME}
```
To access the Kyma Environment Broker endpoints with OAuth2 authorization enabled, use the `/oauth` prefix. For example:
```
/oauth/v2/catalog
```

You must also specify an Authorization Bearer token header:
```
Authorization: Bearer {ACCESS_TOKEN}
```

Follow the tutorial to obtain a new access token for Kyma Environment Broker.

1. Export the credentials of the created client as environment variables. Run:

  ```shell
  export CLIENT_ID="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_id}' | base64 --decode)"
  export CLIENT_SECRET="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_secret}' | base64 --decode)"
  ```

2. Encode your client credentials and export them as an environment variable:

  ```shell
  export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
  ```

3. Get the access token:
```shell
curl -ik -X POST "https://oauth2.$DOMAIN/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=broker:write"
```