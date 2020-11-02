package tokens

import (
	"context"
	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

//go:generate mockery -name=Service
type Service interface {
	CreateToken(ctx context.Context, clientId string, tokenType TokenType) (string, apperrors.AppError)
	Resolve(token string) (TokenData, apperrors.AppError)
	Delete(token string)
}

type tokenService struct {
	generator TokenGenerator
	store     Cache
}

func NewTokenService(store Cache, generator TokenGenerator) *tokenService {
	return &tokenService{
		store:     store,
		generator: generator,
	}
}

func (svc *tokenService) CreateToken(ctx context.Context, clientId string, tokenType TokenType) (string, apperrors.AppError) {
	token, err := svc.generator.NewToken()
	if err != nil {
		return "", err
	}

	tokenData := TokenData{
		Type:     tokenType,
		ClientId: clientId,
	}

	log.C(ctx).Debugf("Storing token for %s with id %s in the cache", tokenData.Type, tokenData.ClientId)
	svc.store.Put(token, tokenData)

	return token, nil
}

func (svc *tokenService) Resolve(token string) (TokenData, apperrors.AppError) {
	tokenData, err := svc.store.Get(token)
	if err != nil {
		return TokenData{}, err.Append("Failed to resolve token")
	}

	return tokenData, nil
}

func (svc *tokenService) Delete(token string) {
	svc.store.Delete(token)
}
