package formationassignment

import (
	"context"
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/director/internal/domain/notifications"
	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationassignment"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
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
	GetAssignmentsForFormation(ctx context.Context, tenantID, formationID string) ([]*model.FormationAssignment, error)
	GetReverseBySourceAndTarget(ctx context.Context, tenantID, formationID, sourceID, targetID string) (*model.FormationAssignment, error)
	List(ctx context.Context, pageSize int, cursor, tenantID string) (*model.FormationAssignmentPage, error)
	ListAllForFormation(ctx context.Context, tenant, formationID string) ([]*model.FormationAssignment, error)
	ListByFormationIDs(ctx context.Context, tenantID string, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error)
	ListByFormationIDsNoPaging(ctx context.Context, tenantID string, formationIDs []string) ([][]*model.FormationAssignment, error)
	ListAllForObject(ctx context.Context, tenant, formationID, objectID string) ([]*model.FormationAssignment, error)
	ListAllForObjectIDs(ctx context.Context, tenant, formationID string, objectIDs []string) ([]*model.FormationAssignment, error)
	ListAllForObjectGlobal(ctx context.Context, objectID string) ([]*model.FormationAssignment, error)
	ListForIDs(ctx context.Context, tenant string, ids []string) ([]*model.FormationAssignment, error)
	Update(ctx context.Context, model *model.FormationAssignment) error
	Delete(ctx context.Context, id, tenantID string) error
	DeleteAssignmentsForObjectID(ctx context.Context, tnt, formationID, objectID string) error
	Exists(ctx context.Context, id, tenantID string) (bool, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
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
	GetParentsRecursivelyByExternalTenant(ctx context.Context, externalTenant string) ([]*model.BusinessTenantMapping, error)
}

// UIDService generates UUIDs for new entities
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

//go:generate mockery --exported --name=labelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelService interface {
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
}

//go:generate mockery --exported --name=constraintEngine --output=automock --outpkg=automock --case=underscore --disable-version-string
type constraintEngine interface {
	EnforceConstraints(ctx context.Context, location formationconstraint.JoinPointLocation, details formationconstraint.JoinPointDetails, formationTemplateID string) error
}

//go:generate mockery --exported --name=statusService --output=automock --outpkg=automock --case=underscore --disable-version-string
type statusService interface {
	UpdateWithConstraints(ctx context.Context, notificationStatusReport *statusreport.NotificationStatusReport, fa *model.FormationAssignment, operation model.FormationOperation) error
	DeleteWithConstraints(ctx context.Context, id string, notificationStatusReport *statusreport.NotificationStatusReport) error
}

//go:generate mockery --exported --name=faNotificationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type faNotificationService interface {
	GenerateFormationAssignmentNotificationExt(ctx context.Context, faRequestMapping, reverseFaRequestMapping *notifications.FormationAssignmentRequestMapping, operation model.FormationOperation) (*webhookclient.FormationAssignmentNotificationRequestExt, error)
	PrepareDetailsForNotificationStatusReturned(ctx context.Context, tenantID string, fa *model.FormationAssignment, operation model.FormationOperation, notificationStatusReport *statusreport.NotificationStatusReport) (*formationconstraint.NotificationStatusReturnedOperationDetails, error)
}

type service struct {
	repo                       FormationAssignmentRepository
	uidSvc                     UIDService
	runtimeContextRepo         runtimeContextRepository
	notificationService        notificationService
	faNotificationService      faNotificationService
	assignmentOperationService assignmentOperationService
	labelService               labelService
	formationRepository        formationRepository
	statusService              statusService
	runtimeTypeLabelKey        string
	applicationTypeLabelKey    string
}

// NewService creates a Formation Assignment service
func NewService(repo FormationAssignmentRepository, uidSvc UIDService, runtimeContextRepo runtimeContextRepository, notificationService notificationService, faNotificationService faNotificationService, assignmentOperationService assignmentOperationService, labelService labelService, formationRepository formationRepository, statusService statusService, runtimeTypeLabelKey, applicationTypeLabelKey string) *service {
	return &service{
		repo:                       repo,
		uidSvc:                     uidSvc,
		runtimeContextRepo:         runtimeContextRepo,
		notificationService:        notificationService,
		faNotificationService:      faNotificationService,
		assignmentOperationService: assignmentOperationService,
		labelService:               labelService,
		formationRepository:        formationRepository,
		statusService:              statusService,
		runtimeTypeLabelKey:        runtimeTypeLabelKey,
		applicationTypeLabelKey:    applicationTypeLabelKey,
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
func (s *service) GetAssignmentsForFormationWithStates(ctx context.Context, tenantID, formationID string, states []string) ([]*model.FormationAssignment, error) {
	formationAssignments, err := s.repo.GetAssignmentsForFormationWithStates(ctx, tenantID, formationID, states)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignments with states for formation with ID: %q and tenant: %q", formationID, tenantID)
	}

	return formationAssignments, nil
}

// GetAssignmentsForFormation retrieves formation assignments matching formation ID `formationID` for tenant with ID `tenantID`
func (s *service) GetAssignmentsForFormation(ctx context.Context, tenantID, formationID string) ([]*model.FormationAssignment, error) {
	formationAssignments, err := s.repo.GetAssignmentsForFormation(ctx, tenantID, formationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignments for formation with ID: %q and tenant: %q", formationID, tenantID)
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

// DeleteAssignmentsForObjectID deletes formation assignments for formation for given objectID
func (s *service) DeleteAssignmentsForObjectID(ctx context.Context, formationID, objectID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.DeleteAssignmentsForObjectID(ctx, tnt, formationID, objectID)
}

// ListFormationAssignmentsForObjectIDs retrieves all Formation Assignment objects for formation with ID `formationID` that have any of the `objectIDs` as source or target
func (s *service) ListFormationAssignmentsForObjectIDs(ctx context.Context, formationID string, objectIDs []string) ([]*model.FormationAssignment, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.ListAllForObjectIDs(ctx, tnt, formationID, objectIDs)
}

// ListAllForObjectGlobal retrieves all Formation Assignment objects that have the `objectID` as source or target across all tenants
func (s *service) ListAllForObjectGlobal(ctx context.Context, objectID string) ([]*model.FormationAssignment, error) {
	return s.repo.ListAllForObjectGlobal(ctx, objectID)
}

// Update updates a Formation Assignment matching ID `id` using `in`
func (s *service) Update(ctx context.Context, id string, fa *model.FormationAssignment) error {
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
	err = s.repo.Update(ctx, fa)
	if apperrors.IsUnauthorizedError(err) {
		return apperrors.NewNotFoundError(resource.FormationAssignment, id)
	}
	if err != nil {
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

	err = s.repo.Delete(ctx, id, tenantID)
	if apperrors.IsUnauthorizedError(err) {
		return apperrors.NewNotFoundError(resource.FormationAssignment, id)
	}
	if err != nil {
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

// GenerateAssignments generates two formation assignments per participant in the formation `formation`.
// For the first formation assignment the source is the objectID and the target is participant's ID.
// For the second assignment the source and target are swapped.
//
// In case of objectType==RUNTIME_CONTEXT formationAssignments for the object and it's parent runtime are not generated.
func (s *service) GenerateAssignments(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation, initialConfigurations model.InitialConfigurations) ([]*model.FormationAssignmentInput, error) {
	allAssignments, err := s.repo.ListAllForFormation(ctx, tnt, formation.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing formation assignments for formation %q", formation.ID)
	}

	appsInFormation := make(map[string]struct{})
	rtmsInFormation := make(map[string]struct{})
	rtmCtxsInFormation := make(map[string]struct{})
	for _, assignment := range allAssignments {
		switch assignment.SourceType {
		case model.FormationAssignmentTypeApplication:
			appsInFormation[assignment.Source] = struct{}{}
		case model.FormationAssignmentTypeRuntime:
			rtmsInFormation[assignment.Source] = struct{}{}
		case model.FormationAssignmentTypeRuntimeContext:
			rtmCtxsInFormation[assignment.Source] = struct{}{}
		}
		switch assignment.TargetType {
		case model.FormationAssignmentTypeApplication:
			appsInFormation[assignment.Target] = struct{}{}
		case model.FormationAssignmentTypeRuntime:
			rtmsInFormation[assignment.Target] = struct{}{}
		case model.FormationAssignmentTypeRuntimeContext:
			rtmCtxsInFormation[assignment.Target] = struct{}{}
		}
	}

	allIDs := make([]string, 0, len(appsInFormation)+len(rtmsInFormation)+len(rtmCtxsInFormation))
	appIDs := make(map[string]bool, len(appsInFormation))
	rtIDs := make(map[string]bool, len(rtmsInFormation))
	rtCtxIDs := make(map[string]bool, len(rtmCtxsInFormation))
	for appID := range appsInFormation {
		allIDs = append(allIDs, appID)
		appIDs[appID] = false
	}
	for rtID := range rtmsInFormation {
		allIDs = append(allIDs, rtID)
		rtIDs[rtID] = false
	}
	for rtCtxID := range rtmCtxsInFormation {
		allIDs = append(allIDs, rtCtxID)
		rtCtxIDs[rtCtxID] = false
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
	assignments := make([]*model.FormationAssignmentInput, 0, (len(allIDs))*2+1)
	for appID, isAssigned := range appIDs {
		if !isAssigned || appID == objectID {
			continue
		}
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(objectID, objectType, formation, model.FormationAssignmentTypeApplication, appID, initialConfigurations)...)
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
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(objectID, objectType, formation, model.FormationAssignmentTypeRuntime, runtimeID, initialConfigurations)...)
	}

	for runtimeCtxID, isAssigned := range rtCtxIDs {
		if !isAssigned || runtimeCtxID == objectID {
			continue
		}
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(objectID, objectType, formation, model.FormationAssignmentTypeRuntimeContext, runtimeCtxID, initialConfigurations)...)
	}

	assignments = append(assignments, &model.FormationAssignmentInput{
		FormationID: formation.ID,
		Source:      objectID,
		SourceType:  model.FormationAssignmentType(objectType),
		Target:      objectID,
		TargetType:  model.FormationAssignmentType(objectType),
		State:       string(model.InitialFormationState),
		Value:       getInitialConfiguration(objectID, objectID, initialConfigurations),
		Error:       nil,
	})

	return assignments, nil
}

// PersistAssignments persists the provided formation assignments
func (s *service) PersistAssignments(ctx context.Context, tnt string, assignments []*model.FormationAssignmentInput) ([]*model.FormationAssignment, error) {
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
func (s *service) GenerateAssignmentsForParticipant(objectID string, objectType graphql.FormationObjectType, formation *model.Formation, participantType model.FormationAssignmentType, participantID string, initialConfigurations model.InitialConfigurations) []*model.FormationAssignmentInput {
	assignments := make([]*model.FormationAssignmentInput, 0, 2)
	assignments = append(assignments, &model.FormationAssignmentInput{
		FormationID: formation.ID,
		Source:      objectID,
		SourceType:  model.FormationAssignmentType(objectType),
		Target:      participantID,
		TargetType:  participantType,
		State:       string(model.InitialAssignmentState),
		Value:       getInitialConfiguration(objectID, participantID, initialConfigurations),
		Error:       nil,
	})
	assignments = append(assignments, &model.FormationAssignmentInput{
		FormationID: formation.ID,
		Source:      participantID,
		SourceType:  participantType,
		Target:      objectID,
		TargetType:  model.FormationAssignmentType(objectType),
		State:       string(model.InitialAssignmentState),
		Value:       getInitialConfiguration(participantID, objectID, initialConfigurations),
		Error:       nil,
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
func (s *service) ProcessFormationAssignments(ctx context.Context, assignmentRequestMappings []*notifications.AssignmentMappingPair, formationAssignmentFunc func(context.Context, *notifications.AssignmentMappingPairWithOperation) (bool, error), formationOperation model.FormationOperation) error {
	var errs *multierror.Error
	alreadyProcessedFAs := make(map[string]bool, 0)
	for _, mapping := range assignmentRequestMappings {
		if alreadyProcessedFAs[mapping.AssignmentReqMapping.FormationAssignment.ID] {
			continue
		}
		mappingWithOperation := &notifications.AssignmentMappingPairWithOperation{
			AssignmentMappingPair: mapping,
			Operation:             formationOperation,
		}
		isReverseProcessed, err := formationAssignmentFunc(ctx, mappingWithOperation)
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "while processing formation assignment with id %q", mapping.AssignmentReqMapping.FormationAssignment.ID))
		}
		if isReverseProcessed {
			alreadyProcessedFAs[mapping.ReverseAssignmentReqMapping.FormationAssignment.ID] = true
		}
	}
	log.C(ctx).Infof("Finished processing %d formation assignments", len(assignmentRequestMappings))

	return errs.ErrorOrNil()
}

// ProcessFormationAssignmentPair prepares and update the `State` and `Config` of the formation assignment based on the response and process the notifications
func (s *service) ProcessFormationAssignmentPair(ctx context.Context, mappingPair *notifications.AssignmentMappingPairWithOperation) (bool, error) {
	var isReverseProcessed bool
	err := s.processFormationAssignmentsWithReverseNotification(ctx, mappingPair, 0, &isReverseProcessed)
	return isReverseProcessed, err
}

func (s *service) processFormationAssignmentsWithReverseNotification(ctx context.Context, mappingPair *notifications.AssignmentMappingPairWithOperation, depth int, isReverseProcessed *bool) error {
	assignmentReqMappingClone := mappingPair.AssignmentReqMapping.Clone()
	var reverseAssignmentReqMappingClone *notifications.FormationAssignmentRequestMapping
	if mappingPair.ReverseAssignmentReqMapping != nil {
		reverseAssignmentReqMappingClone = mappingPair.ReverseAssignmentReqMapping.Clone()
	}
	assignment := assignmentReqMappingClone.FormationAssignment

	if assignment == nil {
		return errors.New("formation assignment is nil")
	}

	logger := log.C(ctx).WithField(log.FieldFormationAssignmentID, assignment.ID)
	ctx = log.ContextWithLogger(ctx, logger)
	log.C(ctx).Infof("Processing formation assignment with ID: %q for formation with ID: %q with Source: %q of Type: %q and Target: %q of Type: %q and State %q", assignment.ID, assignment.FormationID, assignment.Source, assignment.SourceType, assignment.Target, assignment.TargetType, assignment.State)

	if assignment.State == string(model.ReadyAssignmentState) {
		log.C(ctx).Infof("The formation assignment with ID: %q is in %q state. No notifications will be sent for it.", assignment.ID, assignment.State)
		if err := s.assignmentOperationService.Finish(ctx, assignment.ID, assignment.FormationID); err != nil {
			return errors.Wrapf(err, "while finishing %s Operation for assignment with ID: %s", model.FromFormationOperationType(mappingPair.Operation), assignment.ID)
		}
		return nil
	}

	if assignmentReqMappingClone.Request == nil {
		assignment.State = string(model.ReadyAssignmentState)
		if mappingPair.Operation == model.AssignFormation {
			// In case of error in the assignment we want to clear it
			assignment.Error = nil
		}
		log.C(ctx).Infof("In the formation assignment mapping pair, assignment with ID: %q hasn't attached webhook request. Updating the formation assignment to %q state without sending notification", assignment.ID, assignment.State)
		if err := s.Update(ctx, assignment.ID, assignment); err != nil {
			return errors.Wrapf(err, "while updating formation assignment for formation with ID: %q with source: %q and target: %q", assignment.FormationID, assignment.Source, assignment.Target)
		}

		if err := s.assignmentOperationService.Finish(ctx, assignment.ID, assignment.FormationID); err != nil {
			return errors.Wrapf(err, "while finishing %s Operation for assignment with ID: %s that has no notifications", model.Assign, assignment.ID)
		}

		return nil
	}

	extendedRequest, err := s.faNotificationService.GenerateFormationAssignmentNotificationExt(ctx, assignmentReqMappingClone, reverseAssignmentReqMappingClone, mappingPair.Operation)
	if err != nil {
		return errors.Wrap(err, "while creating extended formation assignment request")
	}

	response, err := s.notificationService.SendNotification(ctx, extendedRequest)
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

	var webhookMode graphql.WebhookMode
	if assignmentReqMappingClone.Request.Webhook != nil {
		requestWebhookMode := assignmentReqMappingClone.Request.Webhook.Mode
		if requestWebhookMode != nil {
			webhookMode = *requestWebhookMode
		}
	}

	if err = validateNotificationResponse(response, assignment, mappingPair.Operation, webhookMode); err != nil {
		updateError := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), ClientError, model.CreateErrorAssignmentState)
		if updateError != nil {
			return errors.Wrapf(
				updateError,
				"while updating error state: %s",
				errors.Wrapf(err, "while validating notification response for formation assignment with ID %q", assignment.ID).Error())
		}
		return errors.Wrapf(err, "The provided response is not valid: ")
	}

	notificationStatusReport := newNotificationStatusReportFromWebhookResponse(response, mappingPair.Operation, webhookMode)

	// We are skipping further notification processing in case the webhook has ASYNC CALLBACK mode and the client accepted the notification as we are
	// waiting for async response and will keep the FA state as is. In case of error we want to update the FA with the error.
	if webhookMode == graphql.WebhookModeAsyncCallback && !isErrorState(model.FormationAssignmentState(notificationStatusReport.State)) {
		log.C(ctx).Infof("The webhook with ID: %q in the notification is in %q mode. Waiting for the receiver to report the status on the status API...", assignmentReqMappingClone.Request.Webhook.ID, graphql.WebhookModeAsyncCallback)
		return nil
	}

	if err = s.statusService.UpdateWithConstraints(ctx, notificationStatusReport, assignment, mappingPair.Operation); err != nil {
		return errors.Wrapf(err, "while updating formation assignment with constraints for formation %q with source %q and target %q", assignment.FormationID, assignment.Source, assignment.Target)
	}

	if assignment.State == string(model.ReadyAssignmentState) {
		if err = s.assignmentOperationService.Finish(ctx, assignment.ID, assignment.FormationID); err != nil {
			return errors.Wrapf(err, "while finishing %s Operation for assignment with ID: %s during SYNC processing", model.FromFormationOperationType(mappingPair.Operation), assignment.ID)
		}
	}

	// In case of an error state we should not proceed with processing the reverse notification
	if isErrorState(model.FormationAssignmentState(notificationStatusReport.State)) {
		log.C(ctx).Error(errors.Errorf("Received error from response: %v", notificationStatusReport.Error).Error())
		return nil
	}

	configFromReport := notificationStatusReport.Configuration
	if assignment.Source != assignment.Target && configFromReport != nil && !formationconstraintpkg.IsConfigEmpty(string(configFromReport)) {
		if reverseAssignmentReqMappingClone == nil {
			return nil
		}

		*isReverseProcessed = true

		if depth >= model.NotificationRecursionDepthLimit {
			log.C(ctx).Errorf("Depth limit exceeded for assignments: %q and %q", assignmentReqMappingClone.FormationAssignment.ID, reverseAssignmentReqMappingClone.FormationAssignment.ID)
			return nil
		}

		newAssignmentReqMapping := reverseAssignmentReqMappingClone.Clone()
		newReverseAssignmentReqMapping := assignmentReqMappingClone.Clone()

		if newAssignmentReqMapping.Request != nil {
			newAssignmentReqMapping.Request.Object.SetAssignment(newAssignmentReqMapping.FormationAssignment)
			newAssignmentReqMapping.Request.Object.SetReverseAssignment(newReverseAssignmentReqMapping.FormationAssignment)
		}
		if newReverseAssignmentReqMapping.Request != nil {
			newReverseAssignmentReqMapping.Request.Object.SetAssignment(newReverseAssignmentReqMapping.FormationAssignment)
			newReverseAssignmentReqMapping.Request.Object.SetReverseAssignment(newAssignmentReqMapping.FormationAssignment)
		}

		newAssignmentMappingPair := &notifications.AssignmentMappingPairWithOperation{
			AssignmentMappingPair: &notifications.AssignmentMappingPair{
				AssignmentReqMapping:        newAssignmentReqMapping,
				ReverseAssignmentReqMapping: newReverseAssignmentReqMapping,
			},
			Operation: mappingPair.Operation,
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
func (s *service) CleanupFormationAssignment(ctx context.Context, mappingPair *notifications.AssignmentMappingPairWithOperation) (bool, error) {
	assignment := mappingPair.AssignmentReqMapping.FormationAssignment
	logger := log.C(ctx).WithField(log.FieldFormationAssignmentID, assignment.ID)
	ctx = log.ContextWithLogger(ctx, logger)

	if mappingPair.AssignmentReqMapping.Request == nil {
		if err := s.Delete(ctx, assignment.ID); err != nil {
			if apperrors.IsNotFoundError(err) {
				log.C(ctx).Infof("Assignment with ID %q has already been deleted", assignment.ID)
				return false, nil
			}

			// It is possible that the deletion fails due to some kind of DB constraint, so we will try to update the state
			if updateError := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), TechnicalError, model.DeleteErrorAssignmentState); updateError != nil {
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

	extendedRequest, err := s.faNotificationService.GenerateFormationAssignmentNotificationExt(ctx, mappingPair.AssignmentReqMapping, mappingPair.ReverseAssignmentReqMapping, mappingPair.Operation)
	if err != nil {
		if updateError := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), TechnicalError, model.DeleteErrorAssignmentState); updateError != nil {
			return false, errors.Wrapf(
				updateError,
				"while updating error state: %s",
				errors.Wrapf(err, "while generating notifications for formation assignment with ID: %q", assignment.ID).Error())
		}
		return false, errors.Wrap(err, "while creating extended formation assignment request")
	}

	response, err := s.notificationService.SendNotification(ctx, extendedRequest)
	if err != nil {
		if updateError := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), TechnicalError, model.DeleteErrorAssignmentState); updateError != nil {
			return false, errors.Wrapf(
				updateError,
				"while updating error state: %s",
				errors.Wrapf(err, "while sending notification for formation assignment with ID %q", assignment.ID).Error())
		}
		return false, errors.Wrapf(err, "while sending notification for formation assignment with ID %q", assignment.ID)
	}

	var webhookMode graphql.WebhookMode
	if mappingPair.AssignmentReqMapping.Request.Webhook != nil {
		requestWebhookMode := mappingPair.AssignmentReqMapping.Request.Webhook.Mode
		if requestWebhookMode != nil {
			webhookMode = *requestWebhookMode
		}
	}

	if err = validateNotificationResponse(response, assignment, mappingPair.Operation, webhookMode); err != nil {
		if updateError := s.SetAssignmentToErrorState(ctx, assignment, err.Error(), ClientError, model.DeleteErrorAssignmentState); updateError != nil {
			return false, errors.Wrapf(
				updateError,
				"while updating error state: %s",
				errors.Wrapf(err, "while validating notification response for formation assignment with ID %q", assignment.ID).Error())
		}
		return false, errors.Wrapf(err, "The provided response is not valid: ")
	}

	notificationStatusReport := newNotificationStatusReportFromWebhookResponse(response, mappingPair.Operation, webhookMode)
	stateFromReport := notificationStatusReport.State

	if isErrorState(model.FormationAssignmentState(stateFromReport)) {
		err = s.statusService.UpdateWithConstraints(ctx, notificationStatusReport, assignment, mappingPair.Operation)
		if err != nil && apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("Assignment with ID %q has already been deleted", assignment.ID)
			return false, nil
		} else if err != nil {
			return false, errors.Wrapf(err, "while updating error state for formation assignment with ID %q", assignment.ID)
		}

		return false, errors.Errorf("Received %s assignment state and error: %v", stateFromReport, notificationStatusReport.Error)
	}

	// We are skipping further notification processing in case the webhook has ASYNC CALLBACK mode and the client accepted the notification as we are
	// waiting for async response and will keep the FA state as is.
	if webhookMode == graphql.WebhookModeAsyncCallback {
		log.C(ctx).Infof("The webhook with ID: %q in the notification is in %q mode. Waiting for the receiver to report the status on the status API...", mappingPair.AssignmentReqMapping.Request.Webhook.ID, graphql.WebhookModeAsyncCallback)
		return false, nil
	}

	if stateFromReport == string(model.ReadyAssignmentState) {
		if err = s.statusService.DeleteWithConstraints(ctx, assignment.ID, notificationStatusReport); err != nil {
			if apperrors.IsNotFoundError(err) {
				log.C(ctx).Infof("Assignment with ID %q has already been deleted", assignment.ID)
				return false, nil
			}
			// It is possible that the deletion fails due to some kind of DB constraint, so we will try to update the state
			if updateError := s.SetAssignmentToErrorState(ctx, assignment, "error while deleting assignment", TechnicalError, model.DeleteErrorAssignmentState); updateError != nil {
				if apperrors.IsNotFoundError(updateError) {
					log.C(ctx).Infof("Assignment with ID %q has already been deleted", assignment.ID)
					return false, nil
				}
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

	return false, nil
}

func validateResponseState(newState, previousState string) bool {
	if !model.SupportedFormationAssignmentStates[newState] {
		return false
	}

	// handles synchronous "delete/unassign" statuses
	if previousState == string(model.DeletingAssignmentState) &&
		(newState != string(model.DeleteErrorAssignmentState) && newState != string(model.ReadyAssignmentState) && newState != string(model.DeleteReadyFormationAssignmentState)) {
		return false
	}

	// handles synchronous "create/assign" statuses
	if previousState == string(model.InitialAssignmentState) &&
		(newState != string(model.CreateErrorAssignmentState) && newState != string(model.ConfigPendingAssignmentState) && newState != string(model.ReadyAssignmentState) && newState != string(model.CreateReadyFormationAssignmentState)) {
		return false
	}

	if previousState == string(model.DeleteErrorAssignmentState) &&
		(newState != string(model.DeleteErrorAssignmentState) && newState != string(model.ReadyAssignmentState) && newState != string(model.DeletingAssignmentState) && newState != string(model.DeleteReadyFormationAssignmentState)) {
		return false
	}

	return true
}

func (s *service) SetAssignmentToErrorState(ctx context.Context, assignment *model.FormationAssignment, errorMessage string, errorCode AssignmentErrorCode, state model.FormationAssignmentState) error {
	assignment.State = string(state)
	assignmentError := AssignmentErrorWrapper{AssignmentError{
		Message:   errorMessage,
		ErrorCode: errorCode,
	}}
	marshaled, err := json.Marshal(assignmentError)
	if err != nil {
		return errors.Wrapf(err, "while preparing error message for assignment with ID: %q", assignment.ID)
	}
	assignment.Error = marshaled
	if err := s.Update(ctx, assignment.ID, assignment); err != nil {
		return errors.Wrapf(err, "while updating formation assignment with ID: %s", assignment.ID)
	}
	log.C(ctx).Infof("Assignment with ID: %s set to state: %s", assignment.ID, assignment.State)
	return nil
}

// ResetAssignmentConfigAndError sets the configuration and the error fields of the formation assignment to nil
func ResetAssignmentConfigAndError(assignment *model.FormationAssignment) {
	assignment.Value = nil
	assignment.Error = nil
}

// AssignmentErrorCode represents error code used to differentiate the source of the error
type AssignmentErrorCode int

const (
	// TechnicalError indicates that the reason for the error is technical - for example networking issue
	TechnicalError = 1
	// ClientError indicates that the error was returned from the client
	ClientError = 2
)

// AssignmentError error struct used for storing the errors that occur during the FormationAssignment processing
type AssignmentError struct {
	Message   string              `json:"message"`
	ErrorCode AssignmentErrorCode `json:"errorCode"`
}

// AssignmentErrorWrapper wrapper for AssignmentError
type AssignmentErrorWrapper struct {
	Error AssignmentError `json:"error"`
}

func newNotificationStatusReportFromWebhookResponse(response *webhookdir.Response, operation model.FormationOperation, webhookMode graphql.WebhookMode) *statusreport.NotificationStatusReport {
	var respConfig json.RawMessage
	if response.Config != nil {
		respConfig = []byte(*response.Config)
	}

	var respError string
	if response.Error != nil && *response.Error != "" {
		respError = *response.Error
	}

	return statusreport.NewNotificationStatusReport(respConfig, calculateStateFromWebhookResponse(response, operation, webhookMode), respError)
}

func calculateStateFromWebhookResponse(response *webhookdir.Response, operation model.FormationOperation, webhookMode graphql.WebhookMode) string {
	if response.State != nil && *response.State != "" && *response.State != string(model.CreateReadyFormationAssignmentState) && *response.State != string(model.DeleteReadyFormationAssignmentState) {
		return *response.State
	}

	if response.State != nil && *response.State != "" && (*response.State == string(model.CreateReadyFormationAssignmentState) || *response.State == string(model.DeleteReadyFormationAssignmentState)) {
		return string(model.ReadyAssignmentState)
	}

	if response.Error != nil && *response.Error != "" {
		if operation == model.AssignFormation {
			return string(model.CreateErrorAssignmentState)
		}

		return string(model.DeleteErrorAssignmentState)
	}

	if *response.ActualStatusCode == *response.SuccessStatusCode && webhookMode != graphql.WebhookModeAsyncCallback {
		return string(model.ReadyAssignmentState)
	}

	if operation == model.AssignFormation && webhookMode == graphql.WebhookModeAsyncCallback {
		return string(model.InitialAssignmentState)
	}

	if operation == model.AssignFormation && webhookMode == graphql.WebhookModeSync {
		return string(model.ConfigPendingAssignmentState)
	}
	return string(model.DeletingAssignmentState)
}

func validateNotificationResponse(response *webhookdir.Response, assignment *model.FormationAssignment, operation model.FormationOperation, webhookMode graphql.WebhookMode) error {
	var actualCode int
	if response.ActualStatusCode != nil {
		actualCode = *response.ActualStatusCode
	}
	var incompleteCode int
	if response.IncompleteStatusCode != nil {
		incompleteCode = *response.IncompleteStatusCode
	}
	var successCode int
	if response.SuccessStatusCode != nil {
		successCode = *response.SuccessStatusCode
	}

	var fieldRules []*validation.FieldRules
	fieldRules = append(
		fieldRules,
		validation.Field(&response.State, validation.When(response.State != nil && *response.State != "",
			validation.When(isErrorNotEmpty(response.Error) && operation == model.AssignFormation, validation.In(string(model.CreateErrorAssignmentState))),
			validation.When(isErrorNotEmpty(response.Error) && operation == model.UnassignFormation, validation.In(string(model.DeleteErrorAssignmentState))),
			validation.When(isConfigNotEmpty(response.Config), validation.In(string(model.ReadyAssignmentState), string(model.CreateReadyFormationAssignmentState), string(model.ConfigPendingAssignmentState))),
			validation.When(actualCode == incompleteCode, validation.In(string(model.ConfigPendingAssignmentState))),
			validation.When(actualCode != incompleteCode && actualCode != successCode, validation.In(string(model.DeleteErrorAssignmentState), string(model.CreateErrorAssignmentState))),
			// in case of empty error and configuration
			validation.In(string(model.ReadyAssignmentState), string(model.CreateReadyFormationAssignmentState), string(model.DeleteReadyFormationAssignmentState), string(model.CreateErrorAssignmentState), string(model.DeleteErrorAssignmentState), string(model.ConfigPendingAssignmentState)),
		)),
		validation.Field(&response.Config, validation.When(isConfigNotEmpty(response.Config),
			validation.By(func(val interface{}) error {
				if isErrorNotEmpty(response.Error) {
					return errors.New("Configuration and Error can not be provided at the same time")
				}
				return nil
			}))),
		validation.Field(&response.Error, validation.When(isErrorNotEmpty(response.Error),
			validation.By(func(val interface{}) error {
				if isConfigNotEmpty(response.Config) {
					return errors.New("Configuration and Error can not be provided at the same time")
				}
				return nil
			}))),
	)

	if err := validation.ValidateStruct(response, fieldRules...); err != nil {
		return err
	}

	// it is possible for the state to be empty
	if operation == model.UnassignFormation && webhookMode == graphql.WebhookModeSync && (isConfigNotEmpty(response.Config) || response.ActualStatusCode == response.IncompleteStatusCode) {
		return errors.New("Config propagation is not supported on unassign notifications")
	}

	if response.State != nil && *response.State != "" {
		if isValid := validateResponseState(*response.State, assignment.State); !isValid {
			return errors.Errorf("Invalid transition from state %q to state %s.", assignment.State, *response.State)
		}
	}
	return nil
}

func isConfigNotEmpty(config *string) bool {
	return config != nil && !formationconstraintpkg.IsConfigEmpty(*config)
}

func isErrorNotEmpty(responseError *string) bool {
	return responseError != nil && *responseError != ""
}

func getInitialConfiguration(src, tgt string, initialConfigurations model.InitialConfigurations) json.RawMessage {
	if targetToCfg, ok := initialConfigurations[src]; ok {
		if config, ok := targetToCfg[tgt]; ok {
			return config
		}
	}
	return nil
}
