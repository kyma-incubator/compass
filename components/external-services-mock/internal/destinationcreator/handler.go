package destinationcreator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	clientUserHeaderKey = "CLIENT_USER"
	CertChain           = "e2e-test-destination-cert-mock-cert-chain"
)

var respErrorMsg = "An unexpected error occurred while processing the request"

// Handler is responsible to mock and handle any Destination Creator Service and Destination Service requests
type Handler struct {
	Config                            *Config
	DestinationCreatorSvcDestinations map[string]json.RawMessage
	DestinationCreatorSvcCertificates map[string]json.RawMessage
	DestinationSvcDestinations        map[string]json.RawMessage
	DestinationSvcCertificates        map[string]json.RawMessage
}

// NewHandler creates a new Handler
func NewHandler(config *Config) *Handler {
	return &Handler{
		Config:                            config,
		DestinationCreatorSvcDestinations: make(map[string]json.RawMessage),
		DestinationCreatorSvcCertificates: make(map[string]json.RawMessage),
		DestinationSvcDestinations:        make(map[string]json.RawMessage),
		DestinationSvcCertificates:        make(map[string]json.RawMessage),
	}
}

// Destination Creator Service handlers + helper functions

// CreateDestinations mocks creation of all types of destinations in both Destination Creator Service and Destination Service
func (h *Handler) CreateDestinations(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		respondWithHeader(ctx, writer, fmt.Sprintf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey)), http.StatusUnsupportedMediaType)
	}

	if r.Header.Get(clientUserHeaderKey) == "" {
		respondWithHeader(ctx, writer, fmt.Sprintf("The %q header could not be empty", clientUserHeaderKey), http.StatusBadRequest)
	}

	routeVars := mux.Vars(r)
	if err := h.validateDestinationCreatorPathParams(routeVars, false, true); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while reading destination request body"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	authTypeResult := gjson.GetBytes(bodyBytes, "authenticationType")
	if !authTypeResult.Exists() || authTypeResult.String() == "" {
		err := errors.New("The authenticationType field in the request body is required and it should not be empty")
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	switch destinationcreatorpkg.AuthType(authTypeResult.String()) {
	case destinationcreatorpkg.AuthTypeNoAuth:
		statusCode, err := h.createDesignTimeDestination(ctx, bodyBytes)
		if err != nil {
			httphelpers.RespondWithError(ctx, writer, err, "An unexpected error occurred while creating design time destination", correlationID, statusCode)
			return
		}
		httputils.Respond(writer, statusCode)
	case destinationcreatorpkg.AuthTypeBasic:
		statusCode, err := h.createBasicDestination(ctx, bodyBytes)
		if err != nil {
			httphelpers.RespondWithError(ctx, writer, err, "An unexpected error occurred while creating basic destination", correlationID, statusCode)
			return
		}
		httputils.Respond(writer, statusCode)
	case destinationcreatorpkg.AuthTypeSAMLAssertion:
		statusCode, err := h.createSAMLAssertionDestination(ctx, bodyBytes)
		if err != nil {
			httphelpers.RespondWithError(ctx, writer, err, "An unexpected error occurred while creating SAML assertion destination", correlationID, statusCode)
			return
		}
		httputils.Respond(writer, statusCode)
	case destinationcreatorpkg.AuthTypeClientCertificate:
		statusCode, err := h.createClientCertificateAuthDestination(ctx, bodyBytes)
		if err != nil {
			httphelpers.RespondWithError(ctx, writer, err, "An unexpected error occurred while creating client certificate authentication destination", correlationID, statusCode)
			return
		}
		httputils.Respond(writer, statusCode)
	default:
		err := errors.Errorf("The provided destination authentication type: %s is invalid", authTypeResult.String())
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusInternalServerError)
		return
	}
}

// DeleteDestinations mocks deletion of destinations from both Destination Creator Service and Destination Service
func (h *Handler) DeleteDestinations(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		respondWithHeader(ctx, writer, fmt.Sprintf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey)), http.StatusUnsupportedMediaType)
	}

	if r.Header.Get(clientUserHeaderKey) == "" {
		respondWithHeader(ctx, writer, fmt.Sprintf("The %q header could not be empty", clientUserHeaderKey), http.StatusBadRequest)
	}

	routeVars := mux.Vars(r)
	if err := h.validateDestinationCreatorPathParams(routeVars, true, true); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}
	destinationNameValue := routeVars[h.Config.DestinationAPIConfig.DestinationNameParam]

	_, isDestinationCreatorDestExist := h.DestinationCreatorSvcDestinations[destinationNameValue]
	if !isDestinationCreatorDestExist {
		log.C(ctx).Infof("Destination with name: %q does not exists in the destination creator", destinationNameValue)
	} else {
		delete(h.DestinationCreatorSvcDestinations, destinationNameValue)
		log.C(ctx).Infof("Destination with name: %q was deleted from the destination creator", destinationNameValue)
	}

	_, isDestinationSvcDestExists := h.DestinationSvcDestinations[destinationNameValue]
	if !isDestinationSvcDestExists {
		log.C(ctx).Infof("Destination with name: %q does not exists in the destination service. Returning 204 No Content...", destinationNameValue)
		httputils.Respond(writer, http.StatusNoContent)
	}
	delete(h.DestinationSvcDestinations, destinationNameValue)
	log.C(ctx).Infof("Destination with name: %q was deleted from the destination service", destinationNameValue)

	httputils.Respond(writer, http.StatusNoContent)
}

// CreateCertificate mocks creation of certificate in both Destination Creator Service and Destination Service
func (h *Handler) CreateCertificate(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		respondWithHeader(ctx, writer, fmt.Sprintf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey)), http.StatusUnsupportedMediaType)
	}

	if r.Header.Get(clientUserHeaderKey) == "" {
		respondWithHeader(ctx, writer, fmt.Sprintf("The %q header could not be empty", clientUserHeaderKey), http.StatusBadRequest)
	}

	routeVars := mux.Vars(r)
	if err := h.validateDestinationCreatorPathParams(routeVars, false, false); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while reading destination certificate request body"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	var reqBody CertificateRequestBody
	if err = json.Unmarshal(bodyBytes, &reqBody); err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while unmarshalling destination certificate request body"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	log.C(ctx).Info("Validating destination certificate request body...")
	if err = reqBody.Validate(); err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while validating destination certificate request body"), "Invalid request body", correlationID, http.StatusBadRequest)
		return
	}

	if _, ok := h.DestinationCreatorSvcCertificates[reqBody.Name]; ok {
		log.C(ctx).Infof("Certificate with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		httputils.Respond(writer, http.StatusConflict)
		return
	}

	destinationCertName := reqBody.Name + destinationcreatorpkg.JavaKeyStoreFileExtension
	certResp := CertificateResponseBody{
		FileName:         destinationCertName,
		CommonName:       uuid.New().String(),
		CertificateChain: CertChain,
	}

	certRespBytes, err := json.Marshal(certResp)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while marshalling certificate response body"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	log.C(ctx).Infof("Destination certificate with name: %q added to the destination creator", reqBody.Name)
	h.DestinationCreatorSvcCertificates[reqBody.Name] = certRespBytes

	destSvcCertificateResp := destinationcreator.DestinationSvcCertificateResponse{
		Name:    destinationCertName,
		Content: CertChain,
	}

	destSvcCertificateRespBytes, err := json.Marshal(destSvcCertificateResp)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while marshalling destination certificate response"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	log.C(ctx).Infof("Destination certificate with name: %q added to the destination service", destinationCertName)
	h.DestinationSvcCertificates[destinationCertName] = destSvcCertificateRespBytes

	httputils.RespondWithBody(ctx, writer, http.StatusCreated, certResp)
}

// DeleteCertificate mocks deletion of certificate from both Destination Creator Service and Destination Service
func (h *Handler) DeleteCertificate(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		respondWithHeader(ctx, writer, fmt.Sprintf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey)), http.StatusUnsupportedMediaType)
	}

	if r.Header.Get(clientUserHeaderKey) == "" {
		respondWithHeader(ctx, writer, fmt.Sprintf("The %q header could not be empty", clientUserHeaderKey), http.StatusBadRequest)
	}

	routeVars := mux.Vars(r)
	if err := h.validateDestinationCreatorPathParams(routeVars, true, false); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}
	certNameValue := routeVars[h.Config.CertificateAPIConfig.CertificateNameParam]

	if _, isDestinationCreatorCertExist := h.DestinationCreatorSvcCertificates[certNameValue]; !isDestinationCreatorCertExist {
		log.C(ctx).Infof("Certificate with name: %q does not exists in the destination creator", certNameValue)
	} else {
		delete(h.DestinationCreatorSvcCertificates, certNameValue)
		log.C(ctx).Infof("Certificate with name: %q was deleted from the destination creator", certNameValue)
	}

	if _, isDestinationSvcCertExists := h.DestinationSvcCertificates[certNameValue+destinationcreatorpkg.JavaKeyStoreFileExtension]; !isDestinationSvcCertExists {
		log.C(ctx).Infof("Certificate with name: %q does not exists in the destination service. Returning 204 No Content...", certNameValue)
		httputils.Respond(writer, http.StatusNoContent)
	}
	delete(h.DestinationSvcCertificates, certNameValue+destinationcreatorpkg.JavaKeyStoreFileExtension)
	log.C(ctx).Infof("Certificate with name: %q was deleted from the destination service", certNameValue+destinationcreatorpkg.JavaKeyStoreFileExtension)

	httputils.Respond(writer, http.StatusNoContent)
}

func (h *Handler) createDesignTimeDestination(ctx context.Context, bodyBytes []byte) (int, error) {
	var reqBody DesignTimeDestRequestBody
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "An error occurred while unmarshalling design time destination request body")
	}

	log.C(ctx).Info("Validating design time destination request body...")
	if err := reqBody.Validate(); err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "An error occurred while validating design time destination request body")
	}

	if _, ok := h.DestinationCreatorSvcDestinations[reqBody.Name]; ok {
		log.C(ctx).Infof("Destination with name: %q already exists in the destination creator. Returning 409 Conflict...", reqBody.Name)
		return http.StatusConflict, nil
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination creator", reqBody.Name)
	h.DestinationCreatorSvcDestinations[reqBody.Name] = bodyBytes

	noAuthDest := destinationcreator.NoAuthenticationDestination{
		Name:           reqBody.Name,
		URL:            reqBody.URL,
		Type:           reqBody.Type,
		ProxyType:      reqBody.ProxyType,
		Authentication: reqBody.AuthenticationType,
	}

	noAuthDestBytes, err := json.Marshal(noAuthDest)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "An error occurred while marshalling no authentication destination")
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination service", reqBody.Name)
	h.DestinationSvcDestinations[reqBody.Name] = noAuthDestBytes

	return http.StatusCreated, nil
}

func (h *Handler) createBasicDestination(ctx context.Context, bodyBytes []byte) (int, error) {
	var reqBody BasicDestRequestBody
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "An error occurred while unmarshalling basic destination request body")
	}

	log.C(ctx).Info("Validating basic destination request body...")
	if err := reqBody.Validate(h.Config); err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "An error occurred while validating basic destination request body")
	}

	if _, ok := h.DestinationCreatorSvcDestinations[reqBody.Name]; ok {
		log.C(ctx).Infof("Destination with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		return http.StatusConflict, nil
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination creator", reqBody.Name)
	h.DestinationCreatorSvcDestinations[reqBody.Name] = bodyBytes

	basicAuthDest := destinationcreator.BasicDestination{
		NoAuthenticationDestination: destinationcreator.NoAuthenticationDestination{
			Name:           reqBody.Name,
			Type:           reqBody.Type,
			URL:            reqBody.URL,
			Authentication: reqBody.AuthenticationType,
			ProxyType:      reqBody.ProxyType,
		},
		User:     reqBody.User,
		Password: reqBody.Password,
	}

	basicDestBytes, err := json.Marshal(basicAuthDest)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "An error occurred while marshalling basic destination")
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination service", reqBody.Name)
	h.DestinationSvcDestinations[reqBody.Name] = basicDestBytes

	return http.StatusCreated, nil
}

func (h *Handler) createSAMLAssertionDestination(ctx context.Context, bodyBytes []byte) (int, error) {
	var reqBody SAMLAssertionDestRequestBody
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "An error occurred while unmarshalling SAML assertion destination request body")
	}

	log.C(ctx).Info("Validating SAML assertion destination request body...")
	if err := reqBody.Validate(h.Config); err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "An error occurred while validating SAML assertion destination request body")
	}

	if _, ok := h.DestinationCreatorSvcDestinations[reqBody.Name]; ok {
		log.C(ctx).Infof("Destination with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		return http.StatusConflict, nil
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination creator", reqBody.Name)
	h.DestinationCreatorSvcDestinations[reqBody.Name] = bodyBytes

	samlAssertionAuthDest := destinationcreator.SAMLAssertionDestination{
		NoAuthenticationDestination: destinationcreator.NoAuthenticationDestination{
			Name:           reqBody.Name,
			Type:           reqBody.Type,
			URL:            reqBody.URL,
			Authentication: reqBody.AuthenticationType,
			ProxyType:      reqBody.ProxyType,
		},
		Audience:         reqBody.Audience,
		KeyStoreLocation: reqBody.KeyStoreLocation,
	}

	samlAssertionAuthDestBytes, err := json.Marshal(samlAssertionAuthDest)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "An error occurred while marshalling SAML assertion destination")
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination service", reqBody.Name)
	h.DestinationSvcDestinations[reqBody.Name] = samlAssertionAuthDestBytes

	return http.StatusCreated, nil
}

func (h *Handler) createClientCertificateAuthDestination(ctx context.Context, bodyBytes []byte) (int, error) {
	var reqBody ClientCertificateAuthDestRequestBody
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "An error occurred while unmarshalling client certificate authentication destination request body")
	}

	log.C(ctx).Info("Validating client certificate authentication destination request body...")
	if err := reqBody.Validate(h.Config); err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "An error occurred while validating client certificate authentication destination request body")
	}

	if _, ok := h.DestinationCreatorSvcDestinations[reqBody.Name]; ok {
		log.C(ctx).Infof("Destination with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		return http.StatusConflict, nil
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination creator", reqBody.Name)
	h.DestinationCreatorSvcDestinations[reqBody.Name] = bodyBytes

	clientCertAuthDest := destinationcreator.ClientCertificateAuthenticationDestination{
		NoAuthenticationDestination: destinationcreator.NoAuthenticationDestination{
			Name:           reqBody.Name,
			Type:           reqBody.Type,
			URL:            reqBody.URL,
			Authentication: reqBody.AuthenticationType,
			ProxyType:      reqBody.ProxyType,
		},
		KeyStoreLocation: reqBody.KeyStoreLocation,
	}

	clientCertAuthDestBytes, err := json.Marshal(clientCertAuthDest)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "An error occurred while marshalling client certificate authentication destination")
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination service", reqBody.Name)
	h.DestinationSvcDestinations[reqBody.Name] = clientCertAuthDestBytes

	return http.StatusCreated, nil
}

// CleanupDestinationCertificates is "internal/technical" handler for deleting in-memory certificates mappings
func (h *Handler) CleanupDestinationCertificates(writer http.ResponseWriter, r *http.Request) {
	h.DestinationCreatorSvcCertificates = make(map[string]json.RawMessage)
	h.DestinationSvcCertificates = make(map[string]json.RawMessage)
	log.C(r.Context()).Infof("Destination creator certificates and destination service certificates mappings were successfully deleted")
	httputils.Respond(writer, http.StatusOK)
}

// CleanupDestinations is "internal/technical" handler for deleting in-memory destinations mappings
func (h *Handler) CleanupDestinations(writer http.ResponseWriter, r *http.Request) {
	h.DestinationCreatorSvcDestinations = make(map[string]json.RawMessage)
	h.DestinationSvcDestinations = make(map[string]json.RawMessage)
	log.C(r.Context()).Infof("Destination creator destinations and destination service destinations mappings were successfully deleted")
	httputils.Respond(writer, http.StatusOK)
}

// Destination Service handlers

// GetDestinationByNameFromDestinationSvc mocks getting a single destination by its name from Destination Service
func (h *Handler) GetDestinationByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	destinationNameParamValue, err := validateDestinationSvcPathParams(mux.Vars(r))
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	dest, exists := h.DestinationSvcDestinations[destinationNameParamValue]
	if !exists {
		err := errors.Errorf("Destination with name: %q doest not exists", destinationNameParamValue)
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusNotFound)
		return
	}

	bodyBytes, err := json.Marshal(dest)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred while marshalling destination with name: %s", destinationNameParamValue)
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, errMsg), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(bodyBytes)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while writing response"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}
}

// GetDestinationCertificateByNameFromDestinationSvc mocks getting a single certificate by its name from Destination Service
func (h *Handler) GetDestinationCertificateByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if err := validateAuthorization(ctx, r); err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	certificateNameParamValue, err := validateDestinationSvcPathParams(mux.Vars(r))
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	cert, exists := h.DestinationSvcCertificates[certificateNameParamValue]
	if !exists {
		err := errors.Errorf("Certificate with name: %q doest not exists", certificateNameParamValue)
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusNotFound)
		return
	}

	bodyBytes, err := json.Marshal(cert)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred while marshalling certificate with name: %s", certificateNameParamValue)
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, errMsg), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(bodyBytes)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while writing response"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}
}

func (h *Handler) validateDestinationCreatorPathParams(routeVars map[string]string, isDeleteRequest, isDestinationRequest bool) error {
	regionParam := h.Config.DestinationAPIConfig.RegionParam
	subaccountIDParam := h.Config.DestinationAPIConfig.SubaccountIDParam

	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]
	if regionParamValue == "" || subaccountIDParamValue == "" {
		return errors.Errorf("Missing required parameters: %q or/and %q", regionParam, subaccountIDParam)
	}

	if isDeleteRequest {
		errMsgPattern := "Missing required parameters: %q in case of %s request"
		if isDestinationRequest {
			destinationNameParamValue := h.Config.DestinationAPIConfig.DestinationNameParam
			if destinationNameValue := routeVars[destinationNameParamValue]; destinationNameValue == "" {
				return errors.Errorf(errMsgPattern, destinationNameParamValue, http.MethodDelete)
			}
		} else {
			certificateNameParamValue := h.Config.CertificateAPIConfig.CertificateNameParam
			if destinationNameValue := routeVars[certificateNameParamValue]; destinationNameValue == "" {
				return errors.Errorf(errMsgPattern, certificateNameParamValue, http.MethodDelete)
			}
		}
	}

	return nil
}

func validateDestinationSvcPathParams(routeVars map[string]string) (string, error) {
	nameParamValue := routeVars["name"]
	if nameParamValue == "" {
		return "", errors.New("Missing required parameters: \"name\"")
	}

	return nameParamValue, nil
}

func validateAuthorization(ctx context.Context, r *http.Request) error {
	log.C(ctx).Info("Validating authorization header...")
	authorizationHeaderValue := r.Header.Get(httphelpers.AuthorizationHeaderKey)

	if authorizationHeaderValue == "" {
		return errors.New("Missing authorization header")
	}

	tokenValue := strings.TrimPrefix(authorizationHeaderValue, "Bearer ")
	if tokenValue == "" {
		return errors.New("The token value cannot be empty")
	}

	return nil
}

func respondWithHeader(ctx context.Context, writer http.ResponseWriter, logErrMsg string, statusCode int) {
	log.C(ctx).Error(logErrMsg)
	writer.WriteHeader(statusCode)
	return
}
