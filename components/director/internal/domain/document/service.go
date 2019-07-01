package document

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	GetByID(id string) (*model.Document, error)
	ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.DocumentPage, error)
	Create(item *model.Document) error
	Delete(item *model.Document) error
}

type service struct {
	repo DocumentRepository
}

func NewService(repo DocumentRepository) *service {
	return &service{
		repo: repo,
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

func (s *service) Create(ctx context.Context, id string, applicationID string, in model.DocumentInput) (string, error) {

	document := in.ToDocument(id, applicationID)

	err := s.repo.Create(document)
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
