package webhook

import (
	"encoding/json"
	"net/http"
)

type OperationStatusRequestData struct {
	OK bool
}

type OperationResponseData struct {
	Status string `json:"status"`
}

const (
	OperationPath                     = "webhook/delete/operation"
	DeletePath                        = "webhook/delete"
	OperationResponseStatusOK         = "SUCCEEDED"
	OperationResponseStatusINProgress = "IN_PROGRESS"
)

var isOk bool

func NewDeleteHTTPHandler() func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		if isOk {
			isOk = false
			rw.WriteHeader(http.StatusOK)
		} else {
			rw.WriteHeader(http.StatusLocked)
		}
	}
}

func NewWebHookOperationPostHTTPHandler() func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var okReqData OperationStatusRequestData
		err := json.NewDecoder(r.Body).Decode(&okReqData)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		isOk = okReqData.OK
		rw.WriteHeader(http.StatusOK)
	}
}

func NewWebHookOperationGetHTTPHandler() func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body := OperationResponseData{
			Status: OperationResponseStatusINProgress,
		}
		if isOk {
			body.Status = OperationResponseStatusOK
		}

		operationResponseDataJSON, err := json.Marshal(body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = rw.Write(operationResponseDataJSON)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}
