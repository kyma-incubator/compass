package formationtemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// FormationTemplateConverter missing godoc
//go:generate mockery --name=FormationTemplateConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateConverter interface {
	FromInputGraphQL(in *graphql.FormationTemplateInput) *model.FormationTemplateInput
	ToGraphQL(in *model.FormationTemplate) *graphql.FormationTemplate
	MultipleToGraphQL(in []*model.FormationTemplate) []*graphql.FormationTemplate
	FromModelInputToModel(in *model.FormationTemplateInput, id string) *model.FormationTemplate
}

// FormationTemplateService missing godoc
//go:generate mockery --name=FormationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateService interface {
	Create(ctx context.Context, in *model.FormationTemplateInput) (string, error)
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.FormationTemplatePage, error)
	Update(ctx context.Context, id string, in *model.FormationTemplateInput) error
	Delete(ctx context.Context, id string) error
}

// Resolver is the formation template resolver
type Resolver struct {
	transact persistence.Transactioner

	formationTemplateSvc FormationTemplateService
	converter            FormationTemplateConverter
}

// NewResolver creates formation template resolver
func NewResolver(transact persistence.Transactioner, converter FormationTemplateConverter, formationTemplateSvc FormationTemplateService) *Resolver {
	return &Resolver{
		transact:             transact,
		converter:            converter,
		formationTemplateSvc: formationTemplateSvc,
	}
}

// FormationTemplates missing godoc
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlFormationTemplate := r.converter.MultipleToGraphQL(formationTemplatePage.Data)

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

// FormationTemplate missing godoc
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(formationTemplate), nil
}

// CreateFormationTemplate missing godoc
func (r *Resolver) CreateFormationTemplate(ctx context.Context, in graphql.FormationTemplateInput) (*graphql.FormationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := in.Validate(); err != nil {
		return nil, err
	}

	id, err := r.formationTemplateSvc.Create(ctx, r.converter.FromInputGraphQL(&in))
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created an Formation Template with name %s and id %s", in.Name, id)

	formationTemplate, err := r.formationTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(formationTemplate), nil
}

// DeleteFormationTemplate missing godoc
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(formationTemplate), nil
}

// UpdateFormationTemplate missing godoc
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

	convertedIn := r.converter.FromInputGraphQL(&in)

	err = r.formationTemplateSvc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	formationTemplate, err := r.formationTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(formationTemplate), nil
}
