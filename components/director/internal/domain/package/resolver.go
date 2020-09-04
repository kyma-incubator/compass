package mp_package

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=PackageService -output=automock -outpkg=automock -case=underscore
type PackageService interface {
	Create(ctx context.Context, applicationID string, in model.PackageCreateInput) (string, error)
	Update(ctx context.Context, id string, in model.PackageUpdateInput) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Package, error)
}

//go:generate mockery -name=PackageConverter -output=automock -outpkg=automock -case=underscore
type PackageConverter interface {
	ToGraphQL(in *model.Package) (*graphql.Package, error)
	CreateInputFromGraphQL(in graphql.PackageCreateInput) (model.PackageCreateInput, error)
	UpdateInputFromGraphQL(in graphql.PackageUpdateInput) (*model.PackageUpdateInput, error)
}

//go:generate mockery -name=BundleService -output=automock -outpkg=automock -case=underscore
type BundleService interface {
	Create(ctx context.Context, applicationID string, in model.BundleCreateInput) (string, error)
	Update(ctx context.Context, id string, in model.BundleUpdateInput) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Bundle, error)
	ListForPackage(ctx context.Context, packageID string) ([]*model.Bundle, error)
	GetForPackage(ctx context.Context, id string, packageID string) (*model.Bundle, error)
}

//go:generate mockery -name=BundleConverter -output=automock -outpkg=automock -case=underscore
type BundleConverter interface {
	ToGraphQL(in *model.Bundle) (*graphql.Bundle, error)
	CreateInputFromGraphQL(in graphql.BundleCreateInput) (model.BundleCreateInput, error)
	UpdateInputFromGraphQL(in graphql.BundleUpdateInput) (*model.BundleUpdateInput, error)
	MultipleCreateInputFromGraphQL(in []*graphql.BundleCreateInput) ([]*model.BundleCreateInput, error)
	MultipleToGraphQL(in []*model.Bundle) ([]*graphql.Bundle, error)
}

type Resolver struct {
	transact persistence.Transactioner

	packageSvc PackageService
	bundleSvc  BundleService

	packageConverter PackageConverter
	bundleConverter  BundleConverter
}

func NewResolver(
	transact persistence.Transactioner,
	packageSvc PackageService,
	bundleSvc BundleService,
	packageConverter PackageConverter,
	bundleConv BundleConverter,
) *Resolver {
	return &Resolver{
		transact:         transact,
		packageSvc:       packageSvc,
		bundleSvc:        bundleSvc,
		packageConverter: packageConverter,
		bundleConverter:  bundleConv,
	}
}

func (r *Resolver) AddPackage(ctx context.Context, applicationID string, in graphql.PackageCreateInput) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.packageConverter.CreateInputFromGraphQL(in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting input from GraphQL")
	}

	id, err := r.packageSvc.Create(ctx, applicationID, convertedIn)
	if err != nil {
		return nil, err
	}

	pkg, err := r.packageSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlPackage, err := r.packageConverter.ToGraphQL(pkg)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Package to GraphQL with ID: [%s]", id)
	}

	return gqlPackage, nil
}

func (r *Resolver) UpdatePackage(ctx context.Context, id string, in graphql.PackageUpdateInput) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.packageConverter.UpdateInputFromGraphQL(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Package update input from GraphQL with ID: [%s]", id)
	}

	err = r.packageSvc.Update(ctx, id, *convertedIn)
	if err != nil {
		return nil, err
	}

	pkg, err := r.packageSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlPackage, err := r.packageConverter.ToGraphQL(pkg)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Package to GraphQL with ID: [%s]", id)
	}

	return gqlPackage, nil
}

func (r *Resolver) DeletePackage(ctx context.Context, id string) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	pkg, err := r.packageSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.packageSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	deletedPackage, err := r.packageConverter.ToGraphQL(pkg)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Package to GraphQL with ID: [%s]", id)
	}

	return deletedPackage, nil
}

func (r *Resolver) Bundle(ctx context.Context, obj *graphql.Package, id string) (*graphql.Bundle, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	bundle, err := r.bundleSvc.GetForPackage(ctx, id, obj.ID)
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

	return r.bundleConverter.ToGraphQL(bundle)
}

func (r *Resolver) Bundles(ctx context.Context, obj *graphql.Package) ([]*graphql.Bundle, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	bundlesPage, err := r.bundleSvc.ListForPackage(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.bundleConverter.MultipleToGraphQL(bundlesPage)
}
