package notification

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

type NotificationsConfiguration struct {
	ExternalClientCertTestSecretName      string `envconfig:"EXTERNAL_CLIENT_CERT_TEST_SECRET_NAME"`
	ExternalClientCertTestSecretNamespace string `envconfig:"EXTERNAL_CLIENT_CERT_TEST_SECRET_NAMESPACE"`
	ExternalClientCertCertKey             string `envconfig:"APP_EXTERNAL_CLIENT_CERT_KEY"`
	ExternalClientCertKeyKey              string `envconfig:"APP_EXTERNAL_CLIENT_KEY_KEY"`
	DirectorExternalCertFAAsyncStatusURL  string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASSIGNMENT_ASYNC_STATUS_URL"`
	FormationMappingAsyncResponseDelay    int64  `envconfig:"APP_FORMATION_MAPPING_ASYNC_RESPONSE_DELAY"`
}

type RequestBody struct {
	State         ConfigurationState `json:"state"`
	Configuration json.RawMessage    `json:"configuration,omitempty"`
	Error         string             `json:"error,omitempty"`
}

type ConfigurationState string

const ReadyConfigurationState ConfigurationState = "READY"
const CreateErrorConfigurationState ConfigurationState = "CREATE_ERROR"
const DeleteErrorConfigurationState ConfigurationState = "DELETE_ERROR"

type Handler struct {
	// mappings is a map of string to Response, where the string value currently can be `formationID` or `tenantID`
	// mapped to a particular Response that later will be validated in the E2E tests
	mappings          map[string][]Response
	shouldReturnError bool
	config            NotificationsConfiguration
}

type Response struct {
	Operation     Operation
	ApplicationID *string
	RequestBody   json.RawMessage
}

func NewHandler(notificationConfiguration NotificationsConfiguration) *Handler {
	return &Handler{
		mappings:          make(map[string][]Response),
		shouldReturnError: true,
		config:            notificationConfiguration,
	}
}

func (h *Handler) Patch(writer http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["tenantId"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("missing tenantId in url"), http.StatusBadRequest)
		return
	}

	if _, ok = h.mappings[id]; !ok {
		h.mappings[id] = make([]Response, 0, 1)
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
	mappings := h.mappings[id]
	mappings = append(h.mappings[id], Response{
		Operation:   Assign,
		RequestBody: bodyBytes,
	})
	h.mappings[id] = mappings

	response := struct {
		Config struct {
			Key  string `json:"key"`
			Key2 struct {
				Key string `json:"key"`
			} `json:"key2"`
		}
	}{
		Config: struct {
			Key  string `json:"key"`
			Key2 struct {
				Key string `json:"key"`
			} `json:"key2"`
		}{
			Key: "value",
			Key2: struct {
				Key string `json:"key"`
			}{Key: "value2"},
		},
	}
	httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
}

func (h *Handler) PostFormation(writer http.ResponseWriter, r *http.Request) {
	formationID, ok := mux.Vars(r)["uclFormationId"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("missing uclFormationId in url"), http.StatusBadRequest)
		return
	}

	if _, ok = h.mappings[formationID]; !ok {
		h.mappings[formationID] = make([]Response, 0, 1)
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
	mappings := h.mappings[formationID]
	mappings = append(h.mappings[formationID], Response{
		Operation:   CreateFormation,
		RequestBody: bodyBytes,
	})
	h.mappings[formationID] = mappings

	httputils.Respond(writer, http.StatusOK)
}

func (h *Handler) DeleteFormation(writer http.ResponseWriter, r *http.Request) {
	formationID, ok := mux.Vars(r)["uclFormationId"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("missing uclFormationId in url"), http.StatusBadRequest)
		return
	}

	if _, ok := h.mappings[formationID]; !ok {
		h.mappings[formationID] = make([]Response, 0, 1)
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

	h.mappings[formationID] = append(h.mappings[formationID], Response{
		Operation:   DeleteFormation,
		RequestBody: bodyBytes,
	})

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) RespondWithIncomplete(writer http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["tenantId"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("missing tenantId in url"), http.StatusBadRequest)
		return
	}

	if _, ok = h.mappings[id]; !ok {
		h.mappings[id] = make([]Response, 0, 1)
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

	mappings := h.mappings[id]
	mappings = append(h.mappings[id], Response{
		Operation:   Assign,
		RequestBody: bodyBytes,
	})
	h.mappings[id] = mappings

	if config := gjson.Get(string(bodyBytes), "config").String(); config == "" {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	response := struct {
		Config struct {
			Key  string `json:"key"`
			Key2 struct {
				Key string `json:"key"`
			} `json:"key2"`
		}
	}{
		Config: struct {
			Key  string `json:"key"`
			Key2 struct {
				Key string `json:"key"`
			} `json:"key2"`
		}{
			Key: "value",
			Key2: struct {
				Key string `json:"key"`
			}{Key: "value2"},
		},
	}
	httputils.RespondWithBody(context.TODO(), writer, http.StatusOK, response)
}

func (h *Handler) Delete(writer http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["tenantId"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("missing tenantId in url"), http.StatusBadRequest)
		return
	}
	applicationId, ok := mux.Vars(r)["applicationId"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("missing applicationId in url"), http.StatusBadRequest)
		return
	}

	if _, ok := h.mappings[id]; !ok {
		h.mappings[id] = make([]Response, 0, 1)
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

	h.mappings[id] = append(h.mappings[id], Response{
		Operation:     Unassign,
		ApplicationID: &applicationId,
		RequestBody:   bodyBytes,
	})

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) Async(writer http.ResponseWriter, r *http.Request) {
	responseFunc := func(ctx context.Context, client *http.Client, formationID, formationAssignmentID, config string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		err := h.executeStatusUpdateRequest(client, ReadyConfigurationState, config, formationID, formationAssignmentID)
		if err != nil {
			log.C(ctx).Errorf("while executing status update request: %s", err.Error())
		}
	}
	h.asyncResponse(writer, r, Assign, `{"asyncKey": "asyncValue", "asyncKey2": {"asyncNestedKey": "asyncNestedValue"}}`, responseFunc)

	writer.WriteHeader(http.StatusAccepted)
}

func (h *Handler) AsyncDelete(writer http.ResponseWriter, r *http.Request) {
	responseFunc := func(ctx context.Context, client *http.Client, formationID, formationAssignmentID, config string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		err := h.executeStatusUpdateRequest(client, ReadyConfigurationState, config, formationID, formationAssignmentID)
		if err != nil {
			log.C(ctx).Errorf("while executing status update request: %s", err.Error())
		}
	}
	h.asyncResponse(writer, r, Unassign, "", responseFunc)
}

func (h *Handler) AsyncNoResponseAssign(writer http.ResponseWriter, r *http.Request) {
	h.asyncResponse(writer, r, Assign, "", func(ctx context.Context, client *http.Client, formationID, formationAssignmentID, config string) {})
}

func (h *Handler) AsyncNoResponseUnassign(writer http.ResponseWriter, r *http.Request) {
	h.asyncResponse(writer, r, Unassign, "", func(ctx context.Context, client *http.Client, formationID, formationAssignmentID, config string) {})
}

func (h *Handler) asyncResponse(writer http.ResponseWriter, r *http.Request, operation Operation, config string, responseFunc func(ctx context.Context, client *http.Client, formationID, formationAssignmentID, config string)) {
	ctx := r.Context()
	id, ok := mux.Vars(r)["tenantId"]
	if !ok {
		httphelpers.WriteError(writer, errors.New("missing tenantId in url"), http.StatusBadRequest)
		return
	}
	if _, ok := h.mappings[id]; !ok {
		h.mappings[id] = make([]Response, 0, 1)
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
	response := Response{
		Operation:   operation,
		RequestBody: bodyBytes,
	}
	if r.Method == http.MethodDelete {
		applicationId, ok := mux.Vars(r)["applicationId"]
		if !ok {
			httphelpers.WriteError(writer, errors.New("missing applicationId in url"), http.StatusBadRequest)
			return
		}
		response.ApplicationID = &applicationId
	}

	mappings := h.mappings[id]
	mappings = append(h.mappings[id], response)
	h.mappings[id] = mappings

	formationID := gjson.Get(string(bodyBytes), "ucl-formation-id").String()
	if formationID == "" {
		httputils.RespondWithError(ctx, writer, 500, errors.New("Missing formation ID"))
		return
	}

	formationAssignmentID := gjson.Get(string(bodyBytes), "formation-assignment-id").String()
	if formationAssignmentID == "" {
		httputils.RespondWithError(ctx, writer, 500, errors.New("Missing formation assignment ID"))
		return
	}

	certAuthorizedHTTPClient, err := h.getCertAuthorizedHTTPClient(ctx)
	if err != nil {
		httputils.RespondWithError(ctx, writer, 500, err)
		return
	}

	go responseFunc(ctx, certAuthorizedHTTPClient, formationID, formationAssignmentID, config)

	writer.WriteHeader(http.StatusAccepted)
}

func (h *Handler) AsyncFailOnce(writer http.ResponseWriter, r *http.Request) {
	operation := Assign
	if r.Method == http.MethodPatch {
		operation = Assign
	} else if r.Method == http.MethodDelete {
		operation = Unassign
	}
	responseFunc := func(ctx context.Context, client *http.Client, formationID, formationAssignmentID, config string) {
		time.Sleep(time.Second * time.Duration(h.config.FormationMappingAsyncResponseDelay))
		state := ReadyConfigurationState
		if operation == Assign && h.shouldReturnError {
			state = CreateErrorConfigurationState
			h.shouldReturnError = false
		} else if operation == Unassign && h.shouldReturnError {
			state = DeleteErrorConfigurationState
			h.shouldReturnError = false
		}
		err := h.executeStatusUpdateRequest(client, state, config, formationID, formationAssignmentID)
		if err != nil {
			log.C(ctx).Errorf("while executing status update request: %s", err.Error())
		}
	}
	if h.shouldReturnError {
		config := "test error"
		h.asyncResponse(writer, r, operation, config, responseFunc)
	} else {
		config := `{"asyncKey": "asyncValue", "asyncKey2": {"asyncNestedKey": "asyncNestedValue"}}`
		h.asyncResponse(writer, r, operation, config, responseFunc)
	}

	writer.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetResponses(writer http.ResponseWriter, r *http.Request) {
	if bodyBytes, err := json.Marshal(&h.mappings); err != nil {
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

func (h *Handler) FailOnceResponse(writer http.ResponseWriter, r *http.Request) {
	if h.shouldReturnError {
		id, ok := mux.Vars(r)["tenantId"]
		if !ok {
			httphelpers.WriteError(writer, errors.New("missing tenantId in url"), http.StatusBadRequest)
			return
		}

		if _, ok = h.mappings[id]; !ok {
			h.mappings[id] = make([]Response, 0, 1)
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

		mappings := h.mappings[id]
		if r.Method == http.MethodPatch {
			mappings = append(h.mappings[id], Response{
				Operation:   Assign,
				RequestBody: bodyBytes,
			})
		}

		if r.Method == http.MethodDelete {
			applicationId, ok := mux.Vars(r)["applicationId"]
			if !ok {
				httphelpers.WriteError(writer, errors.New("missing applicationId in url"), http.StatusBadRequest)
				return
			}
			mappings = append(h.mappings[id], Response{
				Operation:     Unassign,
				ApplicationID: &applicationId,
				RequestBody:   bodyBytes,
			})
		}

		h.mappings[id] = mappings

		response := struct {
			Error string `json:"error"`
		}{
			Error: "failed to parse request",
		}
		httputils.RespondWithBody(context.TODO(), writer, http.StatusBadRequest, response)
		h.shouldReturnError = false
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
	h.shouldReturnError = true
	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) Cleanup(writer http.ResponseWriter, r *http.Request) {
	h.mappings = make(map[string][]Response)
	writer.WriteHeader(http.StatusOK)
}

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

func (h *Handler) executeStatusUpdateRequest(certSecuredHTTPClient *http.Client, state ConfigurationState, testConfig, formationID, formationAssignmentID string) error {
	reqBody := RequestBody{
		State: state,
	}
	if testConfig != "" {
		if state == CreateErrorConfigurationState || state == DeleteErrorConfigurationState {
			reqBody.Error = testConfig
		}
		if state == ReadyConfigurationState {
			reqBody.Configuration = json.RawMessage(testConfig)
		}
	}
	marshalBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	formationMappingEndpoint := strings.Replace(h.config.DirectorExternalCertFAAsyncStatusURL, fmt.Sprintf("{%s}", "ucl-formation-id"), formationID, 1)
	formationMappingEndpoint = strings.Replace(formationMappingEndpoint, fmt.Sprintf("{%s}", "ucl-assignment-id"), formationAssignmentID, 1)

	request, err := http.NewRequest(http.MethodPatch, formationMappingEndpoint, bytes.NewBuffer(marshalBody))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")
	_, err = certSecuredHTTPClient.Do(request)
	return err
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
