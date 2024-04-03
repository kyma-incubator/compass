package destinationfetcher

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	pageCountQueryParameter = "$pageCount"
	pageQueryParameter      = "$page"
	pageSizeQueryParameter  = "$pageSize"
)

type Destination struct {
	Name              string `json:"Name"`
	Type              string `json:"Type"`
	URL               string `json:"URL"`
	Authentication    string `json:"Authentication"`
	XCorrelationID    string `json:"x-correlation-id"`
	XSystemTenantID   string `json:"x-system-id"`
	XSystemTenantName string `json:"x-system-name"`
	XSystemType       string `json:"x-system-type"`
}

type DeleteStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type DeleteResponse struct {
	Count   int
	Summary []DeleteStatus
}

type PostResponse struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
	Cause  string `json:"cause,omitempty"`
}

type Handler struct {
	destinations          []Destination
	destinationsSensitive map[string][]byte
}

func NewHandler() *Handler {
	return &Handler{destinationsSensitive: make(map[string][]byte)}
}

func (h *Handler) deleteDestination(name string) {
	for k, v := range h.destinations {
		if v.Name == name {
			h.destinations[k] = h.destinations[len(h.destinations)-1]
			h.destinations = h.destinations[:len(h.destinations)-1]
			return
		}
	}
}

func (h *Handler) DeleteDestination(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	destinationName := mux.Vars(req)["name"]

	if len(destinationName) == 0 {
		http.Error(writer, "Bad request - missing destination name", http.StatusBadRequest)
		return
	}

	deleteResponse := DeleteResponse{Count: 0}

	if _, ok := h.destinationsSensitive[destinationName]; !ok {
		deleteResponse.Summary = append(deleteResponse.Summary, DeleteStatus{
			Name:   destinationName,
			Status: "NOT_FOUND",
			Reason: "Could not find destination",
		})
	}

	delete(h.destinationsSensitive, destinationName)
	deleteResponse.Count = deleteResponse.Count + 1

	deleteResponse.Summary = append(deleteResponse.Summary, DeleteStatus{
		Name:   destinationName,
		Status: "DELETED",
	})

	h.deleteDestination(destinationName)

	responseJSON, err := json.Marshal(deleteResponse)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal response body")
		http.Error(writer, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	if _, err := writer.Write(responseJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
