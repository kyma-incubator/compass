package transformer

import (
	"bytes"
	"html/template"

	"gopkg.in/yaml.v2"
)

//TransformerClient Wrapper for transformer operations
type TransformerClient struct {
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
}

type kubeConfigData struct {
	ContextName      string
	CAData           string
	ServerURL        string
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
}

//NewTransformerClient Create new instance of TransformerClient
func NewTransformerClient(oidcIssuerURL string, oidcClientID string, oidcClientSecret string) *TransformerClient {
	return &TransformerClient{
		OIDCClientID:     oidcClientID,
		OIDCClientSecret: oidcClientSecret,
		OIDCIssuerURL:    oidcIssuerURL,
	}
}

//TransformKubeconfig injects OIDC data into raw kubeconfig structure
func (tc *TransformerClient) TransformKubeconfig(rawKubeCfg string) ([]byte, error) {
	var kubeCfg Kubeconfig
	err := yaml.Unmarshal([]byte(rawKubeCfg), &kubeCfg)
	if err != nil {
		return nil, err
	}

	kcData := tc.extractData(kubeCfg)
	testYaml, err := tc.parseTemplate(*kcData)
	if err != nil {
		return nil, err
	}

	return []byte(testYaml), nil
}

func (tc *TransformerClient) extractData(kubeCfg Kubeconfig) *kubeConfigData {
	return &kubeConfigData{
		ContextName:      kubeCfg.Clusters[0].Name,
		CAData:           kubeCfg.Clusters[0].Cluster.CertificateAuthorityData,
		ServerURL:        kubeCfg.Clusters[0].Cluster.Server,
		OIDCIssuerURL:    tc.OIDCIssuerURL,
		OIDCClientID:     tc.OIDCClientID,
		OIDCClientSecret: tc.OIDCClientSecret,
	}
}

func (tc *TransformerClient) parseTemplate(kcData kubeConfigData) (string, error) {
	var result bytes.Buffer
	t := template.New("kubeconfigParser")
	t, err := t.Parse(KubeconfigTemplate)
	if err != nil {
		return "", err
	}

	err = t.Execute(&result, kcData)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
