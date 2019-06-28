package graphql

import (
	"encoding/json"
)

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

type ApplicationExt struct {
	Application
	Webhooks  []ApplicationWebhook
	Apis      APIDefinitionPage
	EventAPIs EventAPIDefinitionPage
	Documents DocumentPage
}
