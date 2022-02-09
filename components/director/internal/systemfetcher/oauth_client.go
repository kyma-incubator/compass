package systemfetcher

import (
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
)

type oauthClient struct {
	clientID     string
	tokenURL     string
	scopesClaim  string
	tenantHeader string

	c *http.Client
}

// NewOauthClient missing docs
func NewOauthClient(oauthCfg oauth.Config, client *http.Client) *oauthClient {
	return &oauthClient{
		clientID:     oauthCfg.ClientID,
		tokenURL:     oauthCfg.TokenEndpointProtocol + "://" + oauthCfg.TokenBaseHost + oauthCfg.TokenPath,
		scopesClaim:  strings.Join(oauthCfg.ScopesClaim, " "),
		tenantHeader: oauthCfg.TenantHeaderName,
		c:            client,
	}
}

// Do missing docs
func (oc *oauthClient) Do(req *http.Request, tenant string) (*http.Response, error) {
	req = req.WithContext(auth.SaveToContext(req.Context(), &auth.OAuthCredentials{
		ClientID:          oc.clientID,
		TokenURL:          oc.tokenURL,
		Scopes:            oc.scopesClaim,
		AdditionalHeaders: map[string]string{oc.tenantHeader: tenant},
	}))

	return oc.c.Do(req)
}
