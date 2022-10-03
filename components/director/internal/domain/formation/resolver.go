package formation

import (
	"context"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Service missing godoc
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type Service interface {
	Get(ctx context.Context, id string) (*model.Formation, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.FormationPage, error)
	CreateFormation(ctx context.Context, tnt string, formation model.Formation, templateName string) (*model.Formation, error)
	DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
	AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
	UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
}

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	FromGraphQL(i graphql.FormationInput) model.Formation
	ToGraphQL(i *model.Formation) *graphql.Formation
	MultipleToGraphQL(in []*model.Formation) []*graphql.Formation
}

// FormationAssignmentService is responsible for the service-layer FormationAssignment operations
//go:generate mockery --name=FormationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentService interface {
	ListByFormationIDs(ctx context.Context, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error)
	GetForFormation(ctx context.Context, id, formationID string) (*model.FormationAssignment, error)
}

// FormationAssignmentConverter converts FormationAssignment between the model.FormationAssignment service-layer representation and graphql.FormationAssignment.
//go:generate mockery --name=FormationAssignmentConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentConverter interface {
	MultipleToGraphQL(in []*model.FormationAssignment) ([]*graphql.FormationAssignment, error)
	ToGraphQL(in *model.FormationAssignment) (*graphql.FormationAssignment, error)
}

// TenantFetcher calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
//go:generate mockery --name=TenantFetcher --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantFetcher interface {
	FetchOnDemand(tenant, parentTenant string) error
}

// Resolver is the formation resolver
type Resolver struct {
	transact                persistence.Transactioner
	service                 Service
	conv                    Converter
	formationAssignmentSvc  FormationAssignmentService
	formationAssignmentConv FormationAssignmentConverter
	fetcher                 TenantFetcher
}

// NewResolver creates formation resolver
func NewResolver(transact persistence.Transactioner, service Service, conv Converter, formationAssignmentSvc FormationAssignmentService, formationAssignmentConv FormationAssignmentConverter, fetcher TenantFetcher) *Resolver {
	return &Resolver{
		transact:                transact,
		service:                 service,
		conv:                    conv,
		formationAssignmentSvc:  formationAssignmentSvc,
		formationAssignmentConv: formationAssignmentConv,
		fetcher:                 fetcher,
	}
}

// Formation returns a Formation by its id
func (r *Resolver) Formation(ctx context.Context, id string) (*graphql.Formation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formation, err := r.service.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(formation), nil
}

// Formations returns paginated Formations based on first and after
func (r *Resolver) Formations(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.FormationPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationPage, err := r.service.List(ctx, *first, cursor)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	formations := r.conv.MultipleToGraphQL(formationPage.Data)

	return &graphql.FormationPage{
		Data:       formations,
		TotalCount: formationPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(formationPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(formationPage.PageInfo.EndCursor),
			HasNextPage: formationPage.PageInfo.HasNextPage,
		},
	}, nil
}

// CreateFormation creates new formation for the caller tenant
func (r *Resolver) CreateFormation(ctx context.Context, formationInput graphql.FormationInput) (*graphql.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	templateName := model.DefaultTemplateName
	if formationInput.TemplateName != nil && *formationInput.TemplateName != "" {
		templateName = *formationInput.TemplateName
	}

	newFormation, err := r.service.CreateFormation(ctx, tnt, r.conv.FromGraphQL(formationInput), templateName)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation), nil
}

// DeleteFormation deletes the formation from the caller tenant formations
func (r *Resolver) DeleteFormation(ctx context.Context, formation graphql.FormationInput) (*graphql.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	deletedFormation, err := r.service.DeleteFormation(ctx, tnt, r.conv.FromGraphQL(formation))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(deletedFormation), nil
}

// AssignFormation assigns object to the provided formation
func (r *Resolver) AssignFormation(ctx context.Context, objectID string, objectType graphql.FormationObjectType, formation graphql.FormationInput) (*graphql.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if objectType == graphql.FormationObjectTypeTenant {
		if err := r.fetcher.FetchOnDemand(objectID, tnt); err != nil {
			return nil, errors.Wrapf(err, "while trying to create if not exists subaccount %s", objectID)
		}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	newFormation, err := r.service.AssignFormation(ctx, tnt, objectID, objectType, r.conv.FromGraphQL(formation))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation), nil
}

// UnassignFormation unassigns the object from the provided formation
func (r *Resolver) UnassignFormation(ctx context.Context, objectID string, objectType graphql.FormationObjectType, formation graphql.FormationInput) (*graphql.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	newFormation, err := r.service.UnassignFormation(ctx, tnt, objectID, objectType, r.conv.FromGraphQL(formation))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation), nil
}

// FormationAssignments retrieves a page of FormationAssignments for the specified Formation
func (r *Resolver) FormationAssignments(ctx context.Context, obj *graphql.Formation, first *int, after *graphql.PageCursor) (*graphql.FormationAssignmentPage, error) {
	param := dataloader.ParamFormationAssignment{ID: obj.ID, Ctx: ctx, First: first, After: after}
	return dataloader.FormationFor(ctx).FormationAssignmentByID.Load(param)
}

// FormationAssignmentsDataLoader retrieves a page of FormationAssignments for each Formation ID in the keys argument
func (r *Resolver) FormationAssignmentsDataLoader(keys []dataloader.ParamFormationAssignment) ([]*graphql.FormationAssignmentPage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Formations found")}
	}

	ctx := keys[0].Ctx
	formationIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		formationIDs = append(formationIDs, key.ID)
	}

	var cursor string
	if keys[0].After != nil {
		cursor = string(*keys[0].After)
	}

	if keys[0].First == nil {
		return nil, []error{apperrors.NewInvalidDataError("missing required parameter 'first'")}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationAssignmentPages, err := r.formationAssignmentSvc.ListByFormationIDs(ctx, formationIDs, *keys[0].First, cursor)
	if err != nil {
		return nil, []error{err}
	}

	gqlFormationAssignments := make([]*graphql.FormationAssignmentPage, 0, len(formationAssignmentPages))
	for _, page := range formationAssignmentPages {
		fas, err := r.formationAssignmentConv.MultipleToGraphQL(page.Data)
		if err != nil {
			return nil, []error{err}
		}

		gqlFormationAssignments = append(gqlFormationAssignments, &graphql.FormationAssignmentPage{Data: fas, TotalCount: page.TotalCount, PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(page.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(page.PageInfo.EndCursor),
			HasNextPage: page.PageInfo.HasNextPage,
		}})
	}

	if err = tx.Commit(); err != nil {
		return nil, []error{err}
	}

	return gqlFormationAssignments, nil
}

// FormationAssignment missing godoc
func (r *Resolver) FormationAssignment(ctx context.Context, obj *graphql.Formation, id string) (*graphql.FormationAssignment, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Formation cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationAssignment, err := r.formationAssignmentSvc.GetForFormation(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.formationAssignmentConv.ToGraphQL(formationAssignment)
}
