package transformer

import (
	"bytes"
	"html/template"

	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/env"

	"gopkg.in/yaml.v2"
)

//TransformerClient Wrapper for transformer operations
type TransformerClient struct {
	ContextName      string
	CAData           string
	ServerURL        string
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
}

//NewTransformerClient Create new instance of TransformerClient
func NewTransformerClient(rawKubeCfg string) (*TransformerClient, error) {
	var kubeCfg Kubeconfig
	err := yaml.Unmarshal([]byte(rawKubeCfg), &kubeCfg)
	if err != nil {
		return nil, err
	}
	return &TransformerClient{
		ContextName:      kubeCfg.CurrentContext,
		CAData:           kubeCfg.Clusters[0].Cluster.CertificateAuthorityData,
		ServerURL:        kubeCfg.Clusters[0].Cluster.Server,
		OIDCClientID:     env.Config.OIDC.ClientID,
		OIDCClientSecret: env.Config.OIDC.ClientSecret,
		OIDCIssuerURL:    env.Config.OIDC.IssuerURL,
	}, nil
}

//TransformKubeconfig injects OIDC data into raw kubeconfig structure
func (tc *TransformerClient) TransformKubeconfig() ([]byte, error) {

	testYaml, err := tc.parseTemplate()
	if err != nil {
		return nil, err
	}

	return []byte(testYaml), nil
}

func (tc *TransformerClient) parseTemplate() (string, error) {
	var result bytes.Buffer
	t := template.New("kubeconfigParser")
	t, err := t.Parse(KubeconfigTemplate)
	if err != nil {
		return "", err
	}

	err = t.Execute(&result, tc)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
