package token

type ExternalTokenModel struct {
	AppToken     ExternalRuntimeToken `json:"generateApplicationToken"`
	RuntimeToken ExternalRuntimeToken `json:"generateRuntimeToken"`
}

type ExternalRuntimeToken struct {
	Token string `json:"token"`
}

func (t *ExternalTokenModel) Token(tokenType TokenType) string {
	switch tokenType {
	case ApplicationToken:
		return t.AppToken.Token
	case RuntimeToken:
		return t.RuntimeToken.Token
	}
	return ""
}

type TokenType string

const (
	RuntimeToken     TokenType = "Runtime"
	ApplicationToken TokenType = "Application"
)
