package runtime

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	labelPkg "github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// EventingService missing godoc
//go:generate mockery --name=EventingService --output=automock --outpkg=automock --case=underscore
type EventingService interface {
	GetForRuntime(ctx context.Context, runtimeID uuid.UUID) (*model.RuntimeEventingConfiguration, error)
}

// OAuth20Service missing godoc
//go:generate mockery --name=OAuth20Service --output=automock --outpkg=automock --case=underscore
type OAuth20Service interface {
	DeleteMultipleClientCredentials(ctx context.Context, auths []model.SystemAuth) error
}

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore
type RuntimeService interface {
	Create(ctx context.Context, in model.RuntimeInput) (string, error)
	Update(ctx context.Context, id string, in model.RuntimeInput) error
	Get(ctx context.Context, id string) (*model.Runtime, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error)
	ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error)
	DeleteLabel(ctx context.Context, runtimeID string, key string) error
}

// ScenarioAssignmentService missing godoc
//go:generate mockery --name=ScenarioAssignmentService --output=automock --outpkg=automock --case=underscore
type ScenarioAssignmentService interface {
	GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error)
	Delete(ctx context.Context, in model.AutomaticScenarioAssignment) error
}

// RuntimeConverter missing godoc
//go:generate mockery --name=RuntimeConverter --output=automock --outpkg=automock --case=underscore
type RuntimeConverter interface {
	ToGraphQL(in *model.Runtime) *graphql.Runtime
	MultipleToGraphQL(in []*model.Runtime) []*graphql.Runtime
	InputFromGraphQL(in graphql.RuntimeInput) model.RuntimeInput
}

// SystemAuthConverter missing godoc
//go:generate mockery --name=SystemAuthConverter --output=automock --outpkg=automock --case=underscore
type SystemAuthConverter interface {
	ToGraphQL(in *model.SystemAuth) (graphql.SystemAuth, error)
}

// SystemAuthService missing godoc
//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore
type SystemAuthService interface {
	ListForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
}

// BundleInstanceAuthService missing godoc
//go:generate mockery --name=BundleInstanceAuthService --output=automock --outpkg=automock --case=underscore
type BundleInstanceAuthService interface {
	ListByRuntimeID(ctx context.Context, runtimeID string) ([]*model.BundleInstanceAuth, error)
	Update(ctx context.Context, instanceAuth *model.BundleInstanceAuth) error
}

// BundleInstanceAuthService missing godoc
//go:generate mockery --name=SelfRegisterManager --output=automock --outpkg=automock --case=underscore
type SelfRegisterManager interface {
	PrepareRuntimeForSelfRegistration(ctx context.Context, in *graphql.RuntimeInput) error
	CleanupSelfRegisteredRuntime(ctx context.Context, selfRegisterLabelValue string) error
	GetSelfRegDistinguishingLabelKey() string
}

// Resolver missing godoc
type Resolver struct {
	transact                  persistence.Transactioner
	runtimeService            RuntimeService
	scenarioAssignmentService ScenarioAssignmentService
	sysAuthSvc                SystemAuthService
	converter                 RuntimeConverter
	sysAuthConv               SystemAuthConverter
	oAuth20Svc                OAuth20Service
	eventingSvc               EventingService
	bundleInstanceAuthSvc     BundleInstanceAuthService
	selfRegManager            SelfRegisterManager
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, runtimeService RuntimeService, scenarioAssignmentService ScenarioAssignmentService, sysAuthSvc SystemAuthService, oAuthSvc OAuth20Service, conv RuntimeConverter, sysAuthConv SystemAuthConverter, eventingSvc EventingService, bundleInstanceAuthSvc BundleInstanceAuthService, selfRegManager SelfRegisterManager) *Resolver {
	return &Resolver{
		transact:                  transact,
		runtimeService:            runtimeService,
		scenarioAssignmentService: scenarioAssignmentService,
		sysAuthSvc:                sysAuthSvc,
		oAuth20Svc:                oAuthSvc,
		converter:                 conv,
		sysAuthConv:               sysAuthConv,
		eventingSvc:               eventingSvc,
		bundleInstanceAuthSvc:     bundleInstanceAuthSvc,
		selfRegManager:            selfRegManager,
	}
}

// Runtimes missing godoc
// TODO: Proper error handling
func (r *Resolver) Runtimes(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimePage, error) {
	labelFilter := labelfilter.MultipleFromGraphQL(filter)
	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	runtimesPage, err := r.runtimeService.List(ctx, labelFilter, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntimes := r.converter.MultipleToGraphQL(runtimesPage.Data)

	return &graphql.RuntimePage{
		Data:       gqlRuntimes,
		TotalCount: runtimesPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(runtimesPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(runtimesPage.PageInfo.EndCursor),
			HasNextPage: runtimesPage.PageInfo.HasNextPage,
		},
	}, nil
}

// Runtime missing godoc
func (r *Resolver) Runtime(ctx context.Context, id string) (*graphql.Runtime, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	runtime, err := r.runtimeService.Get(ctx, id)
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

	return r.converter.ToGraphQL(runtime), nil
}

// RegisterRuntime missing godoc
func (r *Resolver) RegisterRuntime(ctx context.Context, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	convertedIn := r.converter.InputFromGraphQL(in)

	if err := r.selfRegManager.PrepareRuntimeForSelfRegistration(ctx, &in); err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		didRollback := r.transact.RollbackUnlessCommitted(ctx, tx)
		if didRollback {
			labelVal := str.CastOrEmpty(in.Labels[r.selfRegManager.GetSelfRegDistinguishingLabelKey()])
			r.cleanupAndLogOnError(ctx, labelVal)
		}
	}()

	ctx = persistence.SaveToContext(ctx, tx)

	id, err := r.runtimeService.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	runtime, err := r.runtimeService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntime := r.converter.ToGraphQL(runtime)

	return gqlRuntime, nil
}

// UpdateRuntime missing godoc
func (r *Resolver) UpdateRuntime(ctx context.Context, id string, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	convertedIn := r.converter.InputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.runtimeService.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	runtime, err := r.runtimeService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlRuntime := r.converter.ToGraphQL(runtime)

	return gqlRuntime, nil
}

// DeleteRuntime missing godoc
func (r *Resolver) DeleteRuntime(ctx context.Context, id string) (*graphql.Runtime, error) {
	var selfRegLabelVal string
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		didRollback := r.transact.RollbackUnlessCommitted(ctx, tx)
		if !didRollback { // if we did rollback we should not try to execute the cleanup
			r.cleanupAndLogOnError(ctx, selfRegLabelVal)
		}
	}()

	ctx = persistence.SaveToContext(ctx, tx)

	runtime, err := r.runtimeService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	bundleInstanceAuths, err := r.bundleInstanceAuthSvc.ListByRuntimeID(ctx, runtime.ID)
	if err != nil {
		return nil, err
	}

	selfRegLabel, err := r.runtimeService.GetLabel(ctx, runtime.ID, r.selfRegManager.GetSelfRegDistinguishingLabelKey())
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "while getting self register info label")
		}
		selfRegLabelVal = "" // the deferred cleanup will do nothing
	} else {
		selfRegLabelVal = str.CastOrEmpty(selfRegLabel.Value)
	}

	currentTimestamp := timestamp.DefaultGenerator
	for _, auth := range bundleInstanceAuths {
		if auth.Status.Condition != model.BundleInstanceAuthStatusConditionUnused {
			if err := auth.SetDefaultStatus(model.BundleInstanceAuthStatusConditionUnused, currentTimestamp()); err != nil {
				log.C(ctx).WithError(err).Errorf("while update bundle instance auth status condition: %v", err)
				return nil, err
			}
			if err := r.bundleInstanceAuthSvc.Update(ctx, auth); err != nil {
				log.C(ctx).WithError(err).Errorf("Unable to update bundle instance auth with ID: %s for corresponding bundle with ID: %s: %v", auth.ID, auth.BundleID, err)
				return nil, err
			}
		}
	}

	auths, err := r.sysAuthSvc.ListForObject(ctx, model.RuntimeReference, runtime.ID)
	if err != nil {
		return nil, err
	}

	if err = r.deleteAssociatedScenarioAssignments(ctx, runtime.ID); err != nil {
		return nil, err
	}

	deletedRuntime := r.converter.ToGraphQL(runtime)

	if err = r.runtimeService.Delete(ctx, id); err != nil {
		return nil, err
	}

	if err = r.oAuth20Svc.DeleteMultipleClientCredentials(ctx, auths); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return deletedRuntime, nil
}

// GetLabel missing godoc
func (r *Resolver) GetLabel(ctx context.Context, runtimeID string, key string) (*graphql.Labels, error) {
	if runtimeID == "" {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}
	if key == "" {
		return nil, apperrors.NewInternalError("Runtime label key cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	label, err := r.runtimeService.GetLabel(ctx, runtimeID, key)
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

	resultLabels := make(map[string]interface{})
	resultLabels[key] = label.Value

	var gqlLabels graphql.Labels = resultLabels
	return &gqlLabels, nil
}

// SetRuntimeLabel missing godoc
func (r *Resolver) SetRuntimeLabel(ctx context.Context, runtimeID string, key string, value interface{}) (*graphql.Label, error) {
	// TODO: Use @validation directive on input type instead, after resolving https://github.com/kyma-incubator/compass/issues/515
	gqlLabel := graphql.LabelInput{Key: key, Value: value}
	if err := inputvalidation.Validate(&gqlLabel); err != nil {
		return nil, errors.Wrap(err, "validation error for type LabelInput")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.runtimeService.SetLabel(ctx, &model.LabelInput{
		Key:        key,
		Value:      value,
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   runtimeID,
	})
	if err != nil {
		return nil, err
	}

	label, err := r.runtimeService.GetLabel(ctx, runtimeID, key)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting label with key: [%s]", key)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   label.Key,
		Value: label.Value,
	}, nil
}

// DeleteRuntimeLabel missing godoc
func (r *Resolver) DeleteRuntimeLabel(ctx context.Context, runtimeID string, key string) (*graphql.Label, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	label, err := r.runtimeService.GetLabel(ctx, runtimeID, key)
	if err != nil {
		return nil, err
	}

	err = r.runtimeService.DeleteLabel(ctx, runtimeID, key)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   key,
		Value: label.Value,
	}, nil
}

// Labels missing godoc
func (r *Resolver) Labels(ctx context.Context, obj *graphql.Runtime, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.runtimeService.ListLabels(ctx, obj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") { // TODO: Use custom error and check its type
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})

	for _, label := range itemMap {
		if key == nil || label.Key == *key {
			resultLabels[label.Key] = label.Value
		}
	}

	var gqlLabels graphql.Labels = resultLabels
	return gqlLabels, nil
}

// Auths missing godoc
func (r *Resolver) Auths(ctx context.Context, obj *graphql.Runtime) ([]*graphql.RuntimeSystemAuth, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	sysAuths, err := r.sysAuthSvc.ListForObject(ctx, model.RuntimeReference, obj.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	out := make([]*graphql.RuntimeSystemAuth, 0, len(sysAuths))
	for _, sa := range sysAuths {
		c, err := r.sysAuthConv.ToGraphQL(&sa)
		if err != nil {
			return nil, err
		}
		out = append(out, c.(*graphql.RuntimeSystemAuth))
	}

	return out, nil
}

// EventingConfiguration missing godoc
func (r *Resolver) EventingConfiguration(ctx context.Context, obj *graphql.Runtime) (*graphql.RuntimeEventingConfiguration, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}

	runtimeID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening the transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventingCfg, err := r.eventingSvc.GetForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing configuration for runtime")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing the transaction")
	}

	return eventing.RuntimeEventingConfigurationToGraphQL(eventingCfg), nil
}

// deleteAssociatedScenarioAssignments ensures that scenario assignments which are responsible for creation of certain runtime labels are deleted,
// if runtime doesn't have the scenarios label or is part of a scenario for which no scenario assignment exists => noop
func (r *Resolver) deleteAssociatedScenarioAssignments(ctx context.Context, runtimeID string) error {
	scenariosLbl, err := r.runtimeService.GetLabel(ctx, runtimeID, model.ScenariosKey)
	notFound := apperrors.IsNotFoundError(err)
	if err != nil && !notFound {
		return err
	}

	if notFound {
		return nil
	}

	scenarios, err := labelPkg.ValueToStringsSlice(scenariosLbl.Value)
	if err != nil {
		return err
	}

	for _, scenario := range scenarios {
		scenarioAssignment, err := r.scenarioAssignmentService.GetForScenarioName(ctx, scenario)
		notFound := apperrors.IsNotFoundError(err)
		if err != nil && !notFound {
			return err
		}

		if notFound {
			continue
		}

		if err := r.scenarioAssignmentService.Delete(ctx, scenarioAssignment); err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) cleanupAndLogOnError(ctx context.Context, labelVal string) {
	if err := r.selfRegManager.CleanupSelfRegisteredRuntime(ctx, labelVal); err != nil {
		log.C(ctx).Errorf("An error occured during cleanup of self-registered runtime: %v", err)
	}
}
