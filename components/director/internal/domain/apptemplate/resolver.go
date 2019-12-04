package apptemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"

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
	GetByName(ctx context.Context, name string) (*model.ApplicationTemplate, error)
	List(ctx context.Context, pageSize int, cursor string) (model.ApplicationTemplatePage, error)
	Update(ctx context.Context, id string, in model.ApplicationTemplateInput) error
	Delete(ctx context.Context, id string) error
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, templatePlaceholderValues []*model.ApplicationTemplateValueInput) (string, error)
}

//go:generate mockery -name=ApplicationTemplateConverter -output=automock -outpkg=automock -case=underscore
type ApplicationTemplateConverter interface {
	ToGraphQL(in *model.ApplicationTemplate) (*graphql.ApplicationTemplate, error)
	MultipleToGraphQL(in []*model.ApplicationTemplate) ([]*graphql.ApplicationTemplate, error)
	InputFromGraphQL(in graphql.ApplicationTemplateInput) (model.ApplicationTemplateInput, error)
	ApplicationFromTemplateInputFromGraphQL(in graphql.ApplicationFromTemplateInput) model.ApplicationFromTemplateInput
}

//go:generate mockery -name=ApplicationConverter -output=automock -outpkg=automock -case=underscore
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	CreateInputJSONToGQL(in string) (graphql.ApplicationCreateInput, error)
	CreateInputFromGraphQL(in graphql.ApplicationCreateInput) model.ApplicationCreateInput
}

type Resolver struct {
	transact persistence.Transactioner

	appSvc               application.ApplicationService
	appConverter         ApplicationConverter
	appTemplateSvc       ApplicationTemplateService
	appTemplateConverter ApplicationTemplateConverter
}

func NewResolver(transact persistence.Transactioner, appSvc application.ApplicationService, appConverter ApplicationConverter, appTemplateSvc ApplicationTemplateService, appTemplateConverter ApplicationTemplateConverter) *Resolver {
	return &Resolver{
		transact:             transact,
		appSvc:               appSvc,
		appConverter:         appConverter,
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

	appTemplate, err := r.appTemplateSvc.Get(ctx, id)
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

	out, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application template to graphql")
	}

	return out, nil
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

	gqlAppTemplate, err := r.appTemplateConverter.MultipleToGraphQL(appTemplatePage.Data)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application templates to graphql")
	}
	totalCount := len(gqlAppTemplate)

	return &graphql.ApplicationTemplatePage{
		Data:       gqlAppTemplate,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(appTemplatePage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(appTemplatePage.PageInfo.EndCursor),
			HasNextPage: appTemplatePage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) CreateApplicationTemplate(ctx context.Context, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.appTemplateConverter.InputFromGraphQL(in)
	if err != nil {
		return nil, err
	}

	id, err := r.appTemplateSvc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	appTemplate, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAppTemplate, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application template to graphql")
	}

	return gqlAppTemplate, nil
}

func (r *Resolver) RegisterApplicationFromTemplate(ctx context.Context, in graphql.ApplicationFromTemplateInput) (*graphql.Application, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.appTemplateConverter.ApplicationFromTemplateInputFromGraphQL(in)

	appTemplate, err := r.appTemplateSvc.GetByName(ctx, convertedIn.TemplateName)
	if err != nil {
		return nil, err
	}

	appCreateInputJSON, err := r.appTemplateSvc.PrepareApplicationCreateInputJSON(appTemplate, convertedIn.Values)
	if err != nil {
		return nil, err
	}

	appCreateInputGQL, err := r.appConverter.CreateInputJSONToGQL(appCreateInputJSON)
	if err != nil {
		return nil, err
	}

	if err := appCreateInputGQL.Validate(); err != nil {
		return nil, errors.Wrapf(err, "while validating application input from application template [name=%s]", convertedIn.TemplateName)
	}

	appCreateInputModel := r.appConverter.CreateInputFromGraphQL(appCreateInputGQL)

	id, err := r.appSvc.Create(ctx, appCreateInputModel)
	if err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)
	return gqlApp, nil
}

func (r *Resolver) UpdateApplicationTemplate(ctx context.Context, id string, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.appTemplateConverter.InputFromGraphQL(in)
	if err != nil {
		return nil, err
	}

	err = r.appTemplateSvc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	appTemplate, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAppTemplate, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application template to graphql")
	}

	return gqlAppTemplate, nil
}

func (r *Resolver) DeleteApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appTemplate, err := r.appTemplateSvc.Get(ctx, id)
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

	deletedAppTemplate, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application template to graphql")
	}

	return deletedAppTemplate, nil
}
