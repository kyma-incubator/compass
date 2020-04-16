package dbmodel

import "time"

type LMSTenantDTO struct {
	ID        string
	Name      string
	Region    string
	CreatedAt time.Time
}
