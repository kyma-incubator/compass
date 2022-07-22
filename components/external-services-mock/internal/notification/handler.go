package notification

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

// TODO Unit tests

type Operation string

const (
	// Assign represents the assign operation done on a given formation
	Assign Operation = "assign"
	// Unassign represents the unassign operation done on a given formation
	Unassign Operation = "unassign"
)

type Handler struct {
	mappings map[string][]Response
}

type Response struct {
	Operation     Operation
	ApplicationID *string
	RequestBody   []byte
}

func NewHandler() *Handler {
	return &Handler{
		mappings: make(map[string][]Response),
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

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(bodyBytes)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
	}
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
		//_, err = writer.Write([]byte("{}"))
		_, err = writer.Write(bodyBytes)
		if err != nil {
			httphelpers.WriteError(writer, errors.Wrap(err, "error while writing response"), http.StatusInternalServerError)
			return
		}
	}
}
