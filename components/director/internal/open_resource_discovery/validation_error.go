package ord

const (
	// ErrorSeverity is one of the severity levels of a validation error
	ErrorSeverity = "error"
)

// ValidationError represents a validation error when aggregating and validating ORD documents
type ValidationError struct {
	OrdID       string `json:"ordId"`
	Severity    string `json:"severity"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
