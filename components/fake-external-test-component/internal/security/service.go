package security

import (
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/fake-external-test-component/pkg/model"
)

type Service struct {
	configLogs map[string]model.SecuritEvent
}

func NewService() *Service {
	return &Service{configLogs: make(map[string]model.SecuritEvent)}
}
func (s *Service) Save(change model.SecuritEvent) (string, error) {
	id := uuid.New().String()
	id = "0ef99922-c493-4ccf-b8de-1778ab84ca3a"
	//TODO: any collisions?
	s.configLogs[id] = change

	return id, nil
}

func (s *Service) Get(id string) *model.SecuritEvent {
	val, ok := s.configLogs[id]
	if ok {
		return &val
	}
	return nil
}

func (s *Service) Delete(id string) {
	delete(s.configLogs, id)
}
