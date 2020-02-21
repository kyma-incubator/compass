package mp_package

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/mock"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

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
	CreateInputFromGraphQL(in graphql.PackageCreateInput) (*model.PackageCreateInput, error)
	UpdateInputFromGraphQL(in graphql.PackageUpdateInput) (*model.PackageUpdateInput, error)
}

type Resolver struct {
	transact         persistence.Transactioner
	packageConverter PackageConverter
	packageSvc       PackageService
}

func NewResolver(
	transact persistence.Transactioner,
	packageSvc PackageService,
	packageConverter PackageConverter) *Resolver {
	return &Resolver{
		transact:         transact,
		packageConverter: packageConverter,
		packageSvc:       packageSvc,
	}
}

func (r *Resolver) AddPackage(ctx context.Context, applicationID string, in graphql.PackageCreateInput) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.packageConverter.CreateInputFromGraphQL(in)
	if err != nil {
		return nil, err
	}

	id, err := r.packageSvc.Create(ctx, applicationID, *convertedIn)
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
	defer r.transact.RollbackUnlessCommited(tx)

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

	gqlPkg, err := r.packageConverter.ToGraphQL(pkg)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Package to GraphQL with ID: [%s]", id)
	}

	return gqlPkg, nil
}

func (r *Resolver) DeletePackage(ctx context.Context, id string) (*graphql.Package, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

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

	deletedPkg, err := r.packageConverter.ToGraphQL(pkg)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Package to GraphQL with ID: [%s]", id)
	}

	return deletedPkg, nil
}

// TODO: Replace with real implementation
func (r *Resolver) InstanceAuth(ctx context.Context, obj *graphql.Package, id string) (*graphql.PackageInstanceAuth, error) {
	var condition graphql.PackageInstanceAuthStatusCondition
	switch id {
	case "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb":
		condition = graphql.PackageInstanceAuthStatusConditionSucceeded
	case "cccccccc-cccc-cccc-cccc-cccccccccccc":
		condition = graphql.PackageInstanceAuthStatusConditionFailed
	default:
		condition = graphql.PackageInstanceAuthStatusConditionPending
	}

	return mock.FixPackageInstanceAuth(id, condition), nil
}

// TODO: Replace with real implementation
func (r *Resolver) InstanceAuths(ctx context.Context, obj *graphql.Package) ([]*graphql.PackageInstanceAuth, error) {
	return []*graphql.PackageInstanceAuth{
		mock.FixPackageInstanceAuth("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", graphql.PackageInstanceAuthStatusConditionPending),
		mock.FixPackageInstanceAuth("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", graphql.PackageInstanceAuthStatusConditionSucceeded),
		mock.FixPackageInstanceAuth("cccccccc-cccc-cccc-cccc-cccccccccccc", graphql.PackageInstanceAuthStatusConditionFailed),
	}, nil
}

var packageID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

// TODO: Replace with real implementation
func (r *Resolver) APIDefinitions(ctx context.Context, obj *graphql.Package, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	return &graphql.APIDefinitionPage{
		Data: []*graphql.APIDefinition{
			mock.FixAPIDefinition("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			mock.FixAPIDefinition("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			mock.FixAPIDefinition("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		},
		TotalCount: 3,
	}, nil
}

// TODO: Replace with real implementation
func (r *Resolver) EventDefinitions(ctx context.Context, obj *graphql.Package, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	return &graphql.EventDefinitionPage{
		Data: []*graphql.EventDefinition{
			mock.FixEventDefinition("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			mock.FixEventDefinition("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			mock.FixEventDefinition("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		},
		TotalCount: 3,
	}, nil
}

// TODO: Replace with real implementation
func (r *Resolver) Documents(ctx context.Context, obj *graphql.Package, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	return &graphql.DocumentPage{
		Data: []*graphql.Document{
			mock.FixDocument("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			mock.FixDocument("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			mock.FixDocument("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		},
		TotalCount: 3,
	}, nil
}

// TODO: Replace with real implementation
func (r *Resolver) APIDefinition(ctx context.Context, obj *graphql.Package, id string) (*graphql.APIDefinition, error) {
	return mock.FixAPIDefinition(id), nil
}

// TODO: Replace with real implementation
func (r *Resolver) EventDefinition(ctx context.Context, obj *graphql.Package, id string) (*graphql.EventDefinition, error) {
	return mock.FixEventDefinition(id), nil
}

// TODO: Replace with real implementation
func (r *Resolver) Document(ctx context.Context, obj *graphql.Package, id string) (*graphql.Document, error) {
	return mock.FixDocument(id), nil
}
