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

type handler struct {
	expectedSecret   string
	expectedID       string
	tenantHeaderName string
	signingKey       *rsa.PrivateKey
}

func NewHandler(expectedSecret, expectedID string) *handler {
	return &handler{
		expectedSecret: expectedSecret,
		expectedID:     expectedID,
	}
}

func NewHandlerWithSigningKey(expectedSecret, expectedID, tenantHeaderName string, signingKey *rsa.PrivateKey) *handler {
	return &handler{
		expectedSecret:   expectedSecret,
		expectedID:       expectedID,
		tenantHeaderName: tenantHeaderName,
		signingKey:       signingKey,
	}
}

func (h *handler) Generate(writer http.ResponseWriter, r *http.Request) {
	log.C(r.Context()).Infof("Generate: %+v", r)

	authorization := r.Header.Get("authorization")
	id, secret, err := getBasicCredentials(authorization)
	if err != nil {
		log.C(r.Context()).Errorf("client secret not found in header: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "client secret not found in header"), http.StatusBadRequest)
		return
	}

	if h.expectedID != id || h.expectedSecret != secret {
		log.C(r.Context()).Error("client secret or client id doesn't match expected")
		httphelpers.WriteError(writer, errors.New("client secret or client id doesn't match expected"), http.StatusBadRequest)
		return
	}
	h.GenerateWithoutCredentials(writer, r)
}

func (h *handler) GenerateWithoutCredentials(writer http.ResponseWriter, r *http.Request) {
	claims := map[string]interface{}{}
	log.C(r.Context()).Infof("GenerateWithoutCredentials: %+v", r)

	tenant := r.Header.Get(h.tenantHeaderName)
	claims["x-zid"] = tenant

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(r.Context()).Errorf("while reading request body: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		if form, err := url.ParseQuery(string(body)); err != nil {
			log.C(r.Context()).Errorf("Cannot parse form. Error: %s", err)
		} else {
			client := form.Get("client_id")
			scopes := form.Get("scopes")
			claims["client_id"] = client
			claims["scopes"] = scopes
		}
	} else {
		if len(body) > 0 {
			err = json.Unmarshal(body, &claims)
			if err != nil {
				log.C(r.Context()).WithError(err).Infof("Cannot json unmarshal the request body. Error: %s. Proceeding with empty claims", err)
			}
		}
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
	log.C(r.Context()).Infof("Returning: %s", string(payload))
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
