package model

type GraphqlResponse struct {
	Errors []ErrorMessage `json:"errors"`
	Data   interface{}    `json:"data"`
}

type ErrorMessage struct {
	Message string   `json:"message"`
	Path    []string `json:"path"`
}
