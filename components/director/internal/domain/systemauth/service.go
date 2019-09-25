package systemauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, item model.SystemAuth) error
	ListForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
	Delete(ctx context.Context, tenant string, id string, objectType model.SystemAuthReferenceObjectType) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo       Repository
	uidService UIDService
}

func NewService(repo Repository, uidService UIDService) *service {
	return &service{
		repo:       repo,
		uidService: uidService,
	}
}

func (s *service) Create(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	systemAuth := model.SystemAuth{
		ID:    s.uidService.Generate(),
		Value: authInput.ToAuth(),
	}

	switch objectType {
	case model.ApplicationReference:
		systemAuth.AppID = &objectID
		systemAuth.TenantID = tnt
	case model.RuntimeReference:
		systemAuth.RuntimeID = &objectID
		systemAuth.TenantID = tnt
	case model.IntegrationSystemReference:
		systemAuth.IntegrationSystemID = &objectID
		systemAuth.TenantID = model.IntegrationSystemTenant
	default:
		return "", errors.New("unknown reference object type")
	}

	err = s.repo.Create(ctx, systemAuth)
	if err != nil {
		return "", errors.Wrapf(err, "while creating System Auth for %s", objectType)
	}

	return systemAuth.ID, nil
}

func (s *service) ListForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if objectType == model.IntegrationSystemReference {
		tnt = model.IntegrationSystemTenant
	}

	systemAuths, err := s.repo.ListForObject(ctx, tnt, objectType, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing System Auths for %s with reference ID '%s'", objectType, objectID)
	}

	return systemAuths, nil
}

func (s *service) Delete(ctx context.Context, id string, objectType model.SystemAuthReferenceObjectType) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	if objectType == model.IntegrationSystemReference {
		tnt = model.IntegrationSystemTenant
	}

	err = s.repo.Delete(ctx, tnt, id, objectType)

	return errors.Wrapf(err, "while deleting System Auth with ID '%s' for %s", id, objectType)
}
