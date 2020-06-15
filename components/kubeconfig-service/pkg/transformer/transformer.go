package transformer

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type TransformerClient struct {
	OIDCIssuerURL string
	OIDCClientID string
	OIDCClientSecret string
}

const oidcConfiguration = `
apiVersion: client.authentication.k8s.io/v1beta1
args:
- oidc-login
- get-token
- "--oidc-issuer-url=%s"
- "--oidc-client-id=%s"
- "--oidc-client-secret=%s"
command: kubectl
`

func NewTransformerClient(oidcIssuerURL string, oidcClientID string, oidcClientSecret string) *TransformerClient {
	return &TransformerClient{
		OIDCClientID: oidcClientID,
		OIDCClientSecret: oidcClientSecret,
		OIDCIssuerURL: oidcIssuerURL,
	}
}

//TransformKubeconfig injects OIDC data into raw kubeconfig structure
func (tc *TransformerClient) TransformKubeconfig(rawKubeCfg string) ([]byte, error) {
	var kubeCfg Kubeconfig
	err := yaml.Unmarshal([]byte(rawKubeCfg), &kubeCfg)
	if err != nil {
		return nil, err
	}

	kubeCfg.Users[0].User = map[string]interface{}{
		"exec": fmt.Sprintf(oidcConfiguration, tc.OIDCIssuerURL, tc.OIDCClientID, tc.OIDCClientSecret),
	}

	kubeCfgYaml, err := yaml.Marshal(kubeCfg)
	if err != nil {
		return nil, err
	}

	return kubeCfgYaml, nil
}
