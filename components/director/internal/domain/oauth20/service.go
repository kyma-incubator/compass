package oauth20

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type service struct {
}

func NewService() *service {
	return &service{
	}
}

func (s service) CreateClientCredentials(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) (*model.OAuthCredentialDataInput, error) {
	// generate credentials

	// load proper scopes

	// register in Hydra

	// return creds

	panic("Not implemented")
}