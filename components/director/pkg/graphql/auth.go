package graphql

import (
	"encoding/json"
)

// UnmarshalJSON is used only by integration tests, we have to help graphql client to deal with Credential field
func (a *Auth) UnmarshalJSON(data []byte) error {
	type Alias Auth
	aux := &struct {
		*Alias
		Credential struct {
			*BasicCredentialData
			*OAuthCredentialData
		} `json:"credential"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Credential.BasicCredentialData != nil {
		a.Credential = aux.Credential.BasicCredentialData
	} else {
		a.Credential = aux.Credential.OAuthCredentialData
	}

	return nil
}

func (csrf *CSRFTokenCredentialRequestAuth) UnmarshalJSON(data []byte) error {
	type Alias CSRFTokenCredentialRequestAuth

	aux := &struct {
		*Alias
		Credential struct {
			*BasicCredentialData
			*OAuthCredentialData
		} `json:"credential"`
	}{
		Alias: (*Alias)(csrf),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Credential.BasicCredentialData != nil {
		csrf.Credential = aux.Credential.BasicCredentialData
	} else {
		csrf.Credential = aux.Credential.OAuthCredentialData
	}

	return nil
}
