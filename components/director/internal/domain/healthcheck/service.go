package healthcheck

//go:generate mockery -name=HealthCheckRepository -output=automock -outpkg=automock -case=underscore
type HealthCheckRepository interface {
}

type service struct {
	repo HealthCheckRepository
}

func NewService(repo HealthCheckRepository) *service {
	return &service{repo: repo}
}
