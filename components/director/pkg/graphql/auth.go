package graphql

import (
	"encoding/json"
	"reflect"
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

	cred, err := retrieveCredential(data)
	if err != nil {
		return err
	}

	a.Credential = cred

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

	cred, err := retrieveCredential(data)
	if err != nil {
		return err
	}

	csrf.Credential = cred

	return nil
}

// retrieveCredential checks if any of the structs BasicCredentialData, OAuthCredentialData or CertificateOAuthCredentialData
// has all the data after unmarshalling. The resulting CredentialData is the one struct that has all it's fields full of data
// This is done because CredentialData is an interface and the structs that implement it have conflicting json tag names, so
// they could not be marshalled properly
func retrieveCredential(data []byte) (CredentialData, error) {
	var basicCredential struct {
		BasicCredentialData `json:"credential"`
	}
	if err := json.Unmarshal(data, &basicCredential); err != nil {
		return nil, err
	}

	if isBasic := isCredentialStructFullWithData(basicCredential.BasicCredentialData); isBasic {
		return &basicCredential.BasicCredentialData, nil
	}

	var oauthCredential struct {
		OAuthCredentialData `json:"credential"`
	}
	if err := json.Unmarshal(data, &oauthCredential); err != nil {
		return nil, err
	}

	if isOAuth := isCredentialStructFullWithData(oauthCredential.OAuthCredentialData); isOAuth {
		return &oauthCredential.OAuthCredentialData, nil
	}

	var certOAuthCredential struct {
		CertificateOAuthCredentialData `json:"credential"`
	}
	if err := json.Unmarshal(data, &certOAuthCredential); err != nil {
		return nil, err
	}

	if isCertificateOAuth := isCredentialStructFullWithData(certOAuthCredential.CertificateOAuthCredentialData); isCertificateOAuth {
		return &certOAuthCredential.CertificateOAuthCredentialData, nil
	}

	return nil, nil
}

// isCredentialStructFullWithData checks if any of the fields in the struct is same as an empty struct property or is nil and
// if this tests positive then it is considered that the struct is not in a valid form will properties full of data
func isCredentialStructFullWithData(obj interface{}) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	for i := 0; i < v.NumField(); i++ {
		empty := reflect.New(v.Field(i).Type()).Elem().Interface()
		value := v.Field(i).Interface()
		if reflect.DeepEqual(value, empty) || v.Field(i).Interface() == nil {
			return false
		}
	}

	return true
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
