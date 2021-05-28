package webhook

import (
	"encoding/json"
	"net/http"
)

const (
	OperationPath                     = "/webhook/delete/operation"
	DeletePath                        = "/webhook/delete"
	OperationResponseStatusOK         = "SUCCEEDED"
	OperationResponseStatusINProgress = "IN_PROGRESS"
)

var isInProgress = true

type OperationStatusRequestData struct {
	InProgress bool
}

type OperationResponseData struct {
	Status string `json:"status"`
	Error  error  `json:"error"`
}

func NewDeleteHTTPHandler() func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if isInProgress {
			rw.WriteHeader(http.StatusLocked)
		} else {
			isInProgress = true
			rw.WriteHeader(http.StatusOK)
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

		isInProgress = okReqData.InProgress
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
			Status: OperationResponseStatusOK,
			Error:  nil,
		}
		if isInProgress {
			body.Status = OperationResponseStatusINProgress
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
