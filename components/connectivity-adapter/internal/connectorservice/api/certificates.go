package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type certRequest struct {
	CSR string `json:"csr"`
}

type certificatesHandler struct {
	connectorClientProvider connector.ClientProvider
	logger                  *log.Logger
}

func NewCertificatesHandler(connectorClientProvider connector.ClientProvider, logger *log.Logger) certificatesHandler {
	return certificatesHandler{
		connectorClientProvider: connectorClientProvider,
		logger:                  logger,
	}
}

func (ch *certificatesHandler) SignCSR(w http.ResponseWriter, r *http.Request) {
	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		res.WriteErrorMessage(w, "Failed to read authorization context.", apperrors.CodeForbidden)

		return
	}

	contextLogger := contextLogger(ch.logger, authorizationHeaders.GetSystemAuthID())
	certRequest, err := readCertRequest(r)
	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to read certificate request"), apperrors.CodeWrongInput)

		return
	}

	contextLogger.Info("Generating certificate")

	{
		certificationResult, err := ch.connectorClientProvider.Client(r).SignCSR(r.Context(), certRequest.CSR, authorizationHeaders)
		if err != nil {
			respondWithError(w, contextLogger, errors.Wrap(err, "Failed to sign CSR"), err.Code())

			return
		}

		certResponse := model.ToCertResponse(certificationResult)
		respondWithBody(w, http.StatusCreated, certResponse, contextLogger)
	}
}

func readCertRequest(r *http.Request) (*certRequest, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error while reading request body: %s")
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Errorf("Failed to close response body: %s", err)
		}
	}()

	var certRequest certRequest
	err = json.Unmarshal(b, &certRequest)
	if err != nil {
		return nil, errors.Wrap(err, "error while unmarshalling request body: %s")
	}

	return &certRequest, nil
}

func respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}, logger *log.Entry) {
	respond(w, statusCode)
	err := json.NewEncoder(w).Encode(responseBody)
	if err != nil {
		logger.Errorf("Failed to encode response body: %s", err)
	}
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(HeaderContentType, ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}

func respondWithError(w http.ResponseWriter, contextLogger *log.Entry, err error, appErroCode int) {
	contextLogger.Error(err.Error())
	res.WriteError(w, err, appErroCode)
}

func contextLogger(logger *log.Logger, application string) *log.Entry {
	return logger.WithField("application", application)
}
