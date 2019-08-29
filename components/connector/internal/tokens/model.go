package tokens

type TokenType string

const (
	ApplicationToken TokenType = "Application"
	RuntimeToken     TokenType = "Runtime"
	CSRToken         TokenType = "Certificate"
)

type TokenData struct {
	Type     TokenType
	ClientId string
}
