package ias

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
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

type Handler struct {
	applications Applications
}

func NewHandler() *Handler {
	return &Handler{
		applications: Applications{
			Applications: []Application{
				{
					ID: "app-id-1",
					Authentication: ApplicationAuthentication{
						ClientID: "client-id-1",
						ProvidedAPIs: []ApplicationProvidedAPI{
							{
								Name:        "provided-api-name-1",
								Description: "provided-api-description-1",
							},
						},
					},
				},
				{
					ID: "app-id-1",
					Authentication: ApplicationAuthentication{
						ClientID: "client-id-2",
					},
				},
			},
		},
	}
}

func returnApps(writer http.ResponseWriter, apps Applications) {
	appsBytes, err := json.Marshal(apps)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(appsBytes)
}

func (h *Handler) GetAll(writer http.ResponseWriter, req *http.Request) {
	escapedFilter := req.URL.Query().Get("filter")
	if escapedFilter == "" {
		returnApps(writer, h.applications)
		return
	}

	unescapedFilter, err := url.QueryUnescape(escapedFilter)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(unescapedFilter, "clientId eq ") {
		returnApps(writer, h.applications)
		return
	}

	clientID := strings.TrimPrefix(unescapedFilter, "clientId eq ")
	var apps Applications
	for _, app := range h.applications.Applications {
		if app.Authentication.ClientID == clientID {
			apps.Applications = append(apps.Applications, app)
		}
	}

	returnApps(writer, apps)
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

func (h *Handler) Patch(writer http.ResponseWriter, req *http.Request) {
	appID, ok := mux.Vars(req)["appID"]
	if !ok {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	appUpdate := applicationUpdate{}
	if err := json.NewDecoder(req.Body).Decode(&appUpdate); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(appUpdate.Operations) != 1 ||
		appUpdate.Operations[0].Operation != replaceOp || appUpdate.Operations[0].Path != replacePath {

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
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
