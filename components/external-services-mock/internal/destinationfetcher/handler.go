package destinationfetcher

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetSensitiveData(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	destinationName := mux.Vars(req)["name"]
	log.C(ctx).Infof("Sending sensitive data of destination: %s", destinationName)
	data, ok := destinationsSensitiveData[destinationName]

	if !ok {
		http.Error(writer, "Not Found", http.StatusNotFound)
		return
	}

	if _, err := writer.Write(data); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
	}
}

func (h *Handler) GetSubaccountDestinationsPage(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	writer.Header().Set("Page-Count", "1")
	if _, err := writer.Write(destinations); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write data")
	}
}
