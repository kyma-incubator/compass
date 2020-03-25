package auditlog_test

import (
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
)

const (
	TestTenant = "576b2a22-df6b-454c-99f4-ba12b089a053"
	UUID       = "623d764a-b059-4010-869f-d6d276110a8c"
	User       = "proxy"
)

func fixFactoryMessage() model.ConfigurationChange {
	return model.ConfigurationChange{User: User, AuditlogMetadata: model.AuditlogMetadata{
		Tenant: TestTenant,
		UUID:   UUID,
	}}
}

func fixFactorySecurityEvent() model.SecurityEvent {
	return model.SecurityEvent{User: User, AuditlogMetadata: model.AuditlogMetadata{
		Tenant: TestTenant,
		UUID:   UUID,
	}}
}

func fixLogSuccess(claims proxy.Claims, request, response string) model.ConfigurationChange {
	msg := fixFactoryMessage()
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


