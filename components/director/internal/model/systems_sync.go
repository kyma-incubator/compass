package model

import (
	"time"
)

// SystemSynchronizationTimestamp represents the last synchronization time of a system
type SystemSynchronizationTimestamp struct {
	ID                string
	TenantID          string
	ProductID         string
	LastSyncTimestamp time.Time
}
