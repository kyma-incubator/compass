package uid

import "github.com/google/uuid"

type service struct{}

// NewService missing godoc
func NewService() *service {
	return &service{}
}

// Generate missing godoc
func (s *service) Generate() string {
	return uuid.New().String()
}
