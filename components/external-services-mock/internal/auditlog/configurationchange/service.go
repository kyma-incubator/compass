package configurationchange

import (
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
)

type timestampedConfigurationChange struct {
	model.ConfigurationChange
	CreatedAt time.Time
}

type Service struct {
	configLogs map[string]timestampedConfigurationChange
	mutex      sync.RWMutex
}

func NewService() *Service {
	return &Service{
		configLogs: make(map[string]timestampedConfigurationChange),
		mutex:      sync.RWMutex{},
	}
}
func (s *Service) Save(change model.ConfigurationChange) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.configLogs[change.UUID] = timestampedConfigurationChange{
		ConfigurationChange: change,
		CreatedAt:           time.Now().UTC(),
	}

	return change.UUID, nil
}

func (s *Service) Get(id string) *model.ConfigurationChange {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	val, ok := s.configLogs[id]
	if ok {
		return &val.ConfigurationChange
	}
	return nil
}

func (s *Service) List() []model.ConfigurationChange {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var auditLogs []model.ConfigurationChange
	for _, v := range s.configLogs {
		auditLogs = append(auditLogs, v.ConfigurationChange)
	}
	return auditLogs
}

func (s *Service) SearchByTimestamp(timeFrom, timeTo time.Time) []model.ConfigurationChange {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	configurationChangeSet := make(map[string]model.ConfigurationChange)
	for _, auditLog := range s.configLogs {
		if auditLog.CreatedAt.After(timeFrom) && auditLog.CreatedAt.Before(timeTo) {
			configurationChangeSet[auditLog.UUID] = auditLog.ConfigurationChange
		}
	}

	var out []model.ConfigurationChange
	for _, auditLog := range configurationChangeSet {
		out = append(out, auditLog)
	}

	return out
}

func (s *Service) Delete(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.configLogs, id)
}
