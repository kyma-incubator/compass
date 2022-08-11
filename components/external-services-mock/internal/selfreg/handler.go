package selfreg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/subscription"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
)

const (
	NamePath       = "name"
	responseFormat = `{"%s": "test-prefix-%s"}`
)

type Config struct {
	Path               string `envconfig:"APP_SELF_REGISTER_PATH"`
	NameQueryParam     string `envconfig:"APP_SELF_REGISTER_NAME_QUERY_PARAM"`
	TenantQueryParam   string `envconfig:"APP_SELF_REGISTER_TENANT_QUERY_PARAM"`
	ResponseKey        string `envconfig:"APP_SELF_REGISTER_RESPONSE_KEY"`
	RequestBodyPattern string `envconfig:"APP_SELF_REGISTER_REQUEST_BODY_PATTERN"`
}

type Handler struct {
	c Config
}

func NewSelfRegisterHandler(c Config) *Handler {
	return &Handler{c: c}
}

var SelfRegistrations = make(map[string]string, 0)

func (h *Handler) HandleSelfRegPrep(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(h.c.NameQueryParam)
	if name == "" {
		err := errors.New(fmt.Sprintf("%s query param missing", h.c.NameQueryParam))
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}
	tenant := r.URL.Query().Get(h.c.TenantQueryParam)
	if tenant == "" {
		err := errors.New(fmt.Sprintf("%s query param missing", h.c.TenantQueryParam))
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}

	expectedBody := fmt.Sprintf(h.c.RequestBodyPattern, name)
	equalJSON, err := areEqualJSON(expectedBody, string(body))
	if err != nil {
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	} else if !equalJSON {
		err := errors.New("body does not have the expected structure")
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	response := fmt.Sprintf(responseFormat, h.c.ResponseKey, name)
	if _, err := io.WriteString(w, response); err != nil {
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}

	SelfRegistrations[name] = tenant
}

func (h *Handler) HandleSelfRegCleanup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name, ok := params[NamePath]
	if !ok || name == "" {
		err := errors.New(fmt.Sprintf("%s is missing from path", NamePath))
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}

	providerSubaccount, cloningExists := SelfRegistrations[name]

	subscriberExists := false
	for _, provider := range subscription.Subscriptions {
		if provider == providerSubaccount {
			subscriberExists = true
			break
		}
	}

	if cloningExists && subscriberExists {
		// We swallow their error msg and print only the status code
		err := errors.New("")
		httphelpers.WriteError(w, err, http.StatusConflict)
		return
	}

	delete(SelfRegistrations, name)
	w.WriteHeader(http.StatusOK)
}

func areEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}
