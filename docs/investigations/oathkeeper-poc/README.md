# Securing Compass with OathKeeper

## Setup
Modify `installation/resources/installer-cr-kyma-diet.yaml` and add new component:
```yaml
    - name: "ory"
      namespace: "kyma-system"
```

Modify VirtualService for Gateway component to point it to OathKeeper:
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

Create OathKeeper rule:
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
    - handler: header
      config:
        headers:
          Foo: "bar"
```

Install Compass with `installation/cmd/run.sh` script.

Patch Hydra VirtulServices to use Compass Istio Gateway (in future we have to do it with overrides)
```bash
k apply -f templates/hydra-virtualservice-patch.yaml
```

## Get Access Token

Create OAuthClient CR
```bash
kubectl apply -f templates/oauth-client.yaml
```

> **NOTE**: Instead of CR, you can do simple POST request. Read more on: https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials

Get client_id and client_secret from created secret
```bash
kubectl get secrets -n default sample-client -oyaml
```

Decode client_id and client_secret from the secret:
```bash
export CLIENT_ID=$(echo '{client_id_value}' | base64 -D)
export CLIENT_SECRET=$(echo '{client_secret_value}' | base64 -D)
```

Get Access Token
```bash
export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
export DOMAIN=kyma.local
curl -ik -X POST "https://oauth2.kyma.local/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=scope-a scope-b"
```


## Use Access Token

Use Access Token from response

```bash
curl -ik https://compass-gateway.kyma.local/healthz -H "Authorization: Bearer ${access_token}"
```

If the token is valid and scopes are correct, Gateway will respond "ok".
