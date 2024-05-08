package formationtemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/pkg/errors"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// FormationTemplateConverter converts between the graphql and model
//
//go:generate mockery --name=FormationTemplateConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateConverter interface {
	FromRegisterInputGraphQL(in *graphql.FormationTemplateRegisterInput) (*model.FormationTemplateRegisterInput, error)
	FromUpdateInputGraphQL(in *graphql.FormationTemplateUpdateInput) (*model.FormationTemplateUpdateInput, error)
	ToGraphQL(in *model.FormationTemplate) (*graphql.FormationTemplate, error)
	MultipleToGraphQL(in []*model.FormationTemplate) ([]*graphql.FormationTemplate, error)
	FromModelRegisterInputToModel(in *model.FormationTemplateRegisterInput, id string, tenantID string) *model.FormationTemplate
	FromModelUpdateInputToModel(in *model.FormationTemplateUpdateInput, id string, tenantID string) *model.FormationTemplate
}

// FormationTemplateService represents the FormationTemplate service layer
//
//go:generate mockery --name=FormationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateService interface {
	Create(ctx context.Context, in *model.FormationTemplateRegisterInput) (string, error)
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	List(ctx context.Context, filters []*labelfilter.LabelFilter, name *string, pageSize int, cursor string) (*model.FormationTemplatePage, error)
	Update(ctx context.Context, id string, in *model.FormationTemplateUpdateInput) error
	Delete(ctx context.Context, id string) error
	ListWebhooksForFormationTemplate(ctx context.Context, formationTemplateID string) ([]*model.Webhook, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, formationTemplateID string, key string) (*model.Label, error)
	DeleteLabel(ctx context.Context, formationTemplateID string, key string) error
	ListLabels(ctx context.Context, formationTemplateID string) (map[string]*model.Label, error)
}

// WebhookConverter converts between the graphql and model
//
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookConverter interface {
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

// FormationConstraintService represents the FormationConstraint service layer
//
//go:generate mockery --name=FormationConstraintService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationConstraintService interface {
	ListByFormationTemplateIDs(ctx context.Context, formationTemplateIDs []string) ([][]*model.FormationConstraint, error)
}

// FormationConstraintConverter represents the FormationConstraint converter
//
//go:generate mockery --name=FormationConstraintConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationConstraintConverter interface {
	MultipleToGraphQL(in []*model.FormationConstraint) []*graphql.FormationConstraint
}

// Resolver is the FormationTemplate resolver
type Resolver struct {
	transact persistence.Transactioner

	formationTemplateSvc         FormationTemplateService
	converter                    FormationTemplateConverter
	webhookConverter             WebhookConverter
	formationConstraintSvc       FormationConstraintService
	formationConstraintConverter FormationConstraintConverter
}

// NewResolver creates FormationTemplate resolver
func NewResolver(transact persistence.Transactioner, converter FormationTemplateConverter, formationTemplateSvc FormationTemplateService, webhookConverter WebhookConverter, formationConstraintSvc FormationConstraintService, formationConstraintConverter FormationConstraintConverter) *Resolver {
	return &Resolver{
		transact:                     transact,
		converter:                    converter,
		formationTemplateSvc:         formationTemplateSvc,
		webhookConverter:             webhookConverter,
		formationConstraintSvc:       formationConstraintSvc,
		formationConstraintConverter: formationConstraintConverter,
	}
}

// FormationTemplatesByName pagination lists all FormationTemplates filtered by name and based on `first` and `after`
func (r *Resolver) FormationTemplatesByName(ctx context.Context, name *string, first *int, after *graphql.PageCursor) (*graphql.FormationTemplatePage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter: 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationTemplatePage, err := r.formationTemplateSvc.List(ctx, nil, name, *first, cursor)
	if err != nil {
		return nil, err
	}

	gqlFormationTemplate, err := r.converter.MultipleToGraphQL(formationTemplatePage.Data)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.FormationTemplatePage{
		Data:       gqlFormationTemplate,
		TotalCount: formationTemplatePage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(formationTemplatePage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(formationTemplatePage.PageInfo.EndCursor),
			HasNextPage: formationTemplatePage.PageInfo.HasNextPage,
		},
	}, nil
}

// FormationTemplates pagination lists all FormationTemplates based on `first` and `after`
func (r *Resolver) FormationTemplates(ctx context.Context, filters []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.FormationTemplatePage, error) {
	labelFilters := labelfilter.MultipleFromGraphQL(filters)
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter: 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationTemplatePage, err := r.formationTemplateSvc.List(ctx, labelFilters, nil, *first, cursor)
	if err != nil {
		return nil, err
	}

	gqlFormationTemplate, err := r.converter.MultipleToGraphQL(formationTemplatePage.Data)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.FormationTemplatePage{
		Data:       gqlFormationTemplate,
		TotalCount: formationTemplatePage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(formationTemplatePage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(formationTemplatePage.PageInfo.EndCursor),
			HasNextPage: formationTemplatePage.PageInfo.HasNextPage,
		},
	}, nil
}

// FormationTemplate queries the FormationTemplate matching ID `id`
func (r *Resolver) FormationTemplate(ctx context.Context, id string) (*graphql.FormationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationTemplate, err := r.formationTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlFormationTemplate, err := r.converter.ToGraphQL(formationTemplate)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlFormationTemplate, nil
}

// CreateFormationTemplate creates a FormationTemplate using `in`
func (r *Resolver) CreateFormationTemplate(ctx context.Context, in graphql.FormationTemplateRegisterInput) (*graphql.FormationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = in.Validate(); err != nil {
		return nil, err
	}

	convertedIn, err := r.converter.FromRegisterInputGraphQL(&in)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Debugf("Creating formation template with name: %q", in.Name)
	id, err := r.formationTemplateSvc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	formationTemplate, err := r.formationTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlFormationTemplate, err := r.converter.ToGraphQL(formationTemplate)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return gqlFormationTemplate, nil
}

// DeleteFormationTemplate deletes the FormationTemplate matching ID `id`
func (r *Resolver) DeleteFormationTemplate(ctx context.Context, id string) (*graphql.FormationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationTemplate, err := r.formationTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Debugf("Deleting formation template with ID: %q", id)
	err = r.formationTemplateSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlFormationTemplate, err := r.converter.ToGraphQL(formationTemplate)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlFormationTemplate, nil
}

// UpdateFormationTemplate updates the FormationTemplate matching ID `id` using `in`
func (r *Resolver) UpdateFormationTemplate(ctx context.Context, id string, in graphql.FormationTemplateUpdateInput) (*graphql.FormationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = in.Validate(); err != nil {
		return nil, err
	}

	convertedIn, err := r.converter.FromUpdateInputGraphQL(&in)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Debugf("Updating formation template with ID: %q", id)
	err = r.formationTemplateSvc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	formationTemplate, err := r.formationTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlFormationTemplate, err := r.converter.ToGraphQL(formationTemplate)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlFormationTemplate, nil
}

// SetFormationTemplateLabel add the provided label key and value to a formation template with the specified ID.
// If the label does not exist, it will be created. In case the label is already present, it will be updated with the provided value.
func (r *Resolver) SetFormationTemplateLabel(ctx context.Context, formationTemplateID string, lblInput graphql.LabelInput) (*graphql.Label, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.formationTemplateSvc.SetLabel(ctx, &model.LabelInput{
		Key:        lblInput.Key,
		Value:      lblInput.Value,
		ObjectType: model.FormationTemplateLabelableObject,
		ObjectID:   formationTemplateID,
	})
	if err != nil {
		return nil, err
	}

	label, err := r.formationTemplateSvc.GetLabel(ctx, formationTemplateID, lblInput.Key)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation template label with key: [%s]", lblInput.Key)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   label.Key,
		Value: label.Value,
	}, nil
}

// DeleteFormationTemplateLabel deletes a formation template label with the specified label key and formation template ID
func (r *Resolver) DeleteFormationTemplateLabel(ctx context.Context, formationTemplateID, key string) (*graphql.Label, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	label, err := r.formationTemplateSvc.GetLabel(ctx, formationTemplateID, key)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation template label with key: [%s]", key)
	}

	if err = r.formationTemplateSvc.DeleteLabel(ctx, formationTemplateID, key); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   key,
		Value: label.Value,
	}, nil
}

// Webhooks queries all webhooks related to the 'obj' Formation Template
func (r *Resolver) Webhooks(ctx context.Context, obj *graphql.FormationTemplate) ([]*graphql.Webhook, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Formation Template cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	webhooks, err := r.formationTemplateSvc.ListWebhooksForFormationTemplate(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	gqlWebhooks, err := r.webhookConverter.MultipleToGraphQL(webhooks)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return gqlWebhooks, nil
}

// Labels retrieve all labels for formation template
func (r *Resolver) Labels(ctx context.Context, obj *graphql.FormationTemplate, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Formation Template cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	labels, err := r.formationTemplateSvc.ListLabels(ctx, obj.ID)
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

	for _, l := range labels {
		if key == nil || l.Key == *key {
			resultLabels[l.Key] = l.Value
		}
	}

	var gqlLabels graphql.Labels = resultLabels
	return gqlLabels, nil
}

// FormationConstraint retrieves a FormationConstraint for the specified FormationTemplate
func (r *Resolver) FormationConstraint(ctx context.Context, obj *graphql.FormationTemplate) ([]*graphql.FormationConstraint, error) {
	params := dataloader.ParamFormationConstraint{ID: obj.ID, Ctx: ctx}
	return dataloader.FormationTemplateFor(ctx).FormationConstraintByID.Load(params)
}

// FormationConstraintsDataLoader retrieves Formation Constraints for each FormationTemplate ID in the keys
func (r *Resolver) FormationConstraintsDataLoader(keys []dataloader.ParamFormationConstraint) ([][]*graphql.FormationConstraint, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No formation templates are found")}
	}

	ctx := keys[0].Ctx
	formationTemplateIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		formationTemplateIDs = append(formationTemplateIDs, key.ID)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationConstraintsPerFormation, err := r.formationConstraintSvc.ListByFormationTemplateIDs(ctx, formationTemplateIDs)
	if err != nil {
		return nil, []error{err}
	}

	gqlFormationConstraints := make([][]*graphql.FormationConstraint, 0, len(formationConstraintsPerFormation))

	for _, formationConstraints := range formationConstraintsPerFormation {
		gqlFormationConstraints = append(gqlFormationConstraints, r.formationConstraintConverter.MultipleToGraphQL(formationConstraints))
	}

	if err = tx.Commit(); err != nil {
		return nil, []error{err}
	}

	return gqlFormationConstraints, nil
}
