package destinationcreator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	case destinationcreatorpkg.AuthTypeOAuth2ClientCredentials:
		destinationRequestBody = &OAuth2ClientCredsDestRequestBody{}

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
	log.C(r.Context()).Infof("Destination creator destinations and destination service destinations mappings were successfully deleted")
	httputils.Respond(writer, http.StatusOK)
}

// Destination Service handlers

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

func (h *Handler) buildFindAPIResponse(dest destinationcreator.Destination, r *http.Request, subaccountID, instanceID string) (string, error) {
	if dest.GetType() == destinationcreator.SAMLAssertionDestinationType {
		if usrHeader := r.Header.Get(httphelpers.UserTokenHeaderKey); usrHeader == "" {
			return "", errors.New("when calling destination svc find API for SAML Assertion destination, the `X-user-token` header should be provided")
		}
	}

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
		certName := samlAssertionDest.KeyStoreLocation
		certResponseName, err := h.getCertificateName(certName, subaccountID, instanceID)
		if err != nil {
			return "", err
		}
		findAPIResponse = fmt.Sprintf(FindAPISAMLAssertionDestResponseTemplate, subaccountID, instanceID, samlAssertionDest.Name, samlAssertionDest.Type, samlAssertionDest.URL, samlAssertionDest.Authentication, samlAssertionDest.ProxyType, samlAssertionDest.Audience, samlAssertionDest.KeyStoreLocation, certResponseName)
	case destinationcreator.ClientCertDestinationType:
		clientCertDest, ok := dest.(*destinationcreator.ClientCertificateAuthenticationDestination)
		if !ok {
			return "", errors.New("error while type asserting destination to ClientCertificate one")
		}
		certName := clientCertDest.KeyStoreLocation
		certResponseName, err := h.getCertificateName(certName, subaccountID, instanceID)
		if err != nil {
			return "", err
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

func (h *Handler) GetSensitiveData(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	destinationName := mux.Vars(req)["name"]

	if len(destinationName) == 0 {
		http.Error(writer, "Bad request", http.StatusBadRequest)
		return
	}

	log.C(ctx).Infof("Sending sensitive data of destination: %s", destinationName)
	data, ok := h.DestinationsSensitive[destinationName]

	if !ok {
		http.Error(writer, "Not Found", http.StatusNotFound)
		return
	}

	if _, err := writer.Write(data); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetSubaccountDestinationsPage(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	pageRaw := req.URL.Query().Get(pageQueryParameter)

	destinations := h.DestinationSvcDestinations
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
		//pageNum = 1
		//pageSize = 2

		destinationsArr := make([]Destination, 0)
		for _, destination := range destinations {
			destinationsArr = append(destinationsArr, destination)
		}

		if (pageNum-1)*pageSize > len(destinations) {
			destinations = []Destination{}
		} else if pageNum*pageSize <= len(destinations) {
			destinations = h.DestinationSvcDestinations[((pageNum - 1) * pageSize):(pageNum * pageSize)]
		} else {
			destinations = h.DestinationSvcDestinations[((pageNum - 1) * pageSize):]
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

func (h *Handler) deleteDestination(name string) {
	for k, v := range h.destinations {
		if v.Name == name {
			h.destinations[k] = h.destinations[len(h.destinations)-1]
			h.destinations = h.destinations[:len(h.destinations)-1]
			return
		}
	}
}

func (h *Handler) DeleteDestination(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	destinationName := mux.Vars(req)["name"]

	if len(destinationName) == 0 {
		http.Error(writer, "Bad request - missing destination name", http.StatusBadRequest)
		return
	}

	deleteResponse := DeleteResponse{Count: 0}

	if _, ok := h.DestinationsSensitive[destinationName]; !ok {
		deleteResponse.Summary = append(deleteResponse.Summary, DeleteStatus{
			Name:   destinationName,
			Status: "NOT_FOUND",
			Reason: "Could not find destination",
		})
	}

	delete(h.DestinationsSensitive, destinationName)
	deleteResponse.Count = deleteResponse.Count + 1

	deleteResponse.Summary = append(deleteResponse.Summary, DeleteStatus{
		Name:   destinationName,
		Status: "DELETED",
	})

	h.deleteDestination(destinationName)

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

func validDestinationType(destinationType string) bool {
	return destinationType == "HTTP" || destinationType == "RFC" || destinationType == "MAIL" || destinationType == "LDAP"
}

func (h *Handler) PostDestination(writer http.ResponseWriter, req *http.Request) {
	isMultiStatusCodeEnabled := false
	var destinations []Destination = make([]Destination, 1)
	ctx := req.Context()

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to read request body")
		http.Error(writer, "Missing body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(data, &destinations)
	if err != nil {
		err = json.Unmarshal(data, &destinations[0])
		if err != nil {
			log.C(ctx).WithError(err).Error("Failed to unmarshal request body")
			http.Error(writer, "Invalid json", http.StatusBadRequest)
			return
		}
	}

	var responses []PostResponse
	for _, destination := range destinations {
		if _, ok := h.DestinationsSensitive[destination.Name]; ok {
			responses = append(responses, PostResponse{destination.Name, http.StatusConflict, "Destination name already taken"})
			isMultiStatusCodeEnabled = true
		} else if !validDestinationType(destination.Type) {
			responses = append(responses, PostResponse{destination.Name, http.StatusBadRequest, "Invalid destination type"})
			isMultiStatusCodeEnabled = true
		} else {
			h.destinations = append(h.destinations, destination)
			h.DestinationsSensitive[destination.Name] = []byte(fmt.Sprintf(sensitiveDataTemplate, uuid.NewString(),
				destination.Name, destination.Type, destination.Name))

			responses = append(responses, PostResponse{destination.Name, http.StatusCreated, ""})
		}
	}

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
