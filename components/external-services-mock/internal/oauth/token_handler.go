package oauth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
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
	contentTypeHeader                = "Content-Type"
	contentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"
	contentTypeApplicationJson       = "application/json"

	grantTypeFieldName   = "grant_type"
	credentialsGrantType = "client_credentials"
	passwordGrantType    = "password"
	scopesFieldName      = "scopes"
	claimsKey            = "claims_key"

	clientIDKey     = "client_id"
	clientSecretKey = "client_secret"
	userNameKey     = "username"
	passwordKey     = "password"
	zidKey          = "x-zid"
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

	if r.Header.Get(contentTypeHeader) != contentTypeApplicationURLEncoded {
		log.C(ctx).Errorf("Unsupported media type, expected: application/x-www-form-urlencodedgot: %s", r.Header.Get(contentTypeHeader))
		writer.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if r.FormValue(grantTypeFieldName) != credentialsGrantType && r.FormValue(grantTypeFieldName) != passwordGrantType {
		log.C(ctx).Errorf("The grant_type should be %s or %s but we got: %s", credentialsGrantType, passwordGrantType, r.FormValue(grantTypeFieldName))
		httphelpers.WriteError(writer, errors.New("An error occurred while parsing query"), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while parsing query")
		httphelpers.WriteError(writer, errors.New("An error occurred while parsing query"), http.StatusInternalServerError)
		return
	}

	if r.FormValue(grantTypeFieldName) == credentialsGrantType {
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

	var claims map[string]interface{}
	claimsFunc, ok := h.staticMappingClaims[r.FormValue(claimsKey)]
	if ok { // If the request contains claims key, use the corresponding claims in the static mapping for that key
		claims = claimsFunc()
		claims[zidKey] = r.Header.Get(h.tenantHeaderName)
		respond(writer, r, claims, h.signingKey)
	} else { // If there is no claims key provided use empty claims
		log.C(ctx).Info("Did not find claims key in the request. Proceeding with empty claims...")
		respond(writer, r, claims, h.signingKey)
	}
}

func (h *handler) GenerateWithoutCredentials(writer http.ResponseWriter, r *http.Request) {
	claims := map[string]interface{}{}

	tenant := r.Header.Get(h.tenantHeaderName)
	claims[zidKey] = tenant

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(r.Context()).Errorf("while reading request body: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, &claims)
		if err != nil {
			log.C(r.Context()).WithError(err).Infof("Cannot json unmarshal the request body. Error: %s. Proceeding with empty claims", err)
		}
	}

	respond(writer, r, claims, h.signingKey)
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
		client := form.Get(clientIDKey)
		scopes := form.Get(scopesFieldName)
		tenant := r.Header.Get(h.tenantHeaderName)
		claims[clientIDKey] = client
		claims[scopesFieldName] = scopes
		claims[zidKey] = tenant
	}

	respond(writer, r, claims, h.signingKey)
}

func (h *handler) handleClientCredentialsRequest(r *http.Request) error {
	log.C(r.Context()).Info("Validating client credentials token request...")

	authorization := r.Header.Get("authorization")
	if id, secret, err := getBasicCredentials(authorization); err != nil {
		log.C(r.Context()).Info("Did not find client id or client secret in authorization header. Checking the request body...")
		if r.FormValue(clientIDKey) != h.expectedID || r.FormValue(clientSecretKey) != h.expectedSecret {
			return errors.New("client id or client secret from request body doesn't match the expected one")
		}
	} else if h.expectedID != id || h.expectedSecret != secret {
		return errors.New("client id or client secret from authorization header doesn't match the expected one")
	}

	log.C(r.Context()).Info("Successfully validated client credentials token request")

	return nil
}

func (h *handler) handlePasswordCredentialsRequest(r *http.Request) error {
	log.C(r.Context()).Info("Validating password grant type token request...")
	authorization := r.Header.Get("authorization")
	id, secret, err := getBasicCredentials(authorization)
	if err != nil {
		return errors.Wrap(err, "client secret or client id doesn't match the expected one")
	}

	if id != h.expectedID || secret != h.expectedSecret {
		return errors.New("client secret or client id doesn't match the expected one")
	}

	if r.FormValue(userNameKey) != h.expectedUsername || r.FormValue(passwordKey) != h.expectedPassword {
		return errors.New("username or password doesn't match the expected one")
	}

	log.C(r.Context()).Info("Successfully validated password grant type token request")

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
	writer.Header().Set(contentTypeHeader, contentTypeApplicationJson)
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(payload); err != nil {
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
	encodedCredentials := strings.TrimPrefix(rawData, "Basic ")
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
