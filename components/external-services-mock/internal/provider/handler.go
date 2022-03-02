package provider

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

const compassURL = "https://github.com/kyma-incubator/compass"

type Dependency struct {
	Xsappname string `json:"xsappname"`
}

type handler struct {
	xsappnameClone string
}

func NewHandler() *handler {
	return &handler{}
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

	writer.Header().Set("Content-Type", "text/plain")
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
	log.C(ctx).Info("Configuring subscription dependency...")
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

	if string(body) == "" {
		log.C(ctx).Error("The request body is empty")
		httphelpers.WriteError(writer, errors.New("The request body is empty"), http.StatusInternalServerError)
		return
	}

	h.xsappnameClone = string(body)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(body); err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
	log.C(ctx).Infof("Successfully configured subscription dependency: %s", h.xsappnameClone)
}

// Dependencies is invoked on real environment as part of the subscription request and returns provider's dependencies in the expected format
func (h *handler) Dependencies(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Handling dependency request...")

	deps := []*Dependency{{Xsappname: h.xsappnameClone}}
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
	log.C(ctx).Info("Successfully handled dependency request")
}
