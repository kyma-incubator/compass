package configurationchange

import (
	"strings"
	"sync"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
)

type Service struct {
	configLogs map[string]model.ConfigurationChange
	mutex      sync.RWMutex
}

func NewService() *Service {
	return &Service{
		configLogs: make(map[string]model.ConfigurationChange),
		mutex:      sync.RWMutex{},
	}
}
func (s *Service) Save(change model.ConfigurationChange) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.configLogs[change.UUID] = change

	return change.UUID, nil
}

func (s *Service) Get(id string) *model.ConfigurationChange {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	val, ok := s.configLogs[id]
	if ok {
		return &val
	}
	return nil
}

func (s *Service) List() []model.ConfigurationChange {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var auditLogs []model.ConfigurationChange
	for _, v := range s.configLogs {
		auditLogs = append(auditLogs, v)
	}
	return auditLogs
}

func (s *Service) SearchByString(searchString string) []model.ConfigurationChange {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	configurationChangeSet := make(map[string]model.ConfigurationChange)
	for _, auditLog := range s.configLogs {
		for _, attribute := range auditLog.Attributes {
			if strings.Contains(attribute.New, searchString) {
				configurationChangeSet[auditLog.UUID] = auditLog
			}
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
