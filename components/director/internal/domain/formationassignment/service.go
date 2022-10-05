package formationassignment

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
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
//go:generate mockery --name=FormationAssignmentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentRepository interface {
	Create(ctx context.Context, item *model.FormationAssignment) error
	Get(ctx context.Context, id, tenantID string) (*model.FormationAssignment, error)
	GetForFormation(ctx context.Context, tenantID, id, formationID string) (*model.FormationAssignment, error)
	List(ctx context.Context, pageSize int, cursor, tenantID string) (*model.FormationAssignmentPage, error)
	ListByFormationIDs(ctx context.Context, tenantID string, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error)
	ListAllForObject(ctx context.Context, tenant, formationID, objectID string) ([]*model.FormationAssignment, error)
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
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
	ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Application, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	GetByRuntimeID(ctx context.Context, tenant, runtimeID string) (*model.RuntimeContext, error)
	ListByScenariosAndRuntimeIDs(ctx context.Context, tenant string, scenarios []string, runtimeIDs []string) ([]*model.RuntimeContext, error)
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.RuntimeContext, error)
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
	ExistsByRuntimeID(ctx context.Context, tenant, rtmID string) (bool, error)
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	GetByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	ListAll(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	ListOwnedRuntimes(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Runtime, error)
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.Runtime, error)
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.Runtime, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	OwnerExistsByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (bool, error)
}

// UIDService generates UUIDs for new entities
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	transact                     persistence.Transactioner
	repo                         FormationAssignmentRepository
	uidSvc                       UIDService
	applicationRepository        applicationRepository
	runtimeRepo                  runtimeRepository
	runtimeContextRepo           runtimeContextRepository
	formationAssignmentConverter formationAssignmentConverter
}

// NewService creates a FormationTemplate service
func NewService(transact persistence.Transactioner, repo FormationAssignmentRepository, uidSvc UIDService, applicationRepository applicationRepository, runtimeRepository runtimeRepository, runtimeContextRepo runtimeContextRepository, formationAssignmentConverter formationAssignmentConverter) *service {
	return &service{
		transact:                     transact,
		repo:                         repo,
		uidSvc:                       uidSvc,
		applicationRepository:        applicationRepository,
		runtimeRepo:                  runtimeRepository,
		runtimeContextRepo:           runtimeContextRepo,
		formationAssignmentConverter: formationAssignmentConverter,
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

// ListByFormationIDs retrieves a page of Formation Assignment objects for each Formation
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

// ListFormationAssignmentsForObject retrieves all Formation Assignment objects for formation with ID `formationID` that have `objectID` as source or target
func (s *service) ListFormationAssignmentsForObject(ctx context.Context, formationID, objectID string) ([]*model.FormationAssignment, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.ListAllForObject(ctx, tnt, formationID, objectID)
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

	assignments := make([]*model.FormationAssignment, 0, (len(applications)+len(runtimes)+len(runtimeContexts))*2)
	for _, app := range applications {
		if app.ID == objectID {
			continue
		}
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(tnt, objectID, objectType, formation, graphql.FormationObjectTypeApplication, app)...)
	}
	for _, runtime := range runtimes {
		if runtime.ID == objectID {
			continue
		}
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(tnt, objectID, objectType, formation, graphql.FormationObjectTypeRuntime, runtime)...)
	}
	for _, runtimeCtx := range runtimeContexts {
		if runtimeCtx.ID == objectID {
			continue
		}
		assignments = append(assignments, s.GenerateAssignmentsForParticipant(tnt, objectID, objectType, formation, graphql.FormationObjectTypeRuntimeContext, runtimeCtx)...)
	}
	return assignments, nil
}

func (s *service) GenerateAssignmentsForParticipant(tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation, participantType graphql.FormationObjectType, participant model.Identifiable) []*model.FormationAssignment {
	assignments := make([]*model.FormationAssignment, 0, 2)
	assignments = append(assignments, &model.FormationAssignment{
		FormationID: formation.ID,
		TenantID:    tnt,
		Source:      objectID,
		SourceType:  string(objectType),
		Target:      participant.GetID(),
		TargetType:  string(participantType),
		State:       string(model.ReadyAssignmentState),
		Value:       nil,
	})
	assignments = append(assignments, &model.FormationAssignment{
		FormationID: formation.ID,
		TenantID:    tnt,
		Source:      participant.GetID(),
		SourceType:  string(participantType),
		Target:      objectID,
		TargetType:  string(objectType),
		State:       string(model.ReadyAssignmentState),
		Value:       nil,
	})
	return assignments
}
func (s *service) ProcessFormationAssignments(ctx context.Context, tenant string, formationAssignmentsForObject []*model.FormationAssignment, requests []*webhookclient.Request, responses []*webhookdir.Response, operation func(context.Context, *model.FormationAssignment, *webhookdir.Response) error) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	fmt.Println("HOHOHO")

	assignmentRequestMappings, err := s.matchFormationAssignmentsWithRequests(ctx, tenant, formationAssignmentsForObject, requests, responses)
	if err != nil {
		return errors.Wrap(err, "While mapping formationAssignments to notification requests and responses")
	}
	fmt.Println("SECOND TIME")
	for _, mapping := range assignmentRequestMappings {
		if err := operation(ctx, mapping.FormationAssignment, mapping.Response); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

//TODO maybe add update
func (s *service) CreateOrUpdateFormationAssignment(ctx context.Context, assignment *model.FormationAssignment, response *webhookdir.Response) error {
	if response == nil || *response.ActualStatusCode == *response.SuccessStatusCode {
		assignment.State = string(model.ReadyAssignmentState)
	}
	if response != nil && response.IncompleteStatusCode != nil && *response.ActualStatusCode == *response.IncompleteStatusCode {
		assignment.State = string(model.ConfigPendingAssignmentState)
	}

	fmt.Println("<<<<<<<<<<<<<")
	fmt.Println("RESPONSE: ")
	spew.Dump(response)
	if response != nil && response.Config != nil && *response.Config != "" {
		assignment.Value = []byte(*response.Config)
	}
	if response != nil && response.Error != nil && *response.Error != "" {
		assignment.State = string(model.CreateErrorAssignmentState)
		marshaled, err := json.Marshal(struct{ Error string }{Error: *response.Error})
		if err != nil {
			return errors.Wrapf(err, "While preparing error message from response for assignment with ID %q", assignment.ID)
		}
		assignment.Value = marshaled
	}
	spew.Dump(assignment)
	if _, err := s.Create(ctx, s.formationAssignmentConverter.ToInput(assignment)); err != nil {
		return errors.Wrapf(err, "while creating formation assignment for formation %q with source %q and target %q", assignment.FormationID, assignment.Source, assignment.Target)
	}

	return nil
}

func (s *service) CleanupFormationAssignment(ctx context.Context, assignment *model.FormationAssignment, response *webhookdir.Response) error {
	spew.Dump(response)
	//todo add logs

	if response == nil || *response.ActualStatusCode == *response.SuccessStatusCode {
		if err := s.Delete(ctx, assignment.ID); err != nil {
			return errors.Wrapf(err, "While deleting formation assignment with id %q", assignment.ID)
		}

		return nil
	}

	if response.IncompleteStatusCode != nil && *response.ActualStatusCode == *response.IncompleteStatusCode {
		assignment.State = string(model.DeleteErrorAssignmentState)
		marshaled, err := json.Marshal(struct{ Error string }{Error: "Error while deleting assignment: config propagation is not supported on unassign notifications"})
		if err != nil {
			return errors.Wrapf(err, "While preparing error message for assignment with ID %q", assignment.ID)
		}
		assignment.Value = marshaled
		if err := s.Update(ctx, assignment.ID, s.formationAssignmentConverter.ToInput(assignment)); err != nil {
			return errors.Wrapf(err, "While updating formation assignment with id %q", assignment.ID)
		}

		return nil
	}

	if response.Error != nil && *response.Error != "" {
		assignment.State = string(model.DeleteErrorAssignmentState)
		marshaled, err := json.Marshal(struct{ Error string }{Error: *response.Error})
		if err != nil {
			return errors.Wrapf(err, "While preparing error message from response for assignment with ID %q", assignment.ID)
		}
		assignment.Value = marshaled
		if err := s.Update(ctx, assignment.ID, s.formationAssignmentConverter.ToInput(assignment)); err != nil {
			return errors.Wrapf(err, "While updating formation assignment with id %q", assignment.ID)
		}

		return nil
	}

	return nil
}

type FormationAssignmentRequestMapping struct {
	Request             *webhookclient.Request
	Response            *webhookdir.Response
	FormationAssignment *model.FormationAssignment
}

func (s *service) matchFormationAssignmentsWithRequests(ctx context.Context, tenant string, assignments []*model.FormationAssignment, requests []*webhookclient.Request, responses []*webhookdir.Response) ([]*FormationAssignmentRequestMapping, error) {
	formationAssignmentMapping := make([]*FormationAssignmentRequestMapping, 0, len(assignments))
	for i, assignment := range assignments {
		mappingObject := &FormationAssignmentRequestMapping{
			Request:             nil,
			Response:            nil,
			FormationAssignment: assignments[i],
		}
		target := assignment.Target
		if assignment.TargetType == string(graphql.FormationObjectTypeRuntimeContext) {
			rtmCtx, err := s.runtimeContextRepo.GetByID(ctx, tenant, target)
			if err != nil {
				return nil, err
			}

			target = rtmCtx.RuntimeID
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

			participants := request.Object.GetParticipants()
			for _, id := range participants {
				if assignment.Source == id {
					mappingObject.Request = requests[j]
					mappingObject.Response = responses[j]
					break assignment
				}
			}
		}
		formationAssignmentMapping = append(formationAssignmentMapping, mappingObject)
	}
	return formationAssignmentMapping, nil
}
