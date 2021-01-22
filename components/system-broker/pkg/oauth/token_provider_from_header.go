package oauth

import (
	"context"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pkg/errors"
)

const AuthzHeader = "Authorization"

type TokenProviderFromHeader struct {
	targetURL *url.URL
}

func NewTokenProviderFromHeader(targetURL string) (*TokenProviderFromHeader, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &TokenProviderFromHeader{
		targetURL: parsedURL,
	}, nil
}

func (c *TokenProviderFromHeader) Name() string {
	return "TokenProviderFromHeader"
}

func (c *TokenProviderFromHeader) Matches(ctx context.Context) bool {
	if _, err := getBearerToken(ctx); err != nil {
		log.C(ctx).WithError(err).Warn("while obtaining bearer token")
		return false
	}

	return true
}

func (c *TokenProviderFromHeader) TargetURL() *url.URL {
	return c.targetURL
}

func (c *TokenProviderFromHeader) GetAuthorizationToken(ctx context.Context) (httputils.Token, error) {
	token, err := getBearerToken(ctx)
	if err != nil {
		return httputils.Token{}, errors.Wrapf(err, "while obtaining bearer token from header %s", AuthzHeader)
	}

	tokenResponse := httputils.Token{
		AccessToken: token,
		Expiration:  0,
	}

	log.C(ctx).Info("Successfully unmarshal response oauth token for accessing Director")
	return tokenResponse, nil
}

func getBearerToken(ctx context.Context) (string, error) {
	headers, err := httputils.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Errorf("cannot read headers from context: %s", err.Error())
	}
	reqToken, ok := headers[AuthzHeader]
	if !ok {
		return "", errors.Errorf("cannot read header %s from context", AuthzHeader)
	}

	if reqToken == "" {
		return "", apperrors.NewUnauthorizedError("missing bearer token")
	}

	if !strings.HasPrefix(strings.ToLower(reqToken), "bearer ") {
		return "", apperrors.NewUnauthorizedError("invalid bearer token prefix")
	}

	return strings.TrimPrefix(reqToken, "Bearer "), nil
}
