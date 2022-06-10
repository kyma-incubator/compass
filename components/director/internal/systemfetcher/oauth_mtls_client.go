package systemfetcher

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
)

type oauthMtlsClient struct {
	clientID     string
	tokenURL     string
	scopesClaim  string
	tenantHeader string
	certCache    auth.CertificateCache

	c *http.Client
}

// NewOauthMtlsClient missing docs
func NewOauthMtlsClient(oauthCfg oauth.Config, certCache auth.CertificateCache, client *http.Client) *oauthMtlsClient {
	protocol := oauthCfg.TokenEndpointProtocol + "://"
	tokenParsedURL, err := url.Parse(oauthCfg.TokenBaseURL)
	if err == nil && len(tokenParsedURL.Scheme) != 0 {
		protocol = ""
	}
	return &oauthMtlsClient{
		clientID:     oauthCfg.ClientID,
		certCache:    certCache,
		tokenURL:     protocol + oauthCfg.TokenBaseURL + oauthCfg.TokenPath,
		scopesClaim:  strings.Join(oauthCfg.ScopesClaim, " "),
		tenantHeader: oauthCfg.TenantHeaderName,
		c:            client,
	}
}

// Do missing docs
func (omc *oauthMtlsClient) Do(req *http.Request, tenant string) (*http.Response, error) {
	req = req.WithContext(auth.SaveToContext(req.Context(), &auth.OAuthMtlsCredentials{
		ClientID:          omc.clientID,
		CertCache:         omc.certCache,
		TokenURL:          omc.tokenURL,
		Scopes:            omc.scopesClaim,
		AdditionalHeaders: map[string]string{omc.tenantHeader: tenant},
	}))

	return omc.c.Do(req)
}
