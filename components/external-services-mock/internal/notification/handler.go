package notification

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

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
)

type Handler struct {
	mappings          map[string][]Response
	shouldReturnError bool
}

type Response struct {
	Operation     Operation
	ApplicationID *string
	RequestBody   json.RawMessage
}

func NewHandler() *Handler {
	return &Handler{
		mappings:          make(map[string][]Response),
		shouldReturnError: true,
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

	if config := gjson.Get(string(bodyBytes), "configuration").String(); config == "" {
		writer.WriteHeader(http.StatusNoContent)
	}

	writer.WriteHeader(http.StatusOK)
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
		mappings = append(h.mappings[id], Response{
			Operation:   Assign,
			RequestBody: bodyBytes,
		})
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

	h.Patch(writer, r)
}

func (h *Handler) ResetShouldFail(writer http.ResponseWriter, r *http.Request) {
	h.shouldReturnError = true
	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) Cleanup(writer http.ResponseWriter, r *http.Request) {
	h.mappings = make(map[string][]Response)
	writer.WriteHeader(http.StatusOK)
}
