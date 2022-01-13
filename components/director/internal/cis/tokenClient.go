package cis

import (
	"context"
	"golang.org/x/oauth2/clientcredentials"
)

// FetchToken missing docs
func FetchToken(ctx context.Context, clientID, clientSecret, tokenURL string) (string, error) {
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}
