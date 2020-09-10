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
	Create(ctx context.Context, applicationID string, in model.PackageInput) (string, error)
	Update(ctx context.Context, id string, in model.PackageInput) error
	CreateOrUpdate(ctx context.Context, appID, openDiscoveryID string, in model.PackageInput) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Package, error)
	AssociateBundle(ctx context.Context, id, bundleID string) error
}

//go:generate mockery -name=PackageConverter -output=automock -outpkg=automock -case=underscore
type PackageConverter interface {
	ToGraphQL(in *model.Package) (*graphql.Package, error)
	InputFromGraphQL(in graphql.PackageInput) (model.PackageInput, error)
}

//go:generate mockery -name=BundleService -output=automock -outpkg=automock -case=underscore
type BundleService interface {
	Create(ctx context.Context, applicationID string, in model.BundleInput) (string, error)
	Update(ctx context.Context, id string, in model.BundleInput) error
	CreateOrUpdate(ctx context.Context, appID, openDiscoveryID string, in model.BundleInput) (string, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Bundle, error)
	ListForPackage(ctx context.Context, packageID string, pageSize int, cursor string) (*model.BundlePage, error)
	GetForPackage(ctx context.Context, id string, packageID string) (*model.Bundle, error)
}

//go:generate mockery -name=BundleConverter -output=automock -outpkg=automock -case=underscore
type BundleConverter interface {
	ToGraphQL(in *model.Bundle) (*graphql.Bundle, error)
	InputFromGraphQL(in graphql.BundleInput) (model.BundleInput, error)
	MultipleCreateInputFromGraphQL(in []*graphql.BundleInput) ([]*model.BundleInput, error)
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

func (r *Resolver) AddPackage(ctx context.Context, applicationID string, in graphql.PackageInput) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.packageConverter.InputFromGraphQL(in)
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

func (r *Resolver) UpdatePackage(ctx context.Context, id string, in graphql.PackageInput) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.packageConverter.InputFromGraphQL(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Package update input from GraphQL with ID: [%s]", id)
	}

	err = r.packageSvc.Update(ctx, id, convertedIn)
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

func (r *Resolver) Bundles(ctx context.Context, obj *graphql.Package, first *int, after *graphql.PageCursor) (*graphql.BundlePage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	bundlesPage, err := r.bundleSvc.ListForPackage(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlBundles, err := r.bundleConverter.MultipleToGraphQL(bundlesPage.Data)
	if err != nil {
		return nil, err
	}

	return &graphql.BundlePage{
		Data:       gqlBundles,
		TotalCount: bundlesPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(bundlesPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(bundlesPage.PageInfo.EndCursor),
			HasNextPage: bundlesPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) AssociateBundleWithPackage(ctx context.Context, in graphql.BundlePackageRelationInput) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.packageSvc.AssociateBundle(ctx, in.PackageID, in.BundleID)
	if err != nil {
		return nil, err
	}

	pkg, err := r.packageSvc.Get(ctx, in.PackageID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlPackage, err := r.packageConverter.ToGraphQL(pkg)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Package to GraphQL with ID: [%s]", in.PackageID)
	}

	return gqlPackage, nil
}
