package labeldef

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Resolver missing godoc
type Resolver struct {
	conv          ModelConverter
	srv           Service
	formationsSrv FormationService
	transactioner persistence.Transactioner
}

// NewResolver missing godoc
func NewResolver(transactioner persistence.Transactioner, srv Service, formationSvc FormationService, conv ModelConverter) *Resolver {
	return &Resolver{
		conv:          conv,
		srv:           srv,
		formationsSrv: formationSvc,
		transactioner: transactioner,
	}
}

// ModelConverter missing godoc
//go:generate mockery --name=ModelConverter --output=automock --outpkg=automock --case=underscore
type ModelConverter interface {
	// TODO: Use model.LabelDefinitionInput
	FromGraphQL(input graphql.LabelDefinitionInput, tenant string) (model.LabelDefinition, error)
	ToGraphQL(definition model.LabelDefinition) (graphql.LabelDefinition, error)
}

// Service missing godoc
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore
type Service interface {
	Get(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	GetWithoutCreating(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	List(ctx context.Context, tenant string) ([]model.LabelDefinition, error)
}

// FormationService missing godoc
//go:generate mockery --name=FormationService --output=automock --outpkg=automock --case=underscore
type FormationService interface {
	CreateFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
	DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
}

// CreateLabelDefinition missing godoc
func (r *Resolver) CreateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)

	ld, err := r.conv.FromGraphQL(in, tnt)
	if err != nil {
		return nil, err
	}

	ctx = persistence.SaveToContext(ctx, tx)

	_, err = r.srv.GetWithoutCreating(ctx, tnt, ld.Key)
	if err == nil {
		return nil, errors.New("while creating label definition: Object is not unique [object=labelDefinition]")
	}

	if !strings.Contains(err.Error(), "Object not found") {
		return nil, err
	}

	formations, err := ParseFormationsFromSchema(ld.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing schema")
	}
	for _, formation := range formations {
		_, err := r.formationsSrv.CreateFormation(ctx, tnt, model.Formation{Name: formation})
		if err != nil {
			return nil, errors.Wrap(err, "while creating formation")
		}
	}

	labelDef, err := r.srv.Get(ctx, tnt, ld.Key)

	if err != nil {
		return nil, errors.Wrap(err, "while getting label definition")
	}

	out, err := r.conv.ToGraphQL(*labelDef)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return &out, nil
}

// LabelDefinitions missing godoc
func (r *Resolver) LabelDefinitions(ctx context.Context) ([]*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	defs, err := r.srv.List(ctx, tnt)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Label Definitions")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	out := make([]*graphql.LabelDefinition, 0, len(defs))
	for _, def := range defs {
		c, err := r.conv.ToGraphQL(def)
		if err != nil {
			return nil, err
		}

		out = append(out, &c)
	}
	return out, nil
}

// LabelDefinition missing godoc
func (r *Resolver) LabelDefinition(ctx context.Context, key string) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	def, err := r.srv.Get(ctx, tnt, key)

	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, errors.Wrap(err, "while getting Label Definition")
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}
	if def == nil {
		return nil, apperrors.NewNotFoundError(resource.LabelDefinition, key)
	}
	c, err := r.conv.ToGraphQL(*def)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateLabelDefinition missing godoc
func (r *Resolver) UpdateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	ld, err := r.conv.FromGraphQL(in, tnt)
	if err != nil {
		return nil, err
	}

	storedLd, err := r.srv.Get(ctx, tnt, ld.Key)
	if err != nil {
		return nil, errors.Wrap(err, "while receiving stored label definition")
	}

	formations, err := ParseFormationsFromSchema(ld.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing schema")
	}

	storedFormations, err := ParseFormationsFromSchema(storedLd.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing schema")
	}

	for _, formation := range formations {
		if !containsFormation(storedFormations, formation) {
			_, err := r.formationsSrv.CreateFormation(ctx, tnt, model.Formation{Name: formation})
			if err != nil {
				return nil, errors.Wrap(err, "while creating formation")
			}
		}
	}
	for _, formation := range storedFormations {
		if !containsFormation(formations, formation) {
			_, err := r.formationsSrv.DeleteFormation(ctx, tnt, model.Formation{Name: formation})
			if err != nil {
				return nil, errors.Wrap(err, "while deleting formation")
			}
		}
	}

	updatedLd, err := r.srv.Get(ctx, tnt, in.Key)
	if err != nil {
		return nil, errors.Wrap(err, "while receiving updated label definition")
	}

	out, err := r.conv.ToGraphQL(*updatedLd)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return &out, nil
}

func containsFormation(storedFormations []string, formationToFind string) bool {
	for _, formation := range storedFormations {
		if formation == formationToFind {
			return true
		}
	}
	return false
}
