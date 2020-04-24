package oauth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/dgrijalva/jwt-go"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type handler struct {
	secret string
	id     string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	Jti         string `json:"jti"`
}

func NewHandler() *handler {
	return &handler{}
}

func (h *handler) Generate(writer http.ResponseWriter, r *http.Request) {
	authorizaion := r.Header.Get("authorization")
	fmt.Println(authorizaion)
	id, secret, err := getBasicCredentials(authorizaion)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("client secret not found in header"), http.StatusBadRequest)
		return
	}
	fmt.Println(id, secret)

	//grant_type is in body!
	vars := mux.Vars(r)
	grantType, ok := vars["grant_type"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("invalid request token grant type"), http.StatusBadRequest)
		return
	}
	fmt.Println(grantType)

	token := jwt.New(jwt.SigningMethodNone)
	output, err := token.SigningString()
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while creating oauth token"), http.StatusInternalServerError)
		return
	}

	response := createResponse(output)
	payload, err := json.Marshal(response)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling resposne"), http.StatusInternalServerError)
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
		Scope:       "scope",
		Jti:         "daksdakjsdksahkdja",
	}
}

func getBasicCredentials(rawData string) (string, string, error) {
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
