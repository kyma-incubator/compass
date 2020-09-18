package packageinstanceauth

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

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

func (s *service) Create(ctx context.Context, packageID string, in model.PackageInstanceAuthRequestInput, defaultAuth *model.Auth, requestInputSchema *string) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	log.Debugf("Validating PackageInstanceAuth request input for Package with id %s", packageID)
	err = s.validateInputParamsAgainstSchema(in.InputParams, requestInputSchema)
	if err != nil {
		return "", errors.Wrapf(err, "while validating PackageInstanceAuth request input for Package with id %s", packageID)
	}

	id := s.uidService.Generate()
	log.Debugf("ID %s generated for PackageInstanceAuth for Package with id %s", id, packageID)
	pkgInstAuth := in.ToPackageInstanceAuth(id, packageID, tnt, defaultAuth, nil)

	err = s.setCreationStatusFromAuth(&pkgInstAuth, defaultAuth)
	if err != nil {
		return "", errors.Wrapf(err, "while setting creation status for PackageInstanceAuth with id %s", id)
	}

	err = s.repo.Create(ctx, &pkgInstAuth)
	if err != nil {
		return "", errors.Wrapf(err, "while creating PackageInstanceAuth with id %s for Package with id %s", id, packageID)
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
		return nil, errors.Wrapf(err, "while getting PackageInstanceAuth with id %s", id)
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
		return errors.Wrapf(err, "while getting PackageInstanceAuth with id %s", id)
	}
	if instanceAuth == nil {
		return errors.Errorf("PackageInstanceAuth with id %s not found", id)
	}

	if instanceAuth.Status == nil || instanceAuth.Status.Condition != model.PackageInstanceAuthStatusConditionPending {
		return apperrors.NewInvalidOperationError("auth can be set only on PackageInstanceAuths in PENDING state")
	}

	err = s.setUpdateAuthAndStatus(instanceAuth, in)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, instanceAuth)
	if err != nil {
		return errors.Wrapf(err, "while updating PackageInstanceAuth with ID %s", id)
	}
	return nil
}

func (s *service) RequestDeletion(ctx context.Context, instanceAuth *model.PackageInstanceAuth, defaultPackageInstanceAuth *model.Auth) (bool, error) {
	if instanceAuth == nil {
		return false, apperrors.NewInternalError("PackageInstanceAuth is required to request its deletion")
	}

	if defaultPackageInstanceAuth == nil {
		log.Debugf("Default credentials for PackageInstanceAuth with id %s are not provided.", instanceAuth.ID)

		err := instanceAuth.SetDefaultStatus(model.PackageInstanceAuthStatusConditionUnused, s.timestampGen())
		if err != nil {
			return false, errors.Wrapf(err, "while setting status of PackageInstanceAuth with id %s to '%s'", instanceAuth.ID, model.PackageInstanceAuthStatusConditionUnused)
		}
		log.Infof("Status for PackageInstanceAuth with id %s set to '%s'. Credentials are ready for being deleted by Application or Integration System.", instanceAuth.ID, model.PackageInstanceAuthStatusConditionUnused)

		err = s.repo.Update(ctx, instanceAuth)
		if err != nil {
			return false, errors.Wrapf(err, "while updating PackageInstanceAuth with id %s", instanceAuth.ID)
		}

		return false, nil
	}

	log.Debugf("Default credentials for PackageInstanceAuth with id %s are provided.", instanceAuth.ID)
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

	log.Debugf("Deleting PackageInstanceAuth entity with id %s in db", id)
	err = s.repo.Delete(ctx, tnt, id)

	return errors.Wrapf(err, "while deleting PackageInstanceAuth with id %s", id)
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
		log.Infof("Updating the status of PackageInstanceAuth with id %s to '%s'", instanceAuth.ID, model.PackageInstanceAuthStatusConditionSucceeded)
		err := instanceAuth.SetDefaultStatus(model.PackageInstanceAuthStatusConditionSucceeded, ts)
		if err != nil {
			return errors.Wrapf(err, "while setting status '%s' to PackageInstanceAuth with id %s", model.PackageInstanceAuthStatusConditionSucceeded, instanceAuth.ID)
		}
	}

	return nil
}

func (s *service) setCreationStatusFromAuth(instanceAuth *model.PackageInstanceAuth, defaultAuth *model.Auth) error {
	if instanceAuth == nil {
		return nil
	}

	var condition model.PackageInstanceAuthStatusCondition
	if defaultAuth != nil {
		log.Infof("Default credentials for PackageInstanceAuth with id %s from Package with id %s are provided. Setting creation status to '%s'", instanceAuth.ID, instanceAuth.PackageID, model.PackageInstanceAuthStatusConditionSucceeded)
		condition = model.PackageInstanceAuthStatusConditionSucceeded
	} else {
		log.Infof("Default credentials for PackageInstanceAuth with id %s from Package with id %s are not provided. Setting creation status to '%s'", instanceAuth.ID, instanceAuth.PackageID, model.PackageInstanceAuthStatusConditionPending)
		condition = model.PackageInstanceAuthStatusConditionPending
	}

	err := instanceAuth.SetDefaultStatus(condition, s.timestampGen())
	return errors.Wrapf(err, "while setting default status for PackageInstanceAuth with id %s", instanceAuth.ID)
}

func (s *service) validateInputParamsAgainstSchema(inputParams *string, schema *string) error {
	if schema == nil {
		return nil
	}
	if inputParams == nil {
		return apperrors.NewInvalidDataError("json schema for input parameters was defined for the package but no input parameters were provided")
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
