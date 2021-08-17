package authenticator

func (a *Authenticator) SetJWKSEndpoint(url string) {
	a.jwksEndpoints = []string{url}
}
