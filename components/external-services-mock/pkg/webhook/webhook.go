package webhook

import (
	"encoding/json"
	"net/http"
)

type OKRequestData struct {
	OK bool
}

type OperationResponseData struct {
	Status string `json:"status"`
}

const (
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

func NewWebHookOperationHTTPHandler() func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if isOk {
				operationResponseDataOK := OperationResponseData{
					Status: OperationResponseStatusOK,
				}
				operationResponseDataOKJSON, err := json.Marshal(operationResponseDataOK)
				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}

				_, err = rw.Write(operationResponseDataOKJSON)
				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
				rw.WriteHeader(http.StatusOK)
			} else {
				operationResponseIPData := OperationResponseData{
					Status: OperationResponseStatusINProgress,
				}
				operationResponseDataIPJSON, err := json.Marshal(operationResponseIPData)
				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}

				_, err = rw.Write(operationResponseDataIPJSON)
				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
				rw.WriteHeader(http.StatusOK)
			}
		} else if r.Method == http.MethodPost {
			var okReqData OKRequestData
			err := json.NewDecoder(r.Body).Decode(&okReqData)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			isOk = okReqData.OK
			rw.WriteHeader(http.StatusOK)
		}
	}
}
