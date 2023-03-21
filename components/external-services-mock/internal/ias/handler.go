package ias

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

type Applications struct {
	Applications []Application `json:"applications"`
}

type Application struct {
	ID             string                    `json:"id"`
	Authentication ApplicationAuthentication `json:"urn:sap:identity:application:schemas:extension:sci:1.0:Authentication"`
}

type ApplicationAuthentication struct {
	ClientID     string                   `json:"clientId"`
	ProvidedAPIs []ApplicationProvidedAPI `json:"providedApis"`
	ConsumedAPIs []ApplicationConsumedAPI `json:"consumedApis"`
}

type ApplicationProvidedAPI struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ApplicationConsumedAPI struct {
	Name     string `json:"name"`
	APIName  string `json:"apiName"`
	AppID    string `json:"appId"`
	ClientID string `json:"clientId"`
}

type Config struct {
	ConsumerAppID       string `envconfig:"APP_IAS_ADAPTER_CONSUMER_APP_ID"`
	ConsumerAppClientID string `envconfig:"APP_IAS_ADAPTER_CONSUMER_APP_CLIENT_ID"`
	ProviderAppID       string `envconfig:"APP_IAS_ADAPTER_PROVIDER_APP_ID"`
	ProviderAppClientID string `envconfig:"APP_IAS_ADAPTER_PROVIDER_APP_CLIENT_ID"`
	ProvidedAPIName     string `envconfig:"APP_IAS_ADAPTER_PROVIDED_API_NAME"`
}

type Handler struct {
	applications Applications
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		applications: Applications{
			Applications: []Application{
				{
					ID: cfg.ProviderAppID,
					Authentication: ApplicationAuthentication{
						ClientID: cfg.ProviderAppClientID,
						ProvidedAPIs: []ApplicationProvidedAPI{
							{
								Name:        cfg.ProvidedAPIName,
								Description: cfg.ProvidedAPIName,
							},
						},
					},
				},
				{
					ID: cfg.ConsumerAppID,
					Authentication: ApplicationAuthentication{
						ClientID: cfg.ConsumerAppClientID,
					},
				},
			},
		},
	}
}

func returnApps(writer http.ResponseWriter, request *http.Request, apps Applications) {
	logger := log.C(request.Context())
	appsBytes, err := json.Marshal(apps)
	if err != nil {
		logger.Errorf("Failed to marshal apps: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(appsBytes)
}

func (h *Handler) GetAll(writer http.ResponseWriter, request *http.Request) {
	logger := log.C(request.Context())

	escapedFilter := request.URL.Query().Get("filter")
	if escapedFilter == "" {
		returnApps(writer, request, h.applications)
		return
	}

	unescapedFilter, err := url.QueryUnescape(escapedFilter)
	if err != nil {
		logger.Errorf("Failed to unescape filter query param '%s': %s", escapedFilter, err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(unescapedFilter, "clientId eq ") {
		returnApps(writer, request, h.applications)
		return
	}

	clientID := strings.TrimPrefix(unescapedFilter, "clientId eq ")
	var apps Applications
	for _, app := range h.applications.Applications {
		if app.Authentication.ClientID == clientID {
			apps.Applications = append(apps.Applications, app)
		}
	}

	returnApps(writer, request, apps)
}

type applicationUpdate struct {
	Operations []applicationUpdateOperation `json:"operations"`
}

type updateOperation string

const (
	ReplaceOp updateOperation = "replace"
)

type applicationUpdateOperation struct {
	Operation updateOperation          `json:"op"`
	Path      string                   `json:"path"`
	Value     []ApplicationConsumedAPI `json:"value"`
}

const (
	replaceOp   = "replace"
	replacePath = "/urn:sap:identity:application:schemas:extension:sci:1.0:Authentication/consumedApis"
)

func (h *Handler) Patch(writer http.ResponseWriter, request *http.Request) {
	logger := log.C(request.Context())

	appID, ok := mux.Vars(request)["appID"]
	if !ok {
		logger.Error("Failed to get appID path param")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	appUpdate := applicationUpdate{}
	if err := json.NewDecoder(request.Body).Decode(&appUpdate); err != nil {
		logger.Errorf("Failed to decode app update body: %s", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(appUpdate.Operations) != 1 ||
		appUpdate.Operations[0].Operation != replaceOp || appUpdate.Operations[0].Path != replacePath {

		logger.Errorf("Received invalid update operation %+v", appUpdate.Operations)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	var (
		appIndex int
		app      Application
		appFound bool
	)
	for appIndex, app = range h.applications.Applications {
		if app.ID == appID {
			appFound = true
			break
		}
	}
	if appFound {
		app.Authentication.ConsumedAPIs = appUpdate.Operations[0].Value
		h.applications.Applications[appIndex] = app
	} else {
		logger.Errorf("App with ID '%s' not found", appID)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
