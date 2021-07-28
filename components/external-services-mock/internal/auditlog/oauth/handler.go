package oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type handler struct {
	expectedSecret string
	expectedID     string
}

func NewHandler(expectedSecret, expectedID string) *handler {
	return &handler{
		expectedSecret: expectedSecret,
		expectedID:     expectedID,
	}
}

func (h *handler) Generate(writer http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("authorization")
	id, secret, err := getBasicCredentials(authorization)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("client secret not found in header"), http.StatusBadRequest)
		return
	}

	if h.expectedID != id || h.expectedSecret != secret {
		httphelpers.WriteError(writer, errors.New("client secret or client id doesn't match expected"), http.StatusBadRequest)
		return
	}

	h.GenerateWithoutCredentials(writer, r)
}

func (h *handler) GenerateWithoutCredentials(writer http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodNone)
	output, err := token.SigningString()
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while creating oauth token"), http.StatusInternalServerError)
		return
	}

	response := createResponse(output)
	payload, err := json.Marshal(response)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling response"), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(payload)
	if err != nil {
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
