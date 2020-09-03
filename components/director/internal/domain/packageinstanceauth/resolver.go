package bundleinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	RequestDeletion(ctx context.Context, instanceAuth *model.BundleInstanceAuth, defaultBundleInstanceAuth *model.Auth) (bool, error)
	Create(ctx context.Context, bundleID string, in model.BundleInstanceAuthRequestInput, defaultAuth *model.Auth, requestInputSchema *string) (string, error)
	Get(ctx context.Context, id string) (*model.BundleInstanceAuth, error)
	SetAuth(ctx context.Context, id string, in model.BundleInstanceAuthSetInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToGraphQL(in *model.BundleInstanceAuth) (*graphql.BundleInstanceAuth, error)
	RequestInputFromGraphQL(in graphql.BundleInstanceAuthRequestInput) model.BundleInstanceAuthRequestInput
	SetInputFromGraphQL(in graphql.BundleInstanceAuthSetInput) (model.BundleInstanceAuthSetInput, error)
}

//go:generate mockery -name=BundleService -output=automock -outpkg=automock -case=underscore
type BundleService interface {
	Get(ctx context.Context, id string) (*model.Bundle, error)
	GetByInstanceAuthID(ctx context.Context, instanceAuthID string) (*model.Bundle, error)
}

type Resolver struct {
	transact persistence.Transactioner
	svc      Service
	pkgSvc   BundleService
	conv     Converter
}

func NewResolver(transact persistence.Transactioner, svc Service, pkgSvc BundleService, conv Converter) *Resolver {
	return &Resolver{
		transact: transact,
		svc:      svc,
		pkgSvc:   pkgSvc,
		conv:     conv,
	}
}

var mockRequestTypeKey = "type"
var mockBundleID = "db5d3b2a-cf30-498b-9a66-29e60247c66b"

func (r *Resolver) DeleteBundleInstanceAuth(ctx context.Context, authID string) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	err = r.svc.Delete(ctx, authID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

func (r *Resolver) SetBundleInstanceAuth(ctx context.Context, authID string, in graphql.BundleInstanceAuthSetInput) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.conv.SetInputFromGraphQL(in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting BundleInstanceAuth from GraphQL")
	}

	err = r.svc.SetAuth(ctx, authID, convertedIn)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

func (r *Resolver) RequestBundleInstanceAuthCreation(ctx context.Context, bundleID string, in graphql.BundleInstanceAuthRequestInput) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	pkg, err := r.pkgSvc.Get(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	convertedIn := r.conv.RequestInputFromGraphQL(in)

	instanceAuthID, err := r.svc.Create(ctx, bundleID, convertedIn, pkg.DefaultInstanceAuth, pkg.InstanceAuthRequestInputSchema)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := r.svc.Get(ctx, instanceAuthID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

func (r *Resolver) RequestBundleInstanceAuthDeletion(ctx context.Context, authID string) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	pkg, err := r.pkgSvc.GetByInstanceAuthID(ctx, authID)
	if err != nil {
		return nil, err
	}

	deleted, err := r.svc.RequestDeletion(ctx, instanceAuth, pkg.DefaultInstanceAuth)
	if err != nil {
		return nil, err
	}

	if !deleted {
		instanceAuth, err = r.svc.Get(ctx, authID) // get InstanceAuth once again for new status
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}
