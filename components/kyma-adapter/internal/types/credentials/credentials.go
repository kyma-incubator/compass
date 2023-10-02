package credentials

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types/tenantmapping"
)

// Credentials represents the bundle instance auth credentials
type Credentials interface {
	ToString() string
}

// NewCredentials creates new Credentials
func NewCredentials(configuration tenantmapping.Configuration) Credentials {
	basicCreds := configuration.GetBasicCredentials()
	oauthCreds := configuration.GetOauthCredentials()

	var creds Credentials
	if basicCreds.Username != "" {
		creds = NewBasicCredentials(basicCreds.Username, basicCreds.Password)
	} else {
		creds = NewOauthCredentials(oauthCreds.ClientID, oauthCreds.ClientSecret, oauthCreds.TokenServiceURL)
	}

	return creds
}

// BasicCredentials represents the basic bundle instance auth credentials
type BasicCredentials struct {
	username string
	password string
}

// ToString stringifies BasicCredentials needed in the graphql request
func (c *BasicCredentials) ToString() string {
	return fmt.Sprintf(`basic: { username: "%s", password: "%s" }`, c.username, c.password)
}

// NewBasicCredentials creates new BasicCredentials
func NewBasicCredentials(username, password string) *BasicCredentials {
	return &BasicCredentials{
		username: username,
		password: password,
	}
}

// OauthCredentials represents the basic oauth instance auth credentials
type OauthCredentials struct {
	tokenServiceURL string
	clientID        string
	clientSecret    string
}

// ToString stringifies OauthCredentials needed in the graphql request
func (c *OauthCredentials) ToString() string {
	return fmt.Sprintf(`oauth: { clientId: "%s" clientSecret: "%s" url: "%s"}`, c.clientID, c.clientSecret, c.tokenServiceURL)
}

// NewOauthCredentials creates new OauthCredentials
func NewOauthCredentials(clientID, clientSecret, tokenServiceURL string) *OauthCredentials {
	return &OauthCredentials{
		clientID:        clientID,
		clientSecret:    clientSecret,
		tokenServiceURL: tokenServiceURL,
	}
}
