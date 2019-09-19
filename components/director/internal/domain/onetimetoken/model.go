package onetimetoken

import "github.com/kyma-incubator/compass/components/director/internal/model"

type ConnectorTokenModel struct {
	AppToken     ConnectorToken `json:"generateApplicationToken"`
	RuntimeToken ConnectorToken `json:"generateRuntimeToken"`
}

type ConnectorToken struct {
	Token string `json:"token"`
}

func (t *ConnectorTokenModel) Token(tokenType model.SystemAuthReferenceObjectType) string {
	switch tokenType {
	case model.ApplicationReference:
		return t.AppToken.Token
	case model.RuntimeReference:
		return t.RuntimeToken.Token
	}
	return ""
}
