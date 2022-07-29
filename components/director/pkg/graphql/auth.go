package graphql

import (
	"encoding/json"
)

type credential struct {
	*BasicCredentialData
	*OAuthCredentialData
	*CertificateOAuthCredentialData
}

// OneTimeTokenDTO this a model for transportation of one-time tokens, because the json marshaller cannot unmarshal to either of the types OTTForApp or OTTForRuntime
type OneTimeTokenDTO struct {
	TokenWithURL
	LegacyConnectorURL string `json:"legacyConnectorURL"`
}

// oneTimeTokenDTO is used to hyde TypeName property to consumers
type oneTimeTokenDTO struct {
	*OneTimeTokenDTO
	TypeName string `json:"__typename"`
}

// IsOneTimeToken implements the interface OneTimeToken
func (*OneTimeTokenDTO) IsOneTimeToken() {}

// UnmarshalJSON is used only by integration tests, we have to help graphql client to deal with Credential field
func (a *Auth) UnmarshalJSON(data []byte) error {
	type Alias Auth

	aux := &struct {
		*Alias
		Credential   credential       `json:"credential"`
		OneTimeToken *oneTimeTokenDTO `json:"oneTimeToken"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	a.Credential = retrieveCredential(aux.Credential)
	if aux.OneTimeToken != nil {
		a.OneTimeToken = retrieveOneTimeToken(aux.OneTimeToken)
	}

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
	} else if unmarshaledCredential.OAuthCredentialData != nil {
		return unmarshaledCredential.OAuthCredentialData
	}

	return unmarshaledCredential.CertificateOAuthCredentialData
}

func retrieveOneTimeToken(ottDTO *oneTimeTokenDTO) OneTimeToken {
	switch ottDTO.TypeName {
	case "OneTimeTokenForApplication":
		return &OneTimeTokenForApplication{
			TokenWithURL:       ottDTO.TokenWithURL,
			LegacyConnectorURL: ottDTO.LegacyConnectorURL,
		}
	case "OneTimeTokenForRuntime":
		return &OneTimeTokenForRuntime{
			TokenWithURL: ottDTO.TokenWithURL,
		}
	}
	return ottDTO.OneTimeTokenDTO
}
