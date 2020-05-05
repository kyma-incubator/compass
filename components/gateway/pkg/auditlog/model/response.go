package model

type SuccessResponse struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
