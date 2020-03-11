package auditlog

type GraphqlResponse struct {
	Errors  []ErrorMessage `json:"errors"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
}

type ErrorMessage struct {
	Message string   `json:"message"`
	Path    []string `json:"path"`
}
