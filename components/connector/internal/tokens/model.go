package tokens

func NewTokenResponse(token string) TokenResponse {
	return TokenResponse{
		tokenResponse{
			TokenValue: token,
		},
	}
}

func (r TokenResponse) GetTokenValue() string {
	return r.Token.TokenValue
}

type TokenResponse struct {
	Token tokenResponse `json:"result"`
}

type tokenResponse struct {
	TokenValue string `json:"token"`
}
