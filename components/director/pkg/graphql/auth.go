package graphql

import (
	"encoding/json"
)

type credential struct {
	*BasicCredentialData
	*OAuthCredentialData
}

// UnmarshalJSON is used only by integration tests, we have to help graphql client to deal with Credential field
func (a *Auth) UnmarshalJSON(data []byte) error {
	type Alias Auth
	aux := &struct {
		*Alias
		Credential credential `json:"credential"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	a.Credential = retrieveCredential(aux.Credential)

	return nil
}

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

func retrieveCredential(umarshaledCredential credential) CredentialData {
	if umarshaledCredential.BasicCredentialData != nil {
		return umarshaledCredential.BasicCredentialData
	} else {
		return umarshaledCredential.OAuthCredentialData
	}
}
