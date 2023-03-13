package service

import (
	"context"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

type HealthStorage interface {
	Ping(ctx context.Context) error
}

type HealthService struct {
	Storage HealthStorage
}

func (s HealthService) CheckHealth(ctx context.Context) (types.HealthStatus, error) {
	healthStatus := types.HealthStatus{}
	var statusErrs error

	if err := s.Storage.Ping(ctx); err != nil {
		healthStatus.Storage = types.StatusDown
		statusErrs = errors.Join(statusErrs, errors.Newf("failed to ping the database: %w", err))
	}

	return healthStatus, statusErrs
}
