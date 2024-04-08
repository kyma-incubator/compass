package destinationsvc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	clientUserHeaderKey     = "CLIENT_USER"
	CertChain               = "e2e-test-destination-cert-mock-cert-chain"
	pageCountQueryParameter = "$pageCount"
	pageQueryParameter      = "$page"
	pageSizeQueryParameter  = "$pageSize"
)

var (
	respErrorMsg               = "An unexpected error occurred while processing the request"
	UniqueEntityNameIdentifier = "name_%s_subacc_%s_instance_%s"
	NameIdentifier             = "name_%s"
)

// Handler is responsible to mock and handle any Destination Service requests
type Handler struct {
	Config                     *Config
	DestinationSvcDestinations map[string]destinationcreator.Destination
	DestinationSvcCertificates map[string]json.RawMessage
	DestinationsSensitive      map[string][]byte
}

// NewHandler creates a new Handler
func NewHandler(config *Config) *Handler {
	return &Handler{
		Config:                     config,
		DestinationSvcDestinations: make(map[string]destinationcreator.Destination),
		DestinationSvcCertificates: make(map[string]json.RawMessage),
		DestinationsSensitive:      make(map[string][]byte),
	}
}

// Destination Creator Service handlers + helper functions

// CreateDestinations mocks creation of all types of destinations in both Destination Creator Service and Destination Service
// using the APIs in the Destination Creator component
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

	destinationRequestBody, err := requestBodyToDestination(authTypeResult.String(), bodyBytes)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusInternalServerError)
		return
	}

	statusCode, err := h.createDestination(ctx, destinationRequestBody, subaccountIDParamValue, instanceIDParamValue)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, fmt.Sprintf("An unexpected error occurred while creating %s destination", destinationRequestBody.GetDestinationType()), correlationID, statusCode)
		return
	}
	httputils.Respond(writer, statusCode)
}

// DeleteDestinations mocks deletion of destinations from both Destination Creator Service and Destination Service
// using the APIs in the Destination Creator component
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
	delete(h.DestinationsSensitive, destinationIdentifier)
	log.C(ctx).Infof("Destination with name: %q and identifier: %q was deleted from the destination service", destinationNameValue, destinationIdentifier)

	httputils.Respond(writer, http.StatusNoContent)
}

// CreateCertificate mocks creation of certificate in both Destination Creator Service and Destination Service
// using the APIs in the Destination Creator component
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
// using the APIs in the Destination Creator component
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

func (h *Handler) createDestination(ctx context.Context, reqBody DestinationRequestBody, subaccountID, instanceID string) (int, error) {
	destinationTypeName := reqBody.GetDestinationType()

	log.C(ctx).Infof("Validating %s destination request body...", destinationTypeName)
	if err := reqBody.Validate(); err != nil {
		return http.StatusBadRequest, errors.Wrapf(err, "An error occurred while validating %s destination request body", destinationTypeName)
	}

	destinationIdentifier := reqBody.GetDestinationUniqueIdentifier(subaccountID, instanceID)
	if _, ok := h.DestinationSvcDestinations[destinationIdentifier]; ok {
		log.C(ctx).Infof("Destination with identifier: %q already exists. Returning 409 Conflict...", destinationIdentifier)
		return http.StatusConflict, nil
	}

	destination := reqBody.ToDestination()

	log.C(ctx).Infof("Destination with identifier: %q added to the destination service", destinationIdentifier)
	h.DestinationSvcDestinations[destinationIdentifier] = destination

	findAPIResponse, err := h.getSensitiveDataString(destination, subaccountID, instanceID, true)
	if err != nil {
		return http.StatusBadRequest, errors.Wrapf(err, "An error occurred while building sensitive data for destination %s", destinationTypeName)
	}

	h.DestinationsSensitive[destinationIdentifier] = []byte(findAPIResponse)

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
	h.DestinationSvcDestinations = make(map[string]destinationcreator.Destination)
	h.DestinationsSensitive = make(map[string][]byte)
	log.C(r.Context()).Infof("Destination creator destinations and destination service destinations mappings were successfully deleted")
	httputils.Respond(writer, http.StatusOK)
}

// PostDestination is an "internal/technical" handler for creating Destinations from E2E tests
func (h *Handler) PostDestination(writer http.ResponseWriter, req *http.Request) {
	isMultiStatusCodeEnabled := false
	ctx := req.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to read request body")
		http.Error(writer, "Missing body", http.StatusBadRequest)
		return
	}

	destinationName := gjson.GetBytes(data, "Name").String()
	authTypeResult := gjson.GetBytes(data, "Authentication")
	if !authTypeResult.Exists() || authTypeResult.String() == "" {
		err := errors.New("The authenticationType field in the request body is required and it should not be empty")
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusBadRequest)
		return
	}

	tokenValue, err := validateAuthorization(ctx, req)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	subaccountID, serviceInstanceID, err := extractSubaccountIDAndServiceInstanceIDFromDestinationToken(tokenValue)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusInternalServerError)
		return
	}
	log.C(ctx).Infof("Subaccount ID: %q and service instance ID: %q in the destination token", subaccountID, serviceInstanceID)

	var responses []PostResponse

	destinationRequestBody, err := requestBodyToDestination(authTypeResult.String(), data)
	if err != nil {
		responses = append(responses, PostResponse{destinationName, http.StatusInternalServerError, "Unable to unmarshall destination"})
		isMultiStatusCodeEnabled = true
	}

	destinationIdentifier := destinationRequestBody.GetDestinationUniqueIdentifier(subaccountID, serviceInstanceID)
	if _, ok := h.DestinationSvcDestinations[destinationIdentifier]; ok {
		log.C(ctx).Infof("Destination with identifier: %q already exists. Returning 409 Conflict...", destinationIdentifier)
		responses = append(responses, PostResponse{destinationName, http.StatusConflict, "Destination name already taken"})
		isMultiStatusCodeEnabled = true
	}

	destination := destinationRequestBody.ToDestination()

	findAPIResponse, err := h.getSensitiveDataString(destination, subaccountID, serviceInstanceID, true)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal response body")
		http.Error(writer, "An error occurred while building sensitive data for destination", http.StatusInternalServerError)
		return
	}

	h.DestinationSvcDestinations[destinationIdentifier] = destination
	h.DestinationsSensitive[destinationIdentifier] = []byte(findAPIResponse)

	log.C(ctx).Infof("Destination with identifier: %q added to the destination service", destinationIdentifier)

	responses = append(responses, PostResponse{destinationName, http.StatusCreated, ""})

	if !isMultiStatusCodeEnabled {
		writer.WriteHeader(http.StatusCreated)
		return
	}

	responseJSON, err := json.Marshal(responses)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal response body")
		http.Error(writer, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	if _, err := writer.Write(responseJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusMultiStatus)
}

// DeleteDestination is an "internal/technical" handler for deleting Destinations from E2E tests
func (h *Handler) DeleteDestination(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	destinationName := mux.Vars(req)["name"]

	if len(destinationName) == 0 {
		http.Error(writer, "Bad request - missing destination name", http.StatusBadRequest)
		return
	}

	tokenValue, err := validateAuthorization(ctx, req)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusUnauthorized)
		return
	}

	subaccountID, serviceInstanceID, err := extractSubaccountIDAndServiceInstanceIDFromDestinationToken(tokenValue)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusInternalServerError)
		return
	}
	log.C(ctx).Infof("Deleting Destination with subaccount ID: %q and service instance ID: %q in the destination token", subaccountID, serviceInstanceID)

	identifiers := map[string]string{
		h.Config.DestinationAPIConfig.SubaccountIDParam: subaccountID,
		h.Config.DestinationAPIConfig.InstanceIDParam:   serviceInstanceID,
	}
	destinationIdentifier := h.buildDestinationIdentifier(identifiers, destinationName)
	deleteResponse := DeleteResponse{Count: 0}

	if _, ok := h.DestinationsSensitive[destinationIdentifier]; !ok {
		deleteResponse.Summary = append(deleteResponse.Summary, DeleteStatus{
			Name:   destinationName,
			Status: "NOT_FOUND",
			Reason: "Could not find destination",
		})
	}

	delete(h.DestinationsSensitive, destinationIdentifier)
	delete(h.DestinationSvcDestinations, destinationIdentifier)
	deleteResponse.Count = deleteResponse.Count + 1

	deleteResponse.Summary = append(deleteResponse.Summary, DeleteStatus{
		Name:   destinationName,
		Status: "DELETED",
	})

	responseJSON, err := json.Marshal(deleteResponse)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal response body")
		http.Error(writer, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	if _, err := writer.Write(responseJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Destination Service handlers

// FindDestinationByNameFromDestinationSvc finds a destination by name
func (h *Handler) FindDestinationByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
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
		err := errors.Errorf("destination with name: %q and identifier: %q does not exists", destinationNameParamValue, destinationIdentifier)
		httphelpers.RespondWithError(ctx, writer, err, err.Error(), correlationID, http.StatusNotFound)
		return
	}
	log.C(ctx).Infof("Destination with name: %q and identifier: %q was found in the destination service", destinationNameParamValue, destinationIdentifier)

	findAPIResponse, err := h.buildFindAPIResponse(dest, r, subaccountID, serviceInstanceID)
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, err, respErrorMsg, correlationID, http.StatusInternalServerError)
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write([]byte(findAPIResponse))
	if err != nil {
		httphelpers.RespondWithError(ctx, writer, errors.Wrap(err, "An error occurred while writing response"), respErrorMsg, correlationID, http.StatusInternalServerError)
		return
	}
}

// GetSubaccountDestinationsPage gets a page of destinations
func (h *Handler) GetSubaccountDestinationsPage(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	pageRaw := req.URL.Query().Get(pageQueryParameter)

	destinations := h.destinationsMapToSlice()

	if pageRaw != "" {
		log.C(ctx).Infof("Page %s provided", pageRaw)
		pageNum, err := strconv.Atoi(pageRaw)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("could not convert page %s to int", pageRaw)
			http.Error(writer, "Invalid page number", http.StatusBadRequest)
			return
		}

		pageSizeRaw := req.URL.Query().Get(pageSizeQueryParameter)

		pageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("could not convert pageSize %s to int", pageSizeRaw)
			http.Error(writer, "Invalid page size", http.StatusBadRequest)
			return
		}

		destinationCount := len(h.DestinationSvcDestinations)
		if req.URL.Query().Get(pageCountQueryParameter) == "true" {
			pageCount := destinationCount / pageSize

			if destinationCount%pageSize != 0 {
				pageCount = pageCount + 1
			}

			writer.Header().Set("Page-Count", fmt.Sprintf("%d", pageCount))
		}

		if (pageNum-1)*pageSize > len(h.DestinationSvcDestinations) {
			destinations = []destinationcreator.Destination{}
		} else if pageNum*pageSize <= len(destinations) {
			destinations = destinations[((pageNum - 1) * pageSize):(pageNum * pageSize)]
		} else {
			destinations = destinations[((pageNum - 1) * pageSize):]
		}
	}

	if len(destinations) == 0 {
		if _, err := writer.Write([]byte("[]")); err != nil {
			log.C(ctx).WithError(err).Errorf("Failed to write data")
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	destinationsJSON, err := json.Marshal(destinations)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal destinations")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err := writer.Write(destinationsJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) buildFindAPIResponse(dest destinationcreator.Destination, r *http.Request, subaccountID, instanceID string) (string, error) {
	if dest.GetType() == destinationcreator.SAMLAssertionDestinationType {
		if usrHeader := r.Header.Get(httphelpers.UserTokenHeaderKey); usrHeader == "" {
			return "", errors.New("when calling destination svc find API for SAML Assertion destination, the `X-user-token` header should be provided")
		}
	}

	return h.getSensitiveDataString(dest, subaccountID, instanceID, false)
}

func (h *Handler) getSensitiveDataString(dest destinationcreator.Destination, subaccountID, instanceID string, skipCert bool) (string, error) {
	var findAPIResponse string
	switch dest.GetType() {
	case destinationcreator.DesignTimeDestinationType:
		designTimeDest, ok := dest.(*destinationcreator.NoAuthenticationDestination)
		if !ok {
			return "", errors.New("error while type asserting destination to NoAuth one")
		}
		findAPIResponse = fmt.Sprintf(FindAPINoAuthDestResponseTemplate, subaccountID, instanceID, designTimeDest.Name, designTimeDest.Type, designTimeDest.URL, designTimeDest.Authentication, designTimeDest.ProxyType)
	case destinationcreator.BasicAuthDestinationType:
		basicDest, ok := dest.(*destinationcreator.BasicDestination)
		if !ok {
			return "", errors.New("error while type asserting destination to Basic one")
		}
		findAPIResponse = fmt.Sprintf(FindAPIBasicDestResponseTemplate, subaccountID, instanceID, basicDest.Name, basicDest.Type, basicDest.URL, basicDest.Authentication, basicDest.ProxyType, basicDest.User, basicDest.Password)
	case destinationcreator.SAMLAssertionDestinationType:
		samlAssertionDest, ok := dest.(*destinationcreator.SAMLAssertionDestination)
		if !ok {
			return "", errors.New("error while type asserting destination to SAMLAssertion one")
		}

		var err error
		certResponseName := ""
		if !skipCert {
			certName := samlAssertionDest.KeyStoreLocation
			certResponseName, err = h.getCertificateName(certName, subaccountID, instanceID)
			if err != nil {
				return "", err
			}
		}

		findAPIResponse = fmt.Sprintf(FindAPISAMLAssertionDestResponseTemplate, subaccountID, instanceID, samlAssertionDest.Name, samlAssertionDest.Type, samlAssertionDest.URL, samlAssertionDest.Authentication, samlAssertionDest.ProxyType, samlAssertionDest.Audience, samlAssertionDest.KeyStoreLocation, certResponseName)
	case destinationcreator.ClientCertDestinationType:
		clientCertDest, ok := dest.(*destinationcreator.ClientCertificateAuthenticationDestination)
		if !ok {
			return "", errors.New("error while type asserting destination to ClientCertificate one")
		}

		var err error
		certResponseName := ""
		if !skipCert {
			certName := clientCertDest.KeyStoreLocation
			certResponseName, err = h.getCertificateName(certName, subaccountID, instanceID)
			if err != nil {
				return "", err
			}
		}

		findAPIResponse = fmt.Sprintf(FindAPIClientCertDestResponseTemplate, subaccountID, instanceID, clientCertDest.Name, clientCertDest.Type, clientCertDest.URL, clientCertDest.Authentication, clientCertDest.ProxyType, clientCertDest.KeyStoreLocation, certResponseName)
	case destinationcreator.OAuth2ClientCredentialsType:
		oauth2ClientCredsDest, ok := dest.(*destinationcreator.OAuth2ClientCredentialsDestination)
		if !ok {
			return "", errors.New("error while type asserting destination to OAuth2ClientCredentials one")
		}
		findAPIResponse = fmt.Sprintf(FindAPIOAuth2ClientCredsDestResponseTemplate, subaccountID, instanceID, oauth2ClientCredsDest.Name, oauth2ClientCredsDest.Type, oauth2ClientCredsDest.URL, oauth2ClientCredsDest.Authentication, oauth2ClientCredsDest.ProxyType, oauth2ClientCredsDest.ClientID, oauth2ClientCredsDest.ClientSecret, oauth2ClientCredsDest.TokenServiceURL)
	case destinationcreator.OAuth2mTLSType:
		oauth2mTLSDest, ok := dest.(*destinationcreator.OAuth2mTLSDestination)
		if !ok {
			return "", errors.New("error while type asserting destination to OAuth2mTLS one")
		}
		findAPIResponse = fmt.Sprintf(FindAPIOAuth2mTLSDestResponseTemplate,
			subaccountID,
			instanceID,
			oauth2mTLSDest.Name,
			oauth2mTLSDest.Type,
			oauth2mTLSDest.URL,
			oauth2mTLSDest.Authentication,
			oauth2mTLSDest.ProxyType,
			oauth2mTLSDest.TokenServiceURLType,
			oauth2mTLSDest.ClientID,
			oauth2mTLSDest.TokenServiceURL,
			oauth2mTLSDest.KeyStoreLocation,
		)
	}

	return findAPIResponse, nil
}

func (h *Handler) getCertificateName(certName, subaccountID, instanceID string) (string, error) {
	var certResponse destinationcreator.DestinationSvcCertificateResponse

	certIdentifier := fmt.Sprintf(UniqueEntityNameIdentifier, certName, subaccountID, instanceID)
	cert, exists := h.DestinationSvcCertificates[certIdentifier]
	if !exists {
		return "", errors.Errorf("certificate with name: %q and identifier: %q does not exists in the destination service", certName, certIdentifier)
	}

	err := json.Unmarshal(cert, &certResponse)
	if err != nil {
		return "", errors.Errorf("an error occurred while marshalling certificate with name: %q and identifier: %q", certName, certIdentifier)
	}
	return certResponse.Name, nil
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

func (h *Handler) destinationsMapToSlice() []destinationcreator.Destination {
	dest := make([]destinationcreator.Destination, 0)

	for name, destination := range h.DestinationSvcDestinations {
		empJSON1, err := json.MarshalIndent(destination, "", "  ")
		if err != nil {
			fmt.Println("err", err)
		}
		fmt.Printf("GetSubaccountDestinationsPage name: %s \n %s\n", name, string(empJSON1))

		dest = append(dest, destination)
	}

	return dest
}

// GetSensitiveData mocks getting a sensitive data by destination name from Destination Service
func (h *Handler) GetSensitiveData(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	destinationName := mux.Vars(req)["name"]

	if len(destinationName) == 0 {
		http.Error(writer, "Bad request", http.StatusBadRequest)
		return
	}

	var data []byte
	for name, sensitiveInfo := range h.DestinationsSensitive {
		if strings.Contains(name, GetDestinationPrefixNameIdentifier(destinationName)) {
			data = sensitiveInfo
			break
		}
	}

	log.C(ctx).Infof("Sending sensitive data of destination: %s", destinationName)

	if len(data) == 0 {
		http.Error(writer, "Not Found", http.StatusNotFound)
		return
	}

	if _, err := writer.Write(data); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

func requestBodyToDestination(authType string, bodyBytes []byte) (DestinationRequestBody, error) {
	var destinationRequestBody DestinationRequestBody
	switch destinationcreatorpkg.AuthType(authType) {
	case destinationcreatorpkg.AuthTypeNoAuth:
		destinationRequestBody = &DesignTimeDestRequestBody{}
	case destinationcreatorpkg.AuthTypeBasic:
		destinationRequestBody = &BasicDestRequestBody{}
	case destinationcreatorpkg.AuthTypeSAMLAssertion:
		destinationRequestBody = &SAMLAssertionDestRequestBody{}
	case destinationcreatorpkg.AuthTypeClientCertificate:
		destinationRequestBody = &ClientCertificateAuthDestRequestBody{}
	case destinationcreatorpkg.AuthTypeOAuth2ClientCredentials:
		destinationRequestBody = &OAuth2ClientCredsDestRequestBody{}
	default:
		return nil, errors.Errorf("The provided destination authentication type: %s is invalid", authType)
	}

	destinationTypeName := destinationRequestBody.GetDestinationType()
	if err := json.Unmarshal(bodyBytes, &destinationRequestBody); err != nil {
		return nil, errors.Wrapf(err, "An error occurred while unmarshalling %s destination request body", destinationTypeName)
	}

	fmt.Println("ALEX requestBodyToDestination")
	empJSON1, err := json.MarshalIndent(destinationRequestBody, "", "  ")
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Printf("requestBodyToDestination \n %s\n", string(empJSON1))

	return destinationRequestBody, nil
}
