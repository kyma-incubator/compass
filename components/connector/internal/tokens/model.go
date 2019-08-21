package tokens

type TokenType string

const (
	ApplicationToken TokenType = "Application"
	RuntimeToken     TokenType = "Runtime"
)

type TokenData struct {
	Type     TokenType
	ClientId string
}
