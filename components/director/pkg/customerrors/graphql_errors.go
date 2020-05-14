package customerrors

type GraphqlError struct {
	StatusCode StatusCode    `json:"status_code"`
	Message    string `json:"message"`
}

func (g GraphqlError) Error() string {
	return g.Message
}
