package configuration

import (
	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
)

type Service struct {
	configLogs map[string]model.ConfigurationChange
}

func NewService() *Service {
	return &Service{configLogs: make(map[string]model.ConfigurationChange)}
}
func (s *Service) Save(change model.ConfigurationChange) (string, error) {
	s.configLogs[change.UUID] = change

	return change.UUID, nil
}

func (s *Service) Get(id string) *model.ConfigurationChange {
	val, ok := s.configLogs[id]
	if ok {
		return &val
	}
	return nil
}

func (s *Service) List() []model.ConfigurationChange {
	var auditLogs []model.ConfigurationChange
	for _, v := range s.configLogs {
		auditLogs = append(auditLogs, v)
	}
	return auditLogs
}

func (s *Service) Delete(id string) {
	delete(s.configLogs, id)
}
