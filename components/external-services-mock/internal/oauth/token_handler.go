package oauth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

	if r.FormValue(GrantTypeFieldName) != CredentialsGrantType && r.FormValue(GrantTypeFieldName) != PasswordGrantType {
		log.C(ctx).Errorf("The grant_type should be %s or %s but we got: %s", CredentialsGrantType, PasswordGrantType, r.FormValue(GrantTypeFieldName))
		httphelpers.WriteError(writer, errors.New("An error occurred while parsing query"), http.StatusBadRequest)
		return
	}

	if r.FormValue(GrantTypeFieldName) == CredentialsGrantType {
		if err := h.handleClientCredentialsRequest(r); err != nil {
			log.C(ctx).Error(err)
			httphelpers.WriteError(writer, err, http.StatusBadRequest)
			return
		}
	} else { // Assume it's a password flow because currently we support only client_credentials and password
		if err := h.handlePasswordCredentialsRequest(r); err != nil {
			log.C(ctx).Error(err)
			httphelpers.WriteError(writer, err, http.StatusBadRequest)
			return
		}
	}

	claims := make(map[string]interface{})
	claimsFunc, ok := h.staticMappingClaims[r.FormValue(ClaimsKey)]
	if ok { // If the request contains claims key, use the corresponding claims in the static mapping for that key
		claims = claimsFunc()
	} else { // If there is no claims key provided use empty claims
		log.C(ctx).Info("Did not find claims key in the request. Proceeding with empty claims...")
	}

	claims[ZidKey] = r.Header.Get(h.tenantHeaderName)
	respond(writer, r, claims, h.signingKey)
}

func (h *handler) GenerateWithoutCredentials(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := map[string]interface{}{}

	tenant := r.Header.Get(h.tenantHeaderName)
	claims[ZidKey] = tenant

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).Errorf("while reading request body: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, &claims)
		if err != nil {
			log.C(ctx).WithError(err).Infof("Cannot json unmarshal the request body. Error: %s. Proceeding with empty claims", err)
		}
	}

	respond(writer, r, claims, h.signingKey)
}

func (h *handler) GenerateWithCredentialsFromReqBody(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := map[string]interface{}{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).Errorf("while reading request body: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	if form, err := url.ParseQuery(string(body)); err != nil {
		log.C(ctx).Errorf("Cannot parse form. Error: %s", err)
	} else {
		client := form.Get(ClientIDKey)
		scopes := form.Get(ScopesFieldName)
		tenant := r.Header.Get(h.tenantHeaderName)
		claims[ClientIDKey] = client
		claims[ScopesFieldName] = scopes
		claims[ZidKey] = tenant
	}

	respond(writer, r, claims, h.signingKey)
}

func (h *handler) handleClientCredentialsRequest(r *http.Request) error {
	ctx := r.Context()
	log.C(ctx).Info("Validating client credentials token request...")
	authorization := r.Header.Get("authorization")
	if id, secret, err := getBasicCredentials(authorization); err != nil {
		log.C(ctx).Info("Did not find client_id or client_secret in the authorization header. Checking the request body...")
		id = r.FormValue(ClientIDKey)
		secret = r.FormValue(ClientSecretKey)
		if id != h.expectedID || secret != h.expectedSecret {
			return errors.New(fmt.Sprintf("client_id or client_secret from request body doesn't match the expected one. Expected: %s and %s but we got: %s and %s", h.expectedID, h.expectedSecret, id, secret))
		}
	} else if id != h.expectedID || secret != h.expectedSecret {
		return errors.New(fmt.Sprintf("client_id or client_secret from authorization header doesn't match the expected one. Expected: %s and %s but we got: %s and %s", h.expectedID, h.expectedSecret, id, secret))
	}

	log.C(ctx).Info("Successfully validated client credentials token request")

	return nil
}

func (h *handler) handlePasswordCredentialsRequest(r *http.Request) error {
	ctx := r.Context()
	log.C(ctx).Info("Validating password grant type token request...")
	authorization := r.Header.Get("authorization")
	id, secret, err := getBasicCredentials(authorization)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("client_id or client_secret doesn't match the expected one. Expected: %s and %s but we got: %s and %s", h.expectedID, h.expectedSecret, id, secret))
	}

	if id != h.expectedID || secret != h.expectedSecret {
		return errors.New(fmt.Sprintf("client_id or client_secret doesn't match the expected one. Expected: %s and %s but we got: %s and %s", h.expectedID, h.expectedSecret, id, secret))
	}

	username := r.FormValue(UserNameKey)
	password := r.FormValue(PasswordKey)
	if username != h.expectedUsername || password != h.expectedPassword {
		return errors.New(fmt.Sprintf("username or password doesn't match the expected one. Expected: %s and %s but we got: %s and %s", h.expectedUsername, h.expectedPassword, username, password))
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

func (h *handler) GenerateWithCredentialsFromReqBody(writer http.ResponseWriter, r *http.Request) {
	claims := map[string]interface{}{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(r.Context()).Errorf("while reading request body: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	if form, err := url.ParseQuery(string(body)); err != nil {
		log.C(r.Context()).Errorf("Cannot parse form. Error: %s", err)
	} else {
		client := form.Get("client_id")
		scopes := form.Get("scopes")
		tenant := r.Header.Get(h.tenantHeaderName)
		claims["client_id"] = client
		claims["scopes"] = scopes
		claims["x-zid"] = tenant
	}

	var output string
	if h.signingKey != nil {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(claims))
		output, err = token.SignedString(h.signingKey)
	} else {
		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims(claims))
		output, err = token.SigningString()
	}

	if err != nil {
		log.C(r.Context()).Errorf("while creating oauth token: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while creating oauth token"), http.StatusInternalServerError)
		return
	}

	response := createResponse(output)
	payload, err := json.Marshal(response)
	if err != nil {
		log.C(r.Context()).Errorf("while marshalling response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling response"), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(payload)
	if err != nil {
		log.C(r.Context()).Errorf("while writing response: %s", err.Error())
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
