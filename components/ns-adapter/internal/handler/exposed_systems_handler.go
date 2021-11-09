package handler

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/ns-adapter/internal/httputil"
	"github.com/kyma-incubator/compass/components/ns-adapter/internal/model"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

func NewHandler() *Handler {
	return &Handler{}
}

type Handler struct {
}

//Description	Upsert ExposedSystems is a bulk create-or-update operation on exposed on-premise systems. It takes a list of fully described exposed systems, creates the ones CMP isn't aware of and updates the metadata for the ones it is.
//URL	/api/v1/notifications
//Path Params
//Query Params	reportType=full,delta
//HTTP Method	PUT
//HTTP Headers
//Content-Type: application/json
//HTTP Codes
//204 No Content
//200 OK
//400 Bad Request
//401 Unauthorized
//408 Request Timeout
//500 Internal Server Error
//502 Bad Gateway
//Response Formats	json
//Authentication	TODO
func (a *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := log.C(ctx)

	time.Sleep(time.Minute)
	defer func() {
		if err := req.Body.Close(); err != nil {
			logger.Error("Got error on closing request body", err)
		}
	}()

	decoder := json.NewDecoder(req.Body)
	var reqData model.Report
	err := decoder.Decode(&reqData)
	if err != nil {
		logger.Warnf("Got error on decoding Request Body: %v\n", err)
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, model.NewNSError("failed to parse request body"))
		return
	}

	if err := reqData.Validate(); err != nil {
		logger.Warnf("Got error while validating Request Body: %v\n", err)
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, model.NewNSError(err.Error()))
		return
	}

	reportType := req.URL.Query().Get("reportType")

	if reportType == "" {
		httputil.RespondWithError(ctx, rw, http.StatusBadRequest, model.NewNSError("missing required report type query parameter"))
		return
	}

	if reportType == "full" {

	} else if reportType == "delta" {

	} else {
		//TODO log unknown type
	}

	if err := json.NewEncoder(rw).Encode(model.NewNSError("response test")); err != nil {
		logger.Warnf("Got error on encoding response: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
