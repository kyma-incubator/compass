package formation

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Service missing godoc
//
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type Service interface {
	Get(ctx context.Context, id string) (*model.Formation, error)
	GetFormationByName(ctx context.Context, formationName, tnt string) (*model.Formation, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.FormationPage, error)
	GetGlobalByID(ctx context.Context, id string) (*model.Formation, error)
	ListFormationsForObject(ctx context.Context, objectID string) ([]*model.Formation, error)
	CreateFormation(ctx context.Context, tnt string, formation model.Formation, templateName string) (*model.Formation, error)
	DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
	AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation, initialConfigurations model.InitialConfigurations) (*model.Formation, error)
	UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation, ignoreASA bool) (*model.Formation, error)
	ResynchronizeFormationNotifications(ctx context.Context, formationID string, reset bool) (*model.Formation, error)
	FinalizeDraftFormation(ctx context.Context, formationID string) (*model.Formation, error)
}

// Converter missing godoc
//
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	FromGraphQL(i graphql.FormationInput) model.Formation
	ToGraphQL(i *model.Formation) (*graphql.Formation, error)
	MultipleToGraphQL(in []*model.Formation) ([]*graphql.Formation, error)
}

//go:generate mockery --exported --name=formationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentService interface {
	Delete(ctx context.Context, id string) error
	DeleteAssignmentsForObjectID(ctx context.Context, formationID, objectID string) error
	ListByFormationIDs(ctx context.Context, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error)
	ListByFormationIDsNoPaging(ctx context.Context, formationIDs []string) ([][]*model.FormationAssignment, error)
	GetForFormation(ctx context.Context, id, formationID string) (*model.FormationAssignment, error)
	ListFormationAssignmentsForObjectID(ctx context.Context, formationID, objectID string) ([]*model.FormationAssignment, error)
	ListAllForObjectGlobal(ctx context.Context, objectID string) ([]*model.FormationAssignment, error)
	ProcessFormationAssignments(ctx context.Context, formationAssignmentsForObject []*model.FormationAssignment, requests []*webhookclient.FormationAssignmentNotificationRequestTargetMapping, operation func(context.Context, *formationassignment.AssignmentMappingPairWithOperation) (bool, error), formationOperation model.FormationOperation) error
	ProcessFormationAssignmentPair(ctx context.Context, mappingPair *formationassignment.AssignmentMappingPairWithOperation) (bool, error)
	GenerateAssignments(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation, initialConfigurations model.InitialConfigurations) ([]*model.FormationAssignmentInput, error)
	PersistAssignments(ctx context.Context, tnt string, assignments []*model.FormationAssignmentInput) ([]*model.FormationAssignment, error)
	CleanupFormationAssignment(ctx context.Context, mappingPair *formationassignment.AssignmentMappingPairWithOperation) (bool, error)
	GetAssignmentsForFormation(ctx context.Context, tenantID, formationID string) ([]*model.FormationAssignment, error)
	Update(ctx context.Context, id string, fa *model.FormationAssignment) error
	GetAssignmentsForFormationWithStates(ctx context.Context, tenantID, formationID string, states []string) ([]*model.FormationAssignment, error)
	GetReverseBySourceAndTarget(ctx context.Context, formationID, sourceID, targetID string) (*model.FormationAssignment, error)
}

// FormationAssignmentConverter converts FormationAssignment between the model.FormationAssignment service-layer representation and graphql.FormationAssignment.
//
//go:generate mockery --name=FormationAssignmentConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentConverter interface {
	MultipleToGraphQL(in []*model.FormationAssignment) ([]*graphql.FormationAssignment, error)
	ToGraphQL(in *model.FormationAssignment) (*graphql.FormationAssignment, error)
}

// TenantFetcher calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
//
//go:generate mockery --name=TenantFetcher --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantFetcher interface {
	FetchOnDemand(ctx context.Context, tenant, parentTenant string) error
}

//go:generate mockery --exported --name=tenantSvc --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantSvc interface {
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// Resolver is the formation resolver
type Resolver struct {
	transact                persistence.Transactioner
	service                 Service
	conv                    Converter
	formationAssignmentSvc  formationAssignmentService
	formationAssignmentConv FormationAssignmentConverter
	fetcher                 TenantFetcher
	tenantSvc               tenantSvc
}

// NewResolver creates formation resolver
func NewResolver(transact persistence.Transactioner, service Service, conv Converter, formationAssignmentSvc formationAssignmentService, formationAssignmentConv FormationAssignmentConverter, fetcher TenantFetcher, tenantSvc tenantSvc) *Resolver {
	return &Resolver{
		transact:                transact,
		service:                 service,
		conv:                    conv,
		formationAssignmentSvc:  formationAssignmentSvc,
		formationAssignmentConv: formationAssignmentConv,
		fetcher:                 fetcher,
		tenantSvc:               tenantSvc,
	}
}

func (r *Resolver) getFormation(ctx context.Context, get func(context.Context) (*model.Formation, error)) (*graphql.Formation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formation, err := get(ctx)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(formation)
}

// FormationByName returns a Formation by its name
func (r *Resolver) FormationByName(ctx context.Context, name string) (*graphql.Formation, error) {
	return r.getFormation(ctx, func(ctx context.Context) (*model.Formation, error) {
		tnt, err := tenant.LoadFromContext(ctx)
		if err != nil {
			return nil, err
		}

		return r.service.GetFormationByName(ctx, name, tnt)
	})
}

// Formation returns a Formation by its id
func (r *Resolver) Formation(ctx context.Context, id string) (*graphql.Formation, error) {
	return r.getFormation(ctx, func(ctx context.Context) (*model.Formation, error) {
		return r.service.Get(ctx, id)
	})
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

	formations, err := r.conv.MultipleToGraphQL(formationPage.Data)
	if err != nil {
		return nil, err
	}

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

// FormationsForObject returns all Formations `objectID` is part of
func (r *Resolver) FormationsForObject(ctx context.Context, objectID string) ([]*graphql.Formation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formations, err := r.service.ListFormationsForObject(ctx, objectID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	formationsGQL, err := r.conv.MultipleToGraphQL(formations)
	if err != nil {
		return nil, err
	}

	return formationsGQL, nil
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

	return r.conv.ToGraphQL(newFormation)
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

	return r.conv.ToGraphQL(deletedFormation)
}

// AssignFormation assigns object to the provided formation
func (r *Resolver) AssignFormation(ctx context.Context, objectID string, objectType graphql.FormationObjectType, formation graphql.FormationInput, initialConfigurations []*graphql.InitialConfiguration) (*graphql.Formation, error) {
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

	formationFromDB, err := r.service.GetFormationByName(ctx, formation.Name, tnt)
	if err != nil {
		return nil, err
	}

	assignmentsForFormation, err := r.formationAssignmentSvc.GetAssignmentsForFormation(ctx, tnt, formationFromDB.ID)
	if err != nil {
		return nil, err
	}

	participants := make(map[string]struct{}, len(assignmentsForFormation))
	for _, assignment := range assignmentsForFormation {
		participants[assignment.Source] = struct{}{}
		participants[assignment.Target] = struct{}{}
	}

	initCfgsSourceToTarget := make(model.InitialConfigurations)
	for _, cfg := range initialConfigurations {
		if cfg.SourceID != objectID && cfg.TargetID != objectID {
			return nil, errors.Errorf("Initial Configuration does not contain assigned object %s as \"source\" or \"target\": %v", objectID, cfg)
		}

		_, isSourceParticipant := participants[cfg.SourceID]
		_, isTargetParticipant := participants[cfg.TargetID]
		if (!isSourceParticipant && cfg.SourceID != objectID) || (!isTargetParticipant && cfg.TargetID != objectID) {
			return nil, errors.Errorf("Initial Configurations contains non-participant \"source\" or \"target\": %v", cfg)
		}

		if _, ok := initCfgsSourceToTarget[cfg.SourceID]; !ok {
			initCfgsSourceToTarget[cfg.SourceID] = make(map[string]json.RawMessage)
		}

		initialConfig := json.RawMessage(cfg.Configuration)
		initCfgsSourceToTarget[cfg.SourceID][cfg.TargetID] = initialConfig
	}

	tenantMapping, err := r.tenantSvc.GetTenantByID(ctx, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting parent tenant by internal ID %q...", tnt)
	}
	externalTnt := tenantMapping.ExternalTenant

	if objectType == graphql.FormationObjectTypeTenant {
		if err := r.fetcher.FetchOnDemand(ctx, objectID, externalTnt); err != nil {
			return nil, errors.Wrapf(err, "while trying to create if not exists subaccount %s", objectID)
		}
	}

	newFormation, err := r.service.AssignFormation(ctx, tnt, objectID, objectType, r.conv.FromGraphQL(formation), initCfgsSourceToTarget)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation)
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

	newFormation, err := r.service.UnassignFormation(ctx, tnt, objectID, objectType, r.conv.FromGraphQL(formation), false)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation)
}

// UnassignFormationGlobal unassigns the objectID from the provided formation globally
func (r *Resolver) UnassignFormationGlobal(ctx context.Context, objectID string, objectType graphql.FormationObjectType, formationID string) (*graphql.Formation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formation, err := r.service.GetGlobalByID(ctx, formationID)
	if err != nil {
		return nil, err
	}

	ctx = tenant.SaveToContext(ctx, formation.TenantID, "")

	newFormation, err := r.service.UnassignFormation(ctx, formation.TenantID, objectID, objectType, *formation, false)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation)
}

// FormationAssignments retrieves a page of FormationAssignments for the specified Formation
func (r *Resolver) FormationAssignments(ctx context.Context, obj *graphql.Formation, first *int, after *graphql.PageCursor) (*graphql.FormationAssignmentPage, error) {
	param := dataloader.ParamFormationAssignment{ID: obj.ID, Ctx: ctx, Tenant: obj.TenantID, First: first, After: after}
	return dataloader.FormationFor(ctx).FormationAssignmentByID.Load(param)
}

// FormationAssignmentsDataLoader retrieves a page of FormationAssignments for each Formation ID in the keys argument
// The sub-resolver is referred from tenant scoped resolvers but in some cases it can be referred in non tenant scoped resolver(e.g. formationsForObject)
// In order to work correctly in both cases the formations from the keys are processed grouped by tenant. After the processing the correct order of the
// assignment pages(same as the order of the formations from the keys) must be ensured as the dataloaders depend on the order of the results when resolving the sub-resolvers
func (r *Resolver) FormationAssignmentsDataLoader(keys []dataloader.ParamFormationAssignment) ([]*graphql.FormationAssignmentPage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Formations found")}
	}

	ctx := keys[0].Ctx
	formationIDs := make([]string, 0, len(keys)) // save the order of the formations
	formationIDsByTenant := make(map[string][]string, len(keys))
	for _, key := range keys {
		formationIDs = append(formationIDs, key.ID)
		tnt := key.Tenant
		_, ok := formationIDsByTenant[tnt]
		if !ok {
			formationIDsByTenant[tnt] = make([]string, 0, len(keys))
		}
		formationIDsByTenant[tnt] = append(formationIDsByTenant[tnt], key.ID)
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

	gqlFormationAssignmentPages := make([]*graphql.FormationAssignmentPage, 0, len(keys))

	formationIDToFormationAssignments := make(map[string]*model.FormationAssignmentPage, len(keys))
	for formationTenant, tenantFormationIDs := range formationIDsByTenant {
		ctxWithTenant := tenant.SaveToContext(ctx, formationTenant, "")
		formationAssignmentPages, err := r.formationAssignmentSvc.ListByFormationIDs(ctxWithTenant, tenantFormationIDs, *keys[0].First, cursor) // ListByFormationIDs underneath will map the FAs to the input tenantFormationIDs
		if err != nil {
			return nil, []error{err}
		}

		for i, formationID := range tenantFormationIDs {
			formationIDToFormationAssignments[formationID] = formationAssignmentPages[i] // map the FAs to the formationID of the given tenant; we rely on the index because of the ListByFormationIDs ordering
		}
	}

	for _, formationID := range formationIDs { // loop the initial order of the formations
		page := formationIDToFormationAssignments[formationID] // get the FAs for the given formation regardless of the tenant
		fas, err := r.formationAssignmentConv.MultipleToGraphQL(page.Data)
		if err != nil {
			return nil, []error{err}
		}

		gqlFormationAssignmentPages = append(gqlFormationAssignmentPages, &graphql.FormationAssignmentPage{Data: fas, TotalCount: page.TotalCount, PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(page.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(page.PageInfo.EndCursor),
			HasNextPage: page.PageInfo.HasNextPage,
		}})
	}

	if err = tx.Commit(); err != nil {
		return nil, []error{err}
	}

	return gqlFormationAssignmentPages, nil
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
	ctx = tenant.SaveToContext(ctx, obj.TenantID, "")

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

// Status retrieves a Status for the specified Formation
func (r *Resolver) Status(ctx context.Context, obj *graphql.Formation) (*graphql.FormationStatus, error) {
	param := dataloader.ParamFormationStatus{ID: obj.ID, State: obj.State, Message: obj.Error.Message, ErrorCode: obj.Error.ErrorCode, Ctx: ctx, Tenant: obj.TenantID}
	return dataloader.FormationStatusFor(ctx).FormationStatusByID.Load(param)
}

// StatusDataLoader retrieves a Status for each Formation ID in the keys argument
// The sub-resolver is referred from tenant scoped resolvers but in some cases it can be referred in non tenant scoped resolver(e.g. formationsForObject)
// In order to work correctly in both cases the formations from the keys are processed grouped by tenant. After the processing the correct order of the
// statuses(same as the order of the formations from the keys) must be ensured as the dataloaders depend on the order of the results when resolving the sub-resolvers
func (r *Resolver) StatusDataLoader(keys []dataloader.ParamFormationStatus) ([]*graphql.FormationStatus, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Formations found")}
	}

	ctx := keys[0].Ctx
	formationIDs := make([]string, 0, len(keys)) // save the order of the formations
	formationIDsByTenant := make(map[string][]string, len(keys))
	for _, key := range keys {
		formationIDs = append(formationIDs, key.ID)
		tnt := key.Tenant
		_, ok := formationIDsByTenant[tnt]
		if !ok {
			formationIDsByTenant[tnt] = make([]string, 0, len(keys))
		}
		formationIDsByTenant[tnt] = append(formationIDsByTenant[tnt], key.ID)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationIDToFormationAssignments := make(map[string][]*model.FormationAssignment, len(formationIDs))
	for formationTenant, tenantFormationIDs := range formationIDsByTenant {
		ctxWithTenant := tenant.SaveToContext(ctx, formationTenant, "")
		formationAssignmentsPerFormationForTenant, err := r.formationAssignmentSvc.ListByFormationIDsNoPaging(ctxWithTenant, tenantFormationIDs)
		if err != nil {
			return nil, []error{err}
		}

		for i, formationID := range tenantFormationIDs {
			formationIDToFormationAssignments[formationID] = formationAssignmentsPerFormationForTenant[i] // map the FAs to the formationID of the given tenant; we rely on the index because of the ListByFormationIDs ordering
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, []error{err}
	}

	gqlFormationStatuses := make([]*graphql.FormationStatus, 0, len(formationIDs))
	for i, formationID := range formationIDs {
		formationAssignments := formationIDToFormationAssignments[formationID]

		var condition graphql.FormationStatusCondition
		var formationStatusErrors []*graphql.FormationStatusError

		switch formationState := keys[i].State; formationState {
		case string(model.ReadyFormationState):
			condition = graphql.FormationStatusConditionReady
		case string(model.DraftFormationState):
			condition = graphql.FormationStatusConditionDraft
		case string(model.InitialFormationState), string(model.DeletingFormationState):
			condition = graphql.FormationStatusConditionInProgress
		case string(model.CreateErrorFormationState), string(model.DeleteErrorFormationState):
			condition = graphql.FormationStatusConditionError
			formationStatusErrors = append(formationStatusErrors, &graphql.FormationStatusError{Message: keys[i].Message, ErrorCode: keys[i].ErrorCode})
		}

		if condition != graphql.FormationStatusConditionDraft {
			for _, fa := range formationAssignments {
				if fa.IsInErrorState() {
					condition = graphql.FormationStatusConditionError

					if fa.Error == nil {
						formationStatusErrors = append(formationStatusErrors, &graphql.FormationStatusError{AssignmentID: &fa.ID})
						continue
					}

					var assignmentError formationassignment.AssignmentErrorWrapper
					if err = json.Unmarshal(fa.Error, &assignmentError); err != nil {
						return nil, []error{errors.Wrapf(err, "while unmarshalling formation assignment error with assignment ID %q", fa.ID)}
					}

					formationStatusErrors = append(formationStatusErrors, &graphql.FormationStatusError{
						AssignmentID: &fa.ID,
						Message:      assignmentError.Error.Message,
						ErrorCode:    int(assignmentError.Error.ErrorCode),
					})
				} else if condition != graphql.FormationStatusConditionError && fa.IsInProgressState() {
					condition = graphql.FormationStatusConditionInProgress
				}
			}
		}

		gqlFormationStatuses = append(gqlFormationStatuses, &graphql.FormationStatus{
			Condition: condition,
			Errors:    formationStatusErrors,
		})
	}

	return gqlFormationStatuses, nil
}

// ResynchronizeFormationNotifications sends all notifications that are in error or initial state
func (r *Resolver) ResynchronizeFormationNotifications(ctx context.Context, formationID string, reset *bool) (*graphql.Formation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	shouldReset := false
	if reset != nil {
		shouldReset = *reset
	}

	updatedFormation, err := r.service.ResynchronizeFormationNotifications(ctx, formationID, shouldReset)
	if err != nil {
		return nil, errors.Wrapf(err, "while resynchronizing formation with ID: %s", formationID)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(updatedFormation)
}

// FinalizeDraftFormation changes the formation state to initial and start processing the formation and formation assignment notifications
func (r *Resolver) FinalizeDraftFormation(ctx context.Context, formationID string) (*graphql.Formation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	updatedFormation, err := r.service.FinalizeDraftFormation(ctx, formationID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(updatedFormation)
}
