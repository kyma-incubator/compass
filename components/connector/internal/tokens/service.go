package tokens

import (
	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=Service
type Service interface {
	CreateToken(clientId string, tokenType TokenType) (string, apperrors.AppError)
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

func (svc *tokenService) CreateToken(clientId string, tokenType TokenType) (string, apperrors.AppError) {
	log.Infof("Generating token for %s with id %s", tokenType, clientId)

	token, err := svc.generator.NewToken()
	if err != nil {
		return "", err
	}

	tokenData := TokenData{
		Type:     tokenType,
		ClientId: clientId,
	}

	log.Debugf("Storing token for %s with id %s in the cache", tokenType, clientId)
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
