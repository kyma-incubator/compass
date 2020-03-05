package configuration

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/fake-external-test-component/pkg/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ConfigChangeService -output=automock -outpkg=automock -case=underscore
type ConfigChangeService interface {
	Save(change model.ConfigurationChange) (string, error)
	Get(id string) *model.ConfigurationChange
	Delete(id string)
}

type ConfigChangeHandler struct {
	service ConfigChangeService
}

func NewConfigurationHandler(service ConfigChangeService) *ConfigChangeHandler {
	return &ConfigChangeHandler{service: service}
}

const (
	HeaderContentTypeKey   = "Content-Type"
	HeaderContentTypeValue = "application/json;charset=UTF-8"
)

func (h *ConfigChangeHandler) Save(writer http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var log model.ConfigurationChange
	err := json.NewDecoder(req.Body).Decode(&log)
	if err != nil {
		WriteError(writer, errors.Wrap(err, "while decoding input"), http.StatusInternalServerError)
		return
	}

	id, err := h.service.Save(log)
	if err != nil {
		WriteError(writer, errors.Wrap(err, "while saving configuration change log"), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Got: %+v\n", log)

	response := model.SuccessResponse{ID: id}
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(&response)
	if err != nil {
		err := errors.New("error while encoding response")
		WriteError(writer, err, http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) Get(writer http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		WriteError(writer, errors.New("parameter id not provided"), http.StatusBadRequest)
		return
	}

	val := h.service.Get(id)
	if val == nil {
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	err := json.NewEncoder(writer).Encode(&val)
	if err != nil {
		WriteError(writer, errors.New("error while encoding response"), http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) Delete(writer http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		WriteError(writer, errors.New("parameter id not provided"), http.StatusBadRequest)
	}

	h.service.Delete(id)
}

func WriteError(writer http.ResponseWriter, err error, statusCode int) {
	writer.Header().Set(HeaderContentTypeKey, HeaderContentTypeValue)

	response := model.ErrorResponse{
		Error: err.Error(),
	}

	value, err := json.Marshal(&response)
	if err != nil {
		//TODO: cleanup
		panic(err)
	}
	http.Error(writer, string(value), statusCode)

}
