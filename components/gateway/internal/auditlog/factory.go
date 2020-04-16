package auditlog

import (
	"time"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
)

const (
	LogFormatDate  = "2006-01-02T15:04:05.999Z"
	UserVariable   = "$USER"
	TenantVariable = "$PROVIDER"
)

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
		Metadata: model.Metadata{Tenant: f.tenant,
			Time: logTime,
			UUID: f.uuidSvc.Generate(),
		}}
}

func (f *MessageFactory) CreateSecurityEvent() model.SecurityEvent {
	t := f.timeSvc.Now()
	logTime := t.Format(LogFormatDate)

	return model.SecurityEvent{User: f.user,
		Metadata: model.Metadata{Tenant: f.tenant,
			Time: logTime,
			UUID: f.uuidSvc.Generate(),
		}}
}

func NewMessageFactory(user, tenant string, uuidSvc UUIDService, timeSvc TimeService) *MessageFactory {
	return &MessageFactory{
		user:    user,
		tenant:  tenant,
		uuidSvc: uuidSvc,
		timeSvc: timeSvc,
	}
}
