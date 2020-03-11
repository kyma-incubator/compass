package security

import (
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/mock-external-test-component/pkg/model"
)

type Service struct {
	configLogs map[string]model.SecurityEvent
}

func NewService() *Service {
	return &Service{configLogs: make(map[string]model.SecurityEvent)}
}
func (s *Service) Save(change model.SecurityEvent) (string, error) {
	id := uuid.New().String()
	s.configLogs[id] = change

	return id, nil
}

func (s *Service) Get(id string) *model.SecurityEvent {
	val, ok := s.configLogs[id]
	if ok {
		return &val
	}
	return nil
}

func (s *Service) List() []model.SecurityEvent {
	var auditLogs []model.SecurityEvent
	for _, v := range s.configLogs {
		auditLogs = append(auditLogs, v)
	}
	return auditLogs
}

func (s *Service) Delete(id string) {
	delete(s.configLogs, id)
}
