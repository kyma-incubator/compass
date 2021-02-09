package authenticator

func (a *Authenticator) SetJWKSEndpoint(url string) {
	a.jwksEndpoint = url
}
