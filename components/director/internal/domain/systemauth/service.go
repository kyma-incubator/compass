package systemauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, item model.SystemAuth) error
	GetByID(ctx context.Context, tenant, id string) (*model.SystemAuth, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.SystemAuth, error)
	ListForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
	ListForObjectGlobal(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
	DeleteByIDForObject(ctx context.Context, tenant, id string, objType model.SystemAuthReferenceObjectType) error
	DeleteByIDForObjectGlobal(ctx context.Context, id string, objType model.SystemAuthReferenceObjectType) error
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
	return s.create(ctx, s.uidService.Generate(), objectType, objectID, authInput)
}

func (s *service) CreateWithCustomID(ctx context.Context, id string, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error) {
	return s.create(ctx, id, objectType, objectID, authInput)
}

func (s *service) create(ctx context.Context, id string, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		if !model.IsIntegrationSystemNoTenantFlow(err, objectType) {
			return "", err
		}
	}

	log.C(ctx).Debugf("Tenant %s loaded while creating SystemAuth for %s with id %s", tnt, objectType, objectID)

	systemAuth := model.SystemAuth{
		ID:    id,
		Value: authInput.ToAuth(),
	}

	switch objectType {
	case model.ApplicationReference:
		systemAuth.AppID = &objectID
		systemAuth.TenantID = &tnt
	case model.RuntimeReference:
		systemAuth.RuntimeID = &objectID
		systemAuth.TenantID = &tnt
	case model.IntegrationSystemReference:
		systemAuth.IntegrationSystemID = &objectID
		systemAuth.TenantID = nil
	default:
		return "", apperrors.NewInternalError("unknown reference object type")
	}

	err = s.repo.Create(ctx, systemAuth)
	if err != nil {
		return "", err
	}

	return systemAuth.ID, nil
}

func (s *service) GetByIDForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, authID string) (*model.SystemAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		if !model.IsIntegrationSystemNoTenantFlow(err, objectType) {
			return nil, errors.Wrapf(err, "while loading tenant from context")
		}
	}

	var item *model.SystemAuth

	if objectType == model.IntegrationSystemReference {
		item, err = s.repo.GetByIDGlobal(ctx, authID)
	} else {
		item, err = s.repo.GetByID(ctx, tnt, authID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while getting SystemAuth with ID %s", authID)
	}

	return item, nil
}

func (s *service) GetGlobal(ctx context.Context, id string) (*model.SystemAuth, error) {
	item, err := s.repo.GetByIDGlobal(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting SystemAuth with ID %s", id)
	}

	return item, nil
}

func (s *service) ListForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		if !model.IsIntegrationSystemNoTenantFlow(err, objectType) {
			return nil, err
		}
	}

	var systemAuths []model.SystemAuth

	if objectType == model.IntegrationSystemReference {
		systemAuths, err = s.repo.ListForObjectGlobal(ctx, objectType, objectID)
	} else {
		systemAuths, err = s.repo.ListForObject(ctx, tnt, objectType, objectID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while listing System Auths for %s with reference ID '%s'", objectType, objectID)
	}

	return systemAuths, nil
}

func (s *service) DeleteByIDForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, authID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		if !model.IsIntegrationSystemNoTenantFlow(err, objectType) {
			return err
		}
	}

	if objectType == model.IntegrationSystemReference {
		err = s.repo.DeleteByIDForObjectGlobal(ctx, authID, objectType)
	} else {
		err = s.repo.DeleteByIDForObject(ctx, tnt, authID, objectType)
	}
	if err != nil {
		return errors.Wrapf(err, "while deleting System Auth with ID '%s' for %s", authID, objectType)
	}

	return nil
}
