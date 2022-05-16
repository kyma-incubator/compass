package auditlog

import (
	"time"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
)

//go:generate mockery --name=UUIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UUIDService interface {
	Generate() string
}

//go:generate mockery --name=TimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
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
	logTime := t.Format(model.LogFormatDate)

	return model.ConfigurationChange{User: f.user,
		Metadata: model.Metadata{Tenant: f.tenant,
			Time: logTime,
			UUID: f.uuidSvc.Generate(),
		}}
}

func (f *MessageFactory) CreateSecurityEvent() model.SecurityEvent {
	t := f.timeSvc.Now()
	logTime := t.Format(model.LogFormatDate)

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
