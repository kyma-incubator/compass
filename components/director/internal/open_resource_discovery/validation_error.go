package ord

const (
	ErrorSeverity   = "error"
	WarningSeverity = "warning"
)

type ValidationError struct {
	OrdId       string `json:"ordId"`
	Severity    string `json:"severity"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
