package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func NewIntegrationSystemClient(ctx context.Context, intSystemCredentials *directorSchema.IntSysSystemAuth) (*http.Client, error) {
	oauthCredentialData, ok := intSystemCredentials.Auth.Credential.(*directorSchema.OAuthCredentialData)
	if !ok {
		return nil, fmt.Errorf("while casting integration system credentials to OAuth credentials")
	}

	conf := &clientcredentials.Config{
		ClientID:     oauthCredentialData.ClientID,
		ClientSecret: oauthCredentialData.ClientSecret,
		TokenURL:     oauthCredentialData.URL,
	}

	unsecuredHttpClient := http.DefaultClient
	unsecuredHttpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, unsecuredHttpClient)
	httpClient := conf.Client(ctx)
	httpClient.Timeout = 20 * time.Second

	return httpClient, nil
}
