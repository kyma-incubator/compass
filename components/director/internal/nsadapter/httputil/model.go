package httputil

import (
	"encoding/json"
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
		//TODO
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
	Details []Details `json:"details"`
}

func (d DetailedError) Error() string {
	return d.Message
}

type Detail struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Subaccount string `json:"subaccount"`
	LocationId string `json:"locationId"`
}
