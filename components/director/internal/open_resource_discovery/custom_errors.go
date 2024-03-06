package ord

const (
	// ErrorSeverity is one of the severity levels of a validation error
	ErrorSeverity = "error"
	// WarningSeverity is one of the severity levels of a validation error
	WarningSeverity = "warning"
)

// ValidationError represents a validation error when aggregating and validating ORD documents
type ValidationError struct {
	OrdID       string `json:"ordId"`
	Severity    string `json:"severity"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// RuntimeError represents the message of the runtime errors
type RuntimeError struct {
	Message string `json:"message"`
}

// ProcessingError represents the error containing the validation and runtime errors from processing an operation
type ProcessingError struct {
	ValidationErrors []*ValidationError `json:"validation_errors"`
	RuntimeError     *RuntimeError      `json:"runtime_error"`
}
