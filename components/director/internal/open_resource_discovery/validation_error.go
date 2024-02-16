package ord

const (
	ErrorSeverity = "error"
)

type ValidationError struct {
	OrdId       string `json:"ordId"`
	Severity    string `json:"severity"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
