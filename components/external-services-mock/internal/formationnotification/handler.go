package formationnotification

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/tidwall/gjson"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type Operation string

const (
	// Assign represents the assign operation done on a given formation
	Assign Operation = "assign"
	// Unassign represents the unassign operation done on a given formation
	Unassign Operation = "unassign"
	// CreateFormation represents the create operation on a given formation
	CreateFormation Operation = "createFormation"
	// DeleteFormation represents the delete operation on a given formation
	DeleteFormation Operation = "deleteFormation"
)

type Configuration struct {
	ExternalClientCertTestSecretName            string `envconfig:"EXTERNAL_CLIENT_CERT_TEST_SECRET_NAME"`
	ExternalClientCertTestSecretNamespace       string `envconfig:"EXTERNAL_CLIENT_CERT_TEST_SECRET_NAMESPACE"`
	ExternalClientCertCertKey                   string `envconfig:"APP_EXTERNAL_CLIENT_CERT_KEY"`
	ExternalClientCertKeyKey                    string `envconfig:"APP_EXTERNAL_CLIENT_KEY_KEY"`
	DirectorExternalCertFAAsyncStatusURL        string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASSIGNMENT_ASYNC_STATUS_URL"`
	DirectorExternalCertFormationAsyncStatusURL string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASYNC_STATUS_URL"`
	FormationMappingAsyncResponseDelay          int64  `envconfig:"APP_FORMATION_MAPPING_ASYNC_RESPONSE_DELAY"`
}

// FormationAssignmentRequestBody contains the request input of the formation assignment async status request
type FormationAssignmentRequestBody struct {
	State         FormationAssignmentState `json:"state,omitempty"`
	Configuration json.RawMessage          `json:"configuration,omitempty"`
	Error         string                   `json:"error,omitempty"`
}

// FormationRequestBody contains the request input of the formation async status request
type FormationRequestBody struct {
	State FormationState `json:"state"`
	Error string         `json:"error,omitempty"`
}

// FormationAssignmentResponseBody contains the synchronous formation assignment notification response body
type FormationAssignmentResponseBody struct {
	Config FormationAssignmentResponseConfig
	Error  string `json:"error,omitempty"`
}

// FormationAssignmentResponseBodyWithState contains the synchronous formation assignment notification response body with state in it
type FormationAssignmentResponseBodyWithState struct {
	Config FormationAssignmentResponseConfig
	Error  string                   `json:"error,omitempty"`
	State  FormationAssignmentState `json:"state"`
}

// FormationAssignmentResponseConfig contains the configuration of the formation response body
type FormationAssignmentResponseConfig struct {
	Key  string `json:"key"`
	Key2 struct {
		Key string `json:"key"`
	} `json:"key2"`
}

// KymaMappingsBasicAuthentication contains the basic credentials used in the KymaMappingsOutboundCommunication
type KymaMappingsBasicAuthentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// KymaMappingsOauthAuthentication contains the oauth credentials used in the KymaMappingsOutboundCommunication
type KymaMappingsOauthAuthentication struct {
	TokenServiceUrl string `json:"tokenServiceUrl"`
	ClientId        string `json:"clientId"`
	ClientSecret    string `json:"clientSecret"`
}

// KymaMappingsOutboundCommunication contains the outbound communication used in the KymaMappingsCredentials
type KymaMappingsOutboundCommunication struct {
	BasicAuthentication KymaMappingsBasicAuthentication `json:"basicAuthentication,omitempty"`
	OauthAuthentication KymaMappingsOauthAuthentication `json:"oauth2ClientCredentials,omitempty"`
}

// KymaMappingsCredentials contains the credentials used in the KymaMappingsConfiguration
type KymaMappingsCredentials struct {
	OutboundCommunication KymaMappingsOutboundCommunication `json:"outboundCommunication"`
}

// KymaMappingsConfiguration contains the configuration used in KymaMappingsResponseBody
type KymaMappingsConfiguration struct {
	Credentials KymaMappingsCredentials `json:"credentials"`
}

// KymaMappingsResponseBody contains the state and configuration for the Kyma Tenant Mapping flow
type KymaMappingsResponseBody struct {
	State         string                    `json:"state"`
	Configuration KymaMappingsConfiguration `json:"configuration"`
}

// FormationAssignmentState is a type that represents formation assignments state
type FormationAssignmentState string

// FormationState is a type that represents formation state
type FormationState string

// ReadyAssignmentState indicates that the formation assignment is in a ready state
const ReadyAssignmentState FormationAssignmentState = "READY"

// CreateErrorAssignmentState indicates that an error occurred during the creation of the formation assignment
const CreateErrorAssignmentState FormationAssignmentState = "CREATE_ERROR"

// DeleteErrorAssignmentState indicates that an error occurred during the deletion of the formation assignment
const DeleteErrorAssignmentState FormationAssignmentState = "DELETE_ERROR"

// ConfigPendingAssignmentState indicates that the config is either missing or not finalized in the formation assignment
const ConfigPendingAssignmentState FormationAssignmentState = "CONFIG_PENDING"

// ReadyFormationState indicates that the formation is in a ready state
const ReadyFormationState FormationState = "READY"

// CreateErrorFormationState indicates that an error occurred during the creation of the formation
const CreateErrorFormationState FormationState = "CREATE_ERROR"

// DeleteErrorFormationState indicates that an error occurred during the deletion of the formation
const DeleteErrorFormationState FormationState = "DELETE_ERROR"

// Handler is responsible to mock and handle any formation and formation assignment notification requests
type Handler struct {
	// Mappings is a map of string to Response, where the string value currently can be `formationID` or `tenantID`
	// mapped to a particular Response that later will be validated in the E2E tests
	Mappings          map[string][]Response
	ShouldReturnError bool
	config            Configuration
}

// Response is used to model the response for a given formation or formation assignment notification request.
// It has a metadata fields like Operation and also the request body of the notification request later used for validation in the E2E tests.
type Response struct {
	Operation     Operation
	ApplicationID *string
	RequestBody   json.RawMessage
}

// NewHandler creates a new Handler
func NewHandler(notificationConfiguration Configuration) *Handler {
	return &Handler{
		Mappings:          make(map[string][]Response),
		ShouldReturnError: true,
		config:            notificationConfiguration,
	}
}

// Formation Assignment notifications synchronous handlers

// SyncFAResponseFn is a function type that represents the synchronous formation assignment response function signature
type SyncFAResponseFn func(bodyBytes []byte)

// Patch handles synchronous formation assignment notification requests for Assign operation
func (h *Handler) Patch(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func([]byte) {
		var response interface{} = FormationAssignmentResponseBody{
			Config: FormationAssignmentResponseConfig{
				Key: "value",
				Key2: struct {
					Key string `json:"key"`
				}{Key: "value2"},
			},
		}

		httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
	}

	h.syncFAResponse(ctx, writer, r, responseFunc)
}

// PatchWithState handles synchronous formation assignment notification requests for Assign operation and returns state in the response body
func (h *Handler) PatchWithState(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func([]byte) {
		var response interface{} = FormationAssignmentResponseBodyWithState{
			State: ConfigPendingAssignmentState,
			Config: FormationAssignmentResponseConfig{
				Key: "value",
				Key2: struct {
					Key string `json:"key"`
				}{Key: "value2"},
			},
		}

		httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
	}

	h.syncFAResponse(ctx, writer, r, responseFunc)
}

// RespondWithIncomplete handles synchronous formation assignment notification requests for Assign operation
// that based on the provided config in the request body we return either so called "incomplete" status coe(204) without config in case the config is not provided
// or if the config is provided we just return it with "success" status code(200)
func (h *Handler) RespondWithIncomplete(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func(bodyBytes []byte) {
		if config := gjson.Get(string(bodyBytes), "config").String(); config == "" {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		response := FormationAssignmentResponseBody{
			Config: FormationAssignmentResponseConfig{
				Key: "value",
				Key2: struct {
					Key string `json:"key"`
				}{Key: "value2"},
			},
		}
		httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
	}

	h.syncFAResponse(ctx, writer, r, responseFunc)
}

// RespondWithIncompleteAndDestinationDetails handles synchronous formation assignment notification requests for Assign operation
// that returns destination details if the config in the request body is NOT provided, and if the config is provided returns READY state without configuration
func (h *Handler) RespondWithIncompleteAndDestinationDetails(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func(bodyBytes []byte) {
		if config := gjson.Get(string(bodyBytes), "receiverTenant.configuration").String(); config == "" {
			response := "{\"state\":\"CONFIG_PENDING\",\"configuration\":{\"destinations\":[{\"name\":\"e2e-design-time-destination-name\",\"type\":\"HTTP\",\"description\":\"e2e-design-time-destination description\",\"proxyType\":\"Internet\",\"authentication\":\"NoAuthentication\",\"url\":\"http://e2e-design-time-url-example.com\"}],\"credentials\":{\"inboundCommunication\":{\"basicAuthentication\":{\"correlationIds\":[\"e2e-basic-correlation-ids\"],\"destinations\":[{\"name\":\"e2e-basic-destination-name\",\"description\":\"e2e-basic-destination description\",\"url\":\"http://e2e-basic-url-example.com\",\"authentication\":\"BasicAuthentication\",\"additionalProperties\":{\"e2e-basic-testKey\":\"e2e-basic-testVal\"}}]},\"samlAssertion\":{\"correlationIds\":[\"e2e-saml-correlation-ids\"],\"destinations\":[{\"name\":\"e2e-saml-assertion-destination-name\",\"description\":\"e2e saml assertion destination description\",\"url\":\"http://e2e-saml-url-example.com\",\"additionalProperties\":{\"e2e-samlTestKey\":\"e2e-samlTestVal\"}}]}}},\"additionalProperties\":[{\"propertyName\":\"example-property-name\",\"propertyValue\":\"example-property-value\",\"correlationIds\":[\"correlation-ids\"]}]}}"

			httputils.RespondWithBody(ctx, writer, http.StatusOK, json.RawMessage(response))
			return
		}

		response := FormationAssignmentResponseBodyWithState{
			State: ReadyAssignmentState,
		}
		httputils.RespondWithBody(ctx, writer, http.StatusOK, response)
	}

	h.syncFAResponse(ctx, writer, r, responseFunc)
}

// Delete handles synchronous formation assignment notification requests for Unassign operation
func (h *Handler) Delete(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func([]byte) { writer.WriteHeader(http.StatusOK) }

	h.syncFAResponse(ctx, writer, r, responseFunc)
}

// DestinationDelete handles synchronous formation assignment notification requests for destination deletion during Unassign operation
func (h *Handler) DestinationDelete(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func([]byte) {
		response := FormationAssignmentResponseBodyWithState{
			State: ReadyAssignmentState,
		}
		httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
	}

	h.syncFAResponse(ctx, writer, r, responseFunc)
}

// DeleteWithState handles synchronous formation assignment notification requests for Unassign operation and returns state in the response body
func (h *Handler) DeleteWithState(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func([]byte) {
		response := FormationAssignmentResponseBodyWithState{State: "READY"}
		httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
	}

	h.syncFAResponse(ctx, writer, r, responseFunc)
}

func (h *Handler) FailResponse(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func([]byte) {
		response := FormationAssignmentResponseBody{Error: "failed to parse request"}
		httputils.RespondWithBody(context.TODO(), writer, http.StatusBadRequest, response)
	}
	h.syncFAResponse(ctx, writer, r, responseFunc)
}

// todo::: minor: go doc + the other methods as well
func (h *Handler) FailOnceResponse(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.ShouldReturnError {
		responseFunc := func([]byte) {
			response := FormationAssignmentResponseBody{Error: "failed to parse request"}
			httputils.RespondWithBody(context.TODO(), writer, http.StatusBadRequest, response)
			h.ShouldReturnError = false
		}

		h.syncFAResponse(ctx, writer, r, responseFunc)
		return
	}

	if r.Method == http.MethodPatch {
		h.Patch(writer, r)
	}

	if r.Method == http.MethodDelete {
		h.Delete(writer, r)
	}
}

func (h *Handler) ResetShouldFail(writer http.ResponseWriter, r *http.Request) {
	h.ShouldReturnError = true
	writer.WriteHeader(http.StatusOK)
}

// GetResponses returns the notification data saved in the Mappings
func (h *Handler) GetResponses(writer http.ResponseWriter, r *http.Request) {
	if bodyBytes, err := json.Marshal(&h.Mappings); err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	} else {
		writer.WriteHeader(http.StatusOK)
		_, err = writer.Write(bodyBytes)
		if err != nil {
			httphelpers.WriteError(writer, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
			return
		}
	}
}

// Cleanup deletes/cleanup the notification data saved in the Mappings
func (h *Handler) Cleanup(writer http.ResponseWriter, r *http.Request) {
	h.Mappings = make(map[string][]Response)
	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) syncFAResponse(ctx context.Context, writer http.ResponseWriter, r *http.Request, responseFunc SyncFAResponseFn) {
	routeVars := mux.Vars(r)
	id, ok := routeVars["tenantId"]
	if !ok {
		log.C(ctx).Error("missing tenantId path parameter in the url")
		httphelpers.WriteError(writer, errors.New("missing tenantId path parameter in the url"), http.StatusBadRequest)
		return
	}

	if _, ok = h.Mappings[id]; !ok {
		h.Mappings[id] = make([]Response, 0, 1)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("error while reading request body")
		httphelpers.WriteError(writer, errors.Wrap(err, "error while reading request body"), http.StatusInternalServerError)
		return
	}

	var result interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.C(ctx).WithError(err).Error("body is not a valid JSON")
		httphelpers.WriteError(writer, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}

	mappings := h.Mappings[id]
	if r.Method == http.MethodPatch {
		log.C(ctx).Infof("Adding to formation assignment notifications mappings operation: %s and body: %s", Assign, string(bodyBytes))
		mappings = append(h.Mappings[id], Response{
			Operation:   Assign,
			RequestBody: bodyBytes,
		})
	}

	if r.Method == http.MethodDelete {
		applicationId, ok := routeVars["applicationId"]
		if !ok {
			log.C(ctx).Error("missing tenantId path parameter in the url")
			httphelpers.WriteError(writer, errors.New("missing applicationId in url"), http.StatusBadRequest)
			return
		}
		log.C(ctx).Infof("Adding to formation assignment notifications mappings operation: %s, app ID: %s and body: %s", Unassign, applicationId, string(bodyBytes))
		mappings = append(h.Mappings[id], Response{
			Operation:     Unassign,
			ApplicationID: &applicationId,
			RequestBody:   bodyBytes,
		})
	}

	h.Mappings[id] = mappings

	responseFunc(bodyBytes)
}

// Formation Assignment notifications asynchronous handlers and helper functions

// AsyncFAResponseFn is a function type that represents the formation assignment response function signature
type AsyncFAResponseFn func(client *http.Client, correlationID, formationID, formationAssignmentID, config string)

// AsyncNoopFAResponseFn is an empty implementation of the AsyncFAResponseFn function
var AsyncNoopFAResponseFn = func(client *http.Client, correlationID, formationID, formationAssignmentID, config string) {}

// Async handles asynchronous formation assignment notification requests for Assign operation
func (h *Handler) Async(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func(client *http.Client, correlationID, formationID, formationAssignmentID, config string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		err := h.executeFormationAssignmentStatusUpdateRequest(client, correlationID, ReadyAssignmentState, config, formationID, formationAssignmentID)
		if err != nil {
			log.C(ctx).Errorf("while executing formation assignment status update request: %s", err.Error())
		}
	}
	h.asyncFAResponse(ctx, writer, r, Assign, `{"asyncKey": "asyncValue", "asyncKey2": {"asyncNestedKey": "asyncNestedValue"}}`, responseFunc)
}

// AsyncDestinationPatch handles asynchronous formation assignment notification requests for destination creation during Assign operation
func (h *Handler) AsyncDestinationPatch(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func(client *http.Client, correlationID, formationID, formationAssignmentID, config string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		err := h.executeFormationAssignmentStatusUpdateRequest(client, correlationID, ReadyAssignmentState, config, formationID, formationAssignmentID)
		if err != nil {
			log.C(ctx).Errorf("while executing formation assignment status update request: %s", err.Error())
		}
	}

	config := "{\"credentials\":{\"outboundCommunication\":{\"basicAuthentication\":{\"url\":\"https://e2e-basic-destination-url.com\",\"username\":\"e2e-basic-destination-username\",\"password\":\"e2e-basic-destination-password\"},\"samlAssertion\":{\"url\":\"http://e2e-saml-url-example.com\"}}}}"
	h.asyncFAResponse(ctx, writer, r, Assign, config, responseFunc)
}

// AsyncDelete handles asynchronous formation assignment notification requests for Unassign operation
func (h *Handler) AsyncDelete(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func(client *http.Client, correlationID, formationID, formationAssignmentID, config string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		err := h.executeFormationAssignmentStatusUpdateRequest(client, correlationID, ReadyAssignmentState, config, formationID, formationAssignmentID)
		if err != nil {
			log.C(ctx).Errorf("while executing status update request: %s", err.Error())
		}
	}
	h.asyncFAResponse(ctx, writer, r, Unassign, "", responseFunc)
}

// AsyncDestinationDelete handles asynchronous formation assignment notification requests for destination deletion during Unassign operation
func (h *Handler) AsyncDestinationDelete(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	responseFunc := func(client *http.Client, correlationID, formationID, formationAssignmentID, config string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		err := h.executeFormationAssignmentStatusUpdateRequest(client, correlationID, ReadyAssignmentState, config, formationID, formationAssignmentID)
		if err != nil {
			log.C(ctx).Errorf("while executing status update request: %s", err.Error())
		}
	}
	h.asyncFAResponse(ctx, writer, r, Unassign, "", responseFunc)
}

// AsyncNoResponseAssign handles asynchronous formation assignment notification requests for Assign operation that do not send any request to the formation assignment status API
func (h *Handler) AsyncNoResponseAssign(writer http.ResponseWriter, r *http.Request) {
	h.asyncFAResponse(r.Context(), writer, r, Assign, "", AsyncNoopFAResponseFn)
}

// AsyncNoResponseUnassign handles asynchronous formation assignment notification requests for Unassign operation that do not send any request to the formation assignment status API
func (h *Handler) AsyncNoResponseUnassign(writer http.ResponseWriter, r *http.Request) {
	h.asyncFAResponse(r.Context(), writer, r, Unassign, "", AsyncNoopFAResponseFn)
}

// AsyncFailOnce handles asynchronous formation assignment notification requests for both Assign and Unassign operations by first failing and setting error states. Afterwards the operation succeeds
func (h *Handler) AsyncFailOnce(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	operation := Assign
	if r.Method == http.MethodPatch {
		operation = Assign
	} else if r.Method == http.MethodDelete {
		operation = Unassign
	}
	responseFunc := func(client *http.Client, correlationID, formationID, formationAssignmentID, config string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		state := ReadyAssignmentState
		if operation == Assign && h.ShouldReturnError {
			state = CreateErrorAssignmentState
			h.ShouldReturnError = false
		} else if operation == Unassign && h.ShouldReturnError {
			state = DeleteErrorAssignmentState
			h.ShouldReturnError = false
		}
		err := h.executeFormationAssignmentStatusUpdateRequest(client, correlationID, state, config, formationID, formationAssignmentID)
		if err != nil {
			log.C(ctx).Errorf("while executing status update request: %s", err.Error())
		}
	}
	if h.ShouldReturnError {
		config := "test error"
		h.asyncFAResponse(ctx, writer, r, operation, config, responseFunc)
	} else {
		config := `{"asyncKey": "asyncValue", "asyncKey2": {"asyncNestedKey": "asyncNestedValue"}}`
		h.asyncFAResponse(ctx, writer, r, operation, config, responseFunc)
	}
}

// executeFormationAssignmentStatusUpdateRequest prepares a request with the given inputs and sends it to the formation assignment status API
func (h *Handler) executeFormationAssignmentStatusUpdateRequest(certSecuredHTTPClient *http.Client, correlationID string, state FormationAssignmentState, testConfig, formationID, formationAssignmentID string) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	correlationIDKey := correlation.RequestIDHeaderKey
	ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, &correlationIDKey, &correlationID)

	logger := log.C(ctx).WithField(correlationIDKey, correlationID)
	ctx = log.ContextWithLogger(ctx, logger)

	FAReqBody := FormationAssignmentRequestBody{
		State: state,
	}
	if testConfig != "" {
		if state == CreateErrorAssignmentState || state == DeleteErrorAssignmentState {
			FAReqBody.Error = testConfig
		}
		if state == ReadyAssignmentState {
			FAReqBody.Configuration = json.RawMessage(testConfig)
		}
	}
	marshalBody, err := json.Marshal(FAReqBody)
	if err != nil {
		return err
	}

	FAStatusAPIEndpoint := strings.Replace(h.config.DirectorExternalCertFAAsyncStatusURL, fmt.Sprintf("{%s}", "ucl-formation-id"), formationID, 1)
	FAStatusAPIEndpoint = strings.Replace(FAStatusAPIEndpoint, fmt.Sprintf("{%s}", "ucl-assignment-id"), formationAssignmentID, 1)

	request, err := http.NewRequest(http.MethodPatch, FAStatusAPIEndpoint, bytes.NewBuffer(marshalBody))
	if err != nil {
		return err
	}

	request.Header.Add(correlation.RequestIDHeaderKey, correlationID)
	request.Header.Add(httphelpers.ContentTypeHeaderKey, httphelpers.ContentTypeApplicationJSON)
	log.C(ctx).Infof("Calling status API for formation assignment status update with the following data - formation ID: %s, assignment with ID: %s, state: %s and config: %s", formationID, formationAssignmentID, state, testConfig)
	_, err = certSecuredHTTPClient.Do(request)
	return err
}

// asyncFAResponse handles the incoming formation assignment notification requests and prepare "asynchronous" response through go routine with fixed(configurable) delay that executes the provided `responseFunc` which sends a request to the formation assignment status API
func (h *Handler) asyncFAResponse(ctx context.Context, writer http.ResponseWriter, r *http.Request, operation Operation, config string, responseFunc AsyncFAResponseFn) {
	routeVars := mux.Vars(r)
	id, ok := routeVars["tenantId"]
	if !ok {
		log.C(ctx).Error("missing tenantId path parameter in the url")
		httphelpers.WriteError(writer, errors.New("missing tenantId path parameter in the url"), http.StatusBadRequest)
		return
	}
	if _, ok := h.Mappings[id]; !ok {
		h.Mappings[id] = make([]Response, 0, 1)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("error while reading request body")
		httphelpers.WriteError(writer, errors.Wrap(err, "error while reading request body"), http.StatusInternalServerError)
		return
	}

	var result interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.C(ctx).WithError(err).Error("body is not a valid JSON")
		httphelpers.WriteError(writer, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}
	response := Response{
		Operation:   operation,
		RequestBody: bodyBytes,
	}
	if r.Method == http.MethodDelete {
		applicationId, ok := routeVars["applicationId"]
		if !ok {
			log.C(ctx).Error("missing applicationId path parameter in the url")
			httphelpers.WriteError(writer, errors.New("missing applicationId in url"), http.StatusBadRequest)
			return
		}
		response.ApplicationID = &applicationId
	}

	log.C(ctx).Infof("Adding to formation assignment notifications mappings operation: %s and body: %s", operation, string(bodyBytes))
	h.Mappings[id] = append(h.Mappings[id], response)

	formationID, err := retrieveFormationID(ctx, bodyBytes)
	if err != nil {
		httputils.RespondWithError(ctx, writer, http.StatusInternalServerError, errors.New("Missing formation ID"))
		return
	}

	formationAssignmentID, err := retrieveFormationAssignmentID(ctx, bodyBytes)
	if err != nil {
		httputils.RespondWithError(ctx, writer, http.StatusInternalServerError, errors.New("Missing formation assignment ID"))
		return
	}

	certAuthorizedHTTPClient, err := h.getCertAuthorizedHTTPClient(ctx)
	if err != nil {
		httputils.RespondWithError(ctx, writer, http.StatusInternalServerError, err)
		return
	}

	correlationID := correlation.CorrelationIDFromContext(ctx)
	go responseFunc(certAuthorizedHTTPClient, correlationID, formationID, formationAssignmentID, config)

	writer.WriteHeader(http.StatusAccepted)
}

func retrieveFormationID(ctx context.Context, bodyBytes []byte) (string, error) {
	return retrieveIDFromJSONPath(ctx, bodyBytes, []string{"ucl-formation-id", "context.uclFormationId"})
}

func retrieveFormationAssignmentID(ctx context.Context, bodyBytes []byte) (string, error) {
	return retrieveIDFromJSONPath(ctx, bodyBytes, []string{"formation-assignment-id", "receiverTenant.uclAssignmentId"})
}

func retrieveIDFromJSONPath(ctx context.Context, bodyBytes []byte, jsonPaths []string) (string, error) {
	var found bool
	var id string
	for _, path := range jsonPaths {
		id = gjson.Get(string(bodyBytes), path).String()
		if id == "" {
			log.C(ctx).Warnf("Couldn't find ID at %q path", path)
			continue
		}
		log.C(ctx).Warnf("Successfully find ID at %q path", path)
		found = true
		break
	}

	if !found {
		return "", errors.New("Couldn't find ID in the provided json paths")
	}

	return id, nil
}

// Formation notifications synchronous handlers and helper functions

// PostFormation handles synchronous formation notification requests for CreateFormation operation
func (h *Handler) PostFormation(writer http.ResponseWriter, r *http.Request) {
	h.synchronousFormationResponse(writer, r, CreateFormation)
}

// DeleteFormation handles synchronous formation notification requests for DeleteFormation operation
func (h *Handler) DeleteFormation(writer http.ResponseWriter, r *http.Request) {
	h.synchronousFormationResponse(writer, r, DeleteFormation)
}

// FailOnceFormation handles synchronous formation notification requests for both Create and Delete operations by first failing and setting error states. Afterwards the operation succeeds
func (h *Handler) FailOnceFormation(writer http.ResponseWriter, r *http.Request) {
	operation := CreateFormation
	if r.Method == http.MethodPost {
		operation = CreateFormation
	} else if r.Method == http.MethodDelete {
		operation = DeleteFormation
	}

	if h.ShouldReturnError {
		formationID, ok := mux.Vars(r)["uclFormationId"]
		if !ok {
			httphelpers.WriteError(writer, errors.New("missing uclFormationId in url"), http.StatusBadRequest)
			return
		}

		if _, ok := h.Mappings[formationID]; !ok {
			h.Mappings[formationID] = make([]Response, 0, 1)
		}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			httphelpers.WriteError(writer, errors.Wrap(err, "error while reading request body"), http.StatusInternalServerError)
			return
		}

		var result interface{}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			httphelpers.WriteError(writer, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
			return
		}

		h.Mappings[formationID] = append(h.Mappings[formationID], Response{
			Operation:   operation,
			RequestBody: bodyBytes,
		})

		response := struct {
			Error string `json:"error"`
		}{
			Error: "failed to parse request",
		}
		httputils.RespondWithBody(context.TODO(), writer, http.StatusBadRequest, response)
		h.ShouldReturnError = false
		return
	}
	h.synchronousFormationResponse(writer, r, operation)
}

// synchronousFormationResponse extracts the logic that handles formation notification requests
func (h *Handler) synchronousFormationResponse(writer http.ResponseWriter, r *http.Request, formationOperation Operation) {
	formationID, ok := mux.Vars(r)["uclFormationId"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("missing uclFormationId in url"), http.StatusBadRequest)
		return
	}

	if _, ok := h.Mappings[formationID]; !ok {
		h.Mappings[formationID] = make([]Response, 0, 1)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "error while reading request body"), http.StatusInternalServerError)
		return
	}

	var result interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}

	h.Mappings[formationID] = append(h.Mappings[formationID], Response{
		Operation:   formationOperation,
		RequestBody: bodyBytes,
	})

	writer.WriteHeader(http.StatusOK)
}

// Formation notifications asynchronous handlers and helper functions

// FormationResponseFn is a function type that represents the formation response function signature
type FormationResponseFn func(client *http.Client, correlationID, formationError, formationID string)

// NoopFormationResponseFn is an empty implementation of the FormationResponseFn function
var NoopFormationResponseFn = func(client *http.Client, correlationID, formationError, formationID string) {}

// AsyncPostFormation handles asynchronous formation notification requests for CreateFormation operation.
func (h *Handler) AsyncPostFormation(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	formationResponseFunc := func(client *http.Client, correlationID, formationError, formationID string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		err := h.executeFormationStatusUpdateRequest(client, correlationID, ReadyFormationState, formationError, formationID)
		if err != nil {
			log.C(ctx).Errorf("while executing formation status update request: %s", err.Error())
		}
	}
	h.asyncFormationResponse(ctx, writer, r, CreateFormation, "", formationResponseFunc)
}

// AsyncDeleteFormation handles asynchronous formation notification requests for DeleteFormation operation
func (h *Handler) AsyncDeleteFormation(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	formationResponseFunc := func(client *http.Client, correlationID, formationError, formationID string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		err := h.executeFormationStatusUpdateRequest(client, correlationID, ReadyFormationState, formationError, formationID)
		if err != nil {
			log.C(ctx).Errorf("while executing formation status update request: %s", err.Error())
		}
	}
	h.asyncFormationResponse(ctx, writer, r, DeleteFormation, "", formationResponseFunc)
}

// AsyncFormationFailOnce handles asynchronous formation notification requests for both Create and Delete operations by first failing and setting error states. Afterwards the operation succeeds
func (h *Handler) AsyncFormationFailOnce(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	operation := CreateFormation
	if r.Method == http.MethodPost {
		operation = CreateFormation
	} else if r.Method == http.MethodDelete {
		operation = DeleteFormation
	}
	responseFunc := func(client *http.Client, correlationID, formationError, formationID string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		state := ReadyFormationState
		if r.Method == http.MethodPost && h.ShouldReturnError {
			state = CreateErrorFormationState
			h.ShouldReturnError = false
		} else if r.Method == http.MethodDelete && h.ShouldReturnError {
			state = DeleteErrorFormationState
			h.ShouldReturnError = false
		}
		err := h.executeFormationStatusUpdateRequest(client, correlationID, state, formationError, formationID)
		if err != nil {
			log.C(ctx).Errorf("while executing formation status update request: %s", err.Error())
		}
	}
	if h.ShouldReturnError {
		h.asyncFormationResponse(ctx, writer, r, operation, "failed to parse request", responseFunc)
	} else {
		h.asyncFormationResponse(ctx, writer, r, operation, "", responseFunc)
	}
}

// AsyncNoResponse handles asynchronous formation notification requests that do not send any request to the formation status API
func (h *Handler) AsyncNoResponse(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	operation := CreateFormation
	if r.Method == http.MethodPost {
		operation = CreateFormation
	} else if r.Method == http.MethodDelete {
		operation = DeleteFormation
	}
	h.asyncFormationResponse(ctx, writer, r, operation, "", NoopFormationResponseFn)
}

func (h *Handler) KymaBasicCredentials(writer http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		username := "user"
		password := "pass"
		response := KymaMappingsResponseBody{
			State:         string(ReadyAssignmentState),
			Configuration: KymaMappingsConfiguration{Credentials: KymaMappingsCredentials{OutboundCommunication: KymaMappingsOutboundCommunication{BasicAuthentication: KymaMappingsBasicAuthentication{Username: username, Password: password}}}},
		}

		httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
	} else if r.Method == http.MethodDelete {
		writer.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) KymaOauthCredentials(writer http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		tokenUrl := "url"
		clientId := "id"
		clientSecret := "secret"
		response := KymaMappingsResponseBody{
			State:         string(ReadyAssignmentState),
			Configuration: KymaMappingsConfiguration{Credentials: KymaMappingsCredentials{OutboundCommunication: KymaMappingsOutboundCommunication{OauthAuthentication: KymaMappingsOauthAuthentication{TokenServiceUrl: tokenUrl, ClientId: clientId, ClientSecret: clientSecret}}}},
		}

		httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
	} else if r.Method == http.MethodDelete {
		writer.WriteHeader(http.StatusOK)
	}
}

// executeFormationStatusUpdateRequest prepares a request with the given inputs and sends it to the formation status API
func (h *Handler) executeFormationStatusUpdateRequest(certSecuredHTTPClient *http.Client, correlationID string, formationState FormationState, formationError, formationID string) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	correlationIDKey := correlation.RequestIDHeaderKey
	ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, &correlationIDKey, &correlationID)

	logger := log.C(ctx).WithField(correlationIDKey, correlationID)
	ctx = log.ContextWithLogger(ctx, logger)

	formationReqBody := FormationRequestBody{
		State: formationState,
		Error: formationError,
	}

	marshalBody, err := json.Marshal(formationReqBody)
	if err != nil {
		return err
	}

	formationStatusAPIEndpoint := strings.Replace(h.config.DirectorExternalCertFormationAsyncStatusURL, fmt.Sprintf("{%s}", "ucl-formation-id"), formationID, 1)
	request, err := http.NewRequest(http.MethodPatch, formationStatusAPIEndpoint, bytes.NewBuffer(marshalBody))
	if err != nil {
		return err
	}

	request.Header.Add(correlation.RequestIDHeaderKey, correlationID)
	request.Header.Add(httphelpers.ContentTypeHeaderKey, httphelpers.ContentTypeApplicationJSON)
	log.C(ctx).Infof("Calling status API for formation status update with the following data - formation ID: %s, state: %s and error: %s", formationID, formationState, formationError)
	_, err = certSecuredHTTPClient.Do(request)
	return err
}

// asyncFormationResponse handles the incoming formation notification requests and prepare "asynchronous" response through go routine with fixed(configurable) delay that executes the provided `formationResponseFunc` which sends a request to the formation status API
func (h *Handler) asyncFormationResponse(ctx context.Context, writer http.ResponseWriter, r *http.Request, operation Operation, formationErr string, formationResponseFunc FormationResponseFn) {
	formationID, ok := mux.Vars(r)["uclFormationId"]
	if !ok {
		log.C(ctx).Error("missing uclFormationId path parameter in the url")
		httphelpers.WriteError(writer, errors.New("missing uclFormationId path parameter in the url"), http.StatusBadRequest)
		return
	}

	if _, ok = h.Mappings[formationID]; !ok {
		h.Mappings[formationID] = make([]Response, 0, 1)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("error while reading formation notification request body")
		httphelpers.WriteError(writer, errors.Wrap(err, "error while reading formation notification request body"), http.StatusInternalServerError)
		return
	}

	var result interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.C(ctx).WithError(err).Error("body is not a valid JSON")
		httphelpers.WriteError(writer, errors.Wrap(err, "body is not a valid JSON"), http.StatusBadRequest)
		return
	}

	log.C(ctx).Infof("Adding to formation notifications mappings operation: %s and body: %s", operation, string(bodyBytes))
	h.Mappings[formationID] = append(h.Mappings[formationID], Response{
		Operation:   operation,
		RequestBody: bodyBytes,
	})

	formationIDFromBody := gjson.Get(string(bodyBytes), "details.id").String()
	if formationIDFromBody == "" {
		log.C(ctx).Errorf("Missing formation ID from request body")
		httputils.RespondWithError(ctx, writer, http.StatusInternalServerError, errors.New("Missing formation ID from request body"))
		return
	}

	formationNameFromBody := gjson.Get(string(bodyBytes), "details.name").String()
	if formationNameFromBody == "" {
		log.C(ctx).Errorf("Missing formation name from request body")
		httputils.RespondWithError(ctx, writer, http.StatusInternalServerError, errors.New("Missing formation name from request body"))
		return
	}

	certAuthorizedHTTPClient, err := h.getCertAuthorizedHTTPClient(ctx)
	if err != nil {
		httputils.RespondWithError(ctx, writer, http.StatusInternalServerError, err)
		return
	}

	correlationID := correlation.CorrelationIDFromContext(ctx)
	go formationResponseFunc(certAuthorizedHTTPClient, correlationID, formationErr, formationID)

	writer.WriteHeader(http.StatusAccepted)
}

// Common helper functions for both Formation and Formation Assignment handlers

func (h *Handler) getCertAuthorizedHTTPClient(ctx context.Context) (*http.Client, error) {
	k8sClient, err := kubernetes.NewKubernetesClientSet(ctx, time.Second, time.Minute, time.Minute)
	providerExtCrtTestSecret, err := k8sClient.CoreV1().Secrets(h.config.ExternalClientCertTestSecretNamespace).Get(ctx, h.config.ExternalClientCertTestSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while getting secret with name: %s in namespace: %s", h.config.ExternalClientCertTestSecretName, h.config.ExternalClientCertTestSecretNamespace)
	}

	providerKeyBytes := providerExtCrtTestSecret.Data[h.config.ExternalClientCertKeyKey]
	if len(providerKeyBytes) == 0 {
		return nil, errors.New("The private key could not be empty")
	}

	providerCertChainBytes := providerExtCrtTestSecret.Data[h.config.ExternalClientCertCertKey]
	if len(providerCertChainBytes) == 0 {
		return nil, errors.New("The certificate chain could not be empty")
	}

	privateKey, certChain, err := clientCertPair(providerCertChainBytes, providerKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "while generating client certificate pair")
	}
	certAuthorizedHTTPClient := newCertAuthorizedHTTPClient(privateKey, certChain, true)
	return certAuthorizedHTTPClient, nil
}

func clientCertPair(certChainBytes, privateKeyBytes []byte) (*rsa.PrivateKey, [][]byte, error) {
	certs, err := cert.DecodeCertificates(certChainBytes)
	if err != nil {
		return nil, nil, err
	}

	privateKeyPem, _ := pem.Decode(privateKeyBytes)
	if privateKeyPem == nil {
		return nil, nil, errors.New("Private key should not be nil")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPem.Bytes)
	if err != nil {
		pkcs8PrivateKey, err := x509.ParsePKCS8PrivateKey(privateKeyPem.Bytes)
		if err != nil {
			return nil, nil, err
		}

		var ok bool
		privateKey, ok = pkcs8PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, errors.New("Incorrect type of privateKey")
		}
	}

	tlsCert := cert.NewTLSCertificate(privateKey, certs...)
	return privateKey, tlsCert.Certificate, nil
}

func newCertAuthorizedHTTPClient(key crypto.PrivateKey, rawCertChain [][]byte, skipSSLValidation bool) *http.Client {
	tlsCert := tls.Certificate{
		Certificate: rawCertChain,
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: skipSSLValidation,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Second * 30,
	}

	return httpClient
}