package validator

const (
	ErrorSeverity   = "error"
	WarningSeverity = "warning"
)

type ValidationError struct {
	OrdId       string `json:"ord-id"`
	Severity    string `json:"severity"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
