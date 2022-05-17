package bundleinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// Service missing godoc
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type Service interface {
	RequestDeletion(ctx context.Context, instanceAuth *model.BundleInstanceAuth, defaultBundleInstanceAuth *model.Auth) (bool, error)
	Create(ctx context.Context, bundleID string, in model.BundleInstanceAuthRequestInput, defaultAuth *model.Auth, requestInputSchema *string) (string, error)
	Get(ctx context.Context, id string) (*model.BundleInstanceAuth, error)
	SetAuth(ctx context.Context, id string, in model.BundleInstanceAuthSetInput) error
	Delete(ctx context.Context, id string) error
}

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	ToGraphQL(in *model.BundleInstanceAuth) (*graphql.BundleInstanceAuth, error)
	RequestInputFromGraphQL(in graphql.BundleInstanceAuthRequestInput) model.BundleInstanceAuthRequestInput
	SetInputFromGraphQL(in graphql.BundleInstanceAuthSetInput) (model.BundleInstanceAuthSetInput, error)
}

// BundleService missing godoc
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleService interface {
	Get(ctx context.Context, id string) (*model.Bundle, error)
}

// BundleConverter missing godoc
//go:generate mockery --name=BundleConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleConverter interface {
	ToGraphQL(in *model.Bundle) (*graphql.Bundle, error)
}

// Resolver missing godoc
type Resolver struct {
	transact persistence.Transactioner
	svc      Service
	bndlSvc  BundleService
	conv     Converter
	bndlConv BundleConverter
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, svc Service, bndlSvc BundleService, conv Converter, bndlConv BundleConverter) *Resolver {
	return &Resolver{
		transact: transact,
		svc:      svc,
		bndlSvc:  bndlSvc,
		conv:     conv,
		bndlConv: bndlConv,
	}
}

// BundleByInstanceAuth missing godoc
func (r *Resolver) BundleByInstanceAuth(ctx context.Context, authID string) (*graphql.Bundle, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	bndlInstanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	pkg, err := r.bndlSvc.Get(ctx, bndlInstanceAuth.BundleID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.bndlConv.ToGraphQL(pkg)
}

// BundleInstanceAuth missing godoc
func (r *Resolver) BundleInstanceAuth(ctx context.Context, id string) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	bndlInstanceAuth, err := r.svc.Get(ctx, id)
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

	return r.conv.ToGraphQL(bndlInstanceAuth)
}

// DeleteBundleInstanceAuth missing godoc
func (r *Resolver) DeleteBundleInstanceAuth(ctx context.Context, authID string) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	err = r.svc.Delete(ctx, authID)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("BundleInstanceAuth with id %s successfully deleted", authID)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

// SetBundleInstanceAuth missing godoc
func (r *Resolver) SetBundleInstanceAuth(ctx context.Context, authID string, in graphql.BundleInstanceAuthSetInput) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Setting credentials for BundleInstanceAuth with id %s", authID)

	convertedIn, err := r.conv.SetInputFromGraphQL(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting BundleInstanceAuth with id %s from GraphQL", authID)
	}

	err = r.svc.SetAuth(ctx, authID, convertedIn)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Credentials successfully set for BundleInstanceAuth with id %s", authID)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

// RequestBundleInstanceAuthCreation missing godoc
func (r *Resolver) RequestBundleInstanceAuthCreation(ctx context.Context, bundleID string, in graphql.BundleInstanceAuthRequestInput) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Requesting BundleInstanceAuth creation for Bundle with id %s", bundleID)

	bndl, err := r.bndlSvc.Get(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	convertedIn := r.conv.RequestInputFromGraphQL(in)

	instanceAuthID, err := r.svc.Create(ctx, bundleID, convertedIn, bndl.DefaultInstanceAuth, bndl.InstanceAuthRequestInputSchema)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := r.svc.Get(ctx, instanceAuthID)
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created BundleInstanceAuth with id %s for Bundle with id %s", instanceAuthID, bundleID)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}

// RequestBundleInstanceAuthDeletion missing godoc
func (r *Resolver) RequestBundleInstanceAuthDeletion(ctx context.Context, authID string) (*graphql.BundleInstanceAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Requesting BundleInstanceAuth deletion for BundleInstanceAuth with id %s", authID)

	instanceAuth, err := r.svc.Get(ctx, authID)
	if err != nil {
		return nil, err
	}

	bndl, err := r.bndlSvc.Get(ctx, instanceAuth.BundleID)
	if err != nil {
		return nil, err
	}

	deleted, err := r.svc.RequestDeletion(ctx, instanceAuth, bndl.DefaultInstanceAuth)
	if err != nil {
		return nil, err
	}

	if !deleted {
		instanceAuth, err = r.svc.Get(ctx, authID) // get InstanceAuth once again for new status
		if err != nil {
			return nil, err
		}
	}

	log.C(ctx).Infof("BundleInstanceAuth with id %s successfully deleted.", authID)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(instanceAuth)
}
