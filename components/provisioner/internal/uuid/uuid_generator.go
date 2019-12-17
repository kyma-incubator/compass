package uuid

import (
	"github.com/google/uuid"
)

//go:generate mockery -name=UUIDGenerator
type UUIDGenerator interface {
	New() string
}

type uuidGenerator struct {
}

func NewUUIDGenerator() UUIDGenerator {
	return uuidGenerator{}
}

func (u uuidGenerator) New() string {
	return uuid.New().String()
}
