package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --name=DocumentRepository --output=automock --outpkg=automock --case=underscore
type DocumentRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Document, error)
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.Document, error)
	ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, pageSize int, cursor string) ([]*model.DocumentPage, error)
	Create(ctx context.Context, item *model.Document) error
	Delete(ctx context.Context, tenant, id string) error
}

//go:generate mockery --name=FetchRequestRepository --output=automock --outpkg=automock --case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
	Delete(ctx context.Context, tenant, id string) error
	ListByReferenceObjectIDs(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectIDs []string) ([]*model.FetchRequest, error)
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo             DocumentRepository
	fetchRequestRepo FetchRequestRepository
	uidService       UIDService
	timestampGen     timestamp.Generator
}

func NewService(repo DocumentRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService) *service {
	return &service{
		repo:             repo,
		fetchRequestRepo: fetchRequestRepo,
		uidService:       uidService,
		timestampGen:     timestamp.DefaultGenerator(),
	}
}

func (s *service) Get(ctx context.Context, id string) (*model.Document, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	document, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Document with ID %s", id)
	}

	return document, nil
}

func (s *service) GetForBundle(ctx context.Context, id string, bundleID string) (*model.Document, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	document, err := s.repo.GetForBundle(ctx, tnt, id, bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Document")
	}

	return document, nil
}

func (s *service) CreateInBundle(ctx context.Context, bundleID string, in model.DocumentInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()

	document := in.ToDocumentWithinBundle(id, tnt, bundleID)
	err = s.repo.Create(ctx, document)
	if err != nil {
		return "", errors.Wrap(err, "while creating Document")
	}

	if in.FetchRequest != nil {
		generatedID := s.uidService.Generate()
		fetchRequestID := &generatedID
		fetchRequestModel := in.FetchRequest.ToFetchRequest(s.timestampGen(), *fetchRequestID, tnt, model.DocumentFetchRequestReference, id)
		err := s.fetchRequestRepo.Create(ctx, fetchRequestModel)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for Document %s", id)
		}
	}

	return document.ID, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Document with ID %s", id)
	}

	return nil
}

func (s *service) ListByBundleIDs(ctx context.Context, bundleIDs []string, pageSize int, cursor string) ([]*model.DocumentPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.ListByBundleIDs(ctx, tnt, bundleIDs, pageSize, cursor)
}

func (s *service) ListFetchRequests(ctx context.Context, documentIDs []string) ([]*model.FetchRequest, error) {
	tenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.fetchRequestRepo.ListByReferenceObjectIDs(ctx, tenant, model.DocumentFetchRequestReference, documentIDs)
}
