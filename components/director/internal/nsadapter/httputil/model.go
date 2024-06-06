package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// ErrorResponse on failed notification service request
type ErrorResponse struct {
	Error error `json:"error"`
}

// UnmarshalJSON implements Unmarshaler interface. The method unmarshal the data from b into ErrorResponse structure.
func (e *ErrorResponse) UnmarshalJSON(b []byte) error {
	code, success := gjson.Get(string(b), "error.code").Value().(float64)
	if !success {
		return errors.New("failed to parse code")
	}

	msg, success := gjson.Get(string(b), "error.message").Value().(string)
	if !success {
		return errors.New("failed to parse message")
	}

	e.Error = Error{
		Code:    int(code),
		Message: msg,
	}
	return nil
}

// Error indicates processing failure of on-premise systems
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

// DetailedError indicates partial processing failure of on-premise systems
type DetailedError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []Detail `json:"details"`
}

func (d DetailedError) Error() string {
	return d.Message
}

// Detail contains error details for scc
type Detail struct {
	Message    string `json:"message"`
	Subaccount string `json:"subaccount"`
	LocationID string `json:"locationId"`
}

// GetTimeoutMessage returns ErrorResponse with status code 408 and message "timeout"
func GetTimeoutMessage() []byte {
	marshal, err := json.Marshal(ErrorResponse{Error: Error{
		Code:    http.StatusRequestTimeout,
		Message: "timeout",
	}})
	if err != nil {
		log.D().Errorf("while marshaling error message  %s", err.Error())
	}

	return marshal
}
