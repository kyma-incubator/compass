package systemfetcher

import (
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"net/http"
)

type jwtTokenClient struct {
	certCache                 auth.CertificateCache
	jwtSelfSignCertSecretName string
	c                         *http.Client
}

// NewJwtTokenClient creates a jwt token client
func NewJwtTokenClient(certCache auth.CertificateCache, jwtSelfSignCertSecretName string, client *http.Client) *jwtTokenClient {
	return &jwtTokenClient{
		certCache:                 certCache,
		jwtSelfSignCertSecretName: jwtSelfSignCertSecretName,
		c:                         client,
	}
}

// Do executes a request for jwtTokenClient
func (jtc *jwtTokenClient) Do(req *http.Request, tenant string) (*http.Response, error) {
	req = req.WithContext(auth.SaveToContext(req.Context(), &auth.SelfSignedTokenCredentials{
		CertCache:                 jtc.certCache,
		JwtSelfSignCertSecretName: jtc.jwtSelfSignCertSecretName,
		Claims:                    map[string]interface{}{auth.CustomerIDClaimKey: tenant},
	}))

	return jtc.c.Do(req)
}
