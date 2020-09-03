package bundleinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, item *model.BundleInstanceAuth) error
	GetByID(ctx context.Context, tenantID string, id string) (*model.BundleInstanceAuth, error)
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.BundleInstanceAuth, error)
	ListByBundleID(ctx context.Context, tenantID string, bundleID string) ([]*model.BundleInstanceAuth, error)
	Update(ctx context.Context, item *model.BundleInstanceAuth) error
	Delete(ctx context.Context, tenantID string, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo         Repository
	uidService   UIDService
	timestampGen timestamp.Generator
}

func NewService(repo Repository, uidService UIDService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator(),
	}
}

func (s *service) Create(ctx context.Context, bundleID string, in model.BundleInstanceAuthRequestInput, defaultAuth *model.Auth, requestInputSchema *string) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	err = s.validateInputParamsAgainstSchema(in.InputParams, requestInputSchema)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	pkgInstAuth := in.ToBundleInstanceAuth(id, bundleID, tnt, defaultAuth, nil)

	err = s.setCreationStatusFromAuth(&pkgInstAuth, defaultAuth)
	if err != nil {
		return "", err
	}

	err = s.repo.Create(ctx, &pkgInstAuth)
	if err != nil {
		return "", errors.Wrap(err, "while creating Bundle Instance Auth")
	}

	return id, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Bundle Instance Auth")
	}

	return instanceAuth, nil
}

func (s *service) GetForBundle(ctx context.Context, id string, bundleID string) (*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pkg, err := s.repo.GetForBundle(ctx, tnt, id, bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle Instance Auth with ID: [%s]", id)
	}

	return pkg, nil
}

func (s *service) List(ctx context.Context, bundleID string) ([]*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pkgInstanceAuths, err := s.repo.ListByBundleID(ctx, tnt, bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Bundle Instance Auths")
	}

	return pkgInstanceAuths, nil
}

func (s *service) SetAuth(ctx context.Context, id string, in model.BundleInstanceAuthSetInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrap(err, "while getting Bundle Instance Auth")
	}
	if instanceAuth == nil {
		return errors.Errorf("Bundle Instance Auth with ID %s not found", id)
	}

	if instanceAuth.Status == nil || instanceAuth.Status.Condition != model.BundleInstanceAuthStatusConditionPending {
		return apperrors.NewInvalidOperationError("auth can be set only on Bundle Instance Auths in PENDING state")
	}

	err = s.setUpdateAuthAndStatus(instanceAuth, in)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, instanceAuth)
	if err != nil {
		return errors.Wrapf(err, "while updating Bundle Instance Auth with ID %s", id)
	}
	return nil
}

func (s *service) RequestDeletion(ctx context.Context, instanceAuth *model.BundleInstanceAuth, defaultBundleInstanceAuth *model.Auth) (bool, error) {
	if instanceAuth == nil {
		return false, apperrors.NewInternalError("instance auth is required to request its deletion")
	}

	if defaultBundleInstanceAuth == nil {
		err := instanceAuth.SetDefaultStatus(model.BundleInstanceAuthStatusConditionUnused, s.timestampGen())
		if err != nil {
			return false, errors.Wrapf(err, "while setting status of Instance Auth with ID '%s' to '%s'", instanceAuth.ID, model.BundleInstanceAuthStatusConditionUnused)
		}

		err = s.repo.Update(ctx, instanceAuth)
		if err != nil {
			return false, errors.Wrapf(err, "while updating Bundle Instance Auth with ID %s", instanceAuth.ID)
		}

		return false, nil
	}

	err := s.Delete(ctx, instanceAuth.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, tnt, id)

	return errors.Wrapf(err, "while deleting Bundle Instance Auth with ID %s", id)
}

func (s *service) setUpdateAuthAndStatus(instanceAuth *model.BundleInstanceAuth, in model.BundleInstanceAuthSetInput) error {
	if instanceAuth == nil {
		return nil
	}

	ts := s.timestampGen()

	instanceAuth.Auth = in.Auth.ToAuth()
	instanceAuth.Status = in.Status.ToBundleInstanceAuthStatus(ts)

	// Input validation ensures that status can be nil only when auth was provided, so we can assume SUCCEEDED status
	if instanceAuth.Status == nil {
		err := instanceAuth.SetDefaultStatus(model.BundleInstanceAuthStatusConditionSucceeded, ts)
		if err != nil {
			return errors.Wrapf(err, "while setting Bundle Instance Auth status to: %s", model.BundleInstanceAuthStatusConditionSucceeded)
		}
	}

	return nil
}

func (s *service) setCreationStatusFromAuth(instanceAuth *model.BundleInstanceAuth, defaultAuth *model.Auth) error {
	if instanceAuth == nil {
		return nil
	}

	var condition model.BundleInstanceAuthStatusCondition
	if defaultAuth != nil {
		condition = model.BundleInstanceAuthStatusConditionSucceeded
	} else {
		condition = model.BundleInstanceAuthStatusConditionPending
	}
	err := instanceAuth.SetDefaultStatus(condition, s.timestampGen())
	return errors.Wrap(err, "while setting default status for Bundle Instance Auth")
}

func (s *service) validateInputParamsAgainstSchema(inputParams *string, schema *string) error {
	if schema == nil {
		return nil
	}
	if inputParams == nil {
		return apperrors.NewInvalidDataError("json schema for input parameters was defined for the bundle but no input parameters were provided")
	}

	validator, err := jsonschema.NewValidatorFromStringSchema(*schema)
	if err != nil {
		return errors.Wrapf(err, "while creating JSON Schema validator for schema %+s", *schema)
	}

	result, err := validator.ValidateString(*inputParams)
	if err != nil {
		return errors.Wrapf(err, "while validating value %s against JSON Schema: %s", *inputParams, *schema)
	}
	if !result.Valid {
		return errors.Wrapf(result.Error, "while validating value %s against JSON Schema: %s", *inputParams, *schema)
	}

	return nil
}
