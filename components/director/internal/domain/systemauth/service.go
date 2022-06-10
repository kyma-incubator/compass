package systemauth

import (
	"context"
	"time"

	intModel "github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// Repository missing godoc
//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore --disable-version-string
type Repository interface {
	Create(ctx context.Context, item model.SystemAuth) error
	GetByIDForObject(ctx context.Context, tenant, id string, objType model.SystemAuthReferenceObjectType) (*model.SystemAuth, error)
	GetByIDForObjectGlobal(ctx context.Context, id string, objType model.SystemAuthReferenceObjectType) (*model.SystemAuth, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.SystemAuth, error)
	ListForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
	ListForObjectGlobal(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
	DeleteByIDForObject(ctx context.Context, tenant, id string, objType model.SystemAuthReferenceObjectType) error
	DeleteByIDForObjectGlobal(ctx context.Context, id string, objType model.SystemAuthReferenceObjectType) error
	GetByJSONValue(ctx context.Context, value map[string]interface{}) (*model.SystemAuth, error)
	Update(ctx context.Context, item *model.SystemAuth) error
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo       Repository
	uidService UIDService
}

// NewService missing godoc
func NewService(repo Repository, uidService UIDService) *service {
	return &service{
		repo:       repo,
		uidService: uidService,
	}
}

// Create missing godoc
func (s *service) Create(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *intModel.AuthInput) (string, error) {
	return s.create(ctx, s.uidService.Generate(), objectType, objectID, authInput)
}

// CreateWithCustomID missing godoc
func (s *service) CreateWithCustomID(ctx context.Context, id string, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *intModel.AuthInput) (string, error) {
	return s.create(ctx, id, objectType, objectID, authInput)
}

func (s *service) create(ctx context.Context, id string, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *intModel.AuthInput) (string, error) {
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

// GetByIDForObject missing godoc
func (s *service) GetByIDForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, authID string) (*model.SystemAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		if !model.IsIntegrationSystemNoTenantFlow(err, objectType) {
			return nil, errors.Wrapf(err, "while loading tenant from context")
		}
	}

	var item *model.SystemAuth

	if objectType == model.IntegrationSystemReference {
		item, err = s.repo.GetByIDForObjectGlobal(ctx, authID, objectType)
	} else {
		item, err = s.repo.GetByIDForObject(ctx, tnt, authID, objectType)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while getting SystemAuth with ID %s", authID)
	}

	return item, nil
}

// GetGlobal missing godoc
func (s *service) GetGlobal(ctx context.Context, id string) (*model.SystemAuth, error) {
	item, err := s.repo.GetByIDGlobal(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting SystemAuth with ID %s", id)
	}

	return item, nil
}

// GetByToken get a SystemAuth by a one time token value
func (s *service) GetByToken(ctx context.Context, token string) (*model.SystemAuth, error) {
	return s.repo.GetByJSONValue(ctx, map[string]interface{}{
		"OneTimeToken": map[string]interface{}{
			"Token": token,
			"Used":  false,
		},
	})
}

// InvalidateToken gets a SystemAuth by ID, sets the Used properties for the OTT and updates the model
func (s *service) InvalidateToken(ctx context.Context, id string) (*model.SystemAuth, error) {
	systemAuth, err := s.GetGlobal(ctx, id)
	if err != nil {
		return nil, err
	}

	systemAuth.Value.OneTimeToken.Used = true
	systemAuth.Value.OneTimeToken.UsedAt = time.Now()

	if err := s.repo.Update(ctx, systemAuth); err != nil {
		return nil, err
	}

	return systemAuth, nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, item *model.SystemAuth) error {
	return s.repo.Update(ctx, item)
}

// UpdateValue get SystemAuth by provided id and update it with the given input
func (s *service) UpdateValue(ctx context.Context, id string, item *intModel.Auth) (*model.SystemAuth, error) {
	systemAuth, err := s.repo.GetByIDGlobal(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting System Auth with id '%s'", id)
	}

	systemAuth.Value = item

	if err := s.repo.Update(ctx, systemAuth); err != nil {
		return nil, errors.Wrapf(err, "while updating System Auth with id '%s'", id)
	}

	return systemAuth, nil
}

// ListForObject missing godoc
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

// DeleteByIDForObject missing godoc
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

// DeleteMultipleByIDForObject Deletes multiple system auths at a time
func (s *service) DeleteMultipleByIDForObject(ctx context.Context, systemAuths []model.SystemAuth) error {
	for _, auth := range systemAuths {
		referenceType, err := auth.GetReferenceObjectType()
		if err != nil {
			return errors.Wrapf(err, "while fetching System Auth reference object type for id '%s'", auth.ID)
		}
		if err := s.DeleteByIDForObject(ctx, referenceType, auth.ID); err != nil {
			return errors.Wrapf(err, "while deleting System Auth with reference object type '%s' and id '%s'", referenceType, auth.ID)
		}
	}

	return nil
}
