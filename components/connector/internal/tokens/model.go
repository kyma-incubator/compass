package tokens

func NewCSRTokenResponse(token string) CSRTokenResponse {
	return CSRTokenResponse{
		tokenResponse{
			TokenValue: token,
		},
	}
}

func (r CSRTokenResponse) GetTokenValue() string {
	return r.Token.TokenValue
}

type CSRTokenResponse struct {
	Token tokenResponse `json:"generateCSRToken"`
}

type tokenResponse struct {
	TokenValue string `json:"token"`
}
