package destinationcreator

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strings"
)

const (
	clientUserHeaderKey = "CLIENT_USER"
	certCommonName      = "e2e-test-mock-destination-cert-common-name"
	certChain           = "e2e-test-mock-cert-chain"
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

// todo::: extract the common logic into func

// Destination Creator Service handlers

// CreateDesignTimeDestination todo::: go doc
func (h *Handler) CreateDesignTimeDestination(writer http.ResponseWriter, r *http.Request) {
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while reading design time destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while reading design time destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	var reqBody DesignTimeRequestBody
	if err = json.Unmarshal(bodyBytes, &reqBody); err != nil {
		log.C(ctx).WithError(err).Errorf("while unmarshaling design time destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while unmarshaling design time destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Info("Validating design time destination request body...")
	if err = reqBody.Validate(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating design time destination request body")
		httphelpers.WriteError(writer, errors.Errorf("An error occurred while validating design time destination request body. X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	regionParam := h.Config.DestinationAPIConfig.RegionParam
	subaccountIDParam := h.Config.DestinationAPIConfig.SubaccountIDParam

	routeVars := mux.Vars(r)
	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]

	if regionParamValue == "" || subaccountIDParamValue == "" {
		log.C(ctx).Errorf("Missing required parameters: %q or/and %q", regionParam, subaccountIDParam)
		httphelpers.WriteError(writer, errors.Errorf("Not all of the required parameters - %q or/and %q are provided. X-Request-Id: %s", regionParam, subaccountIDParam, correlationID), http.StatusBadRequest)
		return
	}

	_, ok := h.DestinationCreatorSvcDestinations[reqBody.Name]
	if ok {
		log.C(ctx).Infof("Destination with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		httputils.Respond(writer, http.StatusConflict)
	}

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
		log.C(ctx).WithError(err).Error("while marshalling no authentication destination")
		httphelpers.WriteError(writer, errors.New("while marshalling no authentication destination"), http.StatusInternalServerError)
		return
	}

	h.DestinationSvcDestinations[reqBody.Name] = noAuthDestBytes

	httputils.Respond(writer, http.StatusCreated)
}

// DeleteDesignTimeDestination todo::: go doc
func (h *Handler) DeleteDesignTimeDestination(writer http.ResponseWriter, r *http.Request) {
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while reading design time destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while reading design time destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	var reqBody DesignTimeRequestBody
	if err = json.Unmarshal(bodyBytes, &reqBody); err != nil {
		log.C(ctx).WithError(err).Errorf("while unmarshaling design time destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while unmarshaling design time destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Info("Validating design time destination request body...")
	if err = reqBody.Validate(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating design time destination request body")
		httphelpers.WriteError(writer, errors.Errorf("An error occurred while validating design time destination request body. X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	regionParam := h.Config.DestinationAPIConfig.RegionParam
	subaccountIDParam := h.Config.DestinationAPIConfig.SubaccountIDParam
	destinationNameParam := h.Config.DestinationAPIConfig.DestinationNameParam

	routeVars := mux.Vars(r)
	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]
	destinationNameParamValue := routeVars[destinationNameParam]

	if regionParamValue == "" || subaccountIDParamValue == "" || destinationNameParamValue == "" {
		log.C(ctx).Errorf("Missing required parameters: %q, %q or/and %q", regionParam, subaccountIDParam, destinationNameParam)
		httphelpers.WriteError(writer, errors.Errorf("Not all of the required parameters - %q, %q or/and %q are provided. X-Request-Id: %s", regionParam, subaccountIDParam, destinationNameParam, correlationID), http.StatusBadRequest)
		return
	}

	_, isDestinationCreatorDestExist := h.DestinationCreatorSvcDestinations[reqBody.Name]
	if !isDestinationCreatorDestExist {
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q does not exists in the destination creator", reqBody.Name, reqBody.AuthenticationType)
	} else {
		delete(h.DestinationCreatorSvcDestinations, reqBody.Name)
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q was deleted from the destination creator", reqBody.Name, reqBody.AuthenticationType)
	}

	_, isDestinationSvcDestExists := h.DestinationSvcDestinations[reqBody.Name]
	if !isDestinationSvcDestExists {
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q does not exists in the destination service. Returning 204 No Content...", reqBody.Name, reqBody.AuthenticationType)
		httputils.Respond(writer, http.StatusNoContent)
	}
	delete(h.DestinationSvcDestinations, reqBody.Name)
	log.C(ctx).Infof("Destination with name: %q and authentication type: %q was deleted from the destination service", reqBody.Name, reqBody.AuthenticationType)

	httputils.Respond(writer, http.StatusNoContent)
}

// CreateBasicDestination todo::: go doc
func (h *Handler) CreateBasicDestination(writer http.ResponseWriter, r *http.Request) {
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while reading basic destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while reading basic destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	var reqBody BasicRequestBody
	if err = json.Unmarshal(bodyBytes, &reqBody); err != nil {
		log.C(ctx).WithError(err).Errorf("while unmarshaling basic destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while unmarshaling basic destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Info("Validating basic destination request body...")
	if err = reqBody.Validate(h.Config); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating basic destination request body")
		httphelpers.WriteError(writer, errors.Errorf("An error occurred while validating basic destination request body. X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	regionParam := h.Config.DestinationAPIConfig.RegionParam
	subaccountIDParam := h.Config.DestinationAPIConfig.SubaccountIDParam

	routeVars := mux.Vars(r)
	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]

	if regionParamValue == "" || subaccountIDParamValue == "" {
		log.C(ctx).Errorf("Missing required parameters: %q or/and %q", regionParam, subaccountIDParam)
		httphelpers.WriteError(writer, errors.Errorf("Not all of the required parameters - %q or/and %q are provided. X-Request-Id: %s", regionParam, subaccountIDParam, correlationID), http.StatusBadRequest)
		return
	}

	_, ok := h.DestinationCreatorSvcDestinations[reqBody.Name]
	if ok {
		log.C(ctx).Infof("Destination with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		httputils.Respond(writer, http.StatusConflict)
	}

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
		log.C(ctx).WithError(err).Error("while marshalling basic destination")
		httphelpers.WriteError(writer, errors.New("while marshalling basic destination"), http.StatusInternalServerError)
		return
	}

	h.DestinationSvcDestinations[reqBody.Name] = basicDestBytes

	httputils.Respond(writer, http.StatusCreated)
}

// DeleteBasicDestination todo::: go doc
func (h *Handler) DeleteBasicDestination(writer http.ResponseWriter, r *http.Request) {
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while reading basic destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while reading basic destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	var reqBody BasicRequestBody
	if err = json.Unmarshal(bodyBytes, &reqBody); err != nil {
		log.C(ctx).WithError(err).Errorf("while unmarshaling basic destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while unmarshaling basic destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Info("Validating basic destination request body...")
	if err = reqBody.Validate(h.Config); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating basic destination request body")
		httphelpers.WriteError(writer, errors.Errorf("An error occurred while validating basic destination request body. X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	regionParam := h.Config.DestinationAPIConfig.RegionParam
	subaccountIDParam := h.Config.DestinationAPIConfig.SubaccountIDParam
	destinationNameParam := h.Config.DestinationAPIConfig.DestinationNameParam

	routeVars := mux.Vars(r)
	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]
	destinationNameParamValue := routeVars[destinationNameParam]

	if regionParamValue == "" || subaccountIDParamValue == "" || destinationNameParamValue == "" {
		log.C(ctx).Errorf("Missing required parameters: %q, %q or/and %q", regionParam, subaccountIDParam, destinationNameParam)
		httphelpers.WriteError(writer, errors.Errorf("Not all of the required parameters - %q, %q or/and %q are provided. X-Request-Id: %s", regionParam, subaccountIDParam, destinationNameParam, correlationID), http.StatusBadRequest)
		return
	}

	_, isDestinationCreatorDestExist := h.DestinationCreatorSvcDestinations[reqBody.Name]
	if !isDestinationCreatorDestExist {
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q does not exists in the destination creator", reqBody.Name, reqBody.AuthenticationType)
	} else {
		delete(h.DestinationCreatorSvcDestinations, reqBody.Name)
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q was deleted from the destination creator", reqBody.Name, reqBody.AuthenticationType)
	}

	_, isDestinationSvcDestExists := h.DestinationSvcDestinations[reqBody.Name]
	if !isDestinationSvcDestExists {
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q does not exists in the destination service. Returning 204 No Content...", reqBody.Name, reqBody.AuthenticationType)
		httputils.Respond(writer, http.StatusNoContent)
	}
	delete(h.DestinationSvcDestinations, reqBody.Name)
	log.C(ctx).Infof("Destination with name: %q and authentication type: %q was deleted from the destination service", reqBody.Name, reqBody.AuthenticationType)

	httputils.Respond(writer, http.StatusNoContent)
}

// CreateSAMLAssertionDestination todo::: go doc
func (h *Handler) CreateSAMLAssertionDestination(writer http.ResponseWriter, r *http.Request) {
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while reading SAML assertion destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while reading SAML assertion destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	var reqBody SAMLAssertionRequestBody
	if err = json.Unmarshal(bodyBytes, &reqBody); err != nil {
		log.C(ctx).WithError(err).Errorf("while unmarshaling SAML assertion destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while unmarshaling SAML assertion destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Info("Validating SAML assertion destination request body...")
	if err = reqBody.Validate(h.Config); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating SAML assertion destination request body")
		httphelpers.WriteError(writer, errors.Errorf("An error occurred while validating SAML assertion destination request body. X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	regionParam := h.Config.DestinationAPIConfig.RegionParam
	subaccountIDParam := h.Config.DestinationAPIConfig.SubaccountIDParam

	routeVars := mux.Vars(r)
	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]

	if regionParamValue == "" || subaccountIDParamValue == "" {
		log.C(ctx).Errorf("Missing required parameters: %q or/and %q", regionParam, subaccountIDParam)
		httphelpers.WriteError(writer, errors.Errorf("Not all of the required parameters - %q or/and %q are provided. X-Request-Id: %s", regionParam, subaccountIDParam, correlationID), http.StatusBadRequest)
		return
	}

	_, ok := h.DestinationCreatorSvcDestinations[reqBody.Name]
	if ok {
		log.C(ctx).Infof("Destination with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		httputils.Respond(writer, http.StatusConflict)
	}

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
		log.C(ctx).WithError(err).Error("while marshalling SAML assertion destination")
		httphelpers.WriteError(writer, errors.New("while marshalling SAML assertion destination"), http.StatusInternalServerError)
		return
	}

	h.DestinationSvcDestinations[reqBody.Name] = samlAssertionAuthDestBytes

	httputils.Respond(writer, http.StatusCreated)
}

// DeleteSAMLAssertionDestination todo::: go doc
func (h *Handler) DeleteSAMLAssertionDestination(writer http.ResponseWriter, r *http.Request) {
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while reading SAML assertion destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while reading SAML assertion destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	var reqBody SAMLAssertionRequestBody
	if err = json.Unmarshal(bodyBytes, &reqBody); err != nil {
		log.C(ctx).WithError(err).Errorf("while unmarshaling SAML assertion destination request body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while unmarshaling SAML assertion destination request body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	log.C(ctx).Info("Validating SAML assertion destination request body...")
	if err = reqBody.Validate(h.Config); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while validating SAML assertion destination request body")
		httphelpers.WriteError(writer, errors.Errorf("An error occurred while validating SAML assertion destination request body. X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	regionParam := h.Config.DestinationAPIConfig.RegionParam
	subaccountIDParam := h.Config.DestinationAPIConfig.SubaccountIDParam
	destinationNameParam := h.Config.DestinationAPIConfig.DestinationNameParam

	routeVars := mux.Vars(r)
	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]
	destinationNameParamValue := routeVars[destinationNameParam]

	if regionParamValue == "" || subaccountIDParamValue == "" || destinationNameParamValue == "" {
		log.C(ctx).Errorf("Missing required parameters: %q, %q or/and %q", regionParam, subaccountIDParam, destinationNameParam)
		httphelpers.WriteError(writer, errors.Errorf("Not all of the required parameters - %q, %q or/and %q are provided. X-Request-Id: %s", regionParam, subaccountIDParam, destinationNameParam, correlationID), http.StatusBadRequest)
		return
	}

	_, isDestinationCreatorDestExist := h.DestinationCreatorSvcDestinations[reqBody.Name]
	if !isDestinationCreatorDestExist {
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q does not exists in the destination creator", reqBody.Name, reqBody.AuthenticationType)
	} else {
		delete(h.DestinationCreatorSvcDestinations, reqBody.Name)
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q was deleted from the destination creator", reqBody.Name, reqBody.AuthenticationType)
	}

	_, isDestinationSvcDestExists := h.DestinationSvcDestinations[reqBody.Name]
	if !isDestinationSvcDestExists {
		log.C(ctx).Infof("Destination with name: %q and authentication type: %q does not exists in the destination service. Returning 204 No Content...", reqBody.Name, reqBody.AuthenticationType)
		httputils.Respond(writer, http.StatusNoContent)
	}
	delete(h.DestinationSvcDestinations, reqBody.Name)
	log.C(ctx).Infof("Destination with name: %q and authentication type: %q was deleted from the destination service", reqBody.Name, reqBody.AuthenticationType)

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

	regionParam := h.Config.CertificateAPIConfig.RegionParam
	subaccountIDParam := h.Config.CertificateAPIConfig.SubaccountIDParam

	routeVars := mux.Vars(r)
	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]

	if regionParamValue == "" || subaccountIDParamValue == "" {
		log.C(ctx).Errorf("Missing required parameters: %q or/and %q", regionParam, subaccountIDParam)
		httphelpers.WriteError(writer, errors.Errorf("Not all of the required parameters - %q or/and %q are provided. X-Request-Id: %s", regionParam, subaccountIDParam, correlationID), http.StatusBadRequest)
		return
	}

	_, ok := h.DestinationCreatorSvcCertificates[reqBody.Name]
	if ok {
		log.C(ctx).Infof("Certificate with name: %q already exists. Returning 409 Conflict...", reqBody.Name)
		httputils.Respond(writer, http.StatusConflict)
	}

	certResp := CertificateResponseBody{
		FileName:         reqBody.Name + destinationcreator.JavaKeyStoreFileExtension,
		CommonName:       certCommonName,
		CertificateChain: certChain,
	}

	certRespBytes, err := json.Marshal(certResp)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("while marshaling certificate response body. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Wrapf(err, "while marshaling certificate response body. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	h.DestinationCreatorSvcCertificates[reqBody.Name] = certRespBytes

	destSvcCertificateResp := destinationcreator.DestinationSvcCertificateResponse{
		Name:    reqBody.Name + destinationcreator.JavaKeyStoreFileExtension,
		Content: certChain,
	}

	destSvcCertificateRespBytes, err := json.Marshal(destSvcCertificateResp)
	if err != nil {
		log.C(ctx).WithError(err).Error("while marshalling destination certificate response")
		httphelpers.WriteError(writer, errors.New("while marshalling destination certificate response"), http.StatusInternalServerError)
		return
	}

	h.DestinationSvcCertificates[reqBody.Name] = destSvcCertificateRespBytes

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

	regionParam := h.Config.CertificateAPIConfig.RegionParam
	subaccountIDParam := h.Config.CertificateAPIConfig.SubaccountIDParam
	certNameParam := h.Config.CertificateAPIConfig.CertificateNameParam

	routeVars := mux.Vars(r)
	regionParamValue := routeVars[regionParam]
	subaccountIDParamValue := routeVars[subaccountIDParam]
	certNameParamValue := routeVars[certNameParam]

	if regionParamValue == "" || subaccountIDParamValue == "" || certNameParamValue == "" {
		log.C(ctx).Errorf("Missing required parameters: %q, %q or/and %q", regionParam, subaccountIDParam, certNameParam)
		httphelpers.WriteError(writer, errors.Errorf("Not all of the required parameters - %q, %q or/and %q are provided. X-Request-Id: %s", regionParam, subaccountIDParam, certNameParam, correlationID), http.StatusBadRequest)
		return
	}

	_, isDestinationCreatorCertExist := h.DestinationCreatorSvcCertificates[reqBody.Name]
	if !isDestinationCreatorCertExist {
		log.C(ctx).Infof("Certificate with name: %q does not exists in the destination creator", reqBody.Name)
	} else {
		delete(h.DestinationCreatorSvcCertificates, reqBody.Name)
		log.C(ctx).Infof("Certificate with name: %q was deleted from the destination creator", reqBody.Name)
	}

	_, isDestinationSvcCertExists := h.DestinationSvcCertificates[reqBody.Name]
	if !isDestinationSvcCertExists {
		log.C(ctx).Infof("Certificate with name: %q does not exists in the destination service. Returning 204 No Content...", reqBody.Name)
		httputils.Respond(writer, http.StatusNoContent)
	}
	delete(h.DestinationSvcCertificates, reqBody.Name)
	log.C(ctx).Infof("Certificate with name: %q was deleted from the destination service", reqBody.Name)

	httputils.Respond(writer, http.StatusNoContent)
}

// Destination Service handlers

// GetDestinationByNameFromDestinationSvc todo::: go doc
func (h *Handler) GetDestinationByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	log.C(ctx).Info("Validating authorization header...")
	authorizationHeaderValue := r.Header.Get(httphelpers.AuthorizationHeaderKey)

	if authorizationHeaderValue == "" {
		log.C(ctx).Errorf("missing authorization header. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Errorf("missing authorization header. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	tokenValue := strings.TrimPrefix(authorizationHeaderValue, "Bearer ")
	if tokenValue == "" {
		log.C(ctx).Errorf("the token value cannot be empty. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Errorf("the token value cannot be empty. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	routeVars := mux.Vars(r)
	destinationNameParamValue := routeVars["name"]

	if destinationNameParamValue == "" {
		log.C(ctx).Error("Missing required parameters: \"name\"")
		httphelpers.WriteError(writer, errors.Errorf("Missing required parameters: \"name\". X-Request-Id: %s", correlationID), http.StatusBadRequest)
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

// todo::: delete
//// DeleteDestinationByNameFromDestinationSvc todo::: go doc
//func (h *Handler) DeleteDestinationByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
//	ctx := r.Context()
//	correlationID := correlation.CorrelationIDFromContext(ctx)
//
//	log.C(ctx).Info("Validating authorization header...")
//	authorizationHeaderValue := r.Header.Get(httphelpers.AuthorizationHeaderKey)
//
//	if authorizationHeaderValue == "" {
//		log.C(ctx).Errorf("missing authorization header. X-Request-Id: %s", correlationID)
//		httphelpers.WriteError(writer, errors.Errorf("missing authorization header. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
//		return
//	}
//
//	tokenValue := strings.TrimPrefix(authorizationHeaderValue, "Bearer ")
//	if tokenValue == "" {
//		log.C(ctx).Errorf("the token value cannot be empty. X-Request-Id: %s", correlationID)
//		httphelpers.WriteError(writer, errors.Errorf("the token value cannot be empty. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
//		return
//	}
//
//	routeVars := mux.Vars(r)
//	destinationNameParamValue := routeVars["name"]
//
//	if destinationNameParamValue == "" {
//		log.C(ctx).Error("Missing required parameters: \"name\"")
//		httphelpers.WriteError(writer, errors.Errorf("Missing required parameters: \"name\". X-Request-Id: %s", correlationID), http.StatusBadRequest)
//		return
//	}
//
//	h.DestinationCreatorSvcDestinations = make(map[string]json.RawMessage)
//	httputils.Respond(writer, http.StatusOK)
//}

// GetDestinationCertificateByNameFromDestinationSvc todo::: go doc
func (h *Handler) GetDestinationCertificateByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	log.C(ctx).Info("Validating authorization header...")
	authorizationHeaderValue := r.Header.Get(httphelpers.AuthorizationHeaderKey)

	if authorizationHeaderValue == "" {
		log.C(ctx).Errorf("missing authorization header. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Errorf("missing authorization header. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	tokenValue := strings.TrimPrefix(authorizationHeaderValue, "Bearer ")
	if tokenValue == "" {
		log.C(ctx).Errorf("the token value cannot be empty. X-Request-Id: %s", correlationID)
		httphelpers.WriteError(writer, errors.Errorf("the token value cannot be empty. X-Request-Id: %s", correlationID), http.StatusInternalServerError)
		return
	}

	routeVars := mux.Vars(r)
	certificateNameParamValue := routeVars["name"]

	if certificateNameParamValue == "" {
		log.C(ctx).Error("Missing required parameters: \"name\"")
		httphelpers.WriteError(writer, errors.Errorf("Missing required parameters: \"name\". X-Request-Id: %s", correlationID), http.StatusBadRequest)
		return
	}

	cert, exists := h.DestinationSvcCertificates[certificateNameParamValue]
	if !exists {
		log.C(ctx).Errorf("Destination with name: %q doest not exists", certificateNameParamValue)
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

// todo::: delete
//// DeleteDestinationCertificateByNameFromDestinationSvc todo::: go doc
//func (h *Handler) DeleteDestinationCertificateByNameFromDestinationSvc(writer http.ResponseWriter, r *http.Request) {
//	h.DestinationSvcCertificates = make(map[string]json.RawMessage)
//	httputils.Respond(writer, http.StatusOK)
//}
