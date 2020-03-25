package auditlog

import (
	"time"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
)

const LogFormatDate = "2006-01-02T15:04:05.999Z"

//go:generate mockery -name=UUIDService -output=automock -outpkg=automock -case=underscore
type UUIDService interface {
	Generate() string
}

//go:generate mockery -name=TimeService -output=automock -outpkg=automock -case=underscore
type TimeService interface {
	Now() time.Time
}

type MessageFactory struct {
	user    string
	tenant  string
	uuidSvc UUIDService
	timeSvc TimeService
}

func (f *MessageFactory) CreateConfigurationChange() model.ConfigurationChange {
	t := f.timeSvc.Now()
	logTime := t.Format(LogFormatDate)

	return model.ConfigurationChange{User: f.user,
		AuditlogMetadata: model.AuditlogMetadata{Tenant: f.tenant,
			Time: logTime,
			UUID: f.uuidSvc.Generate(),
		}}
}

func (f *MessageFactory) CreateSecurityEvent() model.SecurityEvent {
	t := f.timeSvc.Now()
	logTime := t.Format(LogFormatDate)

	return model.SecurityEvent{User: f.user,
		AuditlogMetadata: model.AuditlogMetadata{Tenant: f.tenant,
			Time: logTime,
			UUID: f.uuidSvc.Generate(),
		}}
}

func OAuthMessageFactory(uuidSvc UUIDService, tsvc TimeService) *MessageFactory {
	return &MessageFactory{
		user:    "$USER",
		tenant:  "$PROVIDER",
		uuidSvc: uuidSvc,
		timeSvc: tsvc,
	}
}

func BasicAuthMessageFactory(user, tenant string, uuidSvc UUIDService, timeSvc TimeService) *MessageFactory {
	return &MessageFactory{
		user:    user,
		tenant:  tenant,
		uuidSvc: uuidSvc,
		timeSvc: timeSvc,
	}
}
