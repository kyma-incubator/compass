package systemfetcher

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
)

type jwtTokenClient struct {
	keyCache                  credloader.KeysCache
	jwtSelfSignCertSecretName string
	c                         *http.Client
}

// NewJwtTokenClient creates a jwt token client
func NewJwtTokenClient(keyCache credloader.KeysCache, jwtSelfSignCertSecretName string, client *http.Client) *jwtTokenClient {
	return &jwtTokenClient{
		keyCache:                  keyCache,
		jwtSelfSignCertSecretName: jwtSelfSignCertSecretName,
		c:                         client,
	}
}

// Do executes a request for jwtTokenClient
func (jtc *jwtTokenClient) Do(req *http.Request, tenant string) (*http.Response, error) {
	req = req.WithContext(auth.SaveToContext(req.Context(), &auth.SelfSignedTokenCredentials{
		KeysCache:                 jtc.keyCache,
		JwtSelfSignCertSecretName: jtc.jwtSelfSignCertSecretName,
		Claims:                    map[string]interface{}{auth.CustomerIDClaimKey: tenant},
	}))

	return jtc.c.Do(req)
}
