package selfreg

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/subscription"
	"io"
	"net/http"
	"reflect"

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

func (h *Handler) HandleSelfRegPrep(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(h.c.NameQueryParam)
	if name == "" {
		err := errors.New(fmt.Sprintf("%s query param missing", h.c.NameQueryParam))
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}
	if tenant := r.URL.Query().Get(h.c.TenantQueryParam); tenant == "" {
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
}

func (h *Handler) HandleSelfRegCleanup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if name, ok := params[NamePath]; !ok || name == "" {
		err := errors.New(fmt.Sprintf("%s is missing from path", NamePath))
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}
	log.D().Infof("Dump request:")
	spew.Dump(r)
	tenants, exists := r.Header["Tenant"]

	log.D().Infof("Dump Tenants:")
	spew.Dump(tenants)
	if !exists {
		err := errors.New("Tenant is missing in request header")
		httphelpers.WriteError(w, err, http.StatusBadRequest)
		return
	}

	if subscription.Subscriptions[tenants[0]] {
		// We swallow their error msg and print only the status code
		err := errors.New("")
		httphelpers.WriteError(w, err, http.StatusConflict)
		return
	}

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
