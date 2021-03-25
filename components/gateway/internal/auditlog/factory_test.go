package auditlog_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/automock"
	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	"github.com/stretchr/testify/assert"
)

func TestMessageFactory(t *testing.T) {
	t.Run("Security Event", func(t *testing.T) {
		expected := model.SecurityEvent{User: "user", Metadata: model.Metadata{
			UUID:   TestMsgID,
			Tenant: TestTenant,
			Time:   Timestamp_text,
		}}
		timestamp := time.Date(2020, 3, 17, 12, 37, 44, 1093, time.FixedZone("test", 3600))
		uuidSvc, timeSvc := initMocks(TestMsgID, timestamp)

		factory := auditlog.NewMessageFactory("user", TestTenant, uuidSvc, timeSvc)
		//WHEN
		output := factory.CreateSecurityEvent()

		//THEN
		assert.Equal(t, expected, output)
	})

	t.Run("configuration change", func(t *testing.T) {
		expected := model.ConfigurationChange{User: "user", Metadata: model.Metadata{
			UUID:   TestMsgID,
			Time:   Timestamp_text,
			Tenant: TestTenant,
		}}
		timestamp := time.Date(2020, 3, 17, 12, 37, 44, 1093, time.FixedZone("test", 3600))
		uuidSvc, timeSvc := initMocks(TestMsgID, timestamp)

		factory := auditlog.NewMessageFactory("user", TestTenant, uuidSvc, timeSvc)
		//WHEN
		output := factory.CreateConfigurationChange()

		//THEN
		assert.Equal(t, expected, output)
	})
}

func initMocks(msgID string, timestamp time.Time) (auditlog.UUIDService, auditlog.TimeService) {
	uuidSvc := &automock.UUIDService{}
	uuidSvc.On("Generate").Return(msgID).Once()

	timeSvc := &automock.TimeService{}
	timeSvc.On("Now").Return(timestamp).Once()
	return uuidSvc, timeSvc
}
