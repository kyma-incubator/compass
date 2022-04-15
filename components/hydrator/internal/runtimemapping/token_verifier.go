package runtimemapping

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/lestrrat-go/iter/arrayiter"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/form3tech-oss/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
)

const (
	bearerPrefix = "Bearer "

	claimsIssuerKey = "iss"
	jwksKeyIDKey    = "kid"
	jwksURIKey      = "jwks_uri"

	openIDConfigSubpath = ".well-known/openid-configuration"
)

// KeyGetter missing godoc
type KeyGetter interface {
	GetKey(ctx context.Context, token *jwt.Token) (interface{}, error)
}

type tokenVerifier struct {
	keys KeyGetter
}

// NewTokenVerifier missing godoc
func NewTokenVerifier(keys KeyGetter) *tokenVerifier {
	return &tokenVerifier{
		keys: keys,
	}
}

// Verify missing godoc
func (tv *tokenVerifier) Verify(ctx context.Context, token string) (*jwt.MapClaims, error) {
	if token == "" {
		return nil, apperrors.NewUnauthorizedError("token cannot be empty")
	}

	claims, err := tv.verifyToken(ctx, token)
	if err != nil {
		return nil, errors.Wrap(err, "while veryfing the token")
	}

	return claims, nil
}

func (tv *tokenVerifier) verifyToken(ctx context.Context, tokenStr string) (*jwt.MapClaims, error) {
	tokenStr = strings.TrimPrefix(tokenStr, bearerPrefix)
	claims := new(jwt.MapClaims)

	_, err := new(jwt.Parser).ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return tv.keys.GetKey(ctx, t)
	})
	if err != nil {
		return nil, errors.Wrap(err, "while parsing the token with claims")
	}

	return claims, nil
}

type jwksFetch struct {
}

// NewJWKsFetch missing godoc
func NewJWKsFetch() *jwksFetch {
	return &jwksFetch{}
}

// GetKey missing godoc
func (f *jwksFetch) GetKey(ctx context.Context, token *jwt.Token) (interface{}, error) {
	if token == nil {
		return nil, apperrors.NewInternalError("token cannot be nil")
	}

	jwksURI, err := f.getJWKsURI(ctx, *token)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the JWKs URI")
	}

	jwksSet, err := jwk.Fetch(ctx, jwksURI)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching JWKs")
	}

	keyID, err := getKeyID(*token)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the key ID")
	}

	keyIterator := &authenticator.JWTKeyIterator{
		AlgorithmCriteria: func(alg string) bool {
			return token.Method.Alg() == alg
		},
		IDCriteria: func(id string) bool {
			return id == keyID
		},
	}

	if err := arrayiter.Walk(ctx, jwksSet, keyIterator); err != nil {
		return nil, err
	}

	if keyIterator.ResultingKey != nil {
		return keyIterator.ResultingKey, nil
	}

	return nil, apperrors.NewInternalError("unable to find a proper key")
}

func (f *jwksFetch) getJWKsURI(ctx context.Context, token jwt.Token) (string, error) {
	discoveryURL, err := f.getDiscoveryURL(token)
	if err != nil {
		return "", errors.Wrap(err, "while getting the discovery URL")
	}

	res, err := http.Get(discoveryURL)
	if err != nil {
		return "", errors.Wrap(err, "while getting the configuration discovery")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while closing response body: %v", err)
		}
	}()

	var config map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&config); err != nil {
		return "", errors.Wrap(err, "while decoding the configuration discovery response")
	}

	jwksURI, ok := config[jwksURIKey].(string)
	if !ok {
		return "", apperrors.NewInternalError("unable to cast the JWKs URI to a string")
	}

	return jwksURI, nil
}

func (f *jwksFetch) getDiscoveryURL(token jwt.Token) (string, error) {
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return "", apperrors.NewInternalError("unable to cast claims to the MapClaims")
	}

	issuer, err := getTokenIssuer(*claims)
	if err != nil {
		return "", errors.Wrap(err, "while getting the issuer from claims")
	}

	u, err := url.Parse(issuer)
	if err != nil {
		return "", errors.Wrapf(err, "while parsing the issuer URL [issuer=%s]", issuer)
	}

	u.Path = path.Join(u.Path, openIDConfigSubpath)

	return u.String(), nil
}

func getKeyID(token jwt.Token) (string, error) {
	keyID, ok := token.Header[jwksKeyIDKey]
	if !ok {
		return "", apperrors.NewInternalError("unable to find the key ID in the token")
	}

	keyIDStr, ok := keyID.(string)
	if !ok {
		return "", apperrors.NewInternalError("unable to cast the key ID to a string")
	}

	return keyIDStr, nil
}

func getTokenIssuer(claims jwt.MapClaims) (string, error) {
	issuer, ok := claims[claimsIssuerKey]
	if !ok {
		return "", apperrors.NewInternalError("no issuer claim found")
	}

	issuerStr, ok := issuer.(string)
	if !ok {
		return "", apperrors.NewInternalError("unable to cast the issuer to a string")
	}

	return issuerStr, nil
}
