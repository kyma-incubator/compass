package graphql

import (
	"encoding/json"
)

type credential struct {
	*BasicCredentialData
	*OAuthCredentialData
}

type OneTimeTokenDTO struct {
	TokenWithURL
	LegacyConnectorURL string `json:"legacyConnectorURL"`
}

func (*OneTimeTokenDTO) IsOneTimeToken() {}

// UnmarshalJSON is used only by integration tests, we have to help graphql client to deal with Credential field
func (a *Auth) UnmarshalJSON(data []byte) error {
	type Alias Auth

	aux := &struct {
		*Alias
		Credential   credential       `json:"credential"`
		OneTimeToken *OneTimeTokenDTO `json:"oneTimeToken"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	a.Credential = retrieveCredential(aux.Credential)
	a.OneTimeToken = aux.OneTimeToken

	return nil
}

// UnmarshalJSON missing godoc
func (csrf *CSRFTokenCredentialRequestAuth) UnmarshalJSON(data []byte) error {
	type Alias CSRFTokenCredentialRequestAuth

	aux := &struct {
		*Alias
		Credential credential `json:"credential"`
	}{
		Alias: (*Alias)(csrf),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	csrf.Credential = retrieveCredential(aux.Credential)

	return nil
}

func retrieveCredential(unmarshaledCredential credential) CredentialData {
	if unmarshaledCredential.BasicCredentialData != nil {
		return unmarshaledCredential.BasicCredentialData
	}
	return unmarshaledCredential.OAuthCredentialData
}
