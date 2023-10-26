package destinationcreator

import (
	"context"
	"encoding/base64"
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

var (
	respErrorMsg               = "An unexpected error occurred while processing the request"
	UniqueEntityNameIdentifier = "name_%s_subacc_%s_instance_%s"
)

// Handler is responsible to mock and handle any Destination Service requests
type Handler struct {
	Config                     *Config
	DestinationSvcDestinations map[string]json.RawMessage
	DestinationSvcCertificates map[string]json.RawMessage
}

// NewHandler creates a new Handler
func NewHandler(config *Config) *Handler {
	return &Handler{
		Config:                     config,
		DestinationSvcDestinations: make(map[string]json.RawMessage),
		DestinationSvcCertificates: make(map[string]json.RawMessage),
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

	subaccountIDParamValue := routeVars[h.Config.DestinationAPIConfig.SubaccountIDParam]
	instanceIDParamValue := routeVars[h.Config.DestinationAPIConfig.InstanceIDParam]

	var destinationRequestBody DestinationRequestBody
	switch destinationcreatorpkg.AuthType(authTypeResult.String()) {
	case destinationcreatorpkg.AuthTypeNoAuth:
		destinationRequestBody = &DesignTimeDestRequestBody{}
	case destinationcreatorpkg.AuthTypeBasic:
		destinationRequestBody = &BasicDestRequestBody{}
	case destinationcreatorpkg.AuthTypeSAMLAssertion:
		destinationRequestBody = &SAMLAssertionDestRequestBody{}
	case destinationcreatorpkg.AuthTypeClientCertificate:
		destinationRequestBody = &ClientCertificateAuthDestRequestBody{}
	default:
		err := errors.Errorf("The provided destination authentication type: %s is invalid", authTypeResult.String())
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusInternalServerError)
		return
	}

	statusCode, err := h.createDestination(ctx, bodyBytes, destinationRequestBody, subaccountIDParamValue, instanceIDParamValue)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, fmt.Sprintf("An unexpected error occurred while creating %s destination", destinationRequestBody.GetDestinationType()), correlationID, statusCode)
		return
	}
	httputils.Respond(writer, statusCode)
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
	destinationIdentifier := h.buildDestinationIdentifier(routeVars, destinationNameValue)
	_, isDestinationSvcDestExists := h.DestinationSvcDestinations[destinationIdentifier]
	if !isDestinationSvcDestExists {
		log.C(ctx).Infof("Destination with name: %q and identifier: %q does not exists in the destination service. Returning 204 No Content...", destinationNameValue, destinationIdentifier)
		httputils.Respond(writer, http.StatusNoContent)
		return
	}
	delete(h.DestinationSvcDestinations, destinationIdentifier)
	log.C(ctx).Infof("Destination with name: %q and identifier: %q was deleted from the destination service", destinationNameValue, destinationIdentifier)

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

	certName := reqBody.Name + destinationcreatorpkg.JavaKeyStoreFileExtension
	certificateIdentifier := h.buildDestinationCertificateIdentifier(routeVars, certName)
	if _, ok := h.DestinationSvcCertificates[certificateIdentifier]; ok {
		log.C(ctx).Infof("Certificate with name: %q and identifier: %q already exists. Returning 409 Conflict...", certName, certificateIdentifier)
		httputils.Respond(writer, http.StatusConflict)
		return
	}

	certResp := CertificateResponseBody{
		FileName:         certName,
		CommonName:       uuid.New().String(),
		CertificateChain: CertChain,
	}

	destSvcCertificateResp := destinationcreator.DestinationSvcCertificateResponse{
		Name:    certName,
		Content: CertChain,
	}

	destSvcCertificateRespBytes, err := json.Marshal(destSvcCertificateResp)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while marshalling destination certificate response"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}

	log.C(ctx).Infof("Destination certificate with name: %q and identifier: %q added to the destination service", certName, certificateIdentifier)
	h.DestinationSvcCertificates[certificateIdentifier] = destSvcCertificateRespBytes

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
	certNameParamValue := routeVars[h.Config.CertificateAPIConfig.CertificateNameParam]
	certName := certNameParamValue + destinationcreatorpkg.JavaKeyStoreFileExtension

	certificateIdentifier := h.buildDestinationCertificateIdentifier(routeVars, certName)
	if _, isDestinationSvcCertExists := h.DestinationSvcCertificates[certificateIdentifier]; !isDestinationSvcCertExists {
		log.C(ctx).Infof("Certificate with name: %q and identifier: %q does not exists in the destination service. Returning 204 No Content...", certName, certificateIdentifier)
		httputils.Respond(writer, http.StatusNoContent)
		return
	}
	delete(h.DestinationSvcCertificates, certificateIdentifier)
	log.C(ctx).Infof("Certificate with name: %q and identifier: %q was deleted from the destination service", certName, certificateIdentifier)

	httputils.Respond(writer, http.StatusNoContent)
}

func (h *Handler) createDestination(ctx context.Context, bodyBytes []byte, reqBody DestinationRequestBody, subaccountID, instanceID string) (int, error) {
	destinationTypeName := reqBody.GetDestinationType()
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "An error occurred while unmarshalling %s destination request body", destinationTypeName)
	}

	log.C(ctx).Infof("Validating %s destination request body...", destinationTypeName)
	if err := reqBody.Validate(h.Config); err != nil {
		return http.StatusBadRequest, errors.Wrapf(err, "An error occurred while validating %s destination request body", destinationTypeName)
	}

	destinationIdentifier := reqBody.GetDestinationUniqueIdentifier(subaccountID, instanceID)
	if _, ok := h.DestinationSvcDestinations[destinationIdentifier]; ok {
		log.C(ctx).Infof("Destination with identifier: %q already exists. Returning 409 Conflict...", destinationIdentifier)
		return http.StatusConflict, nil
	}

	destinationRawMessage, err := reqBody.ToDestination()
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "An error occurred while marshalling %s destination", destinationTypeName)
	}

	log.C(ctx).Infof("Destination with identifier: %q added to the destination service", destinationIdentifier)
	h.DestinationSvcDestinations[destinationIdentifier] = destinationRawMessage

	return http.StatusCreated, nil
}

// CleanupDestinationCertificates is "internal/technical" handler for deleting in-memory certificates mappings
func (h *Handler) CleanupDestinationCertificates(writer http.ResponseWriter, r *http.Request) {
	h.DestinationSvcCertificates = make(map[string]json.RawMessage)
	log.C(r.Context()).Infof("Destination creator certificates and destination service certificates mappings were successfully deleted")
	httputils.Respond(writer, http.StatusOK)
}

// CleanupDestinations is "internal/technical" handler for deleting in-memory destinations mappings
func (h *Handler) CleanupDestinations(writer http.ResponseWriter, r *http.Request) {
	h.DestinationSvcDestinations = make(map[string]json.RawMessage)
	log.C(r.Context()).Infof("Destination creator destinations and destination service destinations mappings were successfully deleted")
	httputils.Respond(writer, http.StatusOK)
}

// Destination Service handlers

// GetDestinationByNameFromDestinationSvc mocks getting a single destination by its name from Destination Service
func (h *Handler) GetDestinationByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	tokenValue, err := validateAuthorization(ctx, r)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	destinationNameParamValue, err := validateDestinationSvcPathParams(mux.Vars(r))
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	subaccountID, serviceInstanceID, err := extractSubaccountIDAndServiceInstanceIDFromDestinationToken(tokenValue)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusInternalServerError)
		return
	}
	log.C(ctx).Infof("Subaccount ID: %q and service instance ID: %q in the destination token", subaccountID, serviceInstanceID)

	destinationIdentifier := fmt.Sprintf(UniqueEntityNameIdentifier, destinationNameParamValue, subaccountID, serviceInstanceID)
	dest, exists := h.DestinationSvcDestinations[destinationIdentifier]
	if !exists {
		err := errors.Errorf("Destination with name: %q and identifier: %q does not exists", destinationNameParamValue, destinationIdentifier)
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusNotFound)
		return
	}
	log.C(ctx).Infof("Destination with name: %q and identifier: %q was found in the destination service", destinationNameParamValue, destinationIdentifier)

	bodyBytes, err := json.Marshal(dest)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred while marshalling destination with name: %q and identifier: %q", destinationNameParamValue, destinationIdentifier)
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

	tokenValue, err := validateAuthorization(ctx, r)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	certificateNameParamValue, err := validateDestinationSvcPathParams(mux.Vars(r))
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	subaccountID, serviceInstanceID, err := extractSubaccountIDAndServiceInstanceIDFromDestinationToken(tokenValue)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusInternalServerError)
		return
	}
	log.C(ctx).Infof("Subaccount ID: %q and service instance ID: %q in the destination token", subaccountID, serviceInstanceID)

	certificateIdentifier := fmt.Sprintf(UniqueEntityNameIdentifier, certificateNameParamValue, subaccountID, serviceInstanceID)
	cert, exists := h.DestinationSvcCertificates[certificateIdentifier]
	if !exists {
		err := errors.Errorf("Certificate with name: %q and identifier: %q does not exists in the destination service", certificateNameParamValue, certificateIdentifier)
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusNotFound)
		return
	}
	log.C(ctx).Infof("Destination certificate with name: %q and identifier: %q was found in the destination service", certificateNameParamValue, certificateIdentifier)

	bodyBytes, err := json.Marshal(cert)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred while marshalling certificate with name: %q and identifier: %q", certificateNameParamValue, certificateIdentifier)
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

func (h *Handler) buildDestinationCertificateIdentifier(routeVars map[string]string, certName string) string {
	subaccountIDParamValue := routeVars[h.Config.CertificateAPIConfig.SubaccountIDParam]
	instanceIDParamValue := routeVars[h.Config.CertificateAPIConfig.InstanceIDParam]
	return fmt.Sprintf(UniqueEntityNameIdentifier, certName, subaccountIDParamValue, instanceIDParamValue)
}

func (h *Handler) buildDestinationIdentifier(routeVars map[string]string, destinationName string) string {
	subaccountIDParamValue := routeVars[h.Config.DestinationAPIConfig.SubaccountIDParam]
	instanceIDParamValue := routeVars[h.Config.DestinationAPIConfig.InstanceIDParam]
	return fmt.Sprintf(UniqueEntityNameIdentifier, destinationName, subaccountIDParamValue, instanceIDParamValue)
}

func (h *Handler) validateDestinationCreatorPathParams(routeVars map[string]string, isDeleteRequest, isDestinationRequest bool) error {
	var regionParam, subaccountIDParam string
	if isDestinationRequest {
		regionParam = h.Config.DestinationAPIConfig.RegionParam
		subaccountIDParam = h.Config.DestinationAPIConfig.SubaccountIDParam
	} else {
		regionParam = h.Config.CertificateAPIConfig.RegionParam
		subaccountIDParam = h.Config.CertificateAPIConfig.SubaccountIDParam
	}

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

func validateAuthorization(ctx context.Context, r *http.Request) (string, error) {
	log.C(ctx).Info("Validating authorization header...")
	authorizationHeaderValue := r.Header.Get(httphelpers.AuthorizationHeaderKey)

	if authorizationHeaderValue == "" {
		return "", errors.New("Missing authorization header")
	}

	tokenValue := strings.TrimSpace(strings.TrimPrefix(authorizationHeaderValue, "Bearer "))
	if tokenValue == "" {
		return "", errors.New("The token value cannot be empty")
	}

	return tokenValue, nil
}

func extractSubaccountIDAndServiceInstanceIDFromDestinationToken(token string) (string, string, error) {
	// JWT format: <header>.<payload>.<signature>
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		return "", "", errors.New("invalid JWT token format")
	}
	payload := tokenParts[1]

	decodedToken, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", "", errors.Wrapf(err, "An error occurred while decoding JWT token payload")
	}

	data := &struct {
		ExternalAttributes struct {
			SubaccountID      string `json:"subaccountid"`
			ServiceInstanceID string `json:"serviceinstanceid"`
		} `json:"ext_attr"`
	}{}
	if err := json.Unmarshal(decodedToken, data); err != nil {
		return "", "", errors.Wrapf(err, "while unmarhalling destination JWT token")
	}

	if data.ExternalAttributes.SubaccountID == "" {
		return "", "", errors.Errorf("The subaccount ID claim in the token could not be empty")
	}

	return data.ExternalAttributes.SubaccountID, data.ExternalAttributes.ServiceInstanceID, nil
}

func respondWithHeader(ctx context.Context, writer http.ResponseWriter, logErrMsg string, statusCode int) {
	log.C(ctx).Error(logErrMsg)
	writer.WriteHeader(statusCode)
	return
}
