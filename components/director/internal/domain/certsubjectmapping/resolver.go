package certsubjectmapping

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// CertSubjectMappingService is responsible for service-layer certificate subject mapping operations
//go:generate mockery --name=CertSubjectMappingService --output=automock --outpkg=automock --case=underscore --disable-version-string
type CertSubjectMappingService interface {
	Create(ctx context.Context, in *model.CertSubjectMapping) (string, error)
	Get(ctx context.Context, id string) (*model.CertSubjectMapping, error)
	Update(ctx context.Context, in *model.CertSubjectMapping) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.CertSubjectMappingPage, error)
}

// Converter converts between the graphql and internal model
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	ToGraphQL(in *model.CertSubjectMapping) *graphql.CertificateSubjectMapping
	MultipleToGraphQL(in []*model.CertSubjectMapping) []*graphql.CertificateSubjectMapping
	FromGraphql(id string, in graphql.CertificateSubjectMappingInput) *model.CertSubjectMapping
}

// UIDService generates UUIDs for new entities
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// Resolver is an object responsible for resolver-layer operations.
type Resolver struct {
	transact              persistence.Transactioner
	conv                  Converter
	certSubjectMappingSvc CertSubjectMappingService
	uidSvc                UIDService
}

// NewResolver returns a new object responsible for resolver-layer certificate subject mapping operations.
func NewResolver(transact persistence.Transactioner, conv Converter, certSubjectMappingSvc CertSubjectMappingService, uidSvc UIDService) *Resolver {
	return &Resolver{
		transact:              transact,
		conv:                  conv,
		certSubjectMappingSvc: certSubjectMappingSvc,
		uidSvc:                uidSvc,
	}
}

// CertificateSubjectMapping queries the CertificateSubjectMapping matching ID `id`
func (r *Resolver) CertificateSubjectMapping(ctx context.Context, id string) (*graphql.CertificateSubjectMapping, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	csm, err := r.certSubjectMappingSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(csm), nil
}

// CertificateSubjectMappings list all CertificateSubjectMapping with pagination based on `first` and `after`
func (r *Resolver) CertificateSubjectMappings(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.CertificateSubjectMappingPage, error) {
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

	csmPage, err := r.certSubjectMappingSvc.List(ctx, *first, cursor)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlCertSubjectMapping := r.conv.MultipleToGraphQL(csmPage.Data)

	return &graphql.CertificateSubjectMappingPage{
		Data:       gqlCertSubjectMapping,
		TotalCount: csmPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(csmPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(csmPage.PageInfo.EndCursor),
			HasNextPage: csmPage.PageInfo.HasNextPage,
		},
	}, nil
}

// CreateCertificateSubjectMapping creates a CertificateSubjectMapping with the provided input `in`
func (r *Resolver) CreateCertificateSubjectMapping(ctx context.Context, in graphql.CertificateSubjectMappingInput) (*graphql.CertificateSubjectMapping, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = in.Validate(); err != nil {
		return nil, err
	}

	certSubjectMappingID := r.uidSvc.Generate()
	csmID, err := r.certSubjectMappingSvc.Create(ctx, r.conv.FromGraphql(certSubjectMappingID, in))
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created a certificate subject mapping with ID: %s and subject: %s", certSubjectMappingID, in.Subject)

	csm, err := r.certSubjectMappingSvc.Get(ctx, csmID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(csm), nil
}

// UpdateCertificateSubjectMapping updates the CertificateSubjectMapping matching ID `id` using `in`
func (r *Resolver) UpdateCertificateSubjectMapping(ctx context.Context, id string, in graphql.CertificateSubjectMappingInput) (*graphql.CertificateSubjectMapping, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = in.Validate(); err != nil {
		return nil, err
	}

	err = r.certSubjectMappingSvc.Update(ctx, r.conv.FromGraphql(id, in))
	if err != nil {
		return nil, err
	}

	csm, err := r.certSubjectMappingSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(csm), nil
}

// DeleteCertificateSubjectMapping deletes the CertificateSubjectMapping matching ID `id`
func (r *Resolver) DeleteCertificateSubjectMapping(ctx context.Context, id string) (*graphql.CertificateSubjectMapping, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	csm, err := r.certSubjectMappingSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.certSubjectMappingSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(csm), nil
}
