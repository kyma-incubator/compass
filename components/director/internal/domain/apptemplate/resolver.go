package apptemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/pkg/errors"
)

//go:generate mockery -name=ApplicationTemplateService -output=automock -outpkg=automock -case=underscore
type ApplicationTemplateService interface {
	Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error)
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	List(ctx context.Context, pageSize int, cursor string) (model.ApplicationTemplatePage, error)
	Update(ctx context.Context, id string, in model.ApplicationTemplateInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=ApplicationTemplateConverter -output=automock -outpkg=automock -case=underscore
type ApplicationTemplateConverter interface {
	ToGraphQL(in *model.ApplicationTemplate) *graphql.ApplicationTemplate
	MultipleToGraphQL(in []*model.ApplicationTemplate) []*graphql.ApplicationTemplate
	InputFromGraphQL(in graphql.ApplicationTemplateInput) model.ApplicationTemplateInput
}

type Resolver struct {
	transact persistence.Transactioner

	appTemplateSvc       ApplicationTemplateService
	appTemplateConverter ApplicationTemplateConverter
}

func NewResolver(transact persistence.Transactioner, appTemplateSvc ApplicationTemplateService, appTemplateConverter ApplicationTemplateConverter) *Resolver {
	return &Resolver{
		transact:             transact,
		appTemplateSvc:       appTemplateSvc,
		appTemplateConverter: appTemplateConverter,
	}
}

func (r *Resolver) ApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	is, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.appTemplateConverter.ToGraphQL(is), nil
}

func (r *Resolver) ApplicationTemplates(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.ApplicationTemplatePage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appTemplatePage, err := r.appTemplateSvc.List(ctx, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlIntSys := r.appTemplateConverter.MultipleToGraphQL(appTemplatePage.Data)
	totalCount := len(gqlIntSys)

	return &graphql.ApplicationTemplatePage{
		Data:       gqlIntSys,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(appTemplatePage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(appTemplatePage.PageInfo.EndCursor),
			HasNextPage: appTemplatePage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) CreateApplicationTemplate(ctx context.Context, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	convertedIn := r.appTemplateConverter.InputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	id, err := r.appTemplateSvc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	intSys, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlIntSys := r.appTemplateConverter.ToGraphQL(intSys)

	return gqlIntSys, nil
}

func (r *Resolver) UpdateApplicationTemplate(ctx context.Context, id string, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.appTemplateConverter.InputFromGraphQL(in)
	err = r.appTemplateSvc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	intSys, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlIntSys := r.appTemplateConverter.ToGraphQL(intSys)

	return gqlIntSys, nil
}

func (r *Resolver) DeleteApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	intSys, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.appTemplateSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	deletedIntSys := r.appTemplateConverter.ToGraphQL(intSys)

	return deletedIntSys, nil
}
