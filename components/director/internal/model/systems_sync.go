package model

import (
	"time"
)

type SystemSynchronizationTimestamp struct {
	ID                string
	TenantID          string
	ProductID         string
	LastSyncTimestamp time.Time
}
