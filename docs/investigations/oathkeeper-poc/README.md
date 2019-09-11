# Securing Compass with OathKeeper

## Setup
Modify `installation/resources/installer-cr-kyma-diet.yaml` and add new component which configures and installs ORY OathKeeper and ORY Hydra charts (already done on this branch):
```yaml
    - name: "ory"
      namespace: "kyma-system"
```

Modify VirtualService for Gateway component to point it to OathKeeper proxy (already done on this branch):
```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: {{ template "fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    app: {{ template "name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
spec:
  hosts:
    - '{{ .Values.global.gateway.host }}.{{ .Values.global.ingress.domainName }}'
  gateways:
    - {{ .Values.global.istio.gateway.name }}.{{ .Values.global.istio.gateway.namespace }}.svc.cluster.local
  http:
    - match:
        - uri:
            regex: /.*
      route:
        - destination:
            host: ory-oathkeeper-proxy.kyma-system.svc.cluster.local
            port:
              number: 4455
      corsPolicy:
        allowOrigin:
          - "*"
        allowHeaders:
          - "authorization"
          - "content-type"
          - "tenant"
        allowMethods:
          - "GET"
          - "POST"
          - "PUT"
          - "DELETE"
```

Install Compass with `installation/cmd/run.sh` script.

Create OathKeeper rule
```bash
kubectl apply -f ./oathkeeper-rule.yaml
```

This rule configures Oathkeeper's decision engine. There are 4 steps that occur in specific order while handling any incoming HTTP request.
Following fields define those steps:
- `match` field specifies rules for matching HTTP requests (HTTP method, path, host) that should be further processed (unmatched requests are blocked).
- `authenticators` field specifies method of validating user credentials passed with request, in our case we use oauth2 authenticator.
- `authorizer` field defines authorizer which decides if subject passed from authenticator is authorized to perform specific action. We don't perform any subject authorization at this step, so we just use `allow` authorizer.
- `mutators` field describes mutators which add session data to request before forwarding it to upstream. We use two mutators:
  - `hydrator` which communicates with tenant mapping service.
  - `token_id` which crafts signed JWT token with additional claim containing tenant retrieved by `hydrator`.

`upstream` field points to location of server where requests matching this rule will be forwarded to, we are pointing them at Compass Gateway component.
```yaml
apiVersion: oathkeeper.ory.sh/v1alpha1
kind: Rule
metadata:
  name: compass-gateway
  namespace: compass-system
spec:
  description: Test
  upstream:
    url: http://compass-gateway.compass-system.svc.cluster.local:3000
  match:
    methods: ["GET", "POST", "OPTIONS"]
    url: <http|https>://compass-gateway.kyma.local/<.*>
  authenticators:
    - handler: oauth2_introspection
      "config": {
        "required_scope": ["scope-a", "scope-b"]
      }
  authorizer:
    handler: allow
  mutators:
    - handler: hydrator
      config:
        api:
          url: http://compass-healthchecker.compass-system.svc.cluster.local:3000
    - handler: id_token
      config:
        claims: "{\"tenant\": \"{{ print .Extra.tenant }}\"}"
```

Patch Hydra VirtulServices to use Compass Istio Gateway because by default it would point at kyma gateway that is not created in compass-only Kyma configuration (in future we have to do it with overrides)
```bash
kubectl apply -f ./hydra-virtualservice-patch.yaml
```

Patch OathKeeper configmap to enable and configure `id_token` mutator: (in future we have to do it with overrides)
```bash
kubectl apply -f ./oathkeeper-configmap-patch.yaml
```

## Get Access Token

Create OAuth2Client CR. This client automatically generates client OAuth2 credentials (client id and client secret) and registers a client.
```bash
kubectl apply -f ./oauth-client.yaml
```

> **NOTE**: Instead of CR, you can do simple POST request. Read more on: https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials

Get client_id and client_secret from created secret:
```bash
kubectl get secrets -n default sample-client -oyaml
```

Decode client_id and client_secret from the secret:
```bash
export CLIENT_ID=$(echo '{client_id_value}' | base64 -D)
export CLIENT_SECRET=$(echo '{client_secret_value}' | base64 -D)
```

Get Access Token:
```bash
export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
curl -ik -X POST "https://oauth2.kyma.local/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=scope-a scope-b"
```

## Use Access Token

Use Access Token from response:
```bash
curl -ik https://compass-gateway.kyma.local/healthz -H "Authorization: Bearer ${access_token}"
```

If the token is valid and scopes are correct, Gateway will respond "ok".

## See logs

See logs of compass-gateway and compass-healthchecker.

Healthchecker is used as Tenant Mapping Service and it logs request data.
```bash
kubectl logs -n compass-system compass-healthchecker-7f4b9858fd-t7pkr healthchecker
```

Gateway logs request headers.
```bash
kubectl logs -n compass-system compass-gateway-688c856bd8-2x2nc gateway
```

In `Authorization` header you can see that there is valid JWT token with tenant info. Check it on [jwt.io](https://jwt.io/).

In payload you will see something like:
```json
{
  "exp": 1567639275,
  "iat": 1567639215,
  "iss": "https://my-oathkeeper/",
  "jti": "689832e9-e0cd-4af4-95d5-5855132baa3b",
  "nbf": 1567639215,
  "sub": "7077e51f-e2ec-4f69-aefb-650b4bc7bba3", // subject = client_id
  "tenant": "9ac609e1-7487-4aa6-b600-0904b272b11f" // our tenant
}
```
