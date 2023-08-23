package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/tokens"
)

// ScenarioGroup represents scenario group
type ScenarioGroup struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

// OneTimeToken missing godoc
type OneTimeToken struct {
	Token          string
	ConnectorURL   string
	Type           tokens.TokenType
	CreatedAt      time.Time
	Used           bool
	ExpiresAt      time.Time
	UsedAt         time.Time
	ScenarioGroups []ScenarioGroup
}
