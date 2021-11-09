package httputil

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse missing godoc
type ErrorResponse struct {
	Error Error `json:"error"`
}

func GetTimeoutMessage() []byte{
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
