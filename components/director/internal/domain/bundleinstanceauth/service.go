package bundleinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

// Repository missing godoc
//go:generate mockery --name=Repository --output=automock --outpkg=automock --case=underscore
type Repository interface {
	Create(ctx context.Context, item *model.BundleInstanceAuth) error
	GetByID(ctx context.Context, tenantID string, id string) (*model.BundleInstanceAuth, error)
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.BundleInstanceAuth, error)
	ListByBundleID(ctx context.Context, tenantID string, bundleID string) ([]*model.BundleInstanceAuth, error)
	ListByRuntimeID(ctx context.Context, tenantID string, runtimeID string) ([]*model.BundleInstanceAuth, error)
	Update(ctx context.Context, tenant string, item *model.BundleInstanceAuth) error
	Delete(ctx context.Context, tenantID string, id string) error
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo         Repository
	uidService   UIDService
	timestampGen timestamp.Generator
}

// NewService missing godoc
func NewService(repo Repository, uidService UIDService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator,
	}
}

// Create missing godoc
func (s *service) Create(ctx context.Context, bundleID string, in model.BundleInstanceAuthRequestInput, defaultAuth *model.Auth, requestInputSchema *string) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	log.C(ctx).Debugf("Validating BundleInstanceAuth request input for Bundle with id %s", bundleID)
	err = s.validateInputParamsAgainstSchema(in.InputParams, requestInputSchema)
	if err != nil {
		return "", errors.Wrapf(err, "while validating BundleInstanceAuth request input for Bundle with id %s", bundleID)
	}

	con, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	var runtimeID *string
	if con.ConsumerType == consumer.Runtime {
		runtimeID = &con.ConsumerID
	}

	id := s.uidService.Generate()
	log.C(ctx).Debugf("ID %s generated for BundleInstanceAuth for Bundle with id %s", id, bundleID)
	bndlInstAuth := in.ToBundleInstanceAuth(id, bundleID, tnt, defaultAuth, nil, runtimeID, nil)

	err = s.setCreationStatusFromAuth(ctx, &bndlInstAuth, defaultAuth)
	if err != nil {
		return "", errors.Wrapf(err, "while setting creation status for BundleInstanceAuth with id %s", id)
	}

	err = s.repo.Create(ctx, &bndlInstAuth)
	if err != nil {
		return "", errors.Wrapf(err, "while creating BundleInstanceAuth with id %s for Bundle with id %s", id, bundleID)
	}

	return id, nil
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting BundleInstanceAuth with id %s", id)
	}

	return instanceAuth, nil
}

// GetForBundle missing godoc
func (s *service) GetForBundle(ctx context.Context, id string, bundleID string) (*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bndl, err := s.repo.GetForBundle(ctx, tnt, id, bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle Instance Auth with ID: [%s]", id)
	}

	return bndl, nil
}

// List missing godoc
func (s *service) List(ctx context.Context, bundleID string) ([]*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bndlInstanceAuths, err := s.repo.ListByBundleID(ctx, tnt, bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Bundle Instance Auths")
	}

	return bndlInstanceAuths, nil
}

// ListByRuntimeID missing godoc
func (s *service) ListByRuntimeID(ctx context.Context, runtimeID string) ([]*model.BundleInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bndlInstanceAuths, err := s.repo.ListByRuntimeID(ctx, tnt, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Bundle Instance Auths")
	}

	return bndlInstanceAuths, nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, instanceAuth *model.BundleInstanceAuth) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	if err = s.repo.Update(ctx, tnt, instanceAuth); err != nil {
		return errors.Wrap(err, "while updating Bundle Instance Auths")
	}

	return nil
}

// SetAuth missing godoc
func (s *service) SetAuth(ctx context.Context, id string, in model.BundleInstanceAuthSetInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting BundleInstanceAuth with id %s", id)
	}
	if instanceAuth == nil {
		return errors.Errorf("BundleInstanceAuth with id %s not found", id)
	}

	if instanceAuth.Status == nil || instanceAuth.Status.Condition != model.BundleInstanceAuthStatusConditionPending {
		return apperrors.NewInvalidOperationError("auth can be set only on BundleInstanceAuths in PENDING state")
	}

	err = s.setUpdateAuthAndStatus(ctx, instanceAuth, in)
	if err != nil {
		return err
	}

	if err = s.repo.Update(ctx, tnt, instanceAuth); err != nil {
		return errors.Wrapf(err, "while updating BundleInstanceAuth with ID %s", id)
	}
	return nil
}

// RequestDeletion missing godoc
func (s *service) RequestDeletion(ctx context.Context, instanceAuth *model.BundleInstanceAuth, defaultBundleInstanceAuth *model.Auth) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, err
	}
	if instanceAuth == nil {
		return false, apperrors.NewInternalError("BundleInstanceAuth is required to request its deletion")
	}

	if defaultBundleInstanceAuth == nil {
		log.C(ctx).Debugf("Default credentials for BundleInstanceAuth with id %s are not provided.", instanceAuth.ID)

		err := instanceAuth.SetDefaultStatus(model.BundleInstanceAuthStatusConditionUnused, s.timestampGen())
		if err != nil {
			return false, errors.Wrapf(err, "while setting status of BundleInstanceAuth with id %s to '%s'", instanceAuth.ID, model.BundleInstanceAuthStatusConditionUnused)
		}
		log.C(ctx).Infof("Status for BundleInstanceAuth with id %s set to '%s'. Credentials are ready for being deleted by Application or Integration System.", instanceAuth.ID, model.BundleInstanceAuthStatusConditionUnused)

		if err = s.repo.Update(ctx, tnt, instanceAuth); err != nil {
			return false, errors.Wrapf(err, "while updating BundleInstanceAuth with id %s", instanceAuth.ID)
		}

		return false, nil
	}

	log.C(ctx).Debugf("Default credentials for BundleInstanceAuth with id %s are provided.", instanceAuth.ID)
	if err = s.Delete(ctx, instanceAuth.ID); err != nil {
		return false, err
	}

	return true, nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	log.C(ctx).Debugf("Deleting BundleInstanceAuth entity with id %s in db", id)
	err = s.repo.Delete(ctx, tnt, id)

	return errors.Wrapf(err, "while deleting BundleInstanceAuth with id %s", id)
}

func (s *service) setUpdateAuthAndStatus(ctx context.Context, instanceAuth *model.BundleInstanceAuth, in model.BundleInstanceAuthSetInput) error {
	if instanceAuth == nil {
		return nil
	}

	ts := s.timestampGen()

	instanceAuth.Auth = in.Auth.ToAuth()
	instanceAuth.Status = in.Status.ToBundleInstanceAuthStatus(ts)

	// Input validation ensures that status can be nil only when auth was provided, so we can assume SUCCEEDED status
	if instanceAuth.Status == nil {
		log.C(ctx).Infof("Updating the status of BundleInstanceAuth with id %s to '%s'", instanceAuth.ID, model.BundleInstanceAuthStatusConditionSucceeded)
		err := instanceAuth.SetDefaultStatus(model.BundleInstanceAuthStatusConditionSucceeded, ts)
		if err != nil {
			return errors.Wrapf(err, "while setting status '%s' to BundleInstanceAuth with id %s", model.BundleInstanceAuthStatusConditionSucceeded, instanceAuth.ID)
		}
	}

	return nil
}

func (s *service) setCreationStatusFromAuth(ctx context.Context, instanceAuth *model.BundleInstanceAuth, defaultAuth *model.Auth) error {
	if instanceAuth == nil {
		return nil
	}

	var condition model.BundleInstanceAuthStatusCondition
	if defaultAuth != nil {
		log.C(ctx).Infof("Default credentials for BundleInstanceAuth with id %s from Bundle with id %s are provided. Setting creation status to '%s'", instanceAuth.ID, instanceAuth.BundleID, model.BundleInstanceAuthStatusConditionSucceeded)
		condition = model.BundleInstanceAuthStatusConditionSucceeded
	} else {
		log.C(ctx).Infof("Default credentials for BundleInstanceAuth with id %s from Bundle with id %s are not provided. Setting creation status to '%s'", instanceAuth.ID, instanceAuth.BundleID, model.BundleInstanceAuthStatusConditionPending)
		condition = model.BundleInstanceAuthStatusConditionPending
	}

	err := instanceAuth.SetDefaultStatus(condition, s.timestampGen())
	return errors.Wrapf(err, "while setting default status for BundleInstanceAuth with id %s", instanceAuth.ID)
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
