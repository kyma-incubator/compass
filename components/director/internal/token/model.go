package token

type ExternalTokenModel struct {
	GenerateRuntimeToken ExternalRuntimeToken `json:"generateRuntimeToken"`
}

type ExternalRuntimeToken struct {
	Token string `json:"token"`
}
