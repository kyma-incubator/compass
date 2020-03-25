package auditlog_test

import (
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
)

const (
	TestTenant     = "576b2a22-df6b-454c-99f4-ba12b089a053"
	TestMsgID      = "623d764a-b059-4010-869f-d6d276110a8c"
	User           = "proxy"
	Timestamp_text = "2020-03-17T12:37:44Z"
)

func fixFabricatedConfigChangeMsg() model.ConfigurationChange {
	return model.ConfigurationChange{User: User, AuditlogMetadata: model.AuditlogMetadata{
		Tenant: TestTenant,
		UUID:   TestMsgID,
		Time:   Timestamp_text,
	}}
}

func fixFabricatedSecurityEventMsg() model.SecurityEvent {
	return model.SecurityEvent{User: User, AuditlogMetadata: model.AuditlogMetadata{
		Tenant: TestTenant,
		UUID:   TestMsgID,
	}}
}

func fixSuccessConfigChangeMsg(claims proxy.Claims, request, response string) model.ConfigurationChange {
	msg := fixFabricatedConfigChangeMsg()
	msg.Object = model.Object{
		ID: map[string]string{
			"name":           "Config Change",
			"externalTenant": claims.Tenant,
			"apiConsumer":    claims.ConsumerType,
			"consumerID":     claims.ConsumerID,
		}}
	msg.Attributes = []model.Attribute{
		{Name: "request", Old: "", New: request},
		{Name: "response", Old: "", New: response}}

	return msg
}

func fixFilledConfigChangeMsg() model.ConfigurationChange {
	msg := fixFabricatedConfigChangeMsg()
	msg.Object = model.Object{
		ID: map[string]string{
			"name":           "Config Change",
			"externalTenant": "external tenant",
			"apiConsumer":    "application",
			"consumerID":     "consumerID",
		},
	}
	msg.Attributes = []model.Attribute{{Name: "name", Old: "", New: "new value"}}
	return msg
}

func fixFilledSecurityEventMsg() model.SecurityEvent {
	msg := fixFabricatedSecurityEventMsg()
	msg.Data = "test-data"

	return msg
}
