package tenantmapping

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Data struct {
	Subject string      `json:"subject"`
	Extra   interface{} `json:"extra"`
	Header  interface{} `json:"header"`
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Bad request method. Got %s, expected POST", http.StatusBadRequest)
		return
	}

	logBuilder := strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("\nHeaders: %+v", request.Header))

	writer.Header().Set("Content-Type", "application/json")

	var data Data
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		if err == io.EOF {
			http.Error(writer, "Request body is empty", http.StatusBadRequest)
			return
		}

		wrappedErr := errors.Wrap(err, "while decoding request body")
		http.Error(writer, wrappedErr.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		err := request.Body.Close()
		if err != nil {
			wrappedErr := errors.Wrap(err, "while decoding request body")
			log.Error(wrappedErr)
			http.Error(writer, wrappedErr.Error(), http.StatusInternalServerError)
		}
	}()

	logBuilder.WriteString(fmt.Sprintf("\nInput: %+v", data))

	if data.Extra == nil {
		data.Extra = make(map[string]interface{})
	}

	extraMap, ok := data.Extra.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("Incorrect type %T; expected map[string]interface{}\n", data.Extra)
		log.Info(logBuilder.String())
		log.Error(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	extraMap["tenant"] = "9ac609e1-7487-4aa6-b600-0904b272b11f"

	if extraMap["scope"] != nil {
		scopesArray, ok := extraMap["scope"].([]interface{})
		if !ok || len(scopesArray) == 0 {
			h.setScopes(extraMap)
		}
	} else {
		h.setScopes(extraMap)
	}

	data.Extra = extraMap

	logBuilder.WriteString(fmt.Sprintf("\nOutput: %+v\n", data))
	log.Info(logBuilder.String())

	err = json.NewEncoder(writer).Encode(data)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while encoding data")
		log.Error(wrappedErr)
		http.Error(writer, wrappedErr.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) setScopes(extraMap map[string]interface{}) {
	extraMap["scope"] = []string{
		"application:read",
		"application:write",
		"runtime:read",
		"runtime:write",
		"label_definition:read",
		"label_definition:write",
		"health_checks:read",
	}
}
