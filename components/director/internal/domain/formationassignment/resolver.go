package formationassignment

import (
	"context"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

// Resolver is a formation assignment resolver
type Resolver struct {
	transact                     persistence.Transactioner
	appRepo                      applicationRepo
	appConverter                 applicationConverter
	runtimeRepo                  runtimeRepo
	runtimeConverter             runtimeConverter
	runtimeContextRepo           runtimeContextRepo
	runtimeContextConverter      runtimeContextConverter
	assignmentOperationService   assignmentOperationService
	assignmentOperationConverter assignmentOperationConverter
}

//go:generate mockery --name=applicationRepo --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type applicationRepo interface {
	ListAllByIDs(ctx context.Context, tenantID string, ids []string) ([]*model.Application, error)
}

//go:generate mockery --name=runtimeRepo --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type runtimeRepo interface {
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.Runtime, error)
}

//go:generate mockery --name=runtimeContextRepo --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type runtimeContextRepo interface {
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.RuntimeContext, error)
}

//go:generate mockery --name=applicationConverter --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type applicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
}

//go:generate mockery --name=runtimeConverter --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type runtimeConverter interface {
	ToGraphQL(in *model.Runtime) *graphql.Runtime
}

//go:generate mockery --name=runtimeContextConverter --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type runtimeContextConverter interface {
	ToGraphQL(in *model.RuntimeContext) *graphql.RuntimeContext
}

//go:generate mockery --exported --name=assignmentOperationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type assignmentOperationService interface {
	Finish(ctx context.Context, assignmentID, formationID string) error
	ListByFormationAssignmentIDs(ctx context.Context, formationAssignmentIDs []string, pageSize int, cursor string) ([]*model.AssignmentOperationPage, error)
}

//go:generate mockery --exported --name=assignmentOperationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type assignmentOperationConverter interface {
	MultipleToGraphQL(in []*model.AssignmentOperation) []*graphql.AssignmentOperation
}

// NewResolver is a constructor for formation assignment resolver
func NewResolver(transact persistence.Transactioner, appRepo applicationRepo, appConverter applicationConverter, runtimeRepo runtimeRepo, runtimeConverter runtimeConverter, runtimeContextRepo runtimeContextRepo, runtimeContextConverter runtimeContextConverter, operationService assignmentOperationService, operationConverter assignmentOperationConverter) *Resolver {
	return &Resolver{
		transact:                     transact,
		appRepo:                      appRepo,
		appConverter:                 appConverter,
		runtimeRepo:                  runtimeRepo,
		runtimeConverter:             runtimeConverter,
		runtimeContextRepo:           runtimeContextRepo,
		runtimeContextConverter:      runtimeContextConverter,
		assignmentOperationService:   operationService,
		assignmentOperationConverter: operationConverter,
	}
}

// FormationParticipantDataLoader is a dataloader for Source and Target entities of the formation assignment
func (r *Resolver) FormationParticipantDataLoader(params []dataloader.ParamFormationParticipant) ([]graphql.FormationParticipant, []error) {
	if len(params) == 0 {
		return nil, []error{apperrors.NewInternalError("No Formation Assignments found")}
	}

	ctx := params[0].Ctx

	appIDs := make(map[string]struct{}, len(params))
	runtimeIDs := make(map[string]struct{}, len(params))
	runtimeCtxIDs := make(map[string]struct{}, len(params))
	for _, param := range params {
		if param.ParticipantID == "" {
			return nil, []error{apperrors.NewInternalError("Cannot fetch Formation Participant. Participant ID is empty")}
		}
		switch param.ParticipantType {
		case string(model.FormationAssignmentTypeApplication):
			appIDs[param.ParticipantID] = struct{}{}
		case string(model.FormationAssignmentTypeRuntime):
			runtimeIDs[param.ParticipantID] = struct{}{}
		case string(model.FormationAssignmentTypeRuntimeContext):
			runtimeCtxIDs[param.ParticipantID] = struct{}{}
		}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, []error{errors.Wrapf(err, "while loading tenant from context")}
	}

	apps, err := r.appRepo.ListAllByIDs(ctx, tenantID, str.MapToSlice(appIDs))
	if err != nil {
		return nil, []error{errors.Wrapf(err, "while fetching applications")}
	}

	runtimes, err := r.runtimeRepo.ListByIDs(ctx, tenantID, str.MapToSlice(runtimeIDs))
	if err != nil {
		return nil, []error{errors.Wrapf(err, "while fetching runtimes")}
	}

	runtimeContexts, err := r.runtimeContextRepo.ListByIDs(ctx, tenantID, str.MapToSlice(runtimeCtxIDs))
	if err != nil {
		return nil, []error{errors.Wrapf(err, "while fetching runtimeContexts")}
	}

	if err = tx.Commit(); err != nil {
		return nil, []error{err}
	}

	participants := make(map[string]graphql.FormationParticipant, len(apps)+len(runtimes)+len(runtimeContexts))
	for _, app := range apps {
		gqlApp := r.appConverter.ToGraphQL(app)
		participants[gqlApp.ID] = gqlApp
	}
	for _, rt := range runtimes {
		gqlRt := r.runtimeConverter.ToGraphQL(rt)
		participants[gqlRt.ID] = gqlRt
	}
	for _, rtCtx := range runtimeContexts {
		gqlRtCtx := r.runtimeContextConverter.ToGraphQL(rtCtx)
		participants[gqlRtCtx.ID] = gqlRtCtx
	}

	result := make([]graphql.FormationParticipant, 0, len(params))
	for _, param := range params {
		result = append(result, participants[param.ParticipantID])
	}

	return result, nil
}

// TargetEntity is a resolver for the Target of the formation assignment
func (r *Resolver) TargetEntity(ctx context.Context, obj *graphql.FormationAssignment) (graphql.FormationParticipant, error) {
	params := dataloader.ParamFormationParticipant{ID: obj.ID, ParticipantID: obj.Target, ParticipantType: string(obj.TargetType), Ctx: ctx}
	return dataloader.ForTargetFormationParticipant(ctx).FormationParticipantDataloader.Load(params)
}

// SourceEntity is a resolver for the Source of the formation assignment
func (r *Resolver) SourceEntity(ctx context.Context, obj *graphql.FormationAssignment) (graphql.FormationParticipant, error) {
	params := dataloader.ParamFormationParticipant{ID: obj.ID, ParticipantID: obj.Source, ParticipantType: string(obj.SourceType), Ctx: ctx}
	return dataloader.ForSourceFormationParticipant(ctx).FormationParticipantDataloader.Load(params)
}

// AssignmentOperations retrieves a page of FormationAssignments for the specified Formation
func (r *Resolver) AssignmentOperations(ctx context.Context, obj *graphql.FormationAssignment, first *int, after *graphql.PageCursor) (*graphql.AssignmentOperationPage, error) {
	param := dataloader.ParamAssignmentOperation{ID: obj.ID, Ctx: ctx, First: first, After: after}
	return dataloader.FormationAssignmentFor(ctx).AssignmentOperationByID.Load(param)
}

// AssignmentOperationsDataLoader is a dataloader for the Operation entity of the formation assignment
func (r *Resolver) AssignmentOperationsDataLoader(keys []dataloader.ParamAssignmentOperation) ([]*graphql.AssignmentOperationPage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Formation Assignments found")}
	}

	ctx := keys[0].Ctx

	formationAssignmentIDs := make([]string, 0, len(keys))
	for _, param := range keys {
		if param.ID == "" {
			return nil, []error{apperrors.NewInternalError("Cannot fetch Formation Assignment. ID is empty")}
		}
		formationAssignmentIDs = append(formationAssignmentIDs, param.ID)
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

	assignmentOperationsPages, err := r.assignmentOperationService.ListByFormationAssignmentIDs(ctx, formationAssignmentIDs, *keys[0].First, cursor)
	if err != nil {
		return nil, []error{errors.Wrapf(err, "while fetching assignment operations")}
	}

	if err = tx.Commit(); err != nil {
		return nil, []error{err}
	}

	result := make([]*graphql.AssignmentOperationPage, 0, len(assignmentOperationsPages))
	for _, page := range assignmentOperationsPages {
		operations := r.assignmentOperationConverter.MultipleToGraphQL(page.Data)

		result = append(result, &graphql.AssignmentOperationPage{Data: operations, TotalCount: page.TotalCount, PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(page.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(page.PageInfo.EndCursor),
			HasNextPage: page.PageInfo.HasNextPage,
		}})
	}

	return result, nil
}
