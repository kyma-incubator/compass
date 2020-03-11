package runtimemapping

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	bearerPrefix = "Bearer "

	claimsIssuerKey = "iss"
	claimsNameKey   = "name"
	claimsGroupsKey = "groups"
	claimsEmailKey  = "email"
	jwksKeyIDKey    = "kid"
	jwksURIKey      = "jwks_uri"

	openIDConfigSubpath = ".well-known/openid-configuration"
)

type KeyGetter interface {
	GetKey(token *jwt.Token) (interface{}, error)
}

type tokenVerifier struct {
	logger *logrus.Logger
	keys   KeyGetter
}

func NewTokenVerifier(logger *logrus.Logger, keys KeyGetter) *tokenVerifier {
	return &tokenVerifier{
		logger: logger,
		keys:   keys,
	}
}

func (tv *tokenVerifier) Verify(token string) (*jwt.MapClaims, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	claims, err := tv.verifyToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "while veryfing the token")
	}

	return claims, nil
}

func (tv *tokenVerifier) verifyToken(tokenStr string) (*jwt.MapClaims, error) {
	tokenStr = strings.TrimPrefix(tokenStr, bearerPrefix)
	claims := new(jwt.MapClaims)

	_, err := new(jwt.Parser).ParseWithClaims(tokenStr, claims, tv.keys.GetKey)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing the token with claims")
	}

	return claims, nil
}

type jwksFetch struct {
	logger *logrus.Logger
}

func NewJWKsFetch(logger *logrus.Logger) *jwksFetch {
	return &jwksFetch{
		logger: logger,
	}
}

func (f *jwksFetch) GetKey(token *jwt.Token) (interface{}, error) {
	if token == nil {
		return nil, errors.New("token cannot be nil")
	}

	jwksURI, err := f.getJWKsURI(*token)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the JWKs URI")
	}

	jwksSet, err := jwk.Fetch(jwksURI)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching JWKs")
	}

	keyID, err := getKeyID(*token)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the key ID")
	}

	keys := jwksSet.LookupKeyID(keyID)
	for _, key := range keys {
		if key.Algorithm() == token.Method.Alg() {
			return key.Materialize()
		}
	}

	return nil, errors.New("unable to find a proper key")
}

func (f *jwksFetch) getJWKsURI(token jwt.Token) (string, error) {
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
			f.logger.Error(err)
		}
	}()

	var config map[string]interface{}
	if err = json.NewDecoder(res.Body).Decode(&config); err != nil {
		return "", errors.Wrap(err, "while decoding the configuration discovery response")
	}

	jwksURI, ok := config[jwksURIKey].(string)
	if !ok {
		return "", errors.New("unable to cast the JWKs URI to a string")
	}

	return jwksURI, nil
}

func (f *jwksFetch) getDiscoveryURL(token jwt.Token) (string, error) {
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return "", errors.New("unable to cast claims to the MapClaims")
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
		return "", errors.New("unable to find the key ID in the token")
	}

	keyIDStr, ok := keyID.(string)
	if !ok {
		return "", errors.New("unable to cast the key ID to a string")
	}

	return keyIDStr, nil
}

func getTokenIssuer(claims jwt.MapClaims) (string, error) {
	issuer, ok := claims[claimsIssuerKey]
	if !ok {
		return "", errors.New("no issuer claim found")
	}

	issuerStr, ok := issuer.(string)
	if !ok {
		return "", errors.New("unable to cast the issuer to a string")
	}

	return issuerStr, nil
}
