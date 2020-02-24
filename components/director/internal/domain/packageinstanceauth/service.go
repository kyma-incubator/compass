package packageinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, item *model.PackageInstanceAuth) error
	GetByID(ctx context.Context, tenantID string, id string) (*model.PackageInstanceAuth, error)
	GetForPackage(ctx context.Context, tenant string, id string, packageID string) (*model.PackageInstanceAuth, error)
	ListByPackageID(ctx context.Context, tenantID string, packageID string) ([]*model.PackageInstanceAuth, error)
	Update(ctx context.Context, item *model.PackageInstanceAuth) error
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

func (s *service) Create(ctx context.Context, packageID string, in model.PackageInstanceAuthRequestInput, auth *model.Auth, requestInputSchema *string) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	err = s.validateInputParamsAgainstSchema(in.InputParams, requestInputSchema)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	pkgInstAuth := in.ToPackageInstanceAuth(id, packageID, tnt, auth, nil)

	err = s.setCreationStatusFromAuth(&pkgInstAuth, auth)
	if err != nil {
		return "", err
	}

	err = s.repo.Create(ctx, &pkgInstAuth)
	if err != nil {
		return "", errors.Wrap(err, "while creating Package Instance Auth")
	}

	return id, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.PackageInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Package Instance Auth")
	}

	return instanceAuth, nil
}

func (s *service) GetForPackage(ctx context.Context, id string, packageID string) (*model.PackageInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pkg, err := s.repo.GetForPackage(ctx, tnt, id, packageID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Package Instance Auth with ID: [%s]", id)
	}

	return pkg, nil
}

func (s *service) List(ctx context.Context, packageID string) ([]*model.PackageInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pkgInstanceAuths, err := s.repo.ListByPackageID(ctx, tnt, packageID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Package Instance Auths")
	}

	return pkgInstanceAuths, nil
}

func (s *service) SetAuth(ctx context.Context, id string, in model.PackageInstanceAuthSetInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrap(err, "while getting Package Instance Auth")
	}
	if instanceAuth == nil {
		return errors.Errorf("Package Instance Auth with ID %s not found", id)
	}

	if instanceAuth.Status == nil || instanceAuth.Status.Condition != model.PackageInstanceAuthStatusConditionPending {
		return errors.New("auth can be set only on Package Instance Auths in PENDING state")
	}

	err = s.setUpdateAuthAndStatus(instanceAuth, in)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, instanceAuth)
	if err != nil {
		return errors.Wrapf(err, "while updating Package Instance Auth with ID %s", id)
	}
	return nil
}

func (s *service) RequestDeletion(ctx context.Context, instanceAuth *model.PackageInstanceAuth, defaultPackageInstanceAuth *model.Auth) (bool, error) {
	if instanceAuth == nil {
		return false, errors.New("instance auth is required to request its deletion")
	}

	if defaultPackageInstanceAuth == nil {
		err := instanceAuth.SetDefaultStatus(model.PackageInstanceAuthStatusConditionUnused, s.timestampGen())
		if err != nil {
			return false, errors.Wrapf(err, "while setting status of Instance Auth with ID '%s' to '%s'", instanceAuth.ID, model.PackageInstanceAuthStatusConditionUnused)
		}

		err = s.repo.Update(ctx, instanceAuth)
		if err != nil {
			return false, errors.Wrapf(err, "while updating Package Instance Auth with ID %s", instanceAuth.ID)
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

	return errors.Wrapf(err, "while deleting Package Instance Auth with ID %s", id)
}

func (s *service) setUpdateAuthAndStatus(instanceAuth *model.PackageInstanceAuth, in model.PackageInstanceAuthSetInput) error {
	if instanceAuth == nil {
		return nil
	}

	ts := s.timestampGen()

	instanceAuth.Auth = in.Auth.ToAuth()
	instanceAuth.Status = in.Status.ToPackageInstanceAuthStatus(ts)

	// Input validation ensures that status can be nil only when auth was provided, so we can assume SUCCEEDED status
	if instanceAuth.Status == nil {
		err := instanceAuth.SetDefaultStatus(model.PackageInstanceAuthStatusConditionSucceeded, ts)
		if err != nil {
			return errors.Wrapf(err, "while setting Package Instance Auth status to: %s", model.PackageInstanceAuthStatusConditionSucceeded)
		}
	}

	return nil
}

func (s *service) setCreationStatusFromAuth(instanceAuth *model.PackageInstanceAuth, auth *model.Auth) error {
	if instanceAuth == nil {
		return nil
	}

	var condition model.PackageInstanceAuthStatusCondition
	if auth != nil {
		condition = model.PackageInstanceAuthStatusConditionSucceeded
	} else {
		condition = model.PackageInstanceAuthStatusConditionPending
	}
	err := instanceAuth.SetDefaultStatus(condition, s.timestampGen())
	return errors.Wrap(err, "while setting default status for Package Instance Auth")
}

func (s *service) validateInputParamsAgainstSchema(inputParams *string, schema *string) error {
	if schema == nil {
		return nil
	}
	if inputParams == nil {
		return errors.New("json schema for input parameters was defined for the package but no input parameters were provided")
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
