package credentials

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/types/tenantmapping"
)

type Credentials interface {
	ToString() string
}

func NewCredentials(body tenantmapping.Body) Credentials {
	basicCreds := body.GetBasicCredentials()
	oauthCreds := body.GetOauthCredentials()

	var creds Credentials
	if basicCreds.Username != "" {
		creds = NewBasicCredentials(basicCreds.Username, basicCreds.Password)
	} else {
		creds = NewOauthCredentials(oauthCreds.ClientID, oauthCreds.ClientSecret, oauthCreds.TokenServiceURL)
	}

	return creds
}

type BasicCredentials struct {
	username string
	password string
}

func (c *BasicCredentials) ToString() string {
	return fmt.Sprintf(`basic: { username: "%s", password: "%s" }`, c.username, c.password)
}

func NewBasicCredentials(username, password string) *BasicCredentials {
	return &BasicCredentials{
		username: username,
		password: password,
	}
}

type OauthCredentials struct {
	tokenServiceURL string
	clientID        string
	clientSecret    string
}

func (c *OauthCredentials) ToString() string {
	return fmt.Sprintf(`oauth: { clientId: "%s" clientSecret: "%s" url: "%s"}`, c.clientID, c.clientSecret, c.tokenServiceURL)
}

func NewOauthCredentials(clientID, clientSecret, tokenServiceURL string) *OauthCredentials {
	return &OauthCredentials{
		clientID:        clientID,
		clientSecret:    clientSecret,
		tokenServiceURL: tokenServiceURL,
	}
}
