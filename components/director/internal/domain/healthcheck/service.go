package healthcheck

// HealthCheckRepository missing godoc
//go:generate mockery --name=HealthCheckRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type HealthCheckRepository interface {
}

type service struct {
	repo HealthCheckRepository
}

// NewService missing godoc
func NewService(repo HealthCheckRepository) *service {
	return &service{repo: repo}
}
