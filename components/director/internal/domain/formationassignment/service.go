package formationassignment

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/go-multierror"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// FormationAssignmentRepository represents the Formation Assignment repository layer
//
//go:generate mockery --name=FormationAssignmentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentRepository interface {
	Create(ctx context.Context, item *model.FormationAssignment) error
	GetByTargetAndSource(ctx context.Context, target, source, tenantID, formationID string) (*model.FormationAssignment, error)
	Get(ctx context.Context, id, tenantID string) (*model.FormationAssignment, error)
	GetGlobalByID(ctx context.Context, id string) (*model.FormationAssignment, error)
	GetGlobalByIDAndFormationID(ctx context.Context, id, formationID string) (*model.FormationAssignment, error)
	GetForFormation(ctx context.Context, tenantID, id, formationID string) (*model.FormationAssignment, error)
	GetAssignmentsForFormationWithStates(ctx context.Context, tenantID, formationID string, states []string) ([]*model.FormationAssignment, error)
	GetReverseBySourceAndTarget(ctx context.Context, tenantID, formationID, sourceID, targetID string) (*model.FormationAssignment, error)
	List(ctx context.Context, pageSize int, cursor, tenantID string) (*model.FormationAssignmentPage, error)
	ListByFormationIDs(ctx context.Context, tenantID string, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error)
	ListByFormationIDsNoPaging(ctx context.Context, tenantID string, formationIDs []string) ([][]*model.FormationAssignment, error)
	ListAllForObject(ctx context.Context, tenant, formationID, objectID string) ([]*model.FormationAssignment, error)
	ListAllForObjectIDs(ctx context.Context, tenant, formationID string, objectIDs []string) ([]*model.FormationAssignment, error)
	ListForIDs(ctx context.Context, tenant string, ids []string) ([]*model.FormationAssignment, error)
	Update(ctx context.Context, model *model.FormationAssignment) error
	Delete(ctx context.Context, id, tenantID string) error
	Exists(ctx context.Context, id, tenantID string) (bool, error)
}

//go:generate mockery --exported --name=formationAssignmentConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentConverter interface {
	ToInput(assignment *model.FormationAssignment) *model.FormationAssignmentInput
}

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.RuntimeContext, error)
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.Runtime, error)
}

//go:generate mockery --exported --name=webhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookRepository interface {
	GetByIDAndWebhookType(ctx context.Context, tenant, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error)
}

//go:generate mockery --exported --name=webhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
}

//go:generate mockery --exported --name=tenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantRepository interface {
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetCustomerIDParentRecursively(ctx context.Context, tenant string) (string, error)
}

// Used for testing
//nolint
//
//go:generate mockery --exported --name=templateInput --output=automock --outpkg=automock --case=underscore --disable-version-string
type templateInput interface {
	webhookdir.TemplateInput
	GetParticipantsIDs() []string
	GetAssignment() *model.FormationAssignment
	GetReverseAssignment() *model.FormationAssignment
	SetAssignment(*model.FormationAssignment)
	SetReverseAssignment(*model.FormationAssignment)
	Clone() webhookdir.FormationAssignmentTemplateInput
}

// UIDService generates UUIDs for new entities
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo                         FormationAssignmentRepository
	uidSvc                       UIDService
	applicationRepository        applicationRepository
	runtimeRepo                  runtimeRepository
	runtimeContextRepo           runtimeContextRepository
	formationAssignmentConverter formationAssignmentConverter
	notificationService          notificationService
}

// NewService creates a FormationTemplate service
func NewService(repo FormationAssignmentRepository, uidSvc UIDService, applicationRepository applicationRepository, runtimeRepository runtimeRepository, runtimeContextRepo runtimeContextRepository, formationAssignmentConverter formationAssignmentConverter, notificationService notificationService) *service {
	return &service{
		repo:                         repo,
		uidSvc:                       uidSvc,
		applicationRepository:        applicationRepository,
		runtimeRepo:                  runtimeRepository,
		runtimeContextRepo:           runtimeContextRepo,
		formationAssignmentConverter: formationAssignmentConverter,
		notificationService:          notificationService,
	}
}

// Create creates a Formation Assignment using `in`
func (s *service) Create(ctx context.Context, in *model.FormationAssignmentInput) (string, error) {
	formationAssignmentID := s.uidSvc.Generate()
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}
	log.C(ctx).Debugf("ID: %q generated for formation assignment for tenant with ID: %q", formationAssignmentID, tenantID)

	log.C(ctx).Infof("Creating formation assignment with source: %q and source type: %q, and target: %q with target type: %q", in.Source, in.SourceType, in.Target, in.TargetType)
	if err = s.repo.Create(ctx, in.ToModel(formationAssignmentID, tenantID)); err != nil {
		return "", errors.Wrapf(err, "while creating formation assignment for formation with ID: %q", in.FormationID)
	}

	return formationAssignmentID, nil
}

// CreateIfNotExists creates a Formation Assignment using `in`
func (s *service) CreateIfNotExists(ctx context.Context, in *model.FormationAssignmentInput) (string, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	existingEntity, err := s.repo.GetByTargetAndSource(ctx, in.Target, in.Source, tenantID, in.FormationID)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return "", errors.Wrapf(err, "while getting formation assignment by target %q and source %q", in.Target, in.Source)
	}
	if err != nil && apperrors.IsNotFoundError(err) {
		return s.Create(ctx, in)
	}
	return existingEntity.ID, nil
}

// Get queries Formation Assignment matching ID `id`
func (s *service) Get(ctx context.Context, id string) (*model.FormationAssignment, error) {
	log.C(ctx).Infof("Getting formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	fa, err := s.repo.Get(ctx, id, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignment with ID: %q and tenant: %q", id, tenantID)
	}

	return fa, nil
}

// GetAssignmentsForFormationWithStates retrieves formation assignments matching formation ID `formationID` and with state among `states` for tenant with ID `tenantID`
// TODO: Unit Test
func (s *service) GetAssignmentsForFormationWithStates(ctx context.Context, tenantID, formationID string, states []string) ([]*model.FormationAssignment, error) {
	formationAssignments, err := s.repo.GetAssignmentsForFormationWithStates(ctx, tenantID, formationID, states)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignments with error states for formation with ID: %q and tenant: %q", formationID, tenantID)
	}

	return formationAssignments, nil
}

// GetGlobalByID retrieves the formation assignment matching ID `id` globally without tenant parameter
func (s *service) GetGlobalByID(ctx context.Context, id string) (*model.FormationAssignment, error) {
	log.C(ctx).Infof("Getting formation assignment with ID: %q globally", id)

	fa, err := s.repo.GetGlobalByID(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignment with ID: %q globally", id)
	}

	return fa, nil
}

// GetGlobalByIDAndFormationID retrieves the formation assignment matching ID `id` and formation ID `formationID` globally, without tenant parameter
func (s *service) GetGlobalByIDAndFormationID(ctx context.Context, id, formationID string) (*model.FormationAssignment, error) {
	log.C(ctx).Infof("Getting formation assignment with ID: %q and formation ID: %q globally", id, formationID)

	fa, err := s.repo.GetGlobalByIDAndFormationID(ctx, id, formationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignment with ID: %q and formation ID: %q globally", id, formationID)
	}

	return fa, nil
}

// GetForFormation retrieves the Formation Assignment with the provided `id` associated with Formation with id `formationID`
func (s *service) GetForFormation(ctx context.Context, id, formationID string) (*model.FormationAssignment, error) {
	log.C(ctx).Infof("Getting formation assignment for ID: %q and formationID: %q", id, formationID)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	fa, err := s.repo.GetForFormation(ctx, tenantID, id, formationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignment with ID: %q for formation with ID: %q", id, formationID)
	}

	return fa, nil
}

// GetReverseBySourceAndTarget retrieves the Formation Assignment with the provided `id` associated with Formation with id `formationID`
func (s *service) GetReverseBySourceAndTarget(ctx context.Context, formationID, sourceID, targetID string) (*model.FormationAssignment, error) {
	log.C(ctx).Infof("Getting reverse formation assignment for formation ID: %q and source: %q and target: %q", formationID, sourceID, targetID)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	reverseFA, err := s.repo.GetReverseBySourceAndTarget(ctx, tenantID, formationID, sourceID, targetID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting reverse formation assignment for formation ID: %q and source: %q and target: %q", formationID, sourceID, targetID)
	}

	return reverseFA, nil
}

// List pagination lists Formation Assignment based on `pageSize` and `cursor`
func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.FormationAssignmentPage, error) {
	log.C(ctx).Info("Listing formation assignments")

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.List(ctx, pageSize, cursor, tenantID)
}

// ListByFormationIDs retrieves a pages of Formation Assignment objects for each of the provided formation IDs
func (s *service) ListByFormationIDs(ctx context.Context, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error) {
	log.C(ctx).Infof("Listing formation assignment for formation with IDs: %q", formationIDs)

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.ListByFormationIDs(ctx, tnt, formationIDs, pageSize, cursor)
}

func (s *service) ListByFormationIDsNoPaging(ctx context.Context, formationIDs []string) ([][]*model.FormationAssignment, error) {
	log.C(ctx).Infof("Listing all formation assignment for formation with IDs: %q", formationIDs)

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.ListByFormationIDsNoPaging(ctx, tnt, formationIDs)
}

// ListFormationAssignmentsForObjectID retrieves all Formation Assignment objects for formation with ID `formationID` that have `objectID` as source or target
func (s *service) ListFormationAssignmentsForObjectID(ctx context.Context, formationID, objectID string) ([]*model.FormationAssignment, error) {
	log.C(ctx).Infof("Listing formation assignments for object ID: %q and formation ID: %q", objectID, formationID)
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.ListAllForObject(ctx, tnt, formationID, objectID)
}

// ListFormationAssignmentsForObjectIDs retrieves all Formation Assignment objects for formation with ID `formationID` that have any of the `objectIDs` as source or target
func (s *service) ListFormationAssignmentsForObjectIDs(ctx context.Context, formationID string, objectIDs []string) ([]*model.FormationAssignment, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.ListAllForObjectIDs(ctx, tnt, formationID, objectIDs)
}

// Update updates a Formation Assignment matching ID `id` using `in`
func (s *service) Update(ctx context.Context, id string, in *model.FormationAssignmentInput) error {
	log.C(ctx).Infof("Updating formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	if exists, err := s.repo.Exists(ctx, id, tenantID); err != nil {
		return errors.Wrapf(err, "while ensuring formation assignment with ID: %q exists", id)
	} else if !exists {
		return apperrors.NewNotFoundError(resource.FormationAssignment, id)
	}

	if err = s.repo.Update(ctx, in.ToModel(id, tenantID)); err != nil {
		return errors.Wrapf(err, "while updating formation assignment with ID: %q", id)
	}

	return nil
}

// Delete deletes a Formation Assignment matching ID `id`
func (s *service) Delete(ctx context.Context, id string) error {
	log.C(ctx).Infof("Deleting formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	if err := s.repo.Delete(ctx, id, tenantID); err != nil {
		return errors.Wrapf(err, "while deleting formation assignment with ID: %q", id)
	}
	return nil
}

// Exists check if a Formation Assignment with given ID exists
func (s *service) Exists(ctx context.Context, id string) (bool, error) {
	log.C(ctx).Infof("Checking formation assignment existence for ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "while loading tenant from context")
	}

	exists, err := s.repo.Exists(ctx, id, tenantID)
	if err != nil {
		return false, errors.Wrapf(err, "while checking formation assignment existence for ID: %q and tenant: %q", id, tenantID)
	}
	return exists, nil
}

// GenerateAssignments creates and persists two formation assignments per participant in the formation `formation`.
// For the first formation assignment the source is the objectID and the target is participant's ID.
// For the second assignment the source and target are swapped.
//
// In case of objectType==RUNTIME_CONTEXT formationAssignments for the object and it's parent runtime are not generated.
func (s *service) GenerateAssignments(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation) ([]*model.FormationAssignment, error) {
	applications, err := s.applicationRepository.ListByScenariosNoPaging(ctx, tnt, []string{formation.Name})
	if err != nil {
		return nil, err
	}

	runtimes, err := s.runtimeRepo.ListByScenarios(ctx, tnt, []string{formation.Name})
	if err != nil {
		return nil, err
	}

	runtimeContexts, err := s.runtimeContextRepo.ListByScenarios(ctx, tnt, []string{formation.Name})
	if err != nil {
		return nil, err
	}

	allIDs := make([]string, 0, len(applications)+len(runtimes)+len(runtimeContexts))
	appIDs := make(map[string]bool, len(applications))
	rtIDs := make(map[string]bool, len(runtimes))
	rtCtxIDs := make(map[string]bool, len(runtimeContexts))
	for _, app := range applications {
		allIDs = append(allIDs, app.ID)
		appIDs[app.ID] = false
	}
	for _, rt := range runtimes {
		allIDs = append(allIDs, rt.ID)
		rtIDs[rt.ID] = false
	}
	for _, rtCtx := range runtimeContexts {
		allIDs = append(allIDs, rtCtx.ID)
		rtCtxIDs[rtCtx.ID] = false
	}

	allAssignments, err := s.ListFormationAssignmentsForObjectIDs(ctx, formation.ID, allIDs)
	if err != nil {
		return nil, err
	}

	// We should not generate notifications for formation participants that are being unassigned asynchronously
	for _, assignment := range allAssignments {
		if assignment.Source == assignment.Target && assignment.SourceType == assignment.TargetType {
			switch assignment.SourceType {
			case model.FormationAssignmentTypeApplication:
				appIDs[assignment.Source] = true
			case model.FormationAssignmentTypeRuntime:
				rtIDs[assignment.Source] = true
			case model.FormationAssignmentTypeRuntimeContext:
				rtCtxIDs[assignment.Source] = true
			}
		}
	}

	// When assigning an object to a formation we need to create two formation assignments per participant.
	// In the first formation assignment the object we're assigning will be the source and in the second it will be the target
	assignments := make([]*model.FormationAssignmentInput, 0, (len(applications)+len(runtimes)+len(runtimeContexts))*2+1)
	for appID, isAssigned := range appIDs {
		if !isAssigned || appID == objectID {
			continue
		}
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(objectID, objectType, formation, model.FormationAssignmentTypeApplication, appID)...)
	}

	// When runtime context is assigned to formation its parent runtime is unassigned from the formation.
	// There is no need to create formation assignments between the runtime context and the runtime. If such
	// formation assignments were to be created the runtime unassignment from the formation would fail.
	// The reason for this is that the formation assignments are created in one transaction and the runtime
	// unassignment is done in a separate transaction which does not "see" them but will try to delete them.
	parentID := ""
	if objectType == graphql.FormationObjectTypeRuntimeContext {
		rtmCtx, err := s.runtimeContextRepo.GetByID(ctx, tnt, objectID)
		if err != nil {
			return nil, err
		}
		parentID = rtmCtx.RuntimeID
	}
	for runtimeID, isAssigned := range rtIDs {
		if !isAssigned || runtimeID == objectID || runtimeID == parentID {
			continue
		}
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(objectID, objectType, formation, model.FormationAssignmentTypeRuntime, runtimeID)...)
	}

	for runtimeCtxID, isAssigned := range rtCtxIDs {
		if !isAssigned || runtimeCtxID == objectID {
			continue
		}
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(objectID, objectType, formation, model.FormationAssignmentTypeRuntimeContext, runtimeCtxID)...)
	}

	assignments = append(assignments, &model.FormationAssignmentInput{
		FormationID: formation.ID,
		Source:      objectID,
		SourceType:  model.FormationAssignmentType(objectType),
		Target:      objectID,
		TargetType:  model.FormationAssignmentType(objectType),
		State:       string(model.ReadyAssignmentState),
		Value:       nil,
	})

	ids := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		id, err := s.CreateIfNotExists(ctx, assignment)
		if err != nil {
			return nil, errors.Wrapf(err, "while creating formationAssignment for formation %q with source %q of type %q and target %q of type %q", assignment.FormationID, assignment.Source, assignment.SourceType, assignment.Target, assignment.TargetType)
		}
		ids = append(ids, id)
	}

	formationAssignments, err := s.repo.ListForIDs(ctx, tnt, ids)
	if err != nil {
		return nil, errors.Wrap(err, "while listing formationAssignments")
	}

	return formationAssignments, nil
}

// GenerateAssignmentsForParticipant creates in-memory the assignments for two participants in the initial state
func (s *service) GenerateAssignmentsForParticipant(objectID string, objectType graphql.FormationObjectType, formation *model.Formation, participantType model.FormationAssignmentType, participantID string) []*model.FormationAssignmentInput {
	assignments := make([]*model.FormationAssignmentInput, 0, 2)
	assignments = append(assignments, &model.FormationAssignmentInput{
		FormationID: formation.ID,
		Source:      objectID,
		SourceType:  model.FormationAssignmentType(objectType),
		Target:      participantID,
		TargetType:  participantType,
		State:       string(model.InitialAssignmentState),
		Value:       nil,
	})
	assignments = append(assignments, &model.FormationAssignmentInput{
		FormationID: formation.ID,
		Source:      participantID,
		SourceType:  participantType,
		Target:      objectID,
		TargetType:  model.FormationAssignmentType(objectType),
		State:       string(model.InitialAssignmentState),
		Value:       nil,
	})
	return assignments
}

// ProcessFormationAssignments matches the formation assignments with the corresponding notification requests and packs them in FormationAssignmentRequestMapping.
// Each FormationAssignmentRequestMapping is then packed with its reverse in AssignmentMappingPair. Then the provided `formationAssignmentFunc` is executed against the AssignmentMappingPairs.
//
// Assignment and reverseAssignment example
// assignment{source=X, target=Y} - reverseAssignment{source=Y, target=X}
//
// Mapping and reverseMapping example
// mapping{notificationRequest=request, formationAssignment=assignment} - reverseMapping{notificationRequest=reverseRequest, formationAssignment=reverseAssignment}

func (s *service) ProcessFormationAssignments(ctx context.Context, formationAssignmentsForObject []*model.FormationAssignment, runtimeContextIDToRuntimeIDMapping map[string]string, requests []*webhookclient.FormationAssignmentNotificationRequest, formationAssignmentFunc func(context.Context, *AssignmentMappingPair) (bool, error)) error {
	var errs *multierror.Error
	assignmentRequestMappings := s.matchFormationAssignmentsWithRequests(ctx, formationAssignmentsForObject, runtimeContextIDToRuntimeIDMapping, requests)
	alreadyProcessedFAs := make(map[string]bool, 0)
	for _, mapping := range assignmentRequestMappings {
		if alreadyProcessedFAs[mapping.Assignment.FormationAssignment.ID] {
			continue
		}
		isReverseProcessed, err := formationAssignmentFunc(ctx, mapping)
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "while processing formation assignment with id %q", mapping.Assignment.FormationAssignment.ID))
		}
		if isReverseProcessed {
			alreadyProcessedFAs[mapping.ReverseAssignment.FormationAssignment.ID] = true
		}
	}
	log.C(ctx).Infof("Finished processing %d formation assignments", len(formationAssignmentsForObject))

	return errs.ErrorOrNil()
}

// ProcessFormationAssignmentPair prepares and update the `State` and `Config` of the formation assignment based on the response and process the notifications
func (s *service) ProcessFormationAssignmentPair(ctx context.Context, mappingPair *AssignmentMappingPair) (bool, error) {
	var isReverseProcessed bool
	err := s.processFormationAssignmentsWithReverseNotification(ctx, mappingPair, 0, &isReverseProcessed)
	return isReverseProcessed, err
}

func (s *service) processFormationAssignmentsWithReverseNotification(ctx context.Context, mappingPair *AssignmentMappingPair, depth int, isReverseProcessed *bool) error {
	fa := mappingPair.Assignment.FormationAssignment
	log.C(ctx).Infof("Processing formation assignment %q for formation %q with Source: %q of Type: %q and Target: %q of Type: %q and State %q", fa.ID, fa.FormationID, fa.Source, fa.SourceType, fa.Target, fa.TargetType, fa.State)
	assignmentClone := mappingPair.Assignment.Clone()
	var reverseClone *FormationAssignmentRequestMapping
	if mappingPair.ReverseAssignment != nil {
		reverseClone = mappingPair.ReverseAssignment.Clone()
	}
	assignment := assignmentClone.FormationAssignment

	if assignment.State == string(model.ReadyAssignmentState) {
		return nil
	}

	if assignmentClone.Request == nil {
		assignment.State = string(model.ReadyAssignmentState)
		log.C(ctx).Infof("There is no notification to be sent. Setting state for Assignmene with ID %s to %s", assignment.ID, assignment.State)
		if err := s.Update(ctx, assignment.ID, s.formationAssignmentConverter.ToInput(assignment)); err != nil {
			return errors.Wrapf(err, "while updating formation assignment for formation %q with source %q and target %q", assignment.FormationID, assignment.Source, assignment.Target)
		}
		return nil
	}

	response, err := s.notificationService.SendNotification(ctx, assignmentClone.Request)
	if err != nil {
		updateError := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), TechnicalError, model.CreateErrorAssignmentState)
		if updateError != nil {
			return errors.Wrapf(
				updateError,
				"while updating error state: %s",
				errors.Wrapf(err, "while sending notification for formation assignment with ID %q", assignment.ID).Error())
		}
		log.C(ctx).Error(errors.Wrapf(err, "while sending notification for formation assignment with ID %q", assignment.ID).Error())
		return nil
	}

	if response.Error != nil && *response.Error != "" {
		err = s.SetAssignmentToErrorState(ctx, assignment, *response.Error, ClientError, model.CreateErrorAssignmentState)
		if err != nil {
			return errors.Wrapf(err, "while updating error state for formation with ID %q", assignment.ID)
		}

		log.C(ctx).Error(errors.Errorf("Received error from response: %v", *response.Error).Error())
		return nil
	}

	requestWebhookMode := assignmentClone.Request.Webhook.Mode
	if requestWebhookMode != nil && *requestWebhookMode == graphql.WebhookModeAsyncCallback {
		log.C(ctx).Info("The Webhook in the notification is in ASYNC_CALLBACK mode. Waiting for the receiver to report the status on the status API...")
		return nil
	}

	if *response.ActualStatusCode == *response.SuccessStatusCode {
		assignment.State = string(model.ReadyAssignmentState)
		assignment.Value = nil
	}

	if response.IncompleteStatusCode != nil && *response.ActualStatusCode == *response.IncompleteStatusCode {
		assignment.State = string(model.ConfigPendingAssignmentState)
	}

	var shouldSendReverseNotification bool
	if response.Config != nil && *response.Config != "" {
		assignment.Value = []byte(*response.Config)
		shouldSendReverseNotification = true
	}

	storedAssignment, err := s.Get(ctx, assignment.ID)
	if err != nil {
		return errors.Wrapf(err, "while fetching formation assignment with ID: %q", assignment.ID)
	}

	if storedAssignment.State != string(model.ReadyAssignmentState) {
		if err = s.Update(ctx, assignment.ID, s.formationAssignmentConverter.ToInput(assignment)); err != nil {
			return errors.Wrapf(err, "while creating formation assignment for formation %q with source %q and target %q", assignment.FormationID, assignment.Source, assignment.Target)
		}
		log.C(ctx).Infof("Assignment with ID %s was updated with %s state", assignment.ID, assignment.State)
	}

	if shouldSendReverseNotification {
		if reverseClone == nil {
			return nil
		}

		*isReverseProcessed = true

		if depth >= model.NotificationRecursionDepthLimit {
			log.C(ctx).Errorf("Depth limit exceeded for assignments: %q and %q", assignmentClone.FormationAssignment.ID, reverseClone.FormationAssignment.ID)
			return nil
		}

		newAssignment := reverseClone.Clone()
		newReverseAssignment := assignmentClone.Clone()

		if newAssignment.Request != nil {
			newAssignment.Request.Object.SetAssignment(newAssignment.FormationAssignment)
			newAssignment.Request.Object.SetReverseAssignment(newReverseAssignment.FormationAssignment)
		}
		if newReverseAssignment.Request != nil {
			newReverseAssignment.Request.Object.SetAssignment(newReverseAssignment.FormationAssignment)
			newReverseAssignment.Request.Object.SetReverseAssignment(newAssignment.FormationAssignment)
		}

		newAssignmentMappingPair := &AssignmentMappingPair{
			Assignment:        newAssignment,
			ReverseAssignment: newReverseAssignment,
		}

		if err = s.processFormationAssignmentsWithReverseNotification(ctx, newAssignmentMappingPair, depth+1, isReverseProcessed); err != nil {
			return errors.Wrap(err, "while sending reverse notification")
		}
	}

	return nil
}

// CleanupFormationAssignment If the provided mappingPair does not contain notification request the assignment is deleted.
// If the provided pair contains notification request - sends it and adapts the `State` and `Config` of the formation assignment
// based on the response.
// In the case the response is successful it deletes the formation assignment
// In all other cases the `State` and `Config` are updated accordingly
func (s *service) CleanupFormationAssignment(ctx context.Context, mappingPair *AssignmentMappingPair) (bool, error) {
	assignment := mappingPair.Assignment.FormationAssignment
	if mappingPair.Assignment.Request == nil {
		if err := s.Delete(ctx, assignment.ID); err != nil {
			// It is possible that the deletion fails due to some kind of DB constraint, so we will try to update the state
			updateError := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), TechnicalError, model.DeleteErrorAssignmentState)
			if updateError != nil {
				return false, errors.Wrapf(
					updateError,
					"while updating error state: %s",
					errors.Wrapf(err, "while deleting formation assignment with id %q", assignment.ID).Error())
			}
			return false, errors.Wrapf(err, "while deleting formation assignment with id %q", assignment.ID)
		}
		log.C(ctx).Infof("Assignment with ID %s was deleted", assignment.ID)

		return false, nil
	}

	response, err := s.notificationService.SendNotification(ctx, mappingPair.Assignment.Request)
	if err != nil {
		updateError := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), TechnicalError, model.DeleteErrorAssignmentState)
		if updateError != nil {
			return false, errors.Wrapf(
				updateError,
				"while updating error state: %s",
				errors.Wrapf(err, "while sending notification for formation assignment with ID %q", assignment.ID).Error())
		}
		return false, errors.Wrapf(err, "while sending notification for formation assignment with ID %q", assignment.ID)
	}

	requestWebhookMode := mappingPair.Assignment.Request.Webhook.Mode
	if requestWebhookMode != nil && *requestWebhookMode == graphql.WebhookModeAsyncCallback {
		log.C(ctx).Infof("The Webhook in the notification is in %s mode. Updating the assignment state to %s and waiting for the receiver to report the status on the status API...", graphql.WebhookModeAsyncCallback, string(model.DeletingAssignmentState))
		assignment.State = string(model.DeletingAssignmentState)
		if err := s.Update(ctx, assignment.ID, s.formationAssignmentConverter.ToInput(assignment)); err != nil {
			return false, errors.Wrapf(err, "While updating formation assignment with id %q", assignment.ID)
		}
		return false, nil
	}

	if *response.ActualStatusCode == *response.SuccessStatusCode {
		if err = s.Delete(ctx, assignment.ID); err != nil {
			// It is possible that the deletion fails due to some kind of DB constraint, so we will try to update the state
			updateError := s.SetAssignmentToErrorState(ctx, assignment, "error while deleting assignment", TechnicalError, model.DeleteErrorAssignmentState)
			if updateError != nil {
				return false, errors.Wrapf(
					updateError,
					"while updating error state: %s",
					errors.Wrapf(err, "while deleting formation assignment with id %q", assignment.ID).Error())
			}
			return false, errors.Wrapf(err, "while deleting formation assignment with id %q", assignment.ID)
		}
		log.C(ctx).Infof("Assignment with ID %s was deleted", assignment.ID)

		return false, nil
	}

	if response.IncompleteStatusCode != nil && *response.ActualStatusCode == *response.IncompleteStatusCode {
		err = errors.New("Error while deleting assignment: config propagation is not supported on unassign notifications")
		updateErr := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), ClientError, model.DeleteErrorAssignmentState)
		if updateErr != nil {
			return false, errors.Wrapf(updateErr, "while updating error state for formation with ID %q", assignment.ID)
		}
		return false, err
	}

	if response.Error != nil && *response.Error != "" {
		err = s.SetAssignmentToErrorState(ctx, assignment, *response.Error, ClientError, model.DeleteErrorAssignmentState)
		if err != nil {
			return false, errors.Wrapf(err, "while updating error state for formation with ID %q", assignment.ID)
		}
		return false, errors.Errorf("Received error from response: %v", *response.Error)
	}

	return false, nil
}

func (s *service) SetAssignmentToErrorState(ctx context.Context, assignment *model.FormationAssignment, errorMessage string, errorCode AssignmentErrorCode, state model.FormationAssignmentState) error {
	assignment.State = string(state)
	assignmentError := AssignmentErrorWrapper{AssignmentError{
		Message:   errorMessage,
		ErrorCode: errorCode,
	}}
	marshaled, err := json.Marshal(assignmentError)
	if err != nil {
		return errors.Wrapf(err, "While preparing error message for assignment with ID %q", assignment.ID)
	}
	assignment.Value = marshaled
	if err := s.Update(ctx, assignment.ID, s.formationAssignmentConverter.ToInput(assignment)); err != nil {
		return errors.Wrapf(err, "While updating formation assignment with id %q", assignment.ID)
	}
	log.C(ctx).Infof("Assignment with ID %s set to state %s", assignment.ID, assignment.State)
	return nil
}

func (s *service) matchFormationAssignmentsWithRequests(ctx context.Context, assignments []*model.FormationAssignment, runtimeContextIDToRuntimeIDMapping map[string]string, requests []*webhookclient.FormationAssignmentNotificationRequest) []*AssignmentMappingPair {
	formationAssignmentMapping := make([]*FormationAssignmentRequestMapping, 0, len(assignments))
	for i, assignment := range assignments {
		mappingObject := &FormationAssignmentRequestMapping{
			Request:             nil,
			FormationAssignment: assignments[i],
		}
		target := assignment.Target
		if assignment.TargetType == model.FormationAssignmentTypeRuntimeContext {
			log.C(ctx).Infof("Matching for runtime context, fetching associated runtime for runtime context with ID %s", target)

			target = runtimeContextIDToRuntimeIDMapping[assignment.Target]
			log.C(ctx).Infof("Fetched associated runtime with ID %s for runtime context with ID %s", target, assignment.Target)
		}

	assignment:
		for j, request := range requests {
			var objectID string
			if request.Webhook.RuntimeID != nil {
				objectID = *request.Webhook.RuntimeID
			}
			if request.Webhook.ApplicationID != nil {
				objectID = *request.Webhook.ApplicationID
			}

			if objectID != target {
				continue
			}

			participants := request.Object.GetParticipantsIDs()
			for _, id := range participants {
				// We should not generate notifications for self
				if assignment.Source == assignment.Target {
					break assignment
				}
				if assignment.Source == id {
					mappingObject.Request = requests[j]
					break assignment
				}
			}
		}
		formationAssignmentMapping = append(formationAssignmentMapping, mappingObject)
	}

	log.C(ctx).Infof("Mapped %d formation assignments with %d notifications, %d assignments left with no notification", len(assignments), len(requests), len(assignments)-len(requests))
	sourceToTargetToMapping := make(map[string]map[string]*FormationAssignmentRequestMapping)
	for _, mapping := range formationAssignmentMapping {
		if _, ok := sourceToTargetToMapping[mapping.FormationAssignment.Source]; !ok {
			sourceToTargetToMapping[mapping.FormationAssignment.Source] = make(map[string]*FormationAssignmentRequestMapping, len(assignments)/2)
		}
		sourceToTargetToMapping[mapping.FormationAssignment.Source][mapping.FormationAssignment.Target] = mapping
	}
	// Make mapping
	assignmentMappingPairs := make([]*AssignmentMappingPair, 0, len(assignments))

	for _, mapping := range formationAssignmentMapping {
		var reverseMapping *FormationAssignmentRequestMapping
		if mappingsForTarget, ok := sourceToTargetToMapping[mapping.FormationAssignment.Target]; ok {
			if actualReverseMapping, ok := mappingsForTarget[mapping.FormationAssignment.Source]; ok {
				reverseMapping = actualReverseMapping
			}
		}
		assignmentMappingPairs = append(assignmentMappingPairs, &AssignmentMappingPair{
			Assignment:        mapping,
			ReverseAssignment: reverseMapping,
		})
		if mapping.Request != nil {
			mapping.Request.Object.SetAssignment(mapping.FormationAssignment)
			if reverseMapping != nil {
				mapping.Request.Object.SetReverseAssignment(reverseMapping.FormationAssignment)
			}
		}
		if reverseMapping != nil && reverseMapping.Request != nil {
			reverseMapping.Request.Object.SetAssignment(reverseMapping.FormationAssignment)
			reverseMapping.Request.Object.SetReverseAssignment(mapping.FormationAssignment)
		}
	}
	return assignmentMappingPairs
}

// FormationAssignmentRequestMapping represents the mapping between the notification request and formation assignment
type FormationAssignmentRequestMapping struct {
	Request             *webhookclient.FormationAssignmentNotificationRequest
	FormationAssignment *model.FormationAssignment
}

// Clone returns a copy of the FormationAssignmentRequestMapping
func (f *FormationAssignmentRequestMapping) Clone() *FormationAssignmentRequestMapping {
	var request *webhookclient.FormationAssignmentNotificationRequest
	if f.Request != nil {
		request = f.Request.Clone()
	}
	return &FormationAssignmentRequestMapping{
		Request: request,
		FormationAssignment: &model.FormationAssignment{
			ID:          f.FormationAssignment.ID,
			FormationID: f.FormationAssignment.FormationID,
			TenantID:    f.FormationAssignment.TenantID,
			Source:      f.FormationAssignment.Source,
			SourceType:  f.FormationAssignment.SourceType,
			Target:      f.FormationAssignment.Target,
			TargetType:  f.FormationAssignment.TargetType,
			State:       f.FormationAssignment.State,
			Value:       f.FormationAssignment.Value,
		},
	}
}

// AssignmentErrorCode represents error code used to differentiate the source of the error
type AssignmentErrorCode int

const (
	// TechnicalError indicates that the reason for the error is technical - for example networking issue
	TechnicalError = 1
	// ClientError indicates that the error was returned from the client
	ClientError = 2
)

// AssignmentMappingPair represents a pair of FormationAssignmentRequestMapping and its reverse
type AssignmentMappingPair struct {
	Assignment        *FormationAssignmentRequestMapping
	ReverseAssignment *FormationAssignmentRequestMapping
}

// AssignmentError error struct used for storing the errors that occur during the FormationAssignment processing
type AssignmentError struct {
	Message   string              `json:"message"`
	ErrorCode AssignmentErrorCode `json:"errorCode"`
}

// AssignmentErrorWrapper wrapper for AssignmentError
type AssignmentErrorWrapper struct {
	Error AssignmentError `json:"error"`
}
