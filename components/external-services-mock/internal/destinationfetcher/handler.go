package destinationfetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	pageCountQueryParameter = "$pageCount"
	pageQueryParameter      = "$page"
	pageSizeQueryParameter  = "$pageSize"
	deleteQueryParameter    = "$filter"
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

func (h *Handler) GetSensitiveData(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	destinationName := mux.Vars(req)["name"]

	if len(destinationName) == 0 {
		http.Error(writer, "Bad request", http.StatusBadRequest)
		return
	}

	log.C(ctx).Infof("Sending sensitive data of destination: %s", destinationName)
	data, ok := h.destinationsSensitive[destinationName]

	if !ok {
		http.Error(writer, "Not Found", http.StatusNotFound)
		return
	}

	if _, err := writer.Write(data); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetSubaccountDestinationsPage(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	pageRaw := req.URL.Query().Get(pageQueryParameter)

	destinations := h.destinations
	if pageRaw != "" {
		log.C(ctx).Infof("Page %s provided", pageRaw)
		pageNum, err := strconv.Atoi(pageRaw)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("could not convert page %s to int", pageRaw)
			http.Error(writer, "Invalid page number", http.StatusBadRequest)
			return
		}

		pageSizeRaw := req.URL.Query().Get(pageSizeQueryParameter)

		pageSize, err := strconv.Atoi(pageSizeRaw)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("could not convert pageSize %s to int", pageSizeRaw)
			http.Error(writer, "Invalid page size", http.StatusBadRequest)
			return
		}

		destinationCount := len(h.destinations)
		if req.URL.Query().Get(pageCountQueryParameter) == "true" {
			pageCount := destinationCount / pageSize

			if destinationCount%pageSize != 0 {
				pageCount = pageCount + 1
			}

			writer.Header().Set("Page-Count", fmt.Sprintf("%d", pageCount))
		}

		if (pageNum-1)*pageSize > len(h.destinations) {
			destinations = []Destination{}
		} else if pageNum*pageSize <= len(h.destinations) {
			destinations = h.destinations[((pageNum - 1) * pageSize):(pageNum * pageSize)]
		} else {
			destinations = h.destinations[((pageNum - 1) * pageSize):]
		}
	}

	if len(destinations) == 0 {
		if _, err := writer.Write([]byte("[]")); err != nil {
			log.C(ctx).WithError(err).Errorf("Failed to write data")
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	destinationsJSON, err := json.Marshal(destinations)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal destinations")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err := writer.Write(destinationsJSON); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
	filter := req.URL.Query().Get(deleteQueryParameter)

	if len(filter) == 0 {
		http.Error(writer, "Failed to read $filter query parameter", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(filter, "Name in") {
		http.Error(writer, "Invalid $filter format. Only name is supported", http.StatusBadRequest)
		return
	}

	filter = strings.ReplaceAll(filter, " ", "")
	firstBrackerIndex := strings.IndexByte(filter, '(') + 1
	secondBracketIndex := strings.IndexByte(filter, ')')

	destinationNames := strings.Split(filter[firstBrackerIndex:secondBracketIndex], ",")

	deleteResponse := DeleteResponse{Count: 0}

	for _, destinationName := range destinationNames {
		if len(destinationName) < 3 {
			continue
		}

		destinationName = destinationName[1 : len(destinationName)-1]

		if _, ok := h.destinationsSensitive[destinationName]; !ok {
			deleteResponse.Summary = append(deleteResponse.Summary, DeleteStatus{
				Name:   destinationName,
				Status: "NOT_FOUND",
				Reason: "Could not find destination",
			})

			continue
		}

		delete(h.destinationsSensitive, destinationName)
		deleteResponse.Count = deleteResponse.Count + 1

		deleteResponse.Summary = append(deleteResponse.Summary, DeleteStatus{
			Name:   destinationName,
			Status: "DELETED",
		})

		h.deleteDestination(destinationName)
	}

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

func validDestinationType(destinationType string) bool {
	return destinationType == "HTTP" || destinationType == "RFC" || destinationType == "MAIL" || destinationType == "LDAP"
}

func (h *Handler) PostDestination(writer http.ResponseWriter, req *http.Request) {
	isMultiStatusCodeEnabled := false
	var destinations []Destination = make([]Destination, 1)
	ctx := req.Context()

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to read request body")
		http.Error(writer, "Missing body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(data, &destinations)
	if err != nil {
		err = json.Unmarshal(data, &destinations[0])
		if err != nil {
			log.C(ctx).WithError(err).Error("Failed to unmarshal request body")
			http.Error(writer, "Invalid json", http.StatusBadRequest)
			return
		}
	}

	var responses []PostResponse
	for _, destination := range destinations {
		if _, ok := h.destinationsSensitive[destination.Name]; ok {
			responses = append(responses, PostResponse{destination.Name, http.StatusConflict, "Destination name already taken"})
			isMultiStatusCodeEnabled = true
		} else if !validDestinationType(destination.Type) {
			responses = append(responses, PostResponse{destination.Name, http.StatusBadRequest, "Invalid destination type"})
			isMultiStatusCodeEnabled = true
		} else {
			h.destinations = append(h.destinations, destination)
			h.destinationsSensitive[destination.Name] = []byte(fmt.Sprintf(sensitiveDataTemplate, uuid.NewString(),
				destination.Name, destination.Type, destination.Name))

			responses = append(responses, PostResponse{destination.Name, http.StatusCreated, ""})
		}
	}

	if !isMultiStatusCodeEnabled {
		writer.WriteHeader(http.StatusCreated)
		return
	}

	responseJSON, err := json.Marshal(responses)
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

	writer.WriteHeader(http.StatusMultiStatus)
}
