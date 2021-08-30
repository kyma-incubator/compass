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

type TokenData struct {
	Token        string `json:"one_time_token"`
	SystemAuthID string `json:"system_auth_id"`
}
