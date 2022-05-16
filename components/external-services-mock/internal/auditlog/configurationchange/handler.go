package configurationchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const timestampLayout = "2006-01-02T15:04:05"

//go:generate mockery --name=ConfigChangeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ConfigChangeService interface {
	Save(change model.ConfigurationChange) (string, error)
	Get(id string) *model.ConfigurationChange
	List() []model.ConfigurationChange
	SearchByTimestamp(timeFrom, timeTo time.Time) []model.ConfigurationChange
	Delete(id string)
}

type ConfigChangeHandler struct {
	service ConfigChangeService
	logger  *log.Logger
}

func NewConfigurationHandler(service ConfigChangeService, logger *log.Logger) *ConfigChangeHandler {
	return &ConfigChangeHandler{service: service, logger: logger}
}

func (h *ConfigChangeHandler) Save(writer http.ResponseWriter, req *http.Request) {
	defer h.closeBody(req)

	var auditLog model.ConfigurationChange
	err := json.NewDecoder(req.Body).Decode(&auditLog)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while decoding input"), http.StatusInternalServerError)
		return
	}

	err = h.validateInput(auditLog)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while validation configuration change log"), http.StatusBadRequest)
		return
	}

	id, err := h.service.Save(auditLog)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while saving configuration change log"), http.StatusInternalServerError)
		return
	}

	response := model.SuccessResponse{ID: id}
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(&response)
	if err != nil {
		err := errors.New("error while encoding response")
		httphelpers.WriteError(writer, err, http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) Get(writer http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		httphelpers.WriteError(writer, errors.New("parameter [id] not provided"), http.StatusBadRequest)
		return
	}

	val := h.service.Get(id)
	if val == nil {
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	err := json.NewEncoder(writer).Encode(&val)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("error while encoding response"), http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) List(writer http.ResponseWriter, req *http.Request) {
	values := h.service.List()
	if len(values) == 0 {
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	err := json.NewEncoder(writer).Encode(&values)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("error while encoding response"), http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) Delete(writer http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		httphelpers.WriteError(writer, errors.New("parameter [id] not provided"), http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
	h.service.Delete(id)
}

func (h *ConfigChangeHandler) SearchByTimestamp(writer http.ResponseWriter, req *http.Request) {
	timeFromString := req.URL.Query().Get("time_from")
	if timeFromString == "" {
		httphelpers.WriteError(writer, errors.New("parameter [time_from] not provided"), http.StatusBadRequest)
		return
	}

	timeToString := req.URL.Query().Get("time_to")
	if timeFromString == "" {
		httphelpers.WriteError(writer, errors.New("parameter [time_to] not provided"), http.StatusBadRequest)
		return
	}

	from, err := time.Parse(timestampLayout, timeFromString)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("parameter [time_from] is not in proper format"), http.StatusBadRequest)
		return
	}

	to, err := time.Parse(timestampLayout, timeToString)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("parameter [time_to] is not in proper format"), http.StatusBadRequest)
		return
	}

	values := h.service.SearchByTimestamp(from, to)
	if len(values) == 0 {
		http.Error(writer, "", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(writer).Encode(&values)
	if err != nil {
		httphelpers.WriteError(writer, errors.New("error while encoding response"), http.StatusInternalServerError)
		return
	}
}

func (h *ConfigChangeHandler) closeBody(req *http.Request) {
	err := req.Body.Close()
	if err != nil {
		h.logger.Error(errors.Wrap(err, "while closing body"))
	}
}

func (h *ConfigChangeHandler) validateInput(input model.ConfigurationChange) error {
	var invalidData []string

	if input.User != "$USER" {
		invalidData = append(invalidData, fmt.Sprintf("User is not valid, expected $USER, got %s", input.User))
	}

	if input.Tenant != "$PROVIDER" && input.Tenant != "$SUBSCRIBER" {
		invalidData = append(invalidData, fmt.Sprintf("Tenant is not valid, expected $SUBSCRIBER or $PROVIDER, got %s", input.Tenant))
	}

	if _, err := time.Parse(model.LogFormatDate, input.Time); err != nil {
		invalidData = append(invalidData, fmt.Sprintf("Time is not valid with schema: %s, got %s", model.LogFormatDate, input.Time))
	}

	if len(invalidData) != 0 {
		return errors.New(buildValidationErrors(invalidData))
	}
	return nil
}

func buildValidationErrors(bad []string) string {
	builder := strings.Builder{}
	for _, value := range bad {
		builder.WriteString(value)
	}
	return builder.String()
}
