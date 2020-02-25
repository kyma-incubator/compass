package mp_package

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/mock"
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

//go:generate mockery -name=PackageInstanceAuthService -output=automock -outpkg=automock -case=underscore
type PackageInstanceAuthService interface {
	GetForPackage(ctx context.Context, id string, packageID string) (*model.PackageInstanceAuth, error)
	List(ctx context.Context, id string) ([]*model.PackageInstanceAuth, error)
}

//go:generate mockery -name=PackageInstanceAuthConverter -output=automock -outpkg=automock -case=underscore
type PackageInstanceAuthConverter interface {
	ToGraphQL(in *model.PackageInstanceAuth) *graphql.PackageInstanceAuth
	MultipleToGraphQL(in []*model.PackageInstanceAuth) []*graphql.PackageInstanceAuth
}

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	ListForPackage(ctx context.Context, packageID string, pageSize int, cursor string) (*model.APIDefinitionPage, error)
	GetForPackage(ctx context.Context, id string, packageID string) (*model.APIDefinition, error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition
	MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition
}

//go:generate mockery -name=EventService -output=automock -outpkg=automock -case=underscore
type EventService interface {
	ListForPackage(ctx context.Context, packageID string, pageSize int, cursor string) (*model.EventDefinitionPage, error)
	GetForPackage(ctx context.Context, id string, packageID string) (*model.EventDefinition, error)
}

//go:generate mockery -name=EventConverter -output=automock -outpkg=automock -case=underscore
type EventConverter interface {
	ToGraphQL(in *model.EventDefinition) *graphql.EventDefinition
	MultipleToGraphQL(in []*model.EventDefinition) []*graphql.EventDefinition
}

//go:generate mockery -name=DocumentService -output=automock -outpkg=automock -case=underscore
type DocumentService interface {
	ListForPackage(ctx context.Context, packageID string, pageSize int, cursor string) (*model.DocumentPage, error)
	GetForPackage(ctx context.Context, id string, packageID string) (*model.Document, error)
}

//go:generate mockery -name=DocumentConverter -output=automock -outpkg=automock -case=underscore
type DocumentConverter interface {
	ToGraphQL(in *model.Document) *graphql.Document
	MultipleToGraphQL(in []*model.Document) []*graphql.Document
}

type Resolver struct {
	transact persistence.Transactioner

	packageSvc             PackageService
	packageInstanceAuthSvc PackageInstanceAuthService
	apiSvc                 APIService
	eventSvc               EventService
	documentSvc            DocumentService

	packageConverter             PackageConverter
	packageInstanceAuthConverter PackageInstanceAuthConverter
	apiConverter                 APIConverter
	eventConverter               EventConverter
	documentConverter            DocumentConverter
}

func NewResolver(
	transact persistence.Transactioner,
	packageSvc PackageService,
	packageInstanceAuthSvc PackageInstanceAuthService,
	apiSvc APIService,
	eventSvc EventService,
	documentSvc DocumentService,
	packageConverter PackageConverter,
	packageInstanceAuthConverter PackageInstanceAuthConverter,
	apiConv APIConverter,
	eventConv EventConverter,
	documentConv DocumentConverter) *Resolver {
	return &Resolver{
		transact:                     transact,
		packageConverter:             packageConverter,
		packageSvc:                   packageSvc,
		packageInstanceAuthSvc:       packageInstanceAuthSvc,
		apiSvc:                       apiSvc,
		eventSvc:                     eventSvc,
		documentSvc:                  documentSvc,
		packageInstanceAuthConverter: packageInstanceAuthConverter,
		apiConverter:                 apiConv,
		eventConverter:               eventConv,
		documentConverter:            documentConv,
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

func (r *Resolver) InstanceAuth(ctx context.Context, obj *graphql.Package, id string) (*graphql.PackageInstanceAuth, error) {
	if obj == nil {
		return nil, errors.New("Package cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	pkg, err := r.packageInstanceAuthSvc.GetForPackage(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.packageInstanceAuthConverter.ToGraphQL(pkg), nil

}

//TODO Remove mock
func (r *Resolver) InstanceAuthMock(ctx context.Context, obj *graphql.Package, id string) (*graphql.PackageInstanceAuth, error) {
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

func (r *Resolver) InstanceAuths(ctx context.Context, obj *graphql.Package) ([]*graphql.PackageInstanceAuth, error) {
	if obj == nil {
		return nil, errors.New("Package cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	pkgInstanceAuths, err := r.packageInstanceAuthSvc.List(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	out := r.packageInstanceAuthConverter.MultipleToGraphQL(pkgInstanceAuths)

	return out, nil
}

//TODO Remove mock
func (r *Resolver) InstanceAuthsMock(ctx context.Context, obj *graphql.Package) ([]*graphql.PackageInstanceAuth, error) {
	return []*graphql.PackageInstanceAuth{
		mock.FixPackageInstanceAuth("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", graphql.PackageInstanceAuthStatusConditionPending),
		mock.FixPackageInstanceAuth("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", graphql.PackageInstanceAuthStatusConditionSucceeded),
		mock.FixPackageInstanceAuth("cccccccc-cccc-cccc-cccc-cccccccccccc", graphql.PackageInstanceAuthStatusConditionFailed),
	}, nil
}

var packageID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

func (r *Resolver) APIDefinition(ctx context.Context, obj *graphql.Package, id string) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	api, err := r.apiSvc.GetForPackage(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.apiConverter.ToGraphQL(api), nil
}

func (r *Resolver) APIDefinitions(ctx context.Context, obj *graphql.Package, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	apisPage, err := r.apiSvc.ListForPackage(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApis := r.apiConverter.MultipleToGraphQL(apisPage.Data)
	totalCount := len(gqlApis)

	return &graphql.APIDefinitionPage{
		Data:       gqlApis,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(apisPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(apisPage.PageInfo.EndCursor),
			HasNextPage: apisPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) EventDefinition(ctx context.Context, obj *graphql.Package, id string) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventAPI, err := r.eventSvc.GetForPackage(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.eventConverter.ToGraphQL(eventAPI), nil
}

func (r *Resolver) EventDefinitions(ctx context.Context, obj *graphql.Package, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	eventAPIPage, err := r.eventSvc.ListForPackage(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApis := r.eventConverter.MultipleToGraphQL(eventAPIPage.Data)
	totalCount := len(gqlApis)

	return &graphql.EventDefinitionPage{
		Data:       gqlApis,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(eventAPIPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(eventAPIPage.PageInfo.EndCursor),
			HasNextPage: eventAPIPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) Document(ctx context.Context, obj *graphql.Package, id string) (*graphql.Document, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventAPI, err := r.documentSvc.GetForPackage(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.documentConverter.ToGraphQL(eventAPI), nil
}

// TODO: Proper error handling
func (r *Resolver) Documents(ctx context.Context, obj *graphql.Package, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	documentsPage, err := r.documentSvc.ListForPackage(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlDocuments := r.documentConverter.MultipleToGraphQL(documentsPage.Data)
	totalCount := len(gqlDocuments)

	return &graphql.DocumentPage{
		Data:       gqlDocuments,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(documentsPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(documentsPage.PageInfo.EndCursor),
			HasNextPage: documentsPage.PageInfo.HasNextPage,
		},
	}, nil
}
