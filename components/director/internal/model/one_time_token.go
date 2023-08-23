package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/tokens"
)

// OneTimeToken missing godoc
type OneTimeToken struct {
	Token          string
	ConnectorURL   string
	Type           tokens.TokenType
	CreatedAt      time.Time
	Used           bool
	ExpiresAt      time.Time
	UsedAt         time.Time
	ScenarioGroups []string
}
