package provider

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

const compassURL = "https://github.com/kyma-incubator/compass"

var respErrorMsg = "An unexpected error occurred while processing the request"

type Dependency struct {
	Xsappname string `json:"xsappname"`
}

type handler struct {
	xsappnameClones           []string
	directDependencyXsappname string
}

func NewHandler(directDependencyXsappname string) *handler {
	return &handler{
		directDependencyXsappname: directDependencyXsappname,
	}
}

// OnSubscription handles subscription callback request on real environment. When someone is subscribed to the provider tenant this method will be executed
func (h *handler) OnSubscription(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Handling on subscription request...")

	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		log.C(ctx).Errorf("expected %s or %s method but got: %s", http.MethodPut, http.MethodDelete, r.Method)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writer.Header().Set(httphelpers.ContentTypeHeaderKey, "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
	log.C(ctx).Info("Successfully handled on subscription request")
}

// DependenciesConfigure configures a provider dependency that will be used later in subscription request callback
func (h *handler) DependenciesConfigure(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)

	log.C(ctx).Info("Configuring subscription dependencies...")
	if r.Method != http.MethodPost {
		log.C(ctx).Errorf("expected %s method but got: %s", http.MethodPost, r.Method)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).Errorf("while reading request body: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	var deps []string
	if err := json.Unmarshal(body, &deps); err != nil {
		errResp := errors.Wrap(err, "An error occurred while unmarshalling request body")
		httphelpers.RespondWithError(ctx, writer, errResp, errResp.Error(), correlationID, http.StatusInternalServerError)
		return
	}

	h.xsappnameClones = make([]string, 0)
	for _, dep := range deps {
		if dep == "" {
			log.C(ctx).Warnf("The provided dependency is empty. Continue with the next one...")
			continue
		}
		log.C(ctx).Infof("Adding dependency with name: %s to the dependency list", dep)
		h.xsappnameClones = append(h.xsappnameClones, dep)
	}

	if len(h.xsappnameClones) == 0 {
		errResp := errors.New("The dependency list could not be empty")
		httphelpers.RespondWithError(ctx, writer, errResp, errResp.Error(), correlationID, http.StatusInternalServerError)
		return
	}

	writer.Header().Set(httphelpers.ContentTypeHeaderKey, httphelpers.ContentTypeApplicationJSON)
	writer.WriteHeader(http.StatusOK)
	log.C(ctx).Info("Successfully configured subscription dependencies")
}

// Dependencies is invoked on real environment as part of the subscription request and returns provider's dependencies in the expected format
func (h *handler) Dependencies(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := correlation.CorrelationIDFromContext(ctx)
	log.C(ctx).Info("Handling get dependencies request...")

	var deps []*Dependency
	for _, clone := range h.xsappnameClones {
		deps = append(deps, &Dependency{Xsappname: clone})
	}

	if len(deps) == 0 {
		errResp := errors.New("The dependency list could not be empty")
		httphelpers.RespondWithError(ctx, writer, errResp, errResp.Error(), correlationID, http.StatusInternalServerError)
		return
	}

	depsMarshalled, err := json.Marshal(deps)
	if err != nil {
		errResp := errors.Wrap(err, "An error occurred while marshalling subscription dependencies")
		httphelpers.RespondWithError(ctx, writer, errResp, errResp.Error(), correlationID, http.StatusInternalServerError)
		return
	}

	writer.Header().Set(httphelpers.ContentTypeHeaderKey, httphelpers.ContentTypeApplicationJSON)
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(depsMarshalled); err != nil {
		errResp := errors.Wrap(err, "An error occurred while writing response")
		httphelpers.RespondWithError(ctx, writer, errResp, errResp.Error(), correlationID, http.StatusInternalServerError)
		return
	}
	log.C(ctx).Info("Successfully handled get dependencies request")
}

// DependenciesIndirect is invoked on real environment as part of the subscription request where CMP is indirect dependencies and returns provider's dependencies in the expected format
func (h *handler) DependenciesIndirect(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Handling dependency request for indirect dependency subscription...")

	deps := []*Dependency{{Xsappname: h.directDependencyXsappname}}
	depsMarshalled, err := json.Marshal(deps)
	if err != nil {
		log.C(ctx).Errorf("while marshalling subscription dependencies: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling subscription dependencies"), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(depsMarshalled); err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
	log.C(ctx).Infof("Successfully configured indirect subscription dependency: %s", h.directDependencyXsappname)

	log.C(ctx).Info("Successfully handled dependency request for indirect dependency subscription")
}
