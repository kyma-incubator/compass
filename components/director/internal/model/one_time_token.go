package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/tokens"
)

type OneTimeToken struct {
	Token        string
	ConnectorURL string
	Type         tokens.TokenType
	CreatedAt    time.Time
	Used         bool
	UsedAt       time.Time
}
