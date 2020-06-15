package transformer

type Kubeconfig struct {
	APIVersion     string `yaml:"apiVersion"`
	Kind           string `yaml:"kind"`
	CurrentContext string `yaml:"current-context"`
	Clusters       []struct {
		Name    string `yaml:"name"`
		Cluster struct {
			CertificateAuthorityData string `yaml:"certificate-authority-data"`
			Server                   string `yaml:"server"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`
	Contexts []struct {
		Name    string `yaml:"name"`
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
	} `yaml:"contexts"`
	Users []struct {
		Name string                 `yaml:"name"`
		User map[string]interface{} `yaml:"user"`
	} `yaml:"users"`
}

const KubeconfigTemplate = `---
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
      - "--oidc-issuer-url={{ .OIDCIssuerURL }}"
      - "--oidc-client-id={{ .OIDCClientID }}"
      - "--oidc-client-secret={{ .OIDCClientSecret }}"
      command: kubectl`
