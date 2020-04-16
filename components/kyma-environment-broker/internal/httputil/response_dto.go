package httputil

// ErrorDTO represents a json returned in case of error.
// Only Status and RequestID fields are required.
type ErrorDTO struct {
	// Status is original HTTP error code
	Status int `json:"status"`

	// RequestID defines the id of the incoming request
	RequestID string `json:"requestId"`

	// Message is descriptive human-readable error
	Message string `json:"message,omitempty"`

	// Details about the error
	Details string `json:"details,omitempty"`
}
