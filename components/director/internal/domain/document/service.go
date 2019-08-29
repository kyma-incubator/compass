package document

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
)

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(id string) (*model.Document, error)
	ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.DocumentPage, error)
	Create(item *model.Document) error
	Delete(item *model.Document) error
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
	GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error)
	Delete(ctx context.Context, tenant, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
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
	document, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Document with ID %s", id)
	}

	return document, nil
}

func (s *service) List(ctx context.Context, applicationID string, pageSize *int, cursor *string) (*model.DocumentPage, error) {
	return s.repo.ListByApplicationID(applicationID, pageSize, cursor)
}

func (s *service) Create(ctx context.Context, applicationID string, in model.DocumentInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()

	var fetchRequestID *string
	if in.FetchRequest != nil {
		generatedID := s.uidService.Generate()
		fetchRequestID = &generatedID
		fetchRequestModel := in.FetchRequest.ToFetchRequest(s.timestampGen(), *fetchRequestID, tnt, model.DocumentFetchRequestReference, id)
		err := s.fetchRequestRepo.Create(ctx, fetchRequestModel)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for Document %s", id)
		}
	}

	document := in.ToDocument(id, tnt, applicationID, fetchRequestID)
	err = s.repo.Create(document)
	if err != nil {
		return "", errors.Wrap(err, "while creating Document")
	}

	return document.ID, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	document, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Document")
	}

	return s.repo.Delete(document)
}

func (s *service) GetFetchRequest(ctx context.Context, documentID string) (*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.Exists(ctx, tnt, documentID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking if Document exists")
	}
	if !exists {
		return nil, fmt.Errorf("Document with ID %s doesn't exist", documentID)
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.DocumentFetchRequestReference, documentID)
	if err != nil {
		if repo.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting FetchRequest by Document ID %s", documentID)
	}

	return fetchRequest, nil
}
