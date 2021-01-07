package packageinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	RequestDeletion(ctx context.Context, instanceAuth *model.PackageInstanceAuth, defaultPackageInstanceAuth *model.Auth) (bool, error)
	Create(ctx context.Context, packageID string, in model.PackageInstanceAuthRequestInput, defaultAuth *model.Auth, requestInputSchema *string) (string, error)
	Get(ctx context.Context, id string) (*model.PackageInstanceAuth, error)
	SetAuth(ctx context.Context, id string, in model.PackageInstanceAuthSetInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToGraphQL(in *model.PackageInstanceAuth) (*graphql.PackageInstanceAuth, error)
	RequestInputFromGraphQL(in graphql.PackageInstanceAuthRequestInput) model.PackageInstanceAuthRequestInput
	SetInputFromGraphQL(in graphql.PackageInstanceAuthSetInput) (model.PackageInstanceAuthSetInput, error)
}

//go:generate mockery -name=PackageService -output=automock -outpkg=automock -case=underscore
type PackageService interface {
	Get(ctx context.Context, id string) (*model.Package, error)
	GetByInstanceAuthID(ctx context.Context, instanceAuthID string) (*model.Package, error)
}

//go:generate mockery -name=PackageConverter -output=automock -outpkg=automock -case=underscore
type PackageConverter interface {
	ToGraphQL(in *model.Package) (*graphql.Package, error)
}

type Resolver struct {
	transact persistence.Transactioner
	svc      Service
	pkgSvc   PackageService
	conv     Converter
	pkgConv  PackageConverter
}

func NewResolver(transact persistence.Transactioner, svc Service, pkgSvc PackageService, conv Converter, pkgConv PackageConverter) *Resolver {
	return &Resolver{
		transact: transact,
		svc:      svc,
		pkgSvc:   pkgSvc,
		conv:     conv,
		pkgConv:  pkgConv,
	}
}

func (r *Resolver) PackageByInstanceAuth(ctx context.Context, authID string) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	packageInstanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	pkg, err := r.pkgSvc.Get(ctx, packageInstanceAuth.PackageID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.pkgConv.ToGraphQL(pkg)
}

func (r *Resolver) PackageInstanceAuth(ctx context.Context, id string) (*graphql.PackageInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	packageInstanceAuth, err := r.svc.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(packageInstanceAuth)
}

func (r *Resolver) DeletePackageInstanceAuth(ctx context.Context, authID string) (*graphql.PackageInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	err = r.svc.Delete(ctx, authID)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("PackageInstanceAuth with id %s successfully deleted", authID)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

func (r *Resolver) SetPackageInstanceAuth(ctx context.Context, authID string, in graphql.PackageInstanceAuthSetInput) (*graphql.PackageInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Setting credentials for PackageInstanceAuth with id %s", authID)

	convertedIn, err := r.conv.SetInputFromGraphQL(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting PackageInstanceAuth with id %s from GraphQL", authID)
	}

	err = r.svc.SetAuth(ctx, authID, convertedIn)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Credentials successfully set for PackageInstanceAuth with id %s", authID)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

func (r *Resolver) RequestPackageInstanceAuthCreation(ctx context.Context, packageID string, in graphql.PackageInstanceAuthRequestInput) (*graphql.PackageInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Requesting PackageInstanceAuth creation for Package with id %s", packageID)

	pkg, err := r.pkgSvc.Get(ctx, packageID)
	if err != nil {
		return nil, err
	}

	convertedIn := r.conv.RequestInputFromGraphQL(in)

	instanceAuthID, err := r.svc.Create(ctx, packageID, convertedIn, pkg.DefaultInstanceAuth, pkg.InstanceAuthRequestInputSchema)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := r.svc.Get(ctx, instanceAuthID)
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created PackageInstanceAuth with id %s for Package with id %s", instanceAuthID, packageID)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

func (r *Resolver) RequestPackageInstanceAuthDeletion(ctx context.Context, authID string) (*graphql.PackageInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Requesting PackageInstanceAuth deletion for PackageInstanceAuth with id %s", authID)

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	pkg, err := r.pkgSvc.GetByInstanceAuthID(ctx, authID)
	if err != nil {
		return nil, err
	}

	deleted, err := r.svc.RequestDeletion(ctx, instanceAuth, pkg.DefaultInstanceAuth)
	if err != nil {
		return nil, err
	}

	if !deleted {
		instanceAuth, err = r.svc.Get(ctx, authID) // get InstanceAuth once again for new status
		if err != nil {
			return nil, err
		}
	}

	log.C(ctx).Infof("PackageInstanceAuth with id %s successfully deleted.", authID)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}
