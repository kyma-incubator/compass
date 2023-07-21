package destinationcreator

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"strings"
)

const (
	clientUserHeaderKey = "CLIENT_USER"
	certChain           = "e2e-test-destination-cert-mock-cert-chain"
)

// todo::: add detailed comments for the structure/fields/methods/etc..

type Handler struct {
	Config                            *Config
	DestinationCreatorSvcDestinations map[string]json.RawMessage
	DestinationCreatorSvcCertificates map[string]json.RawMessage
	DestinationSvcDestinations        map[string]json.RawMessage
	DestinationSvcCertificates        map[string]json.RawMessage
}

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

// CreateDestinations todo::: go doc
func (h *Handler) CreateDestinations(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		log.C(ctx).Errorf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey))
		writer.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if r.Header.Get(clientUserHeaderKey) == "" {
		log.C(ctx).Errorf("The %q header could not be empty", clientUserHeaderKey)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	routeVars := mux.Vars(r)
	if err := h.validateDestinationCreatorPathParams(routeVars, false, true); err != nil {
		log.C(ctx).Error(err)
		httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while reading destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while reading destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	authTypeResult := gjson.GetBytes(bodyBytes, "authenticationType")
	if !authTypeResult.Exists() || authTypeResult.String() == "" {
		log.C(ctx).Errorf("The authenticationType field is required and it should not be empty. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "The authenticationType field is required and it should not be empty. X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	switch destinationcreator.AuthType(authTypeResult.String()) {
	case destinationcreator.AuthTypeNoAuth:
		statusCode, err := h.createDesignTimeDestination(ctx, bodyBytes, routeVars)
		if err != nil {
			log.C(ctx).Error(err)
			httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), statusCode)
			return
		}
		httputils.Respond(writer, statusCode)
	case destinationcreator.AuthTypeBasic:
		statusCode, err := h.createBasicDestination(ctx, bodyBytes, routeVars)
		if err != nil {
			log.C(ctx).Error(err)
			httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), statusCode)
			return
		}
		httputils.Respond(writer, statusCode)
	case destinationcreator.AuthTypeSAMLAssertion:
		statusCode, err := h.createSAMLAssertionDestination(ctx, bodyBytes, routeVars)
		if err != nil {
			log.C(ctx).Error(err)
			httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), statusCode)
			return
		}
		httputils.Respond(writer, statusCode)
	default:
		log.C(ctx).Errorf("Invalid destination authentication type: %s. X-Request-Id: %s", authTypeResult.String(), correlationID)
		httphelpers.WriteError(writer, errors.Errorf("Invalid destination authentication type: %s. X-Request-Id: %s", authTypeResult.String(), correlationID), http.StatusInternalServerError)
		return
	}
}

// DeleteDestinations todo::: go doc
func (h *Handler) DeleteDestinations(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		log.C(ctx).Errorf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey))
		writer.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if r.Header.Get(clientUserHeaderKey) == "" {
		log.C(ctx).Errorf("The %q header could not be empty", clientUserHeaderKey)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	routeVars := mux.Vars(r)
	if err := h.validateDestinationCreatorPathParams(routeVars, true, true); err != nil {
		log.C(ctx).Error(err)
		httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), http.StatusBadRequest)
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

// CreateCertificate todo::: go doc
func (h *Handler) CreateCertificate(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		log.C(ctx).Errorf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey))
		writer.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if r.Header.Get(clientUserHeaderKey) == "" {
		log.C(ctx).Errorf("The %q header could not be empty", clientUserHeaderKey)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	routeVars := mux.Vars(r)
	if err := h.validateDestinationCreatorPathParams(routeVars, false, false); err != nil {
		log.C(ctx).Error(err)
		httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while reading destination certificate request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while reading destination certificate request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	var reqBody CertificateRequestBody
	if err = json.Unmarshal(bodyBytes, &reqBody); err != nil {
		log.C(ctx).WithError(err).Errorf("while unmarshaling destination certificate request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while unmarshaling destination certificate request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Info("Validating destination certificate request body...")
	if err = reqBody.Validate(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating destination certificate request body")
		httphelpers.WriteError(writer, errors.Errorf("An error occurred while validating destination certificate request body. X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	if _, ok := h.DestinationCreatorSvcCertificates[reqBody.Name]; ok {
		log.C(ctx).Infof("Certificate with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		httputils.Respond(writer, http.StatusConflict)
		return
	}

	destinationCertName := reqBody.Name + destinationcreator.JavaKeyStoreFileExtension
	certResp := CertificateResponseBody{
		FileName:         destinationCertName,
		CommonName:       reqBody.Name,
		CertificateChain: certChain,
	}

	certRespBytes, err := json.Marshal(certResp)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while marshaling certificate response body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while marshaling certificate response body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Infof("Destination certificate with name: %q added to the destination creator", reqBody.Name)
	h.DestinationCreatorSvcCertificates[reqBody.Name] = certRespBytes

	destSvcCertificateResp := destinationcreator.DestinationSvcCertificateResponse{
		Name:    destinationCertName,
		Content: certChain,
	}

	destSvcCertificateRespBytes, err := json.Marshal(destSvcCertificateResp)
	if err != nil {
		log.C(ctx).WithError(err).Error("while marshalling destination certificate response")
		httphelpers.WriteError(writer, errors.New("while marshalling destination certificate response"), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Infof("Destination certificate with name: %q added to the destination service", destinationCertName)
	h.DestinationSvcCertificates[destinationCertName] = destSvcCertificateRespBytes

	httputils.RespondWithBody(ctx, writer, http.StatusCreated, certResp)
}

// DeleteCertificate todo::: go doc
func (h *Handler) DeleteCertificate(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if r.Header.Get(httphelpers.ContentTypeHeaderKey) != httphelpers.ContentTypeApplicationJSON {
		log.C(ctx).Errorf("Unsupported media type, expected: %s got: %s", httphelpers.ContentTypeApplicationJSON, r.Header.Get(httphelpers.ContentTypeHeaderKey))
		writer.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if r.Header.Get(clientUserHeaderKey) == "" {
		log.C(ctx).Errorf("The %q header could not be empty", clientUserHeaderKey)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	routeVars := mux.Vars(r)
	if err := h.validateDestinationCreatorPathParams(routeVars, true, false); err != nil {
		log.C(ctx).Error(err)
		httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), http.StatusBadRequest)
		return
	}
	certNameValue := routeVars[h.Config.CertificateAPIConfig.CertificateNameParam]

	if _, isDestinationCreatorCertExist := h.DestinationCreatorSvcCertificates[certNameValue]; !isDestinationCreatorCertExist {
		log.C(ctx).Infof("Certificate with name: %q does not exists in the destination creator", certNameValue)
	} else {
		delete(h.DestinationCreatorSvcCertificates, certNameValue)
		log.C(ctx).Infof("Certificate with name: %q was deleted from the destination creator", certNameValue)
	}

	if _, isDestinationSvcCertExists := h.DestinationSvcCertificates[certNameValue+destinationcreator.JavaKeyStoreFileExtension]; !isDestinationSvcCertExists {
		log.C(ctx).Infof("Certificate with name: %q does not exists in the destination service. Returning 204 No Content...", certNameValue)
		httputils.Respond(writer, http.StatusNoContent)
	}
	delete(h.DestinationSvcCertificates, certNameValue+destinationcreator.JavaKeyStoreFileExtension)
	log.C(ctx).Infof("Certificate with name: %q was deleted from the destination service", certNameValue+destinationcreator.JavaKeyStoreFileExtension)

	httputils.Respond(writer, http.StatusNoContent)
}

func (h *Handler) createDesignTimeDestination(ctx context.Context, bodyBytes []byte, routeVars map[string]string) (int, error) {
	var reqBody DesignTimeRequestBody
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "while unmarshaling design time destination request body")
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
		Type:           reqBody.Type,
		URL:            reqBody.URL,
		Authentication: reqBody.AuthenticationType,
		ProxyType:      reqBody.ProxyType,
	}

	noAuthDestBytes, err := json.Marshal(noAuthDest)
	if err != nil {
		return http.StatusInternalServerError, errors.New("while marshalling no authentication destination")
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination service", reqBody.Name)
	h.DestinationSvcDestinations[reqBody.Name] = noAuthDestBytes

	return http.StatusCreated, nil
}

// CreateBasicDestination todo::: go doc
func (h *Handler) createBasicDestination(ctx context.Context, bodyBytes []byte, routeVars map[string]string) (int, error) {
	var reqBody BasicRequestBody
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, "while unmarshaling basic destination request body")
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
		return http.StatusInternalServerError, errors.New("An error occurred while marshalling basic destination")
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination service", reqBody.Name)
	h.DestinationSvcDestinations[reqBody.Name] = basicDestBytes

	return http.StatusCreated, nil
}

// CreateSAMLAssertionDestination todo::: go doc
func (h *Handler) createSAMLAssertionDestination(ctx context.Context, bodyBytes []byte, routeVars map[string]string) (int, error) {
	var reqBody SAMLAssertionRequestBody
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "while unmarshaling SAML assertion destination request body")
	}

	log.C(ctx).Info("Validating SAML assertion destination request body...")
	if err := reqBody.Validate(h.Config); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating SAML assertion destination request body")
		return http.StatusBadRequest, errors.Errorf("An error occurred while validating SAML assertion destination request body")
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
		return http.StatusInternalServerError, errors.New("while marshalling SAML assertion destination")
	}

	log.C(ctx).Infof("Destination with name: %q added to the destination service", reqBody.Name)
	h.DestinationSvcDestinations[reqBody.Name] = samlAssertionAuthDestBytes

	return http.StatusCreated, nil
}

// CleanupDestinationCertificates todo::: go doc
func (h *Handler) CleanupDestinationCertificates(writer http.ResponseWriter, r *http.Request) {
	h.DestinationCreatorSvcCertificates = make(map[string]json.RawMessage)
	h.DestinationSvcCertificates = make(map[string]json.RawMessage)
	log.C(r.Context()).Infof("Destination creator certificates and destination service certificates mappings were successfully deleted")
	httputils.Respond(writer, http.StatusOK)
}

// CleanupDestinations todo::: go doc
func (h *Handler) CleanupDestinations(writer http.ResponseWriter, r *http.Request) {
	h.DestinationCreatorSvcDestinations = make(map[string]json.RawMessage)
	h.DestinationSvcDestinations = make(map[string]json.RawMessage)
	log.C(r.Context()).Infof("Destination creator destinations and destination service destinations mappings were successfully deleted")
	httputils.Respond(writer, http.StatusOK)
}

// Destination Service handlers

// GetDestinationByNameFromDestinationSvc todo::: go doc
func (h *Handler) GetDestinationByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if err := validateAuthorization(ctx, r); err != nil {
		log.C(ctx).Error(err)
		httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), http.StatusBadRequest)
		return
	}

	destinationNameParamValue, err := validateDestinationSvcPathParams(mux.Vars(r))
	if err != nil {
		log.C(ctx).Error(err)
		httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), http.StatusBadRequest)
		return
	}

	dest, exists := h.DestinationSvcDestinations[destinationNameParamValue]
	if !exists {
		log.C(ctx).Errorf("Destination with name: %q doest not exists", destinationNameParamValue)
		httphelpers.WriteError(writer, errors.Errorf("Destination with name: %q doest not exists. X-Request-Id: %s", destinationNameParamValue, correlationID), http.StatusNotFound)
		return
	}

	bodyBytes, err := json.Marshal(dest)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(bodyBytes)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
		return
	}
}

// GetDestinationCertificateByNameFromDestinationSvc todo::: go doc
func (h *Handler) GetDestinationCertificateByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	if err := validateAuthorization(ctx, r); err != nil {
		log.C(ctx).Error(err)
		httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), http.StatusBadRequest)
		return
	}

	certificateNameParamValue, err := validateDestinationSvcPathParams(mux.Vars(r))
	if err != nil {
		log.C(ctx).Error(err)
		httphelpers.WriteError(writer, errors.Errorf("%s. X-Request-Id: %s", err.Error(), correlationID), http.StatusBadRequest)
		return
	}

	cert, exists := h.DestinationSvcCertificates[certificateNameParamValue]
	if !exists {
		log.C(ctx).Errorf("Certificate with name: %q doest not exists", certificateNameParamValue)
		httphelpers.WriteError(writer, errors.Errorf("Destination with name: %q doest not exists. X-Request-Id: %s", certificateNameParamValue, correlationID), http.StatusNotFound)
		return
	}

	bodyBytes, err := json.Marshal(cert)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(bodyBytes)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
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
		if isDestinationRequest {
			destinationNameParamValue := h.Config.DestinationAPIConfig.DestinationNameParam
			if destinationNameValue := routeVars[destinationNameParamValue]; destinationNameValue == "" {
				return errors.Errorf("Missing required parameters: %q in case of %s request", destinationNameParamValue, http.MethodDelete)
			}
		} else {
			certificateNameParamValue := h.Config.CertificateAPIConfig.CertificateNameParam
			if destinationNameValue := routeVars[certificateNameParamValue]; destinationNameValue == "" {
				return errors.Errorf("Missing required parameters: %q in case of %s request", certificateNameParamValue, http.MethodDelete)
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
		return errors.New("missing authorization header")
	}

	tokenValue := strings.TrimPrefix(authorizationHeaderValue, "Bearer ")
	if tokenValue == "" {
		return errors.New("the token value cannot be empty")
	}

	return nil
}
