package httputil

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"
)

// ErrorResponse missing godoc
type ErrorResponse struct {
	Error error `json:"error"`
}

func GetTimeoutMessage() []byte {
	marshal, err := json.Marshal(ErrorResponse{Error: Error{
		Code:    http.StatusRequestTimeout,
		Message: "timeout",
	}})
	if err != nil {
		log.D().Errorf("while marshaling error message  %s",  err.Error())
	}

	return marshal
}

// Error missing godoc
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

// Error missing godoc
type DetailedError struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Details []Detail `json:"details"`
}

func (d DetailedError) Error() string {
	return d.Message
}

type Detail struct {
	Message    string `json:"message"`
	Subaccount string `json:"subaccount"`
	LocationId string `json:"locationId"`
}
