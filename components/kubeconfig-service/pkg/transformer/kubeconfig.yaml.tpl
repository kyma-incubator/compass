---
apiVersion: v1
kind: Config
current-context: {{ .ContextName }}
clusters:
- name: {{ .ContextName }}
  cluster:
    certificate-authority-data: {{ .CAData }}
    server: {{ .ServerURL }}
contexts:
- name: {{ .ContextName }}
  context:
    cluster: {{ .ContextName }}
    user: {{ .ContextName }}
users:
- name: {{ .ContextName }}
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - "--oidc-issuer-url={{ .OIDCURL }}"
      - "--oidc-client-id={{ .OIDCClientID }}"
      - "--oidc-client-secret={{ .OIDCSecret }}"
      command: kubectl