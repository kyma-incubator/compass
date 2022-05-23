package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=gqlConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type gqlConverter interface {
	FromInputGraphQL(in graphql.AutomaticScenarioAssignmentSetInput, targetTenantInternalID string) model.AutomaticScenarioAssignment
	ToGraphQL(in model.AutomaticScenarioAssignment, targetTenantExternalID string) graphql.AutomaticScenarioAssignment
}

//go:generate mockery --exported --name=asaService --output=automock --outpkg=automock --case=underscore --disable-version-string
type asaService interface {
	List(ctx context.Context, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error)
	ListForTargetTenant(ctx context.Context, targetTenantInternalID string) ([]*model.AutomaticScenarioAssignment, error)
	GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error)
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantService interface {
	GetExternalTenant(ctx context.Context, id string) (string, error)
	CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

//go:generate mockery --exported --name=tenantFetcher --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantFetcher interface {
	FetchOnDemand(tenant string) error
}

//go:generate mockery --exported --name=formationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationService interface {
	CreateAutomaticScenarioAssignment(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error)
	DeleteManyASAForSameTargetTenant(ctx context.Context, in []*model.AutomaticScenarioAssignment) error
	DeleteAutomaticScenarioAssignment(ctx context.Context, in model.AutomaticScenarioAssignment) error
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, svc asaService, converter gqlConverter, tenantService tenantService, fetcher tenantFetcher, formationSvc formationService) *Resolver {
	return &Resolver{
		transact:      transact,
		svc:           svc,
		converter:     converter,
		tenantService: tenantService,
		fetcher:       fetcher,
		formationSvc:  formationSvc,
	}
}

// Resolver missing godoc
type Resolver struct {
	transact      persistence.Transactioner
	converter     gqlConverter
	svc           asaService
	tenantService tenantService
	fetcher       tenantFetcher
	formationSvc  formationService
}

// CreateAutomaticScenarioAssignment missing godoc
func (r *Resolver) CreateAutomaticScenarioAssignment(ctx context.Context, in graphql.AutomaticScenarioAssignmentSetInput) (*graphql.AutomaticScenarioAssignment, error) {
	if err := r.fetcher.FetchOnDemand(in.Selector.Value); err != nil {
		return nil, errors.Wrapf(err, "while trying to create if not exists subaccount %s", in.Selector.Value)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	targetTenant, err := r.tenantService.GetInternalTenant(ctx, in.Selector.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while converting tenant")
	}

	convertedIn := r.converter.FromInputGraphQL(in, targetTenant)

	out, err := r.formationSvc.CreateAutomaticScenarioAssignment(ctx, convertedIn)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Assignment")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	assignment := r.converter.ToGraphQL(out, in.Selector.Value)

	return &assignment, nil
}

// GetAutomaticScenarioAssignmentForScenarioName missing godoc
func (r *Resolver) GetAutomaticScenarioAssignmentForScenarioName(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	out, err := r.svc.GetForScenarioName(ctx, scenarioName)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Assignment")
	}

	targetTenant, err := r.tenantService.GetExternalTenant(ctx, out.TargetTenantID)
	if err != nil {
		return nil, errors.Wrap(err, "while converting tenant")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	assignment := r.converter.ToGraphQL(out, targetTenant)

	return &assignment, nil
}

// AutomaticScenarioAssignmentsForSelector missing godoc
func (r *Resolver) AutomaticScenarioAssignmentsForSelector(ctx context.Context, in graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	targetTenant, err := r.tenantService.GetInternalTenant(ctx, in.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while converting tenant")
	}

	assignments, err := r.svc.ListForTargetTenant(ctx, targetTenant)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the assignments")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	gqlAssignments := make([]*graphql.AutomaticScenarioAssignment, 0, len(assignments))

	for _, v := range assignments {
		assignment := r.converter.ToGraphQL(*v, in.Value)
		gqlAssignments = append(gqlAssignments, &assignment)
	}

	return gqlAssignments, nil
}

// AutomaticScenarioAssignments missing godoc
func (r *Resolver) AutomaticScenarioAssignments(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.AutomaticScenarioAssignmentPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	page, err := r.svc.List(ctx, *first, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while listing the assignments")
	}

	gqlAssignments := make([]*graphql.AutomaticScenarioAssignment, 0, len(page.Data))

	for _, v := range page.Data {
		targetTenant, err := r.tenantService.GetExternalTenant(ctx, v.TargetTenantID)
		if err != nil {
			return nil, errors.Wrap(err, "while converting tenant")
		}

		assignment := r.converter.ToGraphQL(*v, targetTenant)
		gqlAssignments = append(gqlAssignments, &assignment)
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return &graphql.AutomaticScenarioAssignmentPage{
		Data:       gqlAssignments,
		TotalCount: page.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(page.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(page.PageInfo.EndCursor),
			HasNextPage: page.PageInfo.HasNextPage,
		},
	}, nil
}

// DeleteAutomaticScenarioAssignmentsForSelector missing godoc
func (r *Resolver) DeleteAutomaticScenarioAssignmentsForSelector(ctx context.Context, in graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	targetTenant, err := r.tenantService.GetInternalTenant(ctx, in.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while converting tenant")
	}

	assignments, err := r.svc.ListForTargetTenant(ctx, targetTenant)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting the Assignments for target tenant [%v]", targetTenant)
	}

	err = r.formationSvc.DeleteManyASAForSameTargetTenant(ctx, assignments)
	if err != nil {
		return nil, errors.Wrapf(err, "while deleting the Assignments for target tenant [%v]", targetTenant)
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	gqlAssignments := make([]*graphql.AutomaticScenarioAssignment, 0, len(assignments))

	for _, v := range assignments {
		assignment := r.converter.ToGraphQL(*v, in.Value)
		gqlAssignments = append(gqlAssignments, &assignment)
	}

	return gqlAssignments, nil
}

// DeleteAutomaticScenarioAssignmentForScenario missing godoc
func (r *Resolver) DeleteAutomaticScenarioAssignmentForScenario(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	assignment, err := r.svc.GetForScenarioName(ctx, scenarioName)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting the Assignment for scenario [name=%s]", scenarioName)
	}

	err = r.formationSvc.DeleteAutomaticScenarioAssignment(ctx, assignment)
	if err != nil {
		return nil, errors.Wrapf(err, "while deleting the Assignment for scenario [name=%s]", scenarioName)
	}

	targetTenant, err := r.tenantService.GetExternalTenant(ctx, assignment.TargetTenantID)
	if err != nil {
		return nil, errors.Wrap(err, "while converting tenant")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	gql := r.converter.ToGraphQL(assignment, targetTenant)

	return &gql, nil
}
