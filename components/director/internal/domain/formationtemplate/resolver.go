package formationtemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// FormationTemplateConverter converts between the graphql and model
//go:generate mockery --name=FormationTemplateConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateConverter interface {
	FromInputGraphQL(in *graphql.FormationTemplateInput) (*model.FormationTemplateInput, error)
	ToGraphQL(in *model.FormationTemplate) (*graphql.FormationTemplate, error)
	MultipleToGraphQL(in []*model.FormationTemplate) ([]*graphql.FormationTemplate, error)
	FromModelInputToModel(in *model.FormationTemplateInput, id string, tenantID string) *model.FormationTemplate
}

// FormationTemplateService represents the FormationTemplate service layer
//go:generate mockery --name=FormationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateService interface {
	Create(ctx context.Context, in *model.FormationTemplateInput) (string, error)
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.FormationTemplatePage, error)
	Update(ctx context.Context, id string, in *model.FormationTemplateInput) error
	Delete(ctx context.Context, id string) error
}

// WebhookConverter converts between the graphql and model
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookConverter interface {
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

// WebhookService represents the Webhook service layer
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListForFormationTemplate(ctx context.Context, formationTemplateID string) ([]*model.Webhook, error)
}

// Resolver is the FormationTemplate resolver
type Resolver struct {
	transact persistence.Transactioner

	formationTemplateSvc FormationTemplateService
	converter            FormationTemplateConverter
	webhookConverter     WebhookConverter
	webhookSvc           WebhookService
}

// NewResolver creates FormationTemplate resolver
func NewResolver(transact persistence.Transactioner, converter FormationTemplateConverter, formationTemplateSvc FormationTemplateService, webhookConverter WebhookConverter, webhookSvc WebhookService) *Resolver {
	return &Resolver{
		transact:             transact,
		converter:            converter,
		formationTemplateSvc: formationTemplateSvc,
		webhookConverter:     webhookConverter,
		webhookSvc:           webhookSvc,
	}
}

// FormationTemplates pagination lists all FormationTemplates based on `first` and `after`
func (r *Resolver) FormationTemplates(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.FormationTemplatePage, error) {
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

	formationTemplatePage, err := r.formationTemplateSvc.List(ctx, *first, cursor)
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
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
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
func (r *Resolver) CreateFormationTemplate(ctx context.Context, in graphql.FormationTemplateInput) (*graphql.FormationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = in.Validate(); err != nil {
		return nil, err
	}

	convertedIn, err := r.converter.FromInputGraphQL(&in)
	if err != nil {
		return nil, err
	}

	id, err := r.formationTemplateSvc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created an Formation Template with name %s and id %s", in.Name, id)

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
func (r *Resolver) UpdateFormationTemplate(ctx context.Context, id string, in graphql.FormationTemplateInput) (*graphql.FormationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = in.Validate(); err != nil {
		return nil, err
	}

	convertedIn, err := r.converter.FromInputGraphQL(&in)
	if err != nil {
		return nil, err
	}

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

	webhooks, err := r.webhookSvc.ListForFormationTemplate(ctx, obj.ID)
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
