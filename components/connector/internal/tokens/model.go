package tokens

type CSRTokenResponse struct {
	Data responseData `json:"data"`
}

func NewCSRTokenResponse(token string) CSRTokenResponse {
	return CSRTokenResponse{
		responseData{
			tokenResponse{
				TokenValue: token,
			},
		},
	}
}

func (r CSRTokenResponse) GetTokenValue() string {
	return r.Data.CSRTokenResponse.TokenValue
}

type responseData struct {
	CSRTokenResponse tokenResponse `json:"generateCSRToken"`
}

type tokenResponse struct {
	TokenValue string `json:"token"`
}
