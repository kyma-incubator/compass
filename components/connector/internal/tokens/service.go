package tokens

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

type Token struct {
	TokenData
	CreatedAt time.Time
	Used      bool
}

//go:generate mockery -name=Service
type Service interface {
	CreateToken(ctx context.Context, clientId string, tokenType TokenType) (string, apperrors.AppError)
	Resolve(ctx context.Context, token string) (TokenData, apperrors.AppError)
	Delete(ctx context.Context, token string) error
}

type Repository interface {
	Create(ctx context.Context, token string, tokenData TokenData) error
	Get(ctx context.Context, token string) (Token, apperrors.AppError)
	Invalidate(ctx context.Context, token string) apperrors.AppError
}

type tokenService struct {
	repository Repository
	generator  TokenGenerator

	applicationTokenTTL time.Duration
	runtimeTokenTTL     time.Duration
	csrTokenTTL         time.Duration
}

func NewTokenService(repository Repository, generator TokenGenerator, applicationTokenTTL, runtimeTokenTTL, csrTokenTTL time.Duration) *tokenService {
	return &tokenService{
		generator:           generator,
		repository:          repository,
		applicationTokenTTL: applicationTokenTTL,
		runtimeTokenTTL:     runtimeTokenTTL,
		csrTokenTTL:         csrTokenTTL,
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

	log.C(ctx).Debugf("Storing token for %s with id %s in the db", tokenData.Type, tokenData.ClientId)
	if err := svc.repository.Create(ctx, token, tokenData); err != nil {
		// TODO: Check type assertation
		return "", apperrors.Internal("Could not create token %s", err)
	}

	return token, nil
}

func (svc *tokenService) Resolve(ctx context.Context, token string) (TokenData, apperrors.AppError) {
	tokenModel, err := svc.repository.Get(ctx, token)
	if err != nil {
		return TokenData{}, err.Append("Failed to resolve token")
	}
	if !svc.checkValidity(tokenModel) {
		return TokenData{}, apperrors.BadRequest("Token has expired")
	}

	return tokenModel.TokenData, nil
}

func (svc *tokenService) Delete(ctx context.Context, token string) error {
	return svc.repository.Invalidate(ctx, token)
}

func (svc *tokenService) checkValidity(token Token) bool {
	switch token.Type {
	case ApplicationToken:
		return time.Now().Sub(token.CreatedAt) < svc.applicationTokenTTL
	case RuntimeToken:
		return time.Now().Sub(token.CreatedAt) < svc.runtimeTokenTTL
	case CSRToken:
		return time.Now().Sub(token.CreatedAt) < svc.csrTokenTTL
	default:
		return false
	}
}
