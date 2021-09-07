package httputils

// ErrorResponse missing godoc
type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

// Error missing godoc
type Error struct {
	Message string `json:"message"`
}
