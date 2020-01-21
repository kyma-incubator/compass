package api

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
)

type certRequest struct {
	CSR string `json:"csr"`
}

type certificatesHandler struct {
	client graphql.Client
}

func NewCertificatesHandler(client graphql.Client) certificatesHandler {
	return certificatesHandler{
		client: client,
	}
}

func (ih *certificatesHandler) SignCSR(w http.ResponseWriter, r *http.Request) {
	certRequest, err := readCertRequest(r)
	if err != nil {
		log.Println("Error reading cert request: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
	if clientIdFromToken == "" {
		log.Println("Client Id not provided.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	certificationResult, err := ih.client.SignCSR(certRequest.CSR, map[string]string{})

	if err != nil {
		log.Println("Error getting cert from Connector: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	certResponse := graphql.ToCertResponse(certificationResult)
	respondWithBody(w, http.StatusOK, certResponse)
}

func readCertRequest(r *http.Request) (*certRequest, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error while reading request body: %s")
	}
	defer r.Body.Close()

	var tokenRequest certRequest
	err = json.Unmarshal(b, &tokenRequest)
	if err != nil {
		return nil, errors.Wrap(err, "Error while unmarshalling request body: %s")
	}

	return &tokenRequest, nil
}

func respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) {
	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(HeaderContentType, ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}
