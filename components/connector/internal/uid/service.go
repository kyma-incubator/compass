package uid

import "github.com/google/uuid"

type service struct{}

func NewService() *service {
	return &service{}
}

func (s *service) Generate() string {
	return uuid.New().String()
}
