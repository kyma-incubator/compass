package healthcheck

//go:generate mockery -name=HealthCheckRepository -output=automock -outpkg=automock -case=underscore
type HealthCheckRepository interface {
}

type Service struct {
	repo HealthCheckRepository
}

func NewService(repo HealthCheckRepository) *Service {
	return &Service{repo: repo}
}
