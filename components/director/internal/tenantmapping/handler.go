package tenantmapping

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		log.Infof("Bad request method. Got %s, expected POST", request.Method)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("===")
	log.Printf("Headers:\n%+v\n\n", request.Header)

	writer.Header().Set("Content-Type", "application/json")

	var data Data
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		log.Error(errors.Wrap(err, "while decoding request body"))
		return
	}

	defer func() {
		err := request.Body.Close()
		if err != nil {
			log.Error(errors.Wrap(err, "while closing body"))
		}
	}()

	log.Printf("Input:\n%+v\n\n", data)

	if data.Extra == nil {
		data.Extra = make(map[string]interface{})
	}

	extraMap, ok := data.Extra.(map[string]interface{})
	if !ok {
		log.Error(fmt.Errorf("Incorrect type %T; expected map[string]interface{}\n", data.Extra))
		return
	}

	extraMap["tenant"] = "9ac609e1-7487-4aa6-b600-0904b272b11f"
	data.Extra = extraMap

	log.Printf("Output:\n%+v\n\n", data)
	log.Println("===")

	err = json.NewEncoder(writer).Encode(data)
	if err != nil {
		log.Error(errors.Wrap(err, "while encoding data"))
		return
	}

	err = request.Write(writer)
	if err != nil {
		log.Error(errors.Wrap(err, "while writing request"))
		return
	}
}
