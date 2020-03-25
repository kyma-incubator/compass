package auditlog_test

import (
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/automock"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	timestamp_text = "2020-03-17T12:37:44Z"
	TestMsgId      = "e653cd33-6495-4f97-87db-8c7613df4f82"
)

func TestMessageFactory(t *testing.T) {
	t.Run("Security Event", func(t *testing.T) {
		expected := model.SecurityEvent{User: "user", AuditlogMetadata: model.AuditlogMetadata{
			UUID:   TestMsgId,
			Tenant: TestTenant,
			Time:   timestamp_text,
		}}
		timestamp := time.Date(2020, 3, 17, 12, 37, 44, 1093, time.FixedZone("test", 3600))
		uuidSvc, timeSvc := initMocks(TestMsgId, timestamp)

		factory := auditlog.BasicAuthMessageFactory("user", TestTenant, uuidSvc, timeSvc)
		//WHEN
		output := factory.CreateSecurityEvent()

		//THEN
		assert.Equal(t, expected, output)
	})

	t.Run("Configuration change", func(t *testing.T) {
		expected := model.ConfigurationChange{User: "user", AuditlogMetadata: model.AuditlogMetadata{
			UUID:   TestMsgId,
			Time:   timestamp_text,
			Tenant: TestTenant,
		}}
		timestamp := time.Date(2020, 3, 17, 12, 37, 44, 1093, time.FixedZone("test", 3600))
		uuidSvc, timeSvc := initMocks(TestMsgId, timestamp)

		factory := auditlog.BasicAuthMessageFactory("user", TestTenant, uuidSvc, timeSvc)
		//WHEN
		output := factory.CreateConfigurationChange()

		//THEN
		assert.Equal(t, expected, output)
	})

	t.Run("Configuration change OAuth factory", func(t *testing.T) {
		expected := model.ConfigurationChange{User: "$USER", AuditlogMetadata: model.AuditlogMetadata{
			UUID:   TestMsgId,
			Time:   timestamp_text,
			Tenant: "$PROVIDER",
		}}
		timestamp := time.Date(2020, 3, 17, 12, 37, 44, 1093, time.FixedZone("test", 3600))
		uuidSvc, timeSvc := initMocks(TestMsgId, timestamp)

		factory := auditlog.OAuthMessageFactory(uuidSvc, timeSvc)
		//WHEN
		output := factory.CreateConfigurationChange()

		//THEN
		assert.Equal(t, expected, output)
	})
}

//TODO: use it in factory tests
func initMocks(msgID string, timestamp time.Time) (auditlog.UUIDService, auditlog.TimeService) {
	uuidSvc := &automock.UUIDService{}
	uuidSvc.On("Generate").Return(msgID).Once()

	timeSvc := &automock.TimeService{}
	timeSvc.On("Now").Return(timestamp).Once()
	return uuidSvc, timeSvc
}
