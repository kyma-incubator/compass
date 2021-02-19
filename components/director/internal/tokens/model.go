package tokens

type TokenType string

const (
	ApplicationToken TokenType = "Application"
	RuntimeToken     TokenType = "Runtime"
	CSRToken         TokenType = "Certificate"
)
