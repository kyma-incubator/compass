package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

type CredentialsRequest struct {
	ID              string          `json:"id"`
	ApplicationID   string          `json:"applicationId"`
	ApplicationName string          `json:"applicationName"`
	Type            string          `json:"type"`
	X509Certificate string          `json:"x509Certificate"`
	HcmCloudTenant  *hcmCloudTenant `json:"hcmCloudTenant"`
}

type hcmCloudTenant struct {
	ImmutableCompanyId string `json:"immutableCompanyId"`
	CompanyId          string `json:"companyId"`
	Host               string `json:"host"`
}

func (cr CredentialsRequest) Validate() error {
	if len(cr.ApplicationID) == 0 {
		return errors.New("ApplicationID must not be empty")
	}
	if len(cr.ApplicationName) == 0 {
		return errors.New("ApplicationName must not be empty")
	}
	if len(cr.Type) == 0 {
		return errors.New("Type must not be empty")
	}
	if len(cr.X509Certificate) == 0 {
		return errors.New("X509Certificate must not be empty")
	}
	if cr.HcmCloudTenant == nil {
		return errors.New("HcmCloudTenant must not be empty")
	}
	if len(cr.HcmCloudTenant.ImmutableCompanyId) == 0 {
		return errors.New("ImmutableCompanyId must not be empty")
	}
	if len(cr.HcmCloudTenant.CompanyId) == 0 {
		return errors.New("CompanyId must not be empty")
	}
	if len(cr.HcmCloudTenant.Host) == 0 {
		return errors.New("Host must not be empty")
	}

	return nil
}

type Credentials struct {
	CompanyId   string       `json:"companyId"`
	ApiEndpoint string       `json:"apiEndpoint"`
	OauthClient *oauthClient `json:"oauthClient"`
}

type oauthClient struct {
	Secret         string `json:"secret"`
	TokenServerURL string `json:"tokenServerURL"`
}

type createCredsOperationStatus struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	ErrorID  string `json:"errorId"`
	ErrorMsg string `json:"errorMessage"`
	Location string `json:"location"`
}

var defaultCredentials = Credentials{
	CompanyId:   "myCompanyID",
	ApiEndpoint: "https://api.sfsf.com",
	OauthClient: &oauthClient{
		Secret:         "clientSecret",
		TokenServerURL: "https://api.sfsf.com/oauth/token",
	},
}

type CredentialsHandler struct {
	address              string
	requestedCredentials map[string]bool
}

func NewCredentialsHandler(address string) *CredentialsHandler {
	return &CredentialsHandler{
		address:              address,
		requestedCredentials: map[string]bool{},
	}
}

func (ch *CredentialsHandler) Create(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	body, err := ioutil.ReadAll(r.Body)
	log.C(ctx).Infof("Provided request credentials body is %s", string(body))
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to read requst body: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	credsReqBody := CredentialsRequest{}
	if err := json.Unmarshal(body, &credsReqBody); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to unmarshal requst body: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := credsReqBody.Validate(); err != nil {
		log.C(ctx).WithError(err).Errorf("Missing required fields in body: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(credsReqBody.ID) == 0 {
		credsReqBody.ID = uuid.New().String()
	}
	ch.requestedCredentials[credsReqBody.ID] = true
	rw.Header().Add("Location", fmt.Sprintf("http://localhost:8080/configuration/api/v1/inboundConnections/%s", credsReqBody.ID))
	// TODO uncomment
	//rw.Header().Add("Location", fmt.Sprintf("%s/configuration/api/v1/inboundConnections/%s", ch.address, credsID))
	rw.WriteHeader(http.StatusAccepted)
}

func (ch *CredentialsHandler) Status(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	vars := mux.Vars(r)
	credsID, ok := vars["connectionId"]
	if !ok {
		log.C(ctx)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := createCredsOperationStatus{
		ID:       credsID,
		Status:   "SUCCEEDED",
		Location: fmt.Sprintf("http://localhost:8080/configuration/api/v1/inboundConnections/%s/credentials", credsID),
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to marshal response: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := rw.Write(body); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write response: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Location", fmt.Sprintf("http://localhost:8080/configuration/api/v1/inboundConnections/%s/credentials", credsID))
	rw.Header().Add("Desi", fmt.Sprintf("http://localhost:8080/configuration/api/v1/inboundConnections/%s/credentials", credsID))
}

func (ch *CredentialsHandler) Retrieve(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	vars := mux.Vars(r)
	credsID, ok := vars["connectionId"]
	if !ok {
		log.C(ctx)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	ready, ok := ch.requestedCredentials[credsID]
	if !ok {
		log.C(ctx).Errorf("Credentials with ID %s are not present on server", credsID)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if !ready {
		log.C(ctx).Errorf("Credentials with ID %s are not ready for consumption", credsID)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	resp := defaultCredentials

	body, err := json.Marshal(resp)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to marshal response: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := rw.Write(body); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write response: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (ch *CredentialsHandler) Delete(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	credsID, ok := vars["connectionId"]
	if !ok {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	delete(ch.requestedCredentials,credsID)
	rw.Header().Add("Location", fmt.Sprintf("http://localhost:8080/configuration/api/v1/inboundConnections/%s", credsID))
	rw.WriteHeader(http.StatusAccepted)
}