package oauth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

const (
	AuthorizationHeader              = "Authorization"
	ContentTypeHeader                = "Content-Type"
	ContentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeApplicationJson       = "application/json"

	GrantTypeFieldName   = "grant_type"
	CredentialsGrantType = "client_credentials"
	PasswordGrantType    = "password"
	ScopesFieldName      = "scopes"
	ClaimsKey            = "claims_key"

	ClientIDKey     = "client_id"
	ClientSecretKey = "client_secret"
	UserNameKey     = "username"
	PasswordKey     = "password"
	ZidKey          = "x-zid"
)

type ClaimsGetterFunc func() map[string]interface{}

type handler struct {
	expectedSecret      string
	expectedID          string
	tenantHeaderName    string
	expectedUsername    string
	expectedPassword    string
	signingKey          *rsa.PrivateKey
	staticMappingClaims map[string]ClaimsGetterFunc
}

func NewHandler(expectedSecret, expectedID string) *handler {
	return &handler{
		expectedSecret: expectedSecret,
		expectedID:     expectedID,
	}
}

func NewHandlerWithSigningKey(expectedSecret, expectedID, tenantHeaderName, expectedUsername, expectedPassword string, signingKey *rsa.PrivateKey, staticMappingClaims map[string]ClaimsGetterFunc) *handler {
	return &handler{
		expectedSecret:      expectedSecret,
		expectedID:          expectedID,
		tenantHeaderName:    tenantHeaderName,
		expectedUsername:    expectedUsername,
		expectedPassword:    expectedPassword,
		signingKey:          signingKey,
		staticMappingClaims: staticMappingClaims,
	}
}

func (h *handler) Generate(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Header.Get(ContentTypeHeader) != ContentTypeApplicationURLEncoded {
		log.C(ctx).Errorf("Unsupported media type, expected: application/x-www-form-urlencoded got: %s", r.Header.Get(ContentTypeHeader))
		writer.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while parsing query")
		httphelpers.WriteError(writer, errors.New("An error occurred while parsing query"), http.StatusInternalServerError)
		return
	}

	switch r.FormValue(GrantTypeFieldName) {
	case CredentialsGrantType:
		if err := h.authenticateClientCredentialsRequest(r); err != nil {
			log.C(ctx).Error(err)
			httphelpers.WriteError(writer, err, http.StatusBadRequest)
			return
		}
	case PasswordGrantType:
		if err := h.authenticatePasswordCredentialsRequest(r); err != nil {
			log.C(ctx).Error(err)
			httphelpers.WriteError(writer, err, http.StatusBadRequest)
			return
		}
	default:
		log.C(ctx).Errorf("The grant_type should be %s or %s but we got: %s", CredentialsGrantType, PasswordGrantType, r.FormValue(GrantTypeFieldName))
		httphelpers.WriteError(writer, errors.New("An error occurred while parsing query"), http.StatusBadRequest)
		return
	}

	claims := make(map[string]interface{})
	claimsFunc, ok := h.staticMappingClaims[r.FormValue(ClaimsKey)]
	if ok { // If the request contains claims key, use the corresponding claims in the static mapping for that key
		claims = claimsFunc()
	} else { // If there is no claims key provided use default claims
		log.C(ctx).Info("Did not find claims key in the request. Proceeding with default claims...")
		claims[ClientIDKey] = r.FormValue(ClientIDKey)
		claims[ScopesFieldName] = r.Form.Get(ScopesFieldName)
	}

	claims[ZidKey] = r.Header.Get(h.tenantHeaderName)
	respond(writer, r, claims, h.signingKey)
}

func (h *handler) authenticateClientCredentialsRequest(r *http.Request) error {
	ctx := r.Context()
	log.C(ctx).Info("Validating client credentials token request...")
	authorization := r.Header.Get("authorization")
	if id, secret, err := getBasicCredentials(authorization); err != nil {
		log.C(ctx).Info("Did not find client_id or client_secret in the authorization header. Checking the request body...")
		id = r.FormValue(ClientIDKey)
		secret = r.FormValue(ClientSecretKey)
		if id != h.expectedID || secret != h.expectedSecret {
			return errors.New("client_id or client_secret from request body doesn't match the expected one")
		}
	} else if id != h.expectedID || secret != h.expectedSecret {
		return errors.New("client_id or client_secret from authorization header doesn't match the expected one")
	}

	log.C(ctx).Info("Successfully validated client credentials token request")

	return nil
}

func (h *handler) authenticatePasswordCredentialsRequest(r *http.Request) error {
	ctx := r.Context()
	log.C(ctx).Info("Validating password grant type token request...")
	authorization := r.Header.Get("authorization")
	id, secret, err := getBasicCredentials(authorization)
	if err != nil {
		return errors.Wrap(err, "client_id or client_secret doesn't match the expected one")
	}

	if id != h.expectedID || secret != h.expectedSecret {
		return errors.New("client_id or client_secret doesn't match the expected one")
	}

	username := r.FormValue(UserNameKey)
	password := r.FormValue(PasswordKey)
	if username != h.expectedUsername || password != h.expectedPassword {
		return errors.New("username or password doesn't match the expected one")
	}

	log.C(ctx).Info("Successfully validated password grant type token request")

	return nil
}

func respond(writer http.ResponseWriter, r *http.Request, claims map[string]interface{}, signingKey *rsa.PrivateKey) {
	var (
		output string
		err    error
	)
	if signingKey != nil {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(claims))
		output, err = token.SignedString(signingKey)
	} else {
		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims(claims))
		output, err = token.SigningString()
	}

	ctx := r.Context()
	if err != nil {
		log.C(ctx).Errorf("while creating oauth token: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while creating oauth token"), http.StatusInternalServerError)
		return
	}

	response := createResponse(output)
	payload, err := json.Marshal(response)
	if err != nil {
		log.C(ctx).Errorf("while marshalling response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling response"), http.StatusInternalServerError)
		return
	}
	writer.Header().Set(ContentTypeHeader, ContentTypeApplicationJson)
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(payload); err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
}

func createResponse(token string) TokenResponse {
	return TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresIn:   100000,
	}
}

func getBasicCredentials(rawData string) (id string, secret string, err error) {
	if len(rawData) == 0 {
		return "", "", errors.New("missing authorization header")
	}

	encodedCredentials := strings.TrimPrefix(rawData, "Basic ")
	if len(encodedCredentials) == 0 {
		return "", "", errors.New("the credentials cannot be empty")
	}

	output, err := base64.URLEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return "", "", errors.Wrap(err, "while decoding basic credentials")
	}

	credentials := strings.Split(string(output), ":")
	if len(credentials) != 2 {
		return "", "", errors.New("invalid credential format")
	}

	return credentials[0], credentials[1], nil
}
