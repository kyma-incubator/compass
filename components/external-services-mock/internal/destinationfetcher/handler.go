package destinationfetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	pageCountQueryParameter = "$pageCount"
	pageQueryParameter      = "$page"
	pageSizeQueryParameter  = "$pageSize"
)

type Destination struct {
	Name string
	Type string
}

type Response struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
	Cause  string `json:"cause,omitempty"`
}

type Handler struct {
	destinations          []Destination
	destinationsSensitive map[string][]byte
}

func NewHandler() *Handler {
	return &Handler{}
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
		writer.Write([]byte("[]"))
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

func validDestinationType(destinationType string) bool {
	return destinationType == "HTTP" || destinationType == "RFC" || destinationType == "MAIL" || destinationType == "LDAP"
}

func (h *Handler) PostDestination(writer http.ResponseWriter, req *http.Request) {
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

	var responses []Response
	if len(h.destinationsSensitive) == 0 {
		h.destinationsSensitive = make(map[string][]byte)
	}

	for _, destination := range destinations {
		if _, ok := h.destinationsSensitive[destination.Name]; ok {
			responses = append(responses, Response{destination.Name, http.StatusConflict, "Destination name already taken"})
		} else if !validDestinationType(destination.Type) {
			responses = append(responses, Response{destination.Name, http.StatusBadRequest, "Invalid destination type"})
		} else {
			h.destinations = append(h.destinations, destination)
			h.destinationsSensitive[destination.Name] = []byte(fmt.Sprintf(sensitiveDataTemplate, uuid.NewString(),
				destination.Name, destination.Type, destination.Name))

			responses = append(responses, Response{destination.Name, http.StatusCreated, ""})
		}
	}

	responseJSON, err := json.Marshal(responses)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to marshal response body")
		http.Error(writer, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	writer.Write(responseJSON)
}
